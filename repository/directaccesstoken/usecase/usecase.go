package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
)

type Usecase struct {
	repo directaccesstoken.Repository
}

// New direct access token usecase
func New(repo directaccesstoken.Repository) *Usecase {
	return &Usecase{repo: repo}
}

// Get direct access token by token
func (u *Usecase) Get(ctx context.Context, id uint64) (*model.DirectAccessToken, error) {
	accToken, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, accToken) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "get access token")
	}
	return accToken, nil
}

// FetchList of direct access tokens
func (u *Usecase) FetchList(ctx context.Context, filter *directaccesstoken.Filter, order *directaccesstoken.Order, page *repository.Pagination) ([]*model.DirectAccessToken, error) {
	if !acl.HaveAccessList(ctx, &model.DirectAccessToken{}) {
		acc := session.Account(ctx)
		if !acl.HaveAccessList(ctx, &model.DirectAccessToken{AccountID: acc.ID}) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list access tokens")
		}
		filter.AccountID = []uint64{acc.ID}
	}
	return u.repo.FetchList(ctx, filter, order, page)
}

// Count of direct access tokens
func (u *Usecase) Count(ctx context.Context, filter *directaccesstoken.Filter) (int64, error) {
	if !acl.HaveAccessCount(ctx, &model.DirectAccessToken{}) {
		acc := session.Account(ctx)
		if !acl.HaveAccessCount(ctx, &model.DirectAccessToken{AccountID: acc.ID}) {
			return 0, errors.Wrap(acl.ErrNoPermissions, "count access tokens")
		}
		filter.AccountID = []uint64{acc.ID}
	}
	return u.repo.Count(ctx, filter)
}

// Generate access token
func (u *Usecase) Generate(ctx context.Context, userID, accountID uint64, description string, expiresAt time.Time) (*model.DirectAccessToken, error) {
	if !acl.HaveAccessCreate(ctx, &model.DirectAccessToken{
		UserID:    sql.Null[uint64]{V: userID, Valid: userID > 0},
		AccountID: accountID,
	}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "generate access token")
	}
	return u.repo.Generate(ctx, userID, accountID, description, expiresAt)
}

// Revoke access tokens
func (u *Usecase) Revoke(ctx context.Context, filter *directaccesstoken.Filter) error {
	if !acl.HaveAccessDelete(ctx, &model.DirectAccessToken{}) {
		acc := session.Account(ctx)
		if !acl.HaveAccessList(ctx, &model.DirectAccessToken{AccountID: acc.ID}) {
			return errors.Wrap(acl.ErrNoPermissions, "revoke access tokens")
		}
		filter.AccountID = []uint64{acc.ID}
	}
	return u.repo.Revoke(ctx, filter)
}
