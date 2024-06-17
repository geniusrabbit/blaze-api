package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
)

type Repository struct {
	repository.Repository
}

// New direct access token repository
func New() *Repository {
	return &Repository{}
}

// Get returns direct access token model by ID
func (r *Repository) Get(ctx context.Context, id uint64) (*model.DirectAccessToken, error) {
	object := new(model.DirectAccessToken)
	err := r.Slave(ctx).Model(object).
		Find(object, "id=? AND expires_at>NOW()", id).Error
	if err != nil {
		return nil, err
	}
	return object, nil
}

// GetByToken returns direct access token model by Token
func (r *Repository) GetByToken(ctx context.Context, token string) (*model.DirectAccessToken, error) {
	object := new(model.DirectAccessToken)
	err := r.Slave(ctx).Model(object).
		Find(object, "token=? AND expires_at>NOW()", token).Error
	if err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList returns list of direct access tokens
func (r *Repository) FetchList(ctx context.Context, filter *directaccesstoken.Filter, order *directaccesstoken.Order, page *repository.Pagination) ([]*model.DirectAccessToken, error) {
	objects := make([]*model.DirectAccessToken, 0)
	query := r.Slave(ctx).Model(&model.DirectAccessToken{})
	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
	query = page.PrepareQuery(query)
	err := query.Find(&objects).Error
	if err != nil {
		return nil, err
	}
	return objects, nil
}

// Count returns count of direct access tokens
func (r *Repository) Count(ctx context.Context, filter *directaccesstoken.Filter) (int64, error) {
	var count int64
	query := r.Slave(ctx).Model(&model.DirectAccessToken{})
	query = filter.PrepareQuery(query)
	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Generate creates a new direct access token
func (r *Repository) Generate(ctx context.Context, userID, accountID uint64, description string, expiresAt time.Time) (*model.DirectAccessToken, error) {
	token, err := directaccesstoken.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	object := &model.DirectAccessToken{
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

// Revoke access tokens
func (r *Repository) Revoke(ctx context.Context, filter *directaccesstoken.Filter) error {
	query := r.Master(ctx).Model(&model.DirectAccessToken{})
	query = filter.PrepareQuery(query)
	return query.UpdateColumn("expires_at", time.Now().Add(-time.Hour)).Error
}
