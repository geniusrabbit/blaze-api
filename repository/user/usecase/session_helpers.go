package usecase

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

func sessionUserModel(ctx context.Context) user.Model {
	u, _ := session.UserAccount(ctx)
	return u
}
