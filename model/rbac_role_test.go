package model

import (
	"testing"

	"github.com/geniusrabbit/gosql/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetTitle(t *testing.T) {
	jctx, _ := gosql.NewNullableJSON[map[string]any](map[string]string{
		`object`: `model:User`,
		`cover`:  `system`,
	})
	role := &Role{
		ID:      1,
		Type:    RbacRoleType,
		Name:    `view`,
		Context: *jctx,
	}
	assert.Equal(t, `model:User:view system`, role.GetTitle())
}
