// Package usecase account implementation
package usecase

import (
	"context"
	"fmt"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	historylogModels "github.com/geniusrabbit/blaze-api/repository/historylog/models"
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
func (a *HistoryUsecase) Count(ctx context.Context, opts ...historylog.QOption) (int64, error) {
	if !acl.HaveAccessList(ctx, &historylogModels.HistoryAction{}) {
		return 0, fmt.Errorf("%s: %w", "list log items", acl.ErrNoPermissions)
	}
	return a.repo.Count(ctx, opts...)
}

// FetchList of roles by filter
func (a *HistoryUsecase) FetchList(ctx context.Context, opts ...historylog.QOption) ([]*historylogModels.HistoryAction, error) {
	if !acl.HaveAccessList(ctx, &historylogModels.HistoryAction{}) {
		return nil, fmt.Errorf("%s: %w", "list log items", acl.ErrNoPermissions)
	}
	list, err := a.repo.FetchList(ctx, opts...)
	for _, link := range list {
		if !acl.HaveAccessList(ctx, link) {
			return nil, fmt.Errorf("%s: %w", "list log items", acl.ErrNoPermissions)
		}
	}
	return list, err
}
