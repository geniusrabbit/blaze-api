package models

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/api-template-base/internal/repository/account"
	"github.com/geniusrabbit/api-template-base/model"
)

// FromAccountModel to local graphql model
func FromAccountModel(acc *model.Account) *Account {
	return &Account{
		ID:                acc.ID,
		Status:            ApproveStatusFrom(acc.Approve),
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
func FromAccountModelList(list []*model.Account) []*Account {
	accounts := make([]*Account, 0, len(list))
	for _, u := range list {
		accounts = append(accounts, FromAccountModel(u))
	}
	return accounts
}

func (fl *AccountListFilter) Filter() *account.Filter {
	if fl == nil {
		return nil
	}
	return &account.Filter{
		ID:     fl.ID,
		UserID: fl.UserID,
		Title:  fl.Title,
		Status: xtypes.SliceApply[ApproveStatus](fl.Status, func(st ApproveStatus) model.ApproveStatus {
			return st.ModelStatus()
		}),
	}
}
