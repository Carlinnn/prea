package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon struct {
	saltlen uint
	mem     uint32
	time    uint32
	threads uint8
	keylen  uint32
}

const (
	SaltLen = 32
	Mem     = 1024 * 64
	Time    = 1
	Threads = 4
	KeyLen  = 32
	// MaxPasswordLen is the longest password HashPassword and ComparePassword
	// accept, in bytes. Longer input is rejected before argon2 runs, so a
	// huge body cannot burn CPU and memory on hashing.
	MaxPasswordLen = 128
)

// ErrPasswordTooLong is returned when a password exceeds MaxPasswordLen.
var ErrPasswordTooLong = errors.New("password too long")

func HashPassword(passwd string) (string, error) {
	if err := checkPasswordLen(passwd); err != nil {
		return "", err
	}
	return argon2Encoded(passwd, Argon{SaltLen, Mem, Time, Threads, KeyLen})
}

func ComparePassword(passwd string, hashed string) (bool, error) {
	if err := checkPasswordLen(passwd); err != nil {
		return false, err
	}
	return argon2Compare(passwd, hashed)
}

func checkPasswordLen(passwd string) error {
	if len(passwd) > MaxPasswordLen {
		return ErrPasswordTooLong
	}
	return nil
}

func genSalt(num uint) ([]byte, error) {
	b := make([]byte, num)
	_, err := rand.Read(b)

	return b, err
}

func argon2Encoded(passwd string, c Argon) (enc string, err error) {
	salt, err := genSalt(c.saltlen)
	if err != nil {
		return "", err
	}

	hashed := argon2.IDKey([]byte(passwd), salt, c.time, c.mem, c.threads, c.keylen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hashed)

	enc = fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, c.mem, c.time, c.threads, b64Salt, b64Hash,
	)

	return
}

func argon2Compare(passwd string, hashed string) (bool, error) {
	vals := strings.Split(hashed, "$")

	if len(vals) != 6 {
		return false, ErrInvalidHash
	}

	var version int

	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false, err
	}

	if version != argon2.Version {
		return false, ErrInvalidHashVersion
	}

	var mem uint32
	var time uint32
	var threads uint8

	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &mem, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return false, err
	}

	enc := argon2.IDKey([]byte(passwd), salt, time, mem, threads, uint32(len(hash)))

	if subtle.ConstantTimeCompare(enc, hash) == 1 {
		return true, nil
	}

	return false, nil
}
