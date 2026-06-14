package socialaccount

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
)

// Repository defines the interface for managing social account data operations.
// It provides methods for retrieving, listing, counting, and managing social accounts
// and their associated sessions.
type Repository interface {
	// Get retrieves a social account by its unique identifier.
	// Returns the social account if found, or an error if the operation fails.
	Get(ctx context.Context, id uint64) (*models.AccountSocial, error)

	// FetchList retrieves a list of social accounts based on the provided query options.
	// Returns a slice of social accounts or an error if the operation fails.
	FetchList(ctx context.Context, opts ...QOption) ([]*models.AccountSocial, error)

	// Count returns the total number of social accounts matching the query options.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Disconnect removes the association of a social account by its unique identifier.
	// Returns an error if the operation fails.
	Disconnect(ctx context.Context, id uint64) error

	// FetchSessionList retrieves all sessions for the given social account IDs.
	// Returns a slice of sessions or an error if the operation fails.
	FetchSessionList(ctx context.Context, socialAccountID []uint64) ([]*models.AccountSocialSession, error)
}
