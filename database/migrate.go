package database

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// migration is one file of the migrations dir, <version>_<name>.sql. All of
// it runs on up; what follows a "-- prea:down" line is what down runs.
//
//	-- prea:no-transaction   (optional first line: create index concurrently)
//	alter table users add column active bool not null default true;
//
//	-- prea:down
//	alter table users drop column active;
type migration struct {
	version string // as written: 0.0.2 or 20261013120000
	name    string
	num     int64 // the version as goose's integer
	up      string
	down    string
	noTx    bool
}

func (m migration) String() string { return m.version + "_" + m.name }

var (
	migrationFile = regexp.MustCompile(`^(\d+\.\d+\.\d+|\d+)_([a-z0-9_]+)\.sql$`)
	migrationName = regexp.MustCompile(`^[a-z0-9_]+$`)
	downMarker    = regexp.MustCompile(`(?m)^-- prea:down[ \t]*\r?$`)
)

const noTxMarker = "-- prea:no-transaction"

// readMigrations reads the migrations of dir, ordered by version. A missing
// dir has none; files that aren't .sql are ignored.
func readMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database: migrations: %w", err)
	}
	var migs []migration
	var errs []error
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("database: migrations: %w", err)
		}
		m, err := parseMigration(e.Name(), string(content))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		migs = append(migs, m)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return migs, checkMigrations(migs)
}

func parseMigration(file, content string) (migration, error) {
	parts := migrationFile.FindStringSubmatch(file)
	if parts == nil {
		return migration{}, fmt.Errorf("database: migrations/%s: expected <version>_<name>.sql, e.g. 0.0.1_create_users.sql", file)
	}
	num, err := versionNumber(parts[1])
	if err != nil {
		return migration{}, fmt.Errorf("database: migrations/%s: %w", file, err)
	}
	m := migration{version: parts[1], name: parts[2], num: num, up: content}

	if rest, ok := strings.CutPrefix(strings.TrimLeft(m.up, " \t\r\n"), noTxMarker); ok {
		m.up, m.noTx = rest, true
	}
	if loc := downMarker.FindStringIndex(m.up); loc != nil {
		m.up, m.down = m.up[:loc[0]], m.up[loc[1]:]
		if strings.TrimSpace(m.down) == "" {
			return m, fmt.Errorf("database: migrations/%s: the -- prea:down section is empty", file)
		}
	}
	if strings.TrimSpace(m.up) == "" {
		return m, fmt.Errorf("database: migrations/%s: nothing to run", file)
	}
	return m, nil
}

// versionNumber packs a semver into an integer that sorts like it
// (major*1e12 + minor*1e6 + patch); a plain number is taken as is.
func versionNumber(v string) (int64, error) {
	var n int64
	if parts := strings.Split(v, "."); len(parts) == 3 {
		for _, p := range parts {
			i, err := strconv.ParseInt(p, 10, 64)
			if err != nil || i >= 1_000_000 {
				return 0, fmt.Errorf("version %s: each part must be below 1000000", v)
			}
			n = n*1_000_000 + i
		}
	} else {
		var err error
		if n, err = strconv.ParseInt(v, 10, 64); err != nil {
			return 0, fmt.Errorf("version %s: %w", v, err)
		}
	}
	if n < 1 {
		return 0, fmt.Errorf("version %s must be greater than zero", v)
	}
	return n, nil
}

func isSemver(version string) bool { return strings.Contains(version, ".") }

// checkMigrations sorts migs by version and rejects duplicates and a dir
// mixing semver and timestamp versions.
func checkMigrations(migs []migration) error {
	slices.SortFunc(migs, func(a, b migration) int { return cmp.Compare(a.num, b.num) })
	for i := 1; i < len(migs); i++ {
		a, b := migs[i-1], migs[i]
		switch {
		case a.num == b.num:
			return fmt.Errorf("database: migrations %s and %s have the same version", a, b)
		case isSemver(a.version) != isSemver(b.version):
			return fmt.Errorf("database: migrations %s and %s mix semver and timestamp versions", a, b)
		}
	}
	return nil
}

// nextVersion is the version of a new migration: the last semver with the
// patch bumped (0.0.1 first) or, with format timestamp, now in UTC.
func nextVersion(migs []migration, format string, now time.Time) (string, error) {
	var last string
	if len(migs) > 0 {
		last = migs[len(migs)-1].version
	}
	switch format {
	case "", "semver":
		if last == "" {
			return "0.0.1", nil
		}
		if !isSemver(last) {
			return "", fmt.Errorf("database: the migrations use timestamp versions: set database.migrations.version: timestamp")
		}
		i := strings.LastIndex(last, ".")
		patch, _ := strconv.Atoi(last[i+1:])
		return last[:i+1] + strconv.Itoa(patch+1), nil
	case "timestamp":
		if last != "" && isSemver(last) {
			return "", fmt.Errorf("database: the migrations use semver versions: set database.migrations.version: semver")
		}
		return now.UTC().Format("20060102150405"), nil
	}
	return "", fmt.Errorf("database: migrations.version %q is not semver or timestamp", format)
}

// goose writes m in goose's format. The whole up (and down) is one
// statement, so the file runs as written, $$ bodies included.
func (m migration) goose() []byte {
	var b strings.Builder
	if m.noTx {
		b.WriteString("-- +goose NO TRANSACTION\n")
	}
	b.WriteString("-- +goose Up\n-- +goose StatementBegin\n" + m.up + "\n-- +goose StatementEnd\n")
	if m.down != "" {
		b.WriteString("-- +goose Down\n-- +goose StatementBegin\n" + m.down + "\n-- +goose StatementEnd\n")
	}
	return []byte(b.String())
}
