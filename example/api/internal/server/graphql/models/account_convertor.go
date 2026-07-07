package models

import (
	"github.com/demdxx/xtypes"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
	basemodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// Filter converts extended GraphQL filter to domain filter (includes title).
func (fl *AccountListFilter) Filter() *account.Filter {
	if fl == nil {
		return nil
	}
	return &account.Filter{
		ID:     fl.ID,
		UserID: fl.UserID,
		Title:  fl.Title,
		Status: xtypes.SliceApply(fl.Status, func(st basemodels.ApproveStatus) pkgModels.ApproveStatus {
			return st.ModelStatus()
		}),
	}
}

// Order converts extended GraphQL order to domain order (includes title).
func (ord *AccountListOrder) Order() *account.ListOrder {
	if ord == nil {
		return nil
	}
	return &account.ListOrder{
		ID:        ord.ID.AsOrder(),
		Title:     ord.Title.AsOrder(),
		Status:    ord.Status.AsOrder(),
		CreatedAt: ord.CreatedAt.AsOrder(),
		UpdatedAt: ord.UpdatedAt.AsOrder(),
	}
}

// Filter converts extended user list filter (core + example accountID/roles placeholders).
func (fl *UserListFilter) Filter() *user.ListFilter {
	if fl == nil {
		return nil
	}
	return &user.ListFilter{
		FilterBase:  user.FilterBase{ID: fl.ID},
		FilterEmail: user.FilterEmail{Emails: fl.Emails},
	}
}

// Order converts extended user list order.
func (ord *UserListOrder) Order() *user.ListOrder {
	if ord == nil {
		return nil
	}
	emailOrder := ord.Username
	if ord.Email != nil {
		emailOrder = ord.Email
	}
	regOrder := ord.CreatedAt
	if ord.RegistrationDate != nil {
		regOrder = ord.RegistrationDate
	}
	return &user.ListOrder{
		OrderBase: user.OrderBase{
			ID:        ord.ID.AsOrder(),
			Status:    ord.Status.AsOrder(),
			CreatedAt: regOrder.AsOrder(),
			UpdatedAt: ord.UpdatedAt.AsOrder(),
		},
		OrderEmail: user.OrderEmail{
			Email: emailOrder.AsOrder(),
		},
	}
}
