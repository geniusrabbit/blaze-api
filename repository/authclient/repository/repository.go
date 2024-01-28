// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
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
func (r *Repository) Get(ctx context.Context, id string) (*model.AuthClient, error) {
	object := new(model.AuthClient)
	if err := r.Slave(ctx).Find(object, id).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList returns list of RBAC roles by filter
func (r *Repository) FetchList(ctx context.Context, filter *authclient.Filter) ([]*model.AuthClient, error) {
	var (
		list  []*model.AuthClient
		query = r.Slave(ctx).Model((*model.AuthClient)(nil))
	)
	if filter != nil && len(filter.ID) > 0 {
		query = query.Where(`id IN (?)`, filter.ID)
	}
	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize).Offset(filter.PageSize * filter.Page)
	}
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return list, err
}

// Count returns count of records by filter
func (r *Repository) Count(ctx context.Context, filter *authclient.Filter) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*model.AuthClient)(nil))
	)
	if filter != nil && len(filter.ID) > 0 {
		query = query.Where(`id IN (?)`, filter.ID)
	}
	err := query.Count(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return count, err
}

// Create new object into database
func (r *Repository) Create(ctx context.Context, roleObj *model.AuthClient) (string, error) {
	if roleObj.ID == "" {
		roleObj.ID = newID()
	}
	roleObj.CreatedAt = time.Now()
	roleObj.UpdatedAt = roleObj.CreatedAt
	err := r.Master(ctx).Create(roleObj).Error
	return roleObj.ID, err
}

// Update existing object in database
func (r *Repository) Update(ctx context.Context, id string, roleObj *model.AuthClient) error {
	obj := *roleObj
	obj.ID = id
	return r.Master(ctx).Updates(&obj).Error
}

// Delete delites record by ID
func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.Master(ctx).Model((*model.AuthClient)(nil)).Delete(`id=?`, id).Error
}
