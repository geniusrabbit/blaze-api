package acl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionError(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{
			msg:  "test",
			want: "no permissions: test",
		},
		{
			msg:  "test2",
			want: "no permissions: test2",
		},
	}

	for _, test := range tests {
		err := ErrNoPermissions.WithMessage(test.msg)
		assert.Equal(t, test.want, err.Error())
		assert.True(t, errors.Is(err, ErrNoPermissions))
	}
}
