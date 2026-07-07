package directaccesstoken

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Order is an alias for models.Order
type Order = models.Order

// Filter defines query filters for direct access tokens
type Filter struct {
	ID           []uint64  // Filter by token IDs
	Token        []string  // Filter by token strings
	UserID       []uint64  // Filter by user IDs
	AccountID    []uint64  // Filter by account IDs
	MinExpiresAt time.Time // Minimum expiration time
	MaxExpiresAt time.Time // Maximum expiration time
}

// PrepareQuery applies filter conditions to a GORM query
func (fl *Filter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	if len(fl.ID) > 0 {
		query = query.Where(`id IN (?)`, fl.ID)
	}
	if len(fl.Token) > 0 {
		query = query.Where(`token IN (?)`, fl.Token)
	}
	if len(fl.UserID) > 0 {
		query = query.Where(`user_id IN (?)`, fl.UserID)
	}
	if len(fl.AccountID) > 0 {
		query = query.Where(`account_id IN (?)`, fl.AccountID)
	}
	if !fl.MinExpiresAt.IsZero() {
		query = query.Where(`expires_at >= ?`, fl.MinExpiresAt)
	}
	if !fl.MaxExpiresAt.IsZero() {
		query = query.Where(`expires_at <= ?`, fl.MaxExpiresAt)
	}
	return query
}

// ListOrder defines sort order for query results
type ListOrder struct {
	ID        models.Order // Sort by ID
	Token     models.Order // Sort by token
	UserID    models.Order // Sort by user ID
	AccountID models.Order // Sort by account ID
	CreatedAt models.Order // Sort by creation time
	ExpiresAt models.Order // Sort by expiration time
}

// PrepareQuery applies sort order to a GORM query
func (ord *ListOrder) PrepareQuery(query *gorm.DB) *gorm.DB {
	if ord == nil {
		return query
	}
	query = ord.ID.PrepareQuery(query, "id")
	query = ord.Token.PrepareQuery(query, "token")
	query = ord.UserID.PrepareQuery(query, "user_id")
	query = ord.AccountID.PrepareQuery(query, "account_id")
	query = ord.CreatedAt.PrepareQuery(query, "created_at")
	query = ord.ExpiresAt.PrepareQuery(query, "expires_at")
	return query
}

// AdjustPermissions scopes the filter to the current session account.
// It always overwrites AccountID to ensure results are restricted to the caller's account.
func (fl *Filter) AdjustPermissions(ctx context.Context) error {
	fl.AccountID = []uint64{session.AccountID(ctx)}
	return nil
}

// Pagination is an alias for repository.Pagination
type Pagination = repository.Pagination

// Type aliases for list options.
type (
	QOption     = repository.QOption
	ListOptions = repository.ListOptions
)
