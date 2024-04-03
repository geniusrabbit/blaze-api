// Package usecase account implementation
package usecase

import (
	"context"

	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/option"
)

// Usecase provides bussiness logic for account access
type Usecase struct {
	baseRepo option.Repository
}

// NewUsecase object controller
func NewUsecase(repo option.Repository) *Usecase {
	return &Usecase{
		baseRepo: repo,
	}
}

// Get returns the group by ID if have access
func (a *Usecase) Get(ctx context.Context, name string, otype model.OptionType, targetID uint64) (*model.Option, error) {
	switch {
	case otype == model.UserOptionType && targetID == 0:
		targetID = session.User(ctx).ID
	case otype == model.AccountOptionType && targetID == 0:
		targetID = session.Account(ctx).ID
	}
	targetObj, err := a.baseRepo.Get(ctx, name, otype, targetID)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, targetObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view")
	}
	return targetObj, nil
}

// FetchList of accounts by filter
func (a *Usecase) FetchList(ctx context.Context, filter *option.Filter, order *option.ListOrder, pagination *repository.Pagination) ([]*model.Option, error) {
	if !acl.HaveAccessList(ctx, &model.Option{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list")
	}
	list, err := a.baseRepo.FetchList(ctx, filter, order, pagination)
	for _, obj := range list {
		if !acl.HaveAccessList(ctx, obj) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list")
		}
	}
	return list, err
}

// Count of accounts by filter
func (a *Usecase) Count(ctx context.Context, filter *option.Filter) (int64, error) {
	if !acl.HaveAccessList(ctx, &model.Option{}) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list")
	}
	return a.baseRepo.Count(ctx, filter)
}

// Create new object in database
func (a *Usecase) Set(ctx context.Context, targetObj *model.Option) error {
	var err error
	switch {
	case targetObj.Type == model.UserOptionType && targetObj.TargetID == 0:
		targetObj.TargetID = session.User(ctx).ID
	case targetObj.Type == model.AccountOptionType && targetObj.TargetID == 0:
		targetObj.TargetID = session.Account(ctx).ID
	}
	if !acl.HaveAccessCreate(ctx, targetObj) {
		return errors.Wrap(acl.ErrNoPermissions, "create")
	}
	err = a.baseRepo.Set(ctx, targetObj)
	return err
}

// Delete delites record by ID
func (a *Usecase) Delete(ctx context.Context, name string, otype model.OptionType, targetID uint64) error {
	targetObj, err := a.Get(ctx, name, otype, targetID)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, targetObj) {
		return errors.Wrap(acl.ErrNoPermissions, "delete")
	}
	return a.baseRepo.Delete(ctx, targetObj.Name, targetObj.Type, targetObj.TargetID)
}
