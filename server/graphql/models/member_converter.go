package models

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/guregu/null"
)

// FromMemberModel to local graphql model
func FromMemberModel(ctx context.Context, member *model.AccountMember) *Member {
	if member == nil {
		return nil
	}
	return &Member{
		ID:        member.ID,
		Account:   FromAccountModel(gocast.Or(member.Account, &model.Account{ID: member.AccountID})),
		User:      FromUserModel(gocast.Or(member.User, &model.User{ID: member.UserID})),
		IsAdmin:   member.IsAdmin,
		Status:    ApproveStatusFrom(member.Approve),
		Roles:     FromRBACRoleModelList(ctx, member.Roles),
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}
}

func FromMemberModelList(ctx context.Context, list []*model.AccountMember) []*Member {
	return xtypes.SliceApply(list, func(m *model.AccountMember) *Member {
		return FromMemberModel(ctx, m)
	})
}

func (fl *MemberListFilter) Filter() *account.MemberFilter {
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

func (ord *MemberListOrder) Order() *account.MemberListOrder {
	if ord == nil {
		return nil
	}
	return &account.MemberListOrder{
		ID:        model.Order(ord.ID.AsOrder()),
		AccountID: model.Order(ord.AccountID.AsOrder()),
		UserID:    model.Order(ord.UserID.AsOrder()),
		Status:    model.Order(ord.Status.AsOrder()),
		IsAdmin:   model.Order(ord.IsAdmin.AsOrder()),
		CreatedAt: model.Order(ord.CreatedAt.AsOrder()),
		UpdatedAt: model.Order(ord.UpdatedAt.AsOrder()),
	}
}

func (mem *InviteMemberInput) AllRoles() []string {
	if mem.IsAdmin {
		return xtypes.SliceUnique(append(mem.Roles, account.RoleAdmin))
	}
	return mem.Roles
}

func (mem *MemberInput) AllRoles() []string {
	if mem.IsAdmin {
		return xtypes.SliceUnique(append(mem.Roles, account.RoleAdmin))
	}
	return mem.Roles
}
