package graphql

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	rbacgql "github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql"
	usergql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// FromAccountModel to local graphql model
func FromAccountModel(acc *models.Account) *gqlmodels.Account {
	if acc == nil {
		return nil
	}
	return &gqlmodels.Account{
		ID:                acc.ID,
		Status:            gqlmodels.ApproveStatusFrom(acc.Approve),
		Title:             acc.Title,
		Description:       acc.Description,
		LogoURI:           acc.LogoURI,
		PolicyURI:         acc.PolicyURI,
		TermsOfServiceURI: acc.TermsOfServiceURI,
		ClientURI:         acc.ClientURI,
		Contacts:          acc.Contacts,
		CreatedAt:         acc.CreatedAt,
		UpdatedAt:         acc.UpdatedAt,
	}
}

// FromAccountModelList converts model list to local model list
func FromAccountModelList(list []*models.Account) []*gqlmodels.Account {
	return xtypes.SliceApply(list, FromAccountModel)
}

// FromMemberModel to local graphql model
func FromMemberModel(ctx context.Context, member *models.AccountMember) *gqlmodels.Member {
	if member == nil {
		return nil
	}
	return &gqlmodels.Member{
		ID:        member.ID,
		Account:   FromAccountModel(gocast.Or(member.Account, &model.Account{ID: member.AccountID})),
		User:      usergql.FromUserModel(gocast.Or(member.User, &model.User{ID: member.UserID})),
		IsAdmin:   member.IsAdmin,
		Status:    gqlmodels.ApproveStatusFrom(member.Approve),
		Roles:     rbacgql.FromRBACRoleModelList(ctx, member.Roles),
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}
}

func FromMemberModelList(ctx context.Context, list []*models.AccountMember) []*gqlmodels.Member {
	return xtypes.SliceApply(list, func(m *models.AccountMember) *gqlmodels.Member {
		return FromMemberModel(ctx, m)
	})
}
