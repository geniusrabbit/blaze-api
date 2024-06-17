package directaccesstoken

import (
	"crypto/rand"
)

// GenerateToken generates a random token of specified length in bytes
func GenerateToken(length int) (string, error) {
	// Create a byte slice to hold the random data
	token := make([]byte, length)

	// Read random data into the byte slice
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	// Encode the byte slice into a hexadecimal string
	return base86(token), nil
}

func base86(data []byte) string {
	const base86Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!#$%&()*+-;<=>?@^_`{|}~"
	const base86CharsLen = byte(len(base86Chars))

	// Create a byte slice to hold the encoded data
	encoded := make([]byte, len(data)*2)

	// Encode the data
	for i, b := range data {
		encoded[i*2] = base86Chars[b/base86CharsLen]
		encoded[i*2+1] = base86Chars[b%base86CharsLen]
	}

	// Return the encoded data as a string
	return string(encoded)
}
