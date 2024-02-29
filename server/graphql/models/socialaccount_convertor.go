package models

import (
	"github.com/demdxx/xtypes"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

func FromSocialAccountModel(acc *model.AccountSocial) *SocialAccount {
	return &SocialAccount{
		ID:     acc.ID,
		UserID: acc.UserID,

		SocialID:  acc.SocialID,
		Provider:  acc.Provider,
		Username:  acc.Username,
		Email:     acc.Email,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		Avatar:    acc.Avatar,
		Link:      acc.Link,
		Scope:     acc.Scope,

		Data: *types.MustNullableJSONFrom(acc.Data.Data),

		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
		DeletedAt: DeletedAt(acc.DeletedAt),
	}
}

func FromSocialAccountModelList(list []*model.AccountSocial) []*SocialAccount {
	return xtypes.SliceApply(list, FromSocialAccountModel)
}

func (fl *SocialAccountListFilter) Filter() *socialaccount.Filter {
	if fl == nil {
		return nil
	}
	return &socialaccount.Filter{
		ID:       fl.ID,
		UserID:   fl.UserID,
		Provider: fl.Provider,
		Username: fl.Username,
		Email:    fl.Email,
	}
}

func (ord *SocialAccountListOrder) Order() *socialaccount.Order {
	if ord == nil {
		return nil
	}
	return &socialaccount.Order{
		ID:        ord.ID.AsOrder(),
		UserID:    ord.UserID.AsOrder(),
		Provider:  ord.Provider.AsOrder(),
		Email:     ord.Email.AsOrder(),
		Username:  ord.Username.AsOrder(),
		FirstName: ord.FirstName.AsOrder(),
		LastName:  ord.LastName.AsOrder(),
	}
}
