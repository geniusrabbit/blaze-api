// Package repository implements methods of working with the repository objects
package repository

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
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
func (r *Repository) Count(ctx context.Context, filter *historylog.Filter) (cnt int64, err error) {
	query := r.Slave(ctx).Model((*model.HistoryAction)(nil))
	query = filter.Query(query)
	err = query.Count(&cnt).Error
	return cnt, err
}

// FetchList returns list of history actions log by filter
func (r *Repository) FetchList(ctx context.Context, filter *historylog.Filter, order *historylog.Order, pagination *repository.Pagination) ([]*model.HistoryAction, error) {
	var (
		list  []*model.HistoryAction
		query = r.Slave(ctx).Model((*model.HistoryAction)(nil))
	)
	query = filter.Query(query)
	query = order.Query(query)
	query = pagination.PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}
