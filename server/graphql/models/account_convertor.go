package models

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/account"
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
	return xtypes.SliceApply[*model.Account, *Account](list, FromAccountModel)
}

// Filter converts local graphql model to filter
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

// Model converts local graphql model to model
func (acc *AccountInput) Model(appStatus ...model.ApproveStatus) *model.Account {
	var status model.ApproveStatus
	if len(appStatus) == 0 {
		status = acc.Status.ModelStatus()
	} else {
		status = appStatus[0]
	}
	return &model.Account{
		Approve:           status,
		Title:             s4ptr(acc.Title),
		Description:       s4ptr(acc.Description),
		LogoURI:           s4ptr(acc.LogoURI),
		PolicyURI:         s4ptr(acc.PolicyURI),
		TermsOfServiceURI: s4ptr(acc.TermsOfServiceURI),
		ClientURI:         s4ptr(acc.ClientURI),
		Contacts:          append([]string{}, acc.Contacts...),
	}
}
