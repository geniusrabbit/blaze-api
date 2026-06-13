package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
)

// Repository handles direct access token database operations.
type Repository struct {
	repository.Repository
}

// NewDirectAccessTokenRepository creates and returns a new direct access token repository instance.
func NewDirectAccessTokenRepository() *Repository {
	return &Repository{}
}

// Get retrieves a non-expired direct access token by ID.
func (r *Repository) Get(ctx context.Context, id uint64) (*models.DirectAccessToken, error) {
	object := new(models.DirectAccessToken)
	err := r.Slave(ctx).Model(object).
		Find(object, "id=? AND expires_at>NOW()", id).Error
	if err != nil {
		return nil, err
	}
	return object, nil
}

// GetByToken retrieves a non-expired direct access token by its token value.
func (r *Repository) GetByToken(ctx context.Context, token string) (*models.DirectAccessToken, error) {
	object := new(models.DirectAccessToken)
	err := r.Slave(ctx).Model(object).
		Find(object, "token=? AND expires_at>NOW()", token).Error
	if err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList retrieves a paginated list of direct access tokens with optional filtering and ordering.
func (r *Repository) FetchList(ctx context.Context, filter *directaccesstoken.Filter, order *directaccesstoken.ListOrder, page *repository.Pagination) ([]*models.DirectAccessToken, error) {
	objects := make([]*models.DirectAccessToken, 0)
	query := r.Slave(ctx).Model(&models.DirectAccessToken{})
	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
	query = page.PrepareQuery(query)
	err := query.Find(&objects).Error
	if err != nil {
		return nil, err
	}
	return objects, nil
}

// Count returns the total number of direct access tokens matching the filter criteria.
func (r *Repository) Count(ctx context.Context, filter *directaccesstoken.Filter) (int64, error) {
	var count int64
	query := r.Slave(ctx).Model(&models.DirectAccessToken{})
	query = filter.PrepareQuery(query)
	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Generate creates and stores a new direct access token with the specified parameters.
func (r *Repository) Generate(ctx context.Context, userID, accountID uint64, description string, expiresAt time.Time) (*models.DirectAccessToken, error) {
	token, err := directaccesstoken.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	object := &models.DirectAccessToken{
		Token:       token,
		Description: description,
		UserID:      sql.Null[uint64]{V: userID, Valid: userID > 0},
		AccountID:   accountID,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
	}
	err = r.Master(ctx).Create(object).Error
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Revoke invalidates direct access tokens by setting their expiration to the past.
func (r *Repository) Revoke(ctx context.Context, filter *directaccesstoken.Filter) error {
	query := r.Master(ctx).Model(&models.DirectAccessToken{})
	query = filter.PrepareQuery(query)
	return query.UpdateColumn("expires_at", time.Now().Add(-time.Hour)).Error
}
