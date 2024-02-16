package repository

import (
	"context"
	"errors"

	"github.com/geniusrabbit/blaze-api/elogin"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
	"github.com/guregu/null"
	"gorm.io/gorm"
)

type Repository struct {
	repository.Repository
}

func New() *Repository {
	return &Repository{}
}

// Get account by ID
func (r *Repository) Get(ctx context.Context, id uint64) (*model.AccountSocial, error) {
	var account model.AccountSocial
	err := r.Slave(ctx).First(&account, id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// List accounts by filter
func (r *Repository) List(ctx context.Context, filter *socialauth.Filter) ([]*model.AccountSocial, error) {
	var list []*model.AccountSocial
	query := r.Slave(ctx).Model((*model.AccountSocial)(nil))
	query = filter.PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, nil
}

// Create new account in the database
func (r *Repository) Create(ctx context.Context, account *model.AccountSocial) (uint64, error) {
	err := r.Master(ctx).Create(account).Error
	if err != nil {
		return 0, err
	}
	return account.ID, nil
}

// Update account in the database
func (r *Repository) Update(ctx context.Context, id uint64, account *model.AccountSocial) error {
	return r.Master(ctx).
		Model((*model.AccountSocial)(nil)).
		Where("id = ?", id).
		Updates(account).Error
}

// Token returns the token by social account ID
func (r *Repository) Token(ctx context.Context, id uint64) (*elogin.Token, error) {
	var (
		sess model.AccountSocialSession
		err  = r.Slave(ctx).Model((*model.AccountSocialSession)(nil)).
			Where("account_social_id = ?", id).
			First(&sess).Error
	)
	if err != nil {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &elogin.Token{
		TokenType:    sess.TokenType,
		AccessToken:  sess.AccessToken,
		RefreshToken: sess.RefreshToken,
		ExpiresAt:    sess.ExpiresAt.Time,
	}, nil
}

// SetToken saves the token to the social account
func (r *Repository) SetToken(ctx context.Context, id uint64, token *elogin.Token) error {
	return r.Master(ctx).Model((*model.AccountSocialSession)(nil)).
		Where("account_social_id = ?", id).
		Updates(model.AccountSocialSession{
			AccountSocialID: id,
			TokenType:       token.TokenType,
			AccessToken:     token.AccessToken,
			RefreshToken:    token.RefreshToken,
			ExpiresAt:       null.NewTime(token.ExpiresAt, !token.ExpiresAt.IsZero()),
		}).Error
}
