package generated

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
)

type Repository[T Model[TID], TID comparable] struct {
	repository.Repository
	idField string
}

// NewRepository creates a new repository instance
func NewRepository[T Model[TID], TID comparable]() *Repository[T, TID] {
	return &Repository[T, TID]{idField: getModelIDField(new(T))}
}

// Get returns a campaign by ID
func (r *Repository[T, TID]) Get(ctx context.Context, id TID, qops ...Option) (*T, error) {
	obj := new(T)
	query := r.Slave(ctx).Model(obj)
	query = Options(qops).PrepareQuery(query)
	err := query.First(obj, r.idField+`=?`, id).Error
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// FetchList returns a list of campaigns
func (r *Repository[T, TID]) FetchList(ctx context.Context, qops ...Option) (list []*T, err error) {
	query := r.Slave(ctx).Model((*T)(nil))
	query = Options(qops).PrepareQuery(query)
	err = query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

// Count returns the number of campaigns
func (r *Repository[T, TID]) Count(ctx context.Context, qops ...Option) (count int64, err error) {
	query := r.Slave(ctx).Model((*T)(nil))
	err = Options(qops).PrepareQuery(query).
		Count(&count).Error
	return count, err
}

// Create creates a new campaign
func (r *Repository[T, TID]) Create(ctx context.Context, obj *T, opts ...Option) (TID, error) {
	setModelCreatedAt(obj, time.Now())
	db := Options(opts).PrepareQuery(r.Master(ctx))
	err := db.Create(obj).Error
	return getModelID[TID](obj), err
}

// Update writes the non-zero fields of obj to the object with the ID.
// Zero values (0, "", false, nil) are skipped; use UpdateFields to set them.
// Returns repository.ErrNotFound when the object doesn't exist.
func (r *Repository[T, TID]) Update(ctx context.Context, id TID, obj *T, opts ...Option) error {
	return r.update(ctx, id, obj, nil, opts...)
}

// UpdateFields writes the listed fields (column or struct field names) of obj
// to the object with the ID, zero values included. updated_at is always set.
// Returns repository.ErrNotFound when the object doesn't exist.
func (r *Repository[T, TID]) UpdateFields(ctx context.Context, id TID, obj *T, fields []string, opts ...Option) error {
	if len(fields) == 0 {
		return errNoFields
	}
	return r.update(ctx, id, obj, append(fields, "updated_at"), opts...)
}

func (r *Repository[T, TID]) update(ctx context.Context, id TID, obj *T, fields []string, opts ...Option) error {
	newObj := *obj
	setModelID(&newObj, id)
	setModelUpdatedAt(&newObj, time.Now())
	ctx = historylog.WithPK(ctx, id)
	db := Options(opts).PrepareQuery(r.Master(ctx))
	if len(fields) > 0 {
		db = db.Select(fields)
	}
	res := db.Updates(&newObj)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// MySQL reports changed rows only: tell "nothing changed" from "missing"
		var count int64
		if err := r.Master(ctx).Model((*T)(nil)).Where(r.idField+`=?`, id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return repository.ErrNotFound
		}
	}
	return nil
}

// Delete deletes the object with the ID.
// Returns repository.ErrNotFound when the object doesn't exist.
func (r *Repository[T, TID]) Delete(ctx context.Context, id TID, opts ...Option) error {
	ctx = historylog.WithPK(ctx, id)
	res := Options(opts).PrepareQuery(r.Master(ctx)).Delete(new(T), r.idField+`=?`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

var errNoFields = errors.New("generated: UpdateFields needs at least one field")
