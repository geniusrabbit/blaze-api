package generated

import (
	"context"
	"errors"
	"time"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
	"gorm.io/gorm"
)

type Repository[T any, TID any] struct {
	repository.Repository
	idField string
}

// NewRepository creates a new repository instance
func NewRepository[T any, TID any]() *Repository[T, TID] {
	return &Repository[T, TID]{idField: getModelIDField(new(T))}
}

// Get returns a campaign by ID
func (r *Repository[T, TID]) Get(ctx context.Context, id TID, qops ...Option) (*T, error) {
	obj := new(T)
	query := r.Slave(ctx).Model(obj)
	query = Options(qops).PrepareQuery(query)
	err := query.First(obj, r.idField+`=?`, id).Error
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// FetchList returns a list of campaigns
func (r *Repository[T, TID]) FetchList(ctx context.Context, qops ...Option) (list []*T, err error) {
	query := r.Slave(ctx).Model((*T)(nil))
	query = Options(qops).PrepareQuery(query)
	err = query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

// Count returns the number of campaigns
func (r *Repository[T, TID]) Count(ctx context.Context, qops ...Option) (count int64, err error) {
	query := r.Slave(ctx).Model((*T)(nil))
	err = Options(qops).PrepareQuery(query).
		Count(&count).Error
	return count, err
}

// Create creates a new campaign
func (r *Repository[T, TID]) Create(ctx context.Context, obj *T, opts ...Option) (TID, error) {
	setModelCreatedAt(obj, time.Now())
	setModelApproveStatus(obj, pkgModels.ApproveStatus(pkgModels.PendingApproveStatus))
	db := Options(opts).PrepareQuery(r.Master(ctx))
	err := db.Create(obj).Error
	return getModelID[TID](obj), err
}

// Update updates an existing campaign
func (r *Repository[T, TID]) Update(ctx context.Context, id TID, obj *T, opts ...Option) error {
	newObj := *obj
	setModelID(&newObj, id)
	setModelUpdatedAt(&newObj, time.Now())
	db := Options(opts).PrepareQuery(r.Master(ctx))
	if err := db.Updates(&newObj).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes a campaign by ID
func (r *Repository[T, TID]) Delete(ctx context.Context, id TID, opts ...Option) error {
	obj := new(T)
	return Options(opts).PrepareQuery(r.Master(ctx)).Delete(obj, r.idField+`=?`, id).Error
}
