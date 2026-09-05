package usecase

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

func sessionUserModel(ctx context.Context) user.Model {
	if u, _ := session.UserAccount(ctx); !gocast.IsNil(u) {
		return u
	}
	return nil
}
