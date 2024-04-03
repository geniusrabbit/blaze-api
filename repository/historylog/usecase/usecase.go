// Package usecase account implementation
package usecase

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/pkg/errors"
)

// RoleUsecase provides bussiness logic for account access
type HistoryUsecase struct {
	repo historylog.Repository
}

// NewUsecase object controller
func NewUsecase(repo historylog.Repository) *HistoryUsecase {
	return &HistoryUsecase{
		repo: repo,
	}
}

// Count of roles by filter
func (a *HistoryUsecase) Count(ctx context.Context, filter *historylog.Filter) (int64, error) {
	if !acl.HaveAccessList(ctx, &model.HistoryAction{}) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list log items")
	}
	return a.repo.Count(ctx, filter)
}

// FetchList of roles by filter
func (a *HistoryUsecase) FetchList(ctx context.Context, filter *historylog.Filter, order *historylog.Order, pagination *repository.Pagination) ([]*model.HistoryAction, error) {
	if !acl.HaveAccessList(ctx, &model.HistoryAction{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list log items")
	}
	list, err := a.repo.FetchList(ctx, filter, order, pagination)
	for _, link := range list {
		if !acl.HaveAccessList(ctx, link) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list log items")
		}
	}
	return list, err
}
