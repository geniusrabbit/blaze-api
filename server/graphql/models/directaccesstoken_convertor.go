package models

import (
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
)

func FromDirectAccessToken(token *model.DirectAccessToken) *DirectAccessToken {
	if token == nil {
		return nil
	}
	return &DirectAccessToken{
		ID:          token.ID,
		Token:       token.Token,
		UserID:      gocast.IfThen(token.UserID.Valid, &token.UserID.V, nil),
		AccountID:   token.AccountID,
		Description: token.Description,
		CreatedAt:   token.CreatedAt,
		ExpiresAt:   token.ExpiresAt,
	}
}

func FromDirectAccessTokenModelList(list []*model.DirectAccessToken) []*DirectAccessToken {
	return xtypes.SliceApply(list, func(m *model.DirectAccessToken) *DirectAccessToken {
		return FromDirectAccessToken(m)
	})
}

func (fl *DirectAccessTokenListFilter) Filter() *directaccesstoken.Filter {
	if fl == nil {
		return nil
	}
	return &directaccesstoken.Filter{
		ID:        fl.ID,
		Token:     fl.Token,
		UserID:    fl.UserID,
		AccountID: fl.AccountID,
		MinExpiresAt: gocast.IfThenExec(fl.MinExpiresAt != nil,
			func() time.Time { return *fl.MinExpiresAt }, func() time.Time { return time.Time{} }),
		MaxExpiresAt: gocast.IfThenExec(fl.MaxExpiresAt != nil,
			func() time.Time { return *fl.MaxExpiresAt }, func() time.Time { return time.Time{} }),
	}
}

func (ord *DirectAccessTokenListOrder) Order() *directaccesstoken.Order {
	if ord == nil {
		return nil
	}
	return &directaccesstoken.Order{
		ID:        ord.ID.AsOrder(),
		Token:     ord.Token.AsOrder(),
		UserID:    ord.UserID.AsOrder(),
		AccountID: ord.AccountID.AsOrder(),
		ExpiresAt: ord.ExpiresAt.AsOrder(),
		CreatedAt: ord.CreatedAt.AsOrder(),
	}
}
