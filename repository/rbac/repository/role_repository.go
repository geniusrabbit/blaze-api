// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
)

// Repository DAO which provides functionality of working with RBAC roles
type Repository struct {
	repository.Repository
}

// New role repository
func New() *Repository {
	return &Repository{}
}

// Get returns RBAC role model by ID
func (r *Repository) Get(ctx context.Context, id uint64) (*model.Role, error) {
	object := new(model.Role)
	if err := r.Slave(ctx).Find(object, id).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// GetByName returns RBAC role model by title
func (r *Repository) GetByName(ctx context.Context, title string) (*model.Role, error) {
	object := new(model.Role)
	if err := r.Slave(ctx).Find(object, `name=?`, title).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList returns list of RBAC roles by filter
func (r *Repository) FetchList(ctx context.Context, filter *rbac.Filter, order *rbac.Order, pagination *repository.Pagination) ([]*model.Role, error) {
	var (
		list  []*model.Role
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

// Count returns count of RBAC roles by filter
func (r *Repository) Count(ctx context.Context, filter *rbac.Filter) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*model.Role)(nil))
	)
	query = filter.PrepareQuery(query)
	err := query.Count(&count).Error
	return count, err
}

// Create new object into database
func (r *Repository) Create(ctx context.Context, roleObj *model.Role) (uint64, error) {
	roleObj.CreatedAt = time.Now()
	roleObj.UpdatedAt = roleObj.CreatedAt
	err := r.Master(ctx).Create(roleObj).Error
	return roleObj.ID, err
}

// Update existing object in database
func (r *Repository) Update(ctx context.Context, id uint64, roleObj *model.Role) error {
	obj := *roleObj
	obj.ID = id
	return r.Master(ctx).Updates(&obj).Error
}

// Delete delites record by ID
func (r *Repository) Delete(ctx context.Context, id uint64) error {
	return r.Master(ctx).Model((*model.Role)(nil)).Delete(`id=?`, id).Error
}
