package password

import (
	"crypto/rand"
	"math/big"
)

// tokenAlphabet is URL-safe, so tokens can go into reset links as is.
const tokenAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"

// GenerateResetToken generates a random token for password reset.
// It uses crypto/rand: math/rand tokens are predictable.
func GenerateResetToken(size int) string {
	return RandomString(size, tokenAlphabet)
}

// RandomString returns a cryptographically random string of the given
// length from the alphabet, without modulo bias. It panics if the system
// random source fails, as no safe fallback exists.
func RandomString(size int, alphabet string) string {
	buf := make([]byte, size)
	limit := big.NewInt(int64(len(alphabet)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			panic(err)
		}
		buf[i] = alphabet[n.Int64()]
	}
	return string(buf)
}
