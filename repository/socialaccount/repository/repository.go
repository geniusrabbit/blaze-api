package repository

import (
	"context"
	"errors"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"gorm.io/gorm"
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
	res := r.Slave(ctx).Model(object).Where(`id=?`, id).Find(object)
	if err := res.Error; err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList of social accounts
func (r *Repository) FetchList(ctx context.Context, filter *socialaccount.Filter, order *socialaccount.Order, pagination *repository.Pagination) ([]*model.AccountSocial, error) {
	var (
		list  []*model.AccountSocial
		query = r.Slave(ctx).Model((*model.Role)(nil))
	)
	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
	query = pagination.PrepareQuery(query)
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
		query = r.Slave(ctx).Model((*model.Role)(nil))
	)
	query = filter.PrepareQuery(query)
	err := query.Count(&count).Error
	return count, err
}
