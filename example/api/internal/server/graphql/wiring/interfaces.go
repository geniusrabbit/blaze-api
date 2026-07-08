package wiring

import (
	"context"

	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// EmailPasswordLoginHandler handles the login(email, password, accountID) mutation.
type EmailPasswordLoginHandler interface {
	Login(ctx context.Context, email, password string, accountID ...uint64) (*gqlmodels.SessionToken, error)
}
