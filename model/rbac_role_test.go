package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTitle(t *testing.T) {
	role := &Role{ID: 1, Name: `admin`}
	assert.Equal(t, `admin`, role.GetTitle())
}
