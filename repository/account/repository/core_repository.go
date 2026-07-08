package repository

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	baseRepo "github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
)

type coreAccountRepository[T account.Model] struct {
	baseRepo.Repository
	newModel func() T
}

// NewRepository creates a generic core account repository.
func NewRepository[T account.Model](newModel func() T) account.Repository[T] {
	if newModel == nil {
		panic(`newModel function cannot be nil`)
	}
	return &coreAccountRepository[T]{newModel: newModel}
}

func (r *coreAccountRepository[T]) EmptyObject() T {
	return r.newModel()
}

func (r *coreAccountRepository[T]) Get(ctx context.Context, id uint64) (T, error) {
	var zero T
	object := r.newModel()
	if err := r.Slave(ctx).Find(object, id).Error; err != nil {
		return zero, err
	}
	return object, nil
}

func (r *coreAccountRepository[T]) FetchList(ctx context.Context, opts ...account.QOption) ([]T, error) {
	var (
		list  []T
		query = r.Slave(ctx).Model(r.newModel())
	)
	query = account.ListOptions(opts).PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

func (r *coreAccountRepository[T]) Count(ctx context.Context, opts ...account.QOption) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model(r.newModel())
	)
	query = account.ListOptions(opts).PrepareQuery(query)
	err := query.Count(&count).Error
	return count, err
}

func (r *coreAccountRepository[T]) Create(ctx context.Context, accountObj T) (uint64, error) {
	setAccountApprove(accountObj, pkgModels.UndefinedApproveStatus)
	err := r.Master(ctx).Create(accountObj).Error
	if err != nil {
		return 0, err
	}
	return accountObj.GetID(), nil
}

func (r *coreAccountRepository[T]) Update(ctx context.Context, id uint64, accountObj T) error {
	setAccountID(accountObj, id)
	return r.Master(ctx).Updates(accountObj).Error
}

func (r *coreAccountRepository[T]) Delete(ctx context.Context, id uint64) error {
	return r.Master(ctx).Model(r.newModel()).Delete(`id=?`, id).Error
}

type accountApprove interface {
	SetApprove(pkgModels.ApproveStatus)
}

func setAccountApprove(obj any, status pkgModels.ApproveStatus) {
	if v, ok := obj.(accountApprove); ok {
		v.SetApprove(status)
	}
}

type accountIDSetter interface {
	SetID(uint64)
}

func setAccountID(obj any, id uint64) {
	if v, ok := obj.(accountIDSetter); ok {
		v.SetID(id)
	}
}
