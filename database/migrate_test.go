package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseMigration(t *testing.T) {
	m, err := parseMigration("0.0.2_add_active.sql", "alter table a add b int;\n\n-- prea:down\nalter table a drop b;\n")
	if err != nil {
		t.Fatal(err)
	}
	if m.version != "0.0.2" || m.name != "add_active" || m.num != 2 || m.noTx ||
		m.up != "alter table a add b int;\n\n" || m.down != "\nalter table a drop b;\n" {
		t.Fatalf("%+v", m)
	}
	if m, _ := parseMigration("20261013120000_x.sql", "\n-- prea:no-transaction\ncreate index concurrently i on a (b);"); !m.noTx || m.num != 20261013120000 ||
		strings.Contains(m.up, "prea:") || m.down != "" {
		t.Fatalf("no-transaction: %+v", m)
	}
	if m, _ := parseMigration("0.0.1_x.sql", "select 1; -- prea:down is only a marker on its own line"); m.down != "" {
		t.Fatalf("inline marker: %+v", m)
	}

	for file, want := range map[string]string{
		"create_users.sql":    "expected <version>_<name>.sql",
		"0.0.1_CreateX.sql":   "expected <version>_<name>.sql",
		"0.0_x.sql":           "expected <version>_<name>.sql",
		"0.0.0_x.sql":         "must be greater than zero",
		"1.1000000.0_x.sql":   "each part must be below 1000000",
		"0.0.1_empty.sql":     "nothing to run",
		"0.0.2_emptydown.sql": "the -- prea:down section is empty",
	} {
		content := "select 1;"
		switch file {
		case "0.0.1_empty.sql":
			content = "  \n-- prea:down\nselect 1;"
		case "0.0.2_emptydown.sql":
			content = "select 1;\n-- prea:down\n\n"
		}
		if _, err := parseMigration(file, content); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %q", file, err, want)
		}
	}
}

func TestVersionOrder(t *testing.T) {
	var prev int64
	for _, v := range []string{"0.0.1", "0.0.9", "0.0.10", "0.1.0", "0.99.999999", "1.0.0", "999999.0.0"} {
		n, err := versionNumber(v)
		if err != nil || n <= prev {
			t.Fatalf("%s: %d after %d (%v)", v, n, prev, err)
		}
		prev = n
	}
}

func TestReadMigrations(t *testing.T) {
	write := func(dir string, files map[string]string) string {
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}

	dir := write(t.TempDir(), map[string]string{"0.0.10_c.sql": "select 3;", "0.0.9_b.sql": "select 2;", "0.0.1_a.sql": "select 1;", "README.md": "notes"})
	migs, err := readMigrations(dir)
	if err != nil || len(migs) != 3 || migs[0].name != "a" || migs[2].String() != "0.0.10_c" {
		t.Fatalf("ordered by version: %v %v", migs, err)
	}
	if migs, err := readMigrations(filepath.Join(dir, "missing")); err != nil || migs != nil {
		t.Fatalf("a missing dir has no migrations: %v %v", migs, err)
	}

	for want, files := range map[string]map[string]string{
		"have the same version":         {"0.0.1_a.sql": "select 1;", "1_b.sql": "select 1;"},
		"mix semver and timestamp":      {"0.0.1_a.sql": "select 1;", "20261013120000_b.sql": "select 1;"},
		"expected <version>_<name>.sql": {"bad.sql": "select 1;", "0.0.1_a.sql": ""},
		"0.0.1_a.sql: nothing to run":   {"bad.sql": "select 1;", "0.0.1_a.sql": ""}, // every file error at once
	} {
		if _, err := readMigrations(write(t.TempDir(), files)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("got %v, want %q", err, want)
		}
	}
}

func TestNextVersion(t *testing.T) {
	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.FixedZone("BRT", -3*3600))
	semver := []migration{{version: "0.0.1"}, {version: "0.1.9"}}
	stamp := []migration{{version: "20261013120000"}}
	for _, tc := range []struct {
		migs   []migration
		format string
		want   string
	}{
		{nil, "", "0.0.1"},
		{semver, "semver", "0.1.10"},
		{nil, "timestamp", "20261013150000"},
		{stamp, "timestamp", "20261013150000"},
		{stamp, "", "set database.migrations.version: timestamp"},
		{semver, "timestamp", "set database.migrations.version: semver"},
		{nil, "date", `migrations.version "date" is not semver or timestamp`},
	} {
		got, err := nextVersion(tc.migs, tc.format, now)
		if err != nil {
			got = err.Error()
		}
		if !strings.Contains(got, tc.want) {
			t.Errorf("%v %q: got %q, want %q", tc.migs, tc.format, got, tc.want)
		}
	}
}

func TestGooseFormat(t *testing.T) {
	m := migration{up: "create function f() returns int as $$ select 1; $$ language sql;", down: "drop function f;", noTx: true}
	want := `-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
create function f() returns int as $$ select 1; $$ language sql;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop function f;
-- +goose StatementEnd
`
	if got := string(m.goose()); got != want {
		t.Fatalf("got\n%s", got)
	}
	if got := string(migration{up: "select 1;"}.goose()); strings.Contains(got, "Down") || strings.Contains(got, "NO TRANSACTION") {
		t.Fatalf("no down, in a transaction: %s", got)
	}
}
