package models

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/api-template-base/internal/repository/user"
	"github.com/geniusrabbit/api-template-base/model"
)

// FromUserModel to local graphql model
func FromUserModel(u *model.User) *User {
	return &User{
		ID:        u.ID,
		Username:  u.Email,
		Status:    ApproveStatusFrom(u.Approve),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// FromUserModelList converts model list to local model list
func FromUserModelList(list []*model.User) []*User {
	users := make([]*User, 0, len(list))
	for _, u := range list {
		users = append(users, FromUserModel(u))
	}
	return users
}

func (fl *UserListFilter) Filter() *user.ListFilter {
	if fl == nil {
		return nil
	}
	return &user.ListFilter{
		UserID:    fl.ID,
		AccountID: fl.AccountID,
		Emails:    fl.Emails,
		Roles:     fl.Roles,
	}
}

func (ord *UserListOrder) Order() *user.ListOrder {
	if ord == nil {
		return nil
	}
	return &user.ListOrder{
		ID:        ord.ID.AsOrder(),
		Email:     xtypes.FirstVal(ord.Email, ord.Username).AsOrder(),
		Status:    ord.Status.AsOrder(),
		CreatedAt: ord.CreatedAt.AsOrder(),
		UpdatedAt: ord.UpdatedAt.AsOrder(),
	}
}
