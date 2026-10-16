package security_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/prea-framework/prea/security"
)

func TestHashPassword(t *testing.T) {
	hs, err := security.HashPassword("Renan")
	if err != nil {
		t.Fatal(err)
	}

	cp, err := security.ComparePassword("Renan", hs)
	if err != nil {
		t.Fatal(err)
	}

	if !cp {
		t.Fatalf("expected true")
	}

	cp, err = security.ComparePassword("Naner", hs)
	if err != nil {
		t.Fatal(err)
	}

	if cp {
		t.Fatalf("expected false")
	}
}

func TestPasswordMaxLen(t *testing.T) {
	ok := strings.Repeat("a", security.MaxPasswordLen)
	hs, err := security.HashPassword(ok)
	if err != nil {
		t.Fatalf("MaxPasswordLen must be accepted: %v", err)
	}
	cp, err := security.ComparePassword(ok, hs)
	if err != nil || !cp {
		t.Fatalf("compare at MaxPasswordLen: ok=%v err=%v", cp, err)
	}

	tooLong := ok + "x"
	if _, err := security.HashPassword(tooLong); !errors.Is(err, security.ErrPasswordTooLong) {
		t.Fatalf("HashPassword: got %v, want ErrPasswordTooLong", err)
	}
	if _, err := security.ComparePassword(tooLong, hs); !errors.Is(err, security.ErrPasswordTooLong) {
		t.Fatalf("ComparePassword: got %v, want ErrPasswordTooLong", err)
	}
}
