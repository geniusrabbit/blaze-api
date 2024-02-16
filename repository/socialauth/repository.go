package socialauth

import (
	"context"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/elogin"
	"github.com/geniusrabbit/blaze-api/model"
)

type Filter struct {
	UserID   []uint64
	SocialID []string
	Provider []string
	Email    []string
}

func (fl *Filter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	if len(fl.UserID) > 0 {
		query = query.Where(`user_id IN (?)`, fl.UserID)
	}
	if len(fl.SocialID) > 0 {
		query = query.Where(`social_id IN (?)`, fl.SocialID)
	}
	if len(fl.Provider) > 0 {
		query = query.Where(`provider IN (?)`, fl.Provider)
	}
	if len(fl.Email) > 0 {
		query = query.Where(`email IN (?)`, fl.Email)
	}
	return query
}

type Repository interface {
	Get(ctx context.Context, id uint64) (*model.AccountSocial, error)
	List(ctx context.Context, filter *Filter) ([]*model.AccountSocial, error)
	Create(ctx context.Context, account *model.AccountSocial) (uint64, error)
	Update(ctx context.Context, id uint64, account *model.AccountSocial) error
	Token(ctx context.Context, id uint64) (*elogin.Token, error)
	SetToken(ctx context.Context, id uint64, token *elogin.Token) error
}
