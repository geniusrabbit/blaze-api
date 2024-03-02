package models

import (
	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

func FromSocialAccountModel(acc *model.AccountSocial) *SocialAccount {
	if acc == nil {
		return nil
	}
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

		Data:     *types.MustNullableJSONFrom(acc.Data.Data),
		Sessions: FromSocialAccountSessionModelList(acc.Sessions),

		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
		DeletedAt: DeletedAt(acc.DeletedAt),
	}
}

func FromSocialAccountModelList(list []*model.AccountSocial) []*SocialAccount {
	return xtypes.SliceApply(list, FromSocialAccountModel)
}

func FromSocialAccountSessionModel(sess *model.AccountSocialSession) *SocialAccountSession {
	if sess == nil {
		return nil
	}
	return &SocialAccountSession{
		Name:            sess.Name,
		SocialAccountID: sess.AccountSocialID,

		AccessToken:  sess.AccessToken,
		RefreshToken: sess.RefreshToken,
		Scope:        sess.Scopes,

		ExpiresAt: gocast.IfThen(sess.ExpiresAt.Valid, &sess.ExpiresAt.Time, nil),
		CreatedAt: sess.CreatedAt,
		UpdatedAt: sess.UpdatedAt,
		DeletedAt: DeletedAt(sess.DeletedAt),
	}
}

func FromSocialAccountSessionModelList(list []*model.AccountSocialSession) []*SocialAccountSession {
	return xtypes.SliceApply(list, FromSocialAccountSessionModel)
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
