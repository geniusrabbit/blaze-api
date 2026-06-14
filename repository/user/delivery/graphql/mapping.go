package graphql

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// FromUserModel to local graphql model
func FromUserModel(u *user.User) *gqlmodels.User {
	if u == nil {
		return nil
	}
	return &gqlmodels.User{
		ID:        u.ID,
		Username:  u.Email,
		Status:    gqlmodels.ApproveStatusFrom(u.Approve),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// FromUserModelList converts model list to local model list
func FromUserModelList(list []*user.User) []*gqlmodels.User {
	return xtypes.SliceApply(list, FromUserModel)
}
