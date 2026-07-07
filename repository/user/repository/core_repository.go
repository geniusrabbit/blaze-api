package repository

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	baseRepo "github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var ErrInvalidUserObject = errors.New(`invalid object`)

type coreRepository[T user.Model] struct {
	baseRepo.Repository
	newModel func() T
}

// NewRepository creates a generic core user repository.
// T must be a pointer type (e.g. *models.User) implementing user.Model.
func NewRepository[T user.Model](newModel func() T) user.Repository[T] {
	if newModel == nil {
		panic(`newModel function cannot be nil`)
	}
	return &coreRepository[T]{newModel: newModel}
}

func (r *coreRepository[T]) EmptyObject() T {
	return r.newModel()
}

func (r *coreRepository[T]) Get(ctx context.Context, id uint64) (T, error) {
	object := r.newModel()
	if err := r.Slave(ctx).First(object, id).Error; err != nil {
		var zero T
		return zero, err
	}
	return object, nil
}

func (r *coreRepository[T]) FetchList(ctx context.Context, opts ...user.QOption) ([]T, error) {
	var (
		list  []T
		query = r.Slave(ctx).Model(r.newModel())
	)
	query = user.ListOptions(opts).PrepareQuery(query)
	err := query.Find(&list).Error
	return list, err
}

func (r *coreRepository[T]) Count(ctx context.Context, opts ...user.QOption) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model(r.newModel())
	)
	query = user.ListOptions(opts).PrepareQuery(query)
	err := query.Count(&count).Error
	return count, err
}

func (r *coreRepository[T]) Create(ctx context.Context, userObj T) (uint64, error) {
	setApproveOnModel(userObj, pkgModels.UndefinedApproveStatus)
	err := r.Master(ctx).Create(userObj).Error
	if err != nil {
		return 0, err
	}
	return userObj.GetID(), nil
}

func (r *coreRepository[T]) Update(ctx context.Context, userObj T) error {
	if userObj.GetID() == 0 {
		return ErrInvalidUserObject
	}
	return r.Master(ctx).Select("*").Updates(userObj).Error
}

func (r *coreRepository[T]) Delete(ctx context.Context, id uint64) error {
	res := r.Master(ctx).Delete(r.newModel(), id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
