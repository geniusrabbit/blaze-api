package models

import (
	"github.com/demdxx/xtypes"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// FromUserModel to local graphql model
func FromUserModel(u *user.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		ID:        u.ID,
		Username:  u.Email,
		Status:    ApproveStatusFrom(u.Approve),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// FromUserModelList converts model list to local model list
func FromUserModelList(list []*user.User) []*User {
	return xtypes.SliceApply(list, FromUserModel)
}

// Filter converts local graphql model to filter
func (fl *UserListFilter) Filter() *user.ListFilter {
	if fl == nil {
		return nil
	}
	return &user.ListFilter{
		UserID: fl.ID,
		Emails: fl.Emails,
	}
}

// Order converts local graphql model to order
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

func (usr *UserInput) Model(appStatus ...pkgModels.ApproveStatus) *user.User {
	if usr == nil {
		return nil
	}
	var status pkgModels.ApproveStatus
	if len(appStatus) == 0 {
		status = usr.Status.ModelStatus()
	} else {
		status = appStatus[0]
	}
	return &user.User{
		Email:   s4ptr(usr.Username),
		Approve: status,
	}
}
