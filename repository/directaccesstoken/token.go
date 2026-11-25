package directaccesstoken

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateToken generates a random token from length random bytes.
// The token uses the unpadded base64url alphabet, so it is safe in URLs,
// HTTP headers and double-quoted shell strings.
func GenerateToken(length int) (string, error) {
	token := make([]byte, length)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}
