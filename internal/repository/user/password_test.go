package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Password(t *testing.T) {
	SetSalt([]byte("cm2REbMsJfzVBN8vtd1NPw3UI75ef-zucSdZqqQQMo0"), 0)

	var (
		testPassword = []byte("test")
		hash, err    = PasswordHash(testPassword)
	)
	assert.NoError(t, err)
	assert.NoError(t, ComparePasswords(hash, testPassword))
}
