package historylog

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/option/models"
)

// Filter represents query filters for history log objects.
type Filter struct {
	ID          []uuid.UUID // Filter by history log IDs
	RequestID   []string    // Filter by request IDs
	Name        []string    // Filter by names
	UserID      []uint64    // Filter by user IDs
	AccountID   []uint64    // Filter by account IDs
	ObjectID    []uint64    // Filter by numeric object IDs
	ObjectIDStr []string    // Filter by string object IDs
	ObjectType  []string    // Filter by object types
}

// Query applies the filter conditions to a GORM query.
func (filter *Filter) Query(query *gorm.DB) *gorm.DB {
	if filter == nil {
		return query
	}
	if len(filter.ID) > 0 {
		query = query.Where(`id IN (?)`, filter.ID)
	}
	if len(filter.RequestID) > 0 {
		query = query.Where(`request_id IN (?)`, filter.RequestID)
	}
	if len(filter.Name) > 0 {
		query = query.Where(`name IN (?)`, filter.Name)
	}
	if len(filter.UserID) > 0 {
		query = query.Where(`user_id IN (?)`, filter.UserID)
	}
	if len(filter.AccountID) > 0 {
		query = query.Where(`account_id IN (?)`, filter.AccountID)
	}
	if len(filter.ObjectID) > 0 {
		query = query.Where(`object_id IN (?)`, filter.ObjectID)
	}
	if len(filter.ObjectIDStr) > 0 {
		query = query.Where(`object_ids IN (?)`, filter.ObjectIDStr)
	}
	if len(filter.ObjectType) > 0 {
		query = query.Where(`object_type IN (?)`, filter.ObjectType)
	}
	return query
}

// Order defines sorting options for history log queries.
type Order struct {
	ID          models.Order // Sort by ID
	RequestID   models.Order // Sort by request ID
	Name        models.Order // Sort by name
	UserID      models.Order // Sort by user ID
	AccountID   models.Order // Sort by account ID
	ObjectID    models.Order // Sort by numeric object ID
	ObjectIDStr models.Order // Sort by string object ID
	ObjectType  models.Order // Sort by object type
	ActionAt    models.Order // Sort by action timestamp
}

// Query applies the sorting conditions to a GORM query.
func (o *Order) Query(query *gorm.DB) *gorm.DB {
	if o == nil {
		return query
	}
	query = o.ID.PrepareQuery(query, `id`)
	query = o.RequestID.PrepareQuery(query, `request_id`)
	query = o.Name.PrepareQuery(query, `name`)
	query = o.UserID.PrepareQuery(query, `user_id`)
	query = o.AccountID.PrepareQuery(query, `account_id`)
	query = o.ObjectID.PrepareQuery(query, `object_id`)
	query = o.ObjectIDStr.PrepareQuery(query, `object_ids`)
	query = o.ObjectType.PrepareQuery(query, `object_type`)
	query = o.ActionAt.PrepareQuery(query, `action_at`)
	return query
}

// Type aliases for common repository types.
type (
	Pagination  = repository.Pagination
	QOption     = repository.QOption
	ListOptions = repository.ListOptions
)
