package extra

import (
	"context"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
)

// AccountFilter is a filter for the account ID
type AccountFilter struct {
	AccountID uint64
}

// CurrentAccountFilter returns the current account filter
// Example:
//
//	CurrentAccountFilter(ctx).PrepareQuery(query)
//	query.Where("account_id = ?", session.AccountID(ctx))
func CurrentAccountFilter(ctx context.Context) *AccountFilter {
	return &AccountFilter{AccountID: session.AccountID(ctx)}
}

// PrepareQuery prepares the query for the AccountFilter
func (fl *AccountFilter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	return query.Where("account_id = ?", fl.AccountID)
}
