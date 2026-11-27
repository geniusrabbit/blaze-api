package password

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// MaxLength bounds the accepted password length.
const MaxLength = 4096

// ErrTooLong is returned for passwords longer than MaxLength.
var ErrTooLong = errors.New("password is too long")

// v2Prefix marks hashes of the pre-hashed scheme (see PasswordHash).
const v2Prefix = "v2$"

// salt is the additional noise in the password to improve cryptographic strength
var (
	salt []byte
	cost = bcrypt.DefaultCost
)

// SetSalt for passowrd
func SetSalt(s []byte, c int) {
	salt = s
	if c < 1 {
		c = bcrypt.DefaultCost
	}
	cost = c
}

// prehash maps a password of any length to 43 bytes: HMAC-SHA256 keyed with
// the salt, base64 encoded (bcrypt input must be at most 72 bytes and
// without NUL bytes).
func prehash(pwd []byte) []byte {
	mac := hmac.New(sha256.New, salt)
	_, _ = mac.Write(pwd)
	sum := mac.Sum(nil)
	out := make([]byte, base64.RawStdEncoding.EncodedLen(len(sum)))
	base64.RawStdEncoding.Encode(out, sum)
	return out
}

// PasswordHash returns the bcrypt hash of the pre-hashed password.
//
// The old scheme hashed password+salt directly; bcrypt rejects input over
// 72 bytes, so long passwords (or long salts) failed and an empty hash was
// stored. Hashes of the new scheme start with "v2$"; ComparePasswords
// accepts both.
func PasswordHash(pwd []byte) (string, error) {
	if len(pwd) > MaxLength {
		return "", ErrTooLong
	}
	hash, err := bcrypt.GenerateFromPassword(prehash(pwd), cost)
	if err != nil {
		return "", err
	}
	return v2Prefix + string(hash), nil
}

// ComparePasswords between income password and generated hash.
// It returns bcrypt.ErrMismatchedHashAndPassword for a wrong password.
func ComparePasswords(hashedPwd string, plainPwd []byte) error {
	if len(plainPwd) > MaxLength {
		return bcrypt.ErrMismatchedHashAndPassword
	}
	if hash, ok := strings.CutPrefix(hashedPwd, v2Prefix); ok {
		return bcrypt.CompareHashAndPassword([]byte(hash), prehash(plainPwd))
	}
	// Legacy scheme
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), append(append([]byte{}, plainPwd...), salt...))
}

// NeedsRehash reports whether the hash uses the legacy scheme and should be
// replaced (e.g. after the next successful login).
func NeedsRehash(hashedPwd string) bool {
	return hashedPwd != "" && !strings.HasPrefix(hashedPwd, v2Prefix)
}
