// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/repository/option/models"
)

// Repository DAO which provides functionality of working with RBAC song-tabulatures
type Repository struct {
	repository.Repository
	defaultSystemOptions map[string]any
}

// NewOptionRepository creates a new option repository
func NewOptionRepository(defSysOpts map[string]any) *Repository {
	return &Repository{
		defaultSystemOptions: defSysOpts,
	}
}

// Get returns option by ID
func (r *Repository) Get(ctx context.Context, name string, otype models.OptionType, targetID uint64) (*models.Option, error) {
	object := &option.Option{Name: name, Type: otype, TargetID: targetID}
	res := r.Slave(ctx).Model(object).
		Where(`name=? AND type=? AND target_id=?`, name, otype, targetID).Find(object)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) || errors.Is(res.Error, sql.ErrNoRows) || object.Value.Data == nil {
		if otype == models.SystemOptionType && targetID == 0 && r.defaultSystemOptions != nil {
			object.Name = name
			object.Type = models.SystemOptionType
			object.TargetID = 0

			if value, ok := r.defaultSystemOptions[name]; ok {
				if err := object.Value.SetValue(value); err != nil {
					return nil, err
				}
				return object, nil
			}
		}
	}

	if err := res.Error; err != nil {
		return nil, err
	}
	return object, nil
}

// FetchList returns list of
func (r *Repository) FetchList(ctx context.Context, opts ...option.QOption) ([]*models.Option, error) {
	var (
		list  []*models.Option
		query = r.Slave(ctx).Model((*models.Option)(nil))
	)
	query = option.ListOptions(opts).PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return list, err
}

// Count returns count of records by filter
func (r *Repository) Count(ctx context.Context, opts ...option.QOption) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*models.Option)(nil))
	)
	query = option.ListOptions(opts).PrepareQuery(query)
	err := query.Count(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return count, err
}

// Set new or update object in database
func (r *Repository) Set(ctx context.Context, obj *models.Option) error {
	if obj.CreatedAt.IsZero() {
		obj.CreatedAt = time.Now()
	}
	if obj.UpdatedAt.IsZero() {
		obj.UpdatedAt = obj.CreatedAt
	}
	return r.Master(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}, {Name: "target_id"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at", "deleted_at"}),
	}).Create(obj).Error
}

// Delete delites record by ID
func (r *Repository) Delete(ctx context.Context, name string, otype models.OptionType, targetID uint64) error {
	return r.Master(ctx).Model((*models.Option)(nil)).
		Delete(`type=? AND target_id=? AND name=?`, otype, targetID, name).Error
}
