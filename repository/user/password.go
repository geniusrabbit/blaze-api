package user

import (
	"golang.org/x/crypto/bcrypt"
)

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

// ComparePasswords between income password and generated hash
func ComparePasswords(hashedPwd string, plainPwd []byte) error {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	return bcrypt.CompareHashAndPassword(byteHash, append(plainPwd, salt...))
}

// PasswordHash process and return processed password to bcrypt hash
func PasswordHash(pwd []byte) (string, error) {
	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(append(pwd, salt...), cost)
	if err != nil {
		return "", err
	}

	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash), nil
}
