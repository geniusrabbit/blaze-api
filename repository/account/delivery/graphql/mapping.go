package graphql

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"
	"github.com/guregu/null"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/account"
	rbac_graphql "github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql"
	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// FromMemberModel to local graphql model
func FromMemberModel[TUser user.Model, TAccount account.Model](
	ctx context.Context,
	member *account.Member[TUser, TAccount],
	accounts account.Usecase[TUser, TAccount],
	users user.Repository[TUser],
) *gqlmodels.Member {
	if member == nil {
		return nil
	}
	return &gqlmodels.Member{
		ID:        member.ID,
		AccountID: member.AccountID,
		UserID:    member.UserID,
		IsAdmin:   member.IsAdmin,
		Status:    gqlmodels.ApproveStatusFrom(member.Approve),
		Roles:     rbac_graphql.FromRBACRoleModelList(ctx, member.Roles),
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}
}

func FromMemberModelList[TUser user.Model, TAccount account.Model](
	ctx context.Context,
	list []*account.Member[TUser, TAccount],
	accounts account.Usecase[TUser, TAccount],
	users user.Repository[TUser],
) []*gqlmodels.Member {
	return xtypes.SliceApply(list, func(m *account.Member[TUser, TAccount]) *gqlmodels.Member {
		return FromMemberModel(ctx, m, accounts, users)
	})
}

func FromMemberGQLFilter(fl *gqlmodels.MemberListFilter) *account.MemberFilter {
	if fl == nil {
		return nil
	}
	return &account.MemberFilter{
		ID:        fl.ID,
		AccountID: fl.AccountID,
		UserID:    fl.UserID,
		IsAdmin:   gocast.IfThen(fl.IsAdmin != nil, null.BoolFromPtr(fl.IsAdmin), null.Bool{}),
	}
}

func FromMemberGQLOrder(ord *gqlmodels.MemberListOrder) *account.MemberListOrder {
	if ord == nil {
		return nil
	}
	return &account.MemberListOrder{
		ID:        pkgModels.Order(ord.ID.AsOrder()),
		AccountID: pkgModels.Order(ord.AccountID.AsOrder()),
		UserID:    pkgModels.Order(ord.UserID.AsOrder()),
		Status:    pkgModels.Order(ord.Status.AsOrder()),
		IsAdmin:   pkgModels.Order(ord.IsAdmin.AsOrder()),
		CreatedAt: pkgModels.Order(ord.CreatedAt.AsOrder()),
		UpdatedAt: pkgModels.Order(ord.UpdatedAt.AsOrder()),
	}
}

func InviteMemberAllRoles(mem *gqlmodels.InviteMemberInput) []string {
	if mem.IsAdmin {
		return xtypes.SliceUnique(append(mem.Roles, account.RoleAdmin))
	}
	return mem.Roles
}

func MemberAllRoles(mem *gqlmodels.MemberInput) []string {
	if mem.IsAdmin {
		return xtypes.SliceUnique(append(mem.Roles, account.RoleAdmin))
	}
	return mem.Roles
}
