// Package usecase provides business logic for option management
package usecase

import (
	"context"

	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/repository/option/models"
)

// Usecase implements business logic for option access and management
type Usecase struct {
	baseRepo option.Repository
}

// NewUsecase creates and returns a new Usecase instance
func NewUsecase(repo option.Repository) *Usecase {
	return &Usecase{
		baseRepo: repo,
	}
}

// Get retrieves an option by name, type, and target ID with permission checks
func (a *Usecase) Get(ctx context.Context, name string, otype models.OptionType, targetID uint64) (*models.Option, error) {
	switch {
	case otype == models.UserOptionType && targetID == 0:
		targetID = session.User(ctx).GetID()
	case otype == models.AccountOptionType && targetID == 0:
		targetID = session.AccountID(ctx)
	}
	targetObj, err := a.baseRepo.Get(ctx, name, otype, targetID)
	if err != nil {
		return nil, err
	}
	if !acl.HaveObjectPermissions(ctx, targetObj, acl.PermGet+`.*`) {
		return nil, acl.ErrNoPermissions.WithMessage("get")
	}
	return targetObj, nil
}

// FetchList retrieves a list of options filtered and ordered with permission checks
func (a *Usecase) FetchList(ctx context.Context, opts ...option.QOption) ([]*models.Option, error) {
	if !acl.HaveAccessList(ctx, &models.Option{}) {
		return nil, acl.ErrNoPermissions.WithMessage("list")
	}
	list, err := a.baseRepo.FetchList(ctx, opts...)
	for _, obj := range list {
		if !acl.HaveAccessList(ctx, obj) {
			return nil, acl.ErrNoPermissions.WithMessage("list")
		}
	}
	return list, err
}

// Count returns the total count of options matching the filter with permission checks
func (a *Usecase) Count(ctx context.Context, opts ...option.QOption) (int64, error) {
	if !acl.HaveAccessList(ctx, &models.Option{}) {
		return 0, acl.ErrNoPermissions.WithMessage("list")
	}
	return a.baseRepo.Count(ctx, opts...)
}

// Set creates or updates an option with permission checks
func (a *Usecase) Set(ctx context.Context, targetObj *models.Option) error {
	switch {
	case targetObj.Type == models.UserOptionType && targetObj.TargetID == 0:
		targetObj.TargetID = session.User(ctx).GetID()
	case targetObj.Type == models.AccountOptionType && targetObj.TargetID == 0:
		targetObj.TargetID = session.AccountID(ctx)
	}
	if !acl.HaveObjectPermissions(ctx, targetObj, acl.PermSet+`.*`) {
		return acl.ErrNoPermissions.WithMessage("set")
	}
	return a.baseRepo.Set(ctx, targetObj)
}

// SetOption sets an option value by name, type, and target ID
func (a *Usecase) SetOption(ctx context.Context, name string, otype models.OptionType, targetID uint64, value any) error {
	obj := &models.Option{
		Type:     otype,
		TargetID: targetID,
		Name:     name,
	}
	if err := obj.Value.SetValue(value); err != nil {
		return errors.Wrap(err, "set option value")
	}
	return a.baseRepo.Set(ctx, obj)
}

// Delete removes an option with permission checks
func (a *Usecase) Delete(ctx context.Context, name string, otype models.OptionType, targetID uint64) error {
	targetObj, err := a.Get(ctx, name, otype, targetID)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, targetObj) {
		return acl.ErrNoPermissions.WithMessage("delete")
	}
	return a.baseRepo.Delete(
		historylog.WithPK(ctx, targetObj.Name),
		targetObj.Name,
		targetObj.Type,
		targetObj.TargetID,
	)
}
