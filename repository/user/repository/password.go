package repository

import (
	"math/rand"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/repository/user"
)

func (r *Repository) hashAndSalt(pwd []byte) string {
	hash, err := user.PasswordHash(pwd)
	if err != nil {
		zap.L().Error("GenerateFromPassword", zap.Error(err))
	}
	return hash
}

func (r *Repository) comparePasswords(hashedPwd string, plainPwd []byte) bool {
	err := user.ComparePasswords(hashedPwd, plainPwd)
	if err != nil {
		zap.L().Error("CompareHashAndPassword", zap.Error(err))
		return false
	}
	return true
}

// RandomPassword returns a random password of the given length
func RandomPassword(length int) string {
	const (
		digits   = "0123456789"
		specials = "~=+%^*/()[]{}/!@#$?|"
		all      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz" +
			digits + specials
	)
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) { buf[i], buf[j] = buf[j], buf[i] })
	return string(buf)
}
