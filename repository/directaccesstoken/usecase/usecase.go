package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
)

// Usecase handles business logic for direct access tokens.
type Usecase struct {
	repo directaccesstoken.Repository
}

// New creates a new direct access token usecase instance.
func New(repo directaccesstoken.Repository) *Usecase {
	return &Usecase{repo: repo}
}

// Get retrieves a direct access token by ID with permission checks.
func (u *Usecase) Get(ctx context.Context, id uint64) (*models.DirectAccessToken, error) {
	accToken, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, accToken) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "get access token")
	}
	return accToken, nil
}

// FetchList retrieves a filtered and paginated list of direct access tokens.
func (u *Usecase) FetchList(ctx context.Context, filter *directaccesstoken.Filter, order *directaccesstoken.ListOrder, page *repository.Pagination) ([]*models.DirectAccessToken, error) {
	if !acl.HaveAccessList(ctx, &models.DirectAccessToken{}) {
		acc := session.Account(ctx)
		if !acl.HaveAccessList(ctx, &models.DirectAccessToken{AccountID: acc.ID}) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list access tokens")
		}
		filter.AccountID = []uint64{acc.ID}
	}
	return u.repo.FetchList(ctx, filter, order, page)
}

// Count returns the total count of direct access tokens matching the filter.
func (u *Usecase) Count(ctx context.Context, filter *directaccesstoken.Filter) (int64, error) {
	if !acl.HaveAccessCount(ctx, &models.DirectAccessToken{}) {
		acc := session.Account(ctx)
		if !acl.HaveAccessCount(ctx, &models.DirectAccessToken{AccountID: acc.ID}) {
			return 0, errors.Wrap(acl.ErrNoPermissions, "count access tokens")
		}
		filter.AccountID = []uint64{acc.ID}
	}
	return u.repo.Count(ctx, filter)
}

// Generate creates a new direct access token.
func (u *Usecase) Generate(ctx context.Context, userID, accountID uint64, description string, expiresAt time.Time) (*models.DirectAccessToken, error) {
	if !acl.HaveAccessCreate(ctx, &models.DirectAccessToken{
		UserID:    sql.Null[uint64]{V: userID, Valid: userID > 0},
		AccountID: accountID,
	}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "generate access token")
	}
	return u.repo.Generate(ctx, userID, accountID, description, expiresAt)
}

// Revoke revokes direct access tokens matching the filter criteria.
func (u *Usecase) Revoke(ctx context.Context, filter *directaccesstoken.Filter) error {
	if !acl.HaveAccessDelete(ctx, &models.DirectAccessToken{}) {
		acc := session.Account(ctx)
		if !acl.HaveAccessList(ctx, &models.DirectAccessToken{AccountID: acc.ID}) {
			return errors.Wrap(acl.ErrNoPermissions, "revoke access tokens")
		}
		filter.AccountID = []uint64{acc.ID}
	}
	return u.repo.Revoke(ctx, filter)
}
