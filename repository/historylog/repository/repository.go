// Package repository implements methods of working with the repository objects
package repository

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	historylogModels "github.com/geniusrabbit/blaze-api/repository/historylog/models"
)

// Repository DAO which provides functionality of working with changelogs
type Repository struct {
	repository.Repository
}

// New history action repository
func New() *Repository {
	return &Repository{}
}

// Count returns count of history actions log by filter
func (r *Repository) Count(ctx context.Context, opts ...historylog.QOption) (cnt int64, err error) {
	query := r.Slave(ctx).Model((*historylogModels.HistoryAction)(nil))
	query = historylog.ListOptions(opts).PrepareQuery(query)
	err = query.Count(&cnt).Error
	return cnt, err
}

// FetchList returns list of history actions log by filter
func (r *Repository) FetchList(ctx context.Context, opts ...historylog.QOption) ([]*historylogModels.HistoryAction, error) {
	var (
		list  []*historylogModels.HistoryAction
		query = r.Slave(ctx).Model((*historylogModels.HistoryAction)(nil))
	)
	query = historylog.ListOptions(opts).PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}
