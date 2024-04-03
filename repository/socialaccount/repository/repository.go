package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
)

// Repository for social account
type Repository struct {
	repository.Repository
}

// New social account repository
func New() *Repository {
	return &Repository{}
}

// Get social account by ID
func (r *Repository) Get(ctx context.Context, id uint64) (*model.AccountSocial, error) {
	object := &model.AccountSocial{}
	res := r.Slave(ctx).Model(object).
		Preload(clause.Associations).
		Where(`id=?`, id).Find(object)
	if err := res.Error; err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList of social accounts
func (r *Repository) FetchList(ctx context.Context, filter *socialaccount.Filter, order *socialaccount.Order, pagination *repository.Pagination) ([]*model.AccountSocial, error) {
	var (
		list  []*model.AccountSocial
		query = r.Slave(ctx).Model((*model.AccountSocial)(nil))
	)
	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
	query = pagination.PrepareQuery(query)
	query = query.Preload(clause.Associations)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

// Count of social accounts
func (r *Repository) Count(ctx context.Context, filter *socialaccount.Filter) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*model.AccountSocial)(nil))
	)
	query = filter.PrepareQuery(query)
	err := query.Count(&count).Error
	return count, err
}

// Disconnect social account by ID
func (r *Repository) Disconnect(ctx context.Context, id uint64) error {
	return database.ContextTransactionExec(ctx, func(ctx context.Context, tx *gorm.DB) error {
		if err := tx.Delete(&model.AccountSocialSession{}, `account_social_id=?`, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.AccountSocial{}, `id=?`, id).Error; err != nil {
			return err
		}
		return nil
	})
}

// FetchSessionList of social account
func (r *Repository) FetchSessionList(ctx context.Context, socialAccountID []uint64) ([]*model.AccountSocialSession, error) {
	var (
		list  []*model.AccountSocialSession
		query = r.Slave(ctx).Model((*model.AccountSocialSession)(nil))
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
