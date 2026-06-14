package directaccesstoken

import (
	"context"
	"time"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go

// Usecase defines the business logic operations for direct access tokens.
type Usecase interface {
	// Get retrieves a direct access token by its ID.
	Get(ctx context.Context, id uint64) (*DirectAccessToken, error)

	// FetchList retrieves a list of direct access tokens with optional filtering, ordering, and pagination.
	FetchList(ctx context.Context, opts ...QOption) ([]*DirectAccessToken, error)

	// Count returns the total number of direct access tokens matching the query options.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Generate creates a new direct access token for the specified user and account.
	Generate(ctx context.Context, userID, accountID uint64, description string, expiresAt time.Time) (*DirectAccessToken, error)

	// Revoke deactivates direct access tokens matching the query options.
	Revoke(ctx context.Context, opts ...QOption) error
}
