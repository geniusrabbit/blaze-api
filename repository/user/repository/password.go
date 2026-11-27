package repository

import (
	"crypto/rand"
	"errors"
	"math/big"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/geniusrabbit/blaze-api/repository/user/password"
)

func hashAndSalt(pwd []byte) (string, error) {
	return password.PasswordHash(pwd)
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	err := password.ComparePasswords(hashedPwd, plainPwd)
	if err != nil && !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		// A wrong password is not an error worth logging; a broken hash is
		zap.L().Error("CompareHashAndPassword", zap.Error(err))
	}
	return err == nil
}

// RandomPassword returns a cryptographically random password of the given
// length (at least 2) with at least one digit and one special character.
func RandomPassword(length int) string {
	const (
		digits   = "0123456789"
		specials = "~=+%^*/()[]{}/!@#$?|"
		all      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz" +
			digits + specials
	)
	length = max(length, 2)
	buf := []byte(password.RandomString(1, digits) + password.RandomString(1, specials) +
		password.RandomString(length-2, all))
	// Fisher-Yates shuffle with crypto/rand
	for i := len(buf) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			panic(err)
		}
		buf[i], buf[j.Int64()] = buf[j.Int64()], buf[i]
	}
	return string(buf)
}
