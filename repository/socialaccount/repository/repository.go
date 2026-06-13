package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
)

// Repository for social account
type Repository struct {
	repository.Repository
}

// NewSocaccRepository social account repository
func NewSocaccRepository() *Repository {
	return &Repository{}
}

// Get social account by ID
func (r *Repository) Get(ctx context.Context, id uint64) (*models.AccountSocial, error) {
	object := &models.AccountSocial{}
	res := r.Slave(ctx).Model(object).
		Preload(clause.Associations).
		Where(`id=?`, id).Find(object)
	if err := res.Error; err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList of social accounts
func (r *Repository) FetchList(ctx context.Context, opts ...socialaccount.QOption) ([]*models.AccountSocial, error) {
	var (
		list  []*models.AccountSocial
		query = r.Slave(ctx).Model((*models.AccountSocial)(nil))
	)
	query = socialaccount.ListOptions(opts).PrepareQuery(query)
	query = query.Preload(clause.Associations)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

// Count of social accounts
func (r *Repository) Count(ctx context.Context, opts ...socialaccount.QOption) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*models.AccountSocial)(nil))
	)
	query = socialaccount.ListOptions(opts).PrepareQuery(query)
	err := query.Count(&count).Error
	return count, err
}

// Disconnect social account by ID
func (r *Repository) Disconnect(ctx context.Context, id uint64) error {
	return database.ContextTransactionExec(ctx, func(ctx context.Context, tx *gorm.DB) error {
		if err := tx.Delete(&models.AccountSocialSession{}, `account_social_id=?`, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.AccountSocial{}, `id=?`, id).Error; err != nil {
			return err
		}
		return nil
	})
}

// FetchSessionList of social account
func (r *Repository) FetchSessionList(ctx context.Context, socialAccountID []uint64) ([]*models.AccountSocialSession, error) {
	var (
		list  []*models.AccountSocialSession
		query = r.Slave(ctx).Model((*models.AccountSocialSession)(nil))
	)
	if len(socialAccountID) > 0 {
		query = query.Where(`account_social_id IN (?)`, socialAccountID)
	}
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}
