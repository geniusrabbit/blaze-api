// Package usecase user managing
package usecase

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// UserUsecase provides business logic for user access.
type UserUsecase[T user.Model] struct {
	userRepo user.Repository[T]
}

// NewUsecase creates generic core user usecase.
func NewUsecase[T user.Model](repo user.Repository[T]) user.Usecase[T] {
	return &UserUsecase[T]{userRepo: repo}
}

func (a *UserUsecase[T]) EmptyObject() T {
	return a.userRepo.EmptyObject()
}

func (a *UserUsecase[T]) Get(ctx context.Context, id uint64) (T, error) {
	var zero T
	currentUser := sessionUserModel(ctx)
	if currentUser != nil && currentUser.GetID() == id {
		if typed, ok := any(currentUser).(T); ok {
			if !acl.HaveAccessView(ctx, typed) {
				return zero, acl.ErrNoPermissions
			}
			return typed, nil
		}
	}
	userObj, err := a.userRepo.Get(ctx, id)
	if err != nil {
		return zero, err
	}
	if !acl.HaveAccessView(ctx, userObj) {
		return zero, acl.ErrNoPermissions
	}
	return userObj, nil
}

func (a *UserUsecase[T]) FetchList(ctx context.Context, opts ...user.QOption) ([]T, error) {
	probe := a.EmptyObject()
	if !acl.HaveAccessList(ctx, probe) {
		current := sessionUserModel(ctx)
		if current == nil || !acl.HaveAccessList(ctx, current) {
			return nil, acl.ErrNoPermissions
		}
		if err := adjustFetchListPermissions(ctx, opts...); err != nil {
			return nil, err
		}
	}
	return a.userRepo.FetchList(ctx, opts...)
}

func (a *UserUsecase[T]) Count(ctx context.Context, opts ...user.QOption) (int64, error) {
	probe := a.EmptyObject()
	if !acl.HaveAccessCount(ctx, probe) {
		current := sessionUserModel(ctx)
		if current == nil || !acl.HaveAccessCount(ctx, current) {
			return 0, acl.ErrNoPermissions
		}
		if err := adjustFetchListPermissions(ctx, opts...); err != nil {
			return 0, err
		}
	}
	return a.userRepo.Count(ctx, opts...)
}

func (a *UserUsecase[T]) Create(ctx context.Context, userObj T) (uint64, error) {
	if !acl.HaveAccessCreate(ctx, userObj) {
		return 0, acl.ErrNoPermissions
	}
	return a.userRepo.Create(ctx, userObj)
}

func (a *UserUsecase[T]) Update(ctx context.Context, userObj T) error {
	if !acl.HaveAccessUpdate(ctx, userObj) {
		return acl.ErrNoPermissions
	}
	return a.userRepo.Update(historylog.WithPK(ctx, userObj.GetID()), userObj)
}

func (a *UserUsecase[T]) Delete(ctx context.Context, id uint64) error {
	userObj, err := a.userRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, userObj) {
		return acl.ErrNoPermissions
	}
	return a.userRepo.Delete(historylog.WithPK(ctx, id), id)
}

func adjustFetchListPermissions(ctx context.Context, opts ...user.QOption) error {
	adjusted := false
	for _, opt := range opts {
		if adjuster, ok := opt.(repository.QueryPermissionAdjuster); ok {
			if err := adjuster.AdjustPermissions(ctx); err != nil {
				return err
			}
			adjusted = true
		}
	}
	if !adjusted {
		return acl.ErrNoPermissions
	}
	return nil
}
