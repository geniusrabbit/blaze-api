package password

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLongPasswords(t *testing.T) {
	// Before the fix bcrypt rejected password+salt over 72 bytes and an
	// empty hash was stored, locking the account.
	SetSalt([]byte("cm2REbMsJfzVBN8vtd1NPw3UI75ef-zucSdZqqQQMo0"), bcrypt.MinCost)
	t.Cleanup(func() { SetSalt(nil, 0) })

	long := bytes.Repeat([]byte("p"), 200)
	hash, err := PasswordHash(long)
	if err != nil {
		t.Fatalf("hash a 200-byte password: %v", err)
	}
	if !strings.HasPrefix(hash, v2Prefix) || NeedsRehash(hash) {
		t.Fatalf("hash = %q, want the v2 scheme", hash)
	}
	if err := ComparePasswords(hash, long); err != nil {
		t.Fatalf("compare: %v", err)
	}
	// The whole password matters, not only the first 72 bytes
	if err := ComparePasswords(hash, append(bytes.Repeat([]byte("p"), 199), 'x')); !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("different long password: err = %v", err)
	}

	if _, err := PasswordHash(bytes.Repeat([]byte("p"), MaxLength+1)); !errors.Is(err, ErrTooLong) {
		t.Fatalf("over MaxLength: err = %v", err)
	}
}

func TestLegacyHashesStillVerify(t *testing.T) {
	SetSalt([]byte("legacy-salt"), bcrypt.MinCost)
	t.Cleanup(func() { SetSalt(nil, 0) })

	legacy, err := bcrypt.GenerateFromPassword([]byte("secret-pw"+"legacy-salt"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := ComparePasswords(string(legacy), []byte("secret-pw")); err != nil {
		t.Fatalf("legacy hash: %v", err)
	}
	if err := ComparePasswords(string(legacy), []byte("wrong")); !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("legacy hash, wrong password: err = %v", err)
	}
	if !NeedsRehash(string(legacy)) {
		t.Fatal("legacy hash should need a rehash")
	}
}

func TestGenerateResetToken(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}
	for range 50 {
		token := GenerateResetToken(64)
		if len(token) != 64 || strings.Trim(token, tokenAlphabet) != "" {
			t.Fatalf("token %q: wrong length or alphabet", token)
		}
		if seen[token] {
			t.Fatal("duplicate token")
		}
		seen[token] = true
	}
}
