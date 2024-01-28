// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/option"
)

// Repository DAO which provides functionality of working with RBAC song-tabulatures
type Repository struct {
	repository.Repository
}

// New role repository
func New() *Repository {
	return &Repository{}
}

// Get returns option by ID
func (r *Repository) Get(ctx context.Context, name string, otype model.OptionType, targetID uint64) (*model.Option, error) {
	object := &model.Option{Name: name, Type: otype, TargetID: targetID}
	res := r.Slave(ctx).Model(object).Where(`name=? AND type=? AND target_id=?`, name, otype, targetID).Find(object)
	if err := res.Error; err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList returns list of
func (r *Repository) FetchList(ctx context.Context, filter *option.Filter, order *option.ListOrder, pagination *repository.Pagination) ([]*model.Option, error) {
	var (
		list  []*model.Option
		query = r.Slave(ctx).Model((*model.Option)(nil))
	)
	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
	query = pagination.PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return list, err
}

// Count returns count of records by filter
func (r *Repository) Count(ctx context.Context, filter *option.Filter) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*model.Option)(nil))
	)
	query = filter.PrepareQuery(query)
	err := query.Count(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return count, err
}

// Set new or update object in database
func (r *Repository) Set(ctx context.Context, obj *model.Option) error {
	if obj.CreatedAt.IsZero() {
		obj.CreatedAt = time.Now()
		obj.UpdatedAt = obj.CreatedAt
	}
	return r.Master(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}, {Name: "target_id"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at", "deleted_at"}),
	}).Create(obj).Error
}

// Delete delites record by ID
func (r *Repository) Delete(ctx context.Context, name string, otype model.OptionType, targetID uint64) error {
	return r.Master(ctx).Model((*model.Option)(nil)).
		Delete(`type=? AND target_id=? AND name=?`, otype, targetID, name).Error
}
