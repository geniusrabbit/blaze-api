package models

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// FromUserModel to local graphql model
func FromUserModel(u *model.User) *User {
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
func FromUserModelList(list []*model.User) []*User {
	return xtypes.SliceApply(list, FromUserModel)
}

// Filter converts local graphql model to filter
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

func (usr *UserInput) Model(appStatus ...model.ApproveStatus) *model.User {
	if usr == nil {
		return nil
	}
	var status model.ApproveStatus
	if len(appStatus) == 0 {
		status = usr.Status.ModelStatus()
	} else {
		status = appStatus[0]
	}
	return &model.User{
		Email:   s4ptr(usr.Username),
		Approve: status,
	}
}
