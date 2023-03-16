package session

import (
	"context"
	"testing"

	scs "github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	ctx := context.Background()
	ctx = WithSession(ctx, &scs.SessionManager{})
	assert.NotNil(t, Get(ctx))
}
