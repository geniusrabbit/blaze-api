package socialaccount

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
)

// Usecase defines the business logic operations for social accounts.
type Usecase interface {
	// Get retrieves a single social account by ID.
	Get(ctx context.Context, id uint64) (*models.AccountSocial, error)

	// FetchList retrieves a list of social accounts with optional query parameters.
	FetchList(ctx context.Context, opts ...QOption) ([]*models.AccountSocial, error)

	// Count returns the total number of social accounts matching the query options.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Disconnect removes a social account connection by ID.
	Disconnect(ctx context.Context, id uint64) (*models.AccountSocial, error)

	// FetchSessionList retrieves sessions for the given social account IDs.
	FetchSessionList(ctx context.Context, socialAccountID []uint64) ([]*models.AccountSocialSession, error)
}
