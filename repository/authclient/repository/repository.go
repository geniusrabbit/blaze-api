// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/models"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
)

// Repository DAO which provides functionality of working with RBAC roles
type Repository struct {
	repository.Repository
}

// NewAuthclientRepository creates a new instance of the AuthClient repository
func NewAuthclientRepository() *Repository {
	return &Repository{}
}

// Get returns RBAC role model by ID
func (r *Repository) Get(ctx context.Context, id string) (*models.AuthClient, error) {
	object := new(models.AuthClient)
	if err := r.Slave(ctx).Find(object, id).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList returns list of RBAC roles by filter
func (r *Repository) FetchList(ctx context.Context, opts ...authclient.QOption) (list []*models.AuthClient, err error) {
	query := r.Slave(ctx).Model((*models.AuthClient)(nil))
	query = authclient.ListOptions(opts).PrepareQuery(query)
	err = query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return list, err
}

// Count returns count of records by filter
func (r *Repository) Count(ctx context.Context, opts ...authclient.QOption) (count int64, err error) {
	query := r.Slave(ctx).Model((*models.AuthClient)(nil))
	query = authclient.ListOptions(opts).PrepareQuery(query)
	err = query.Count(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return count, err
}

// Create new object into database
func (r *Repository) Create(ctx context.Context, roleObj *models.AuthClient, message string) (string, error) {
	if roleObj.ID == "" {
		roleObj.ID = newID()
	}
	roleObj.CreatedAt = time.Now()
	roleObj.UpdatedAt = roleObj.CreatedAt
	err := r.Master(
		historylog.WithMessage(
			historylog.WithPK(ctx, roleObj.ID),
			message,
		),
	).Create(roleObj).Error
	return roleObj.ID, err
}

// Update existing object in database
func (r *Repository) Update(ctx context.Context, id string, roleObj *models.AuthClient, message string) error {
	obj := *roleObj
	obj.ID = id
	return r.Master(
		historylog.WithMessage(
			historylog.WithPK(ctx, obj.ID),
			message,
		),
	).Updates(&obj).Error
}

// Delete delites record by ID
func (r *Repository) Delete(ctx context.Context, id, message string) error {
	return r.Master(
		historylog.WithMessage(
			historylog.WithPK(ctx, id),
			message,
		),
	).Model((*models.AuthClient)(nil)).Delete(`id=?`, id).Error
}
