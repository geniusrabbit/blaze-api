package password

import (
	"bytes"
	"math/rand"
)

const tokenAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-$@"

// GenerateResetToken generates a random token for password reset
func GenerateResetToken(size int) string {
	var randomString bytes.Buffer
	for i := 0; i < size; i++ {
		randomString.WriteByte(tokenAlphabet[rand.Intn(len(tokenAlphabet))])
	}
	return randomString.String()
}
