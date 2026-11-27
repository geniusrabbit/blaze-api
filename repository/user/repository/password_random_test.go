package repository

import (
	"strings"
	"testing"
)

func TestRandomPassword(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}
	for range 50 {
		pwd := RandomPassword(16)
		if len(pwd) != 16 || !strings.ContainsAny(pwd, "0123456789") || !strings.ContainsAny(pwd, "~=+%^*/()[]{}/!@#$?|") {
			t.Fatalf("password %q doesn't meet the rules", pwd)
		}
		if seen[pwd] {
			t.Fatal("duplicate password")
		}
		seen[pwd] = true
	}
	if len(RandomPassword(1)) != 2 {
		t.Fatal("length below 2 isn't raised to 2")
	}
}

func TestHashErrorIsReturned(t *testing.T) {
	t.Parallel()

	// Before the fix a failed hash became "" and was stored
	if _, err := hashAndSalt(make([]byte, 5000)); err == nil {
		t.Fatal("hashing an over-long password didn't fail")
	}
}
