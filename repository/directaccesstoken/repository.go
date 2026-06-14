package directaccesstoken

import (
	"context"
	"time"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go

// Repository defines the interface for managing direct access tokens.
type Repository interface {
	// Get retrieves a direct access token by its ID.
	Get(ctx context.Context, id uint64) (*DirectAccessToken, error)

	// GetByToken retrieves a direct access token by its token string.
	GetByToken(ctx context.Context, token string) (*DirectAccessToken, error)

	// FetchList retrieves a paginated list of direct access tokens matching the filter and order criteria.
	FetchList(ctx context.Context, opts ...QOption) ([]*DirectAccessToken, error)

	// Count returns the total count of direct access tokens matching the filter criteria.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Generate creates and stores a new direct access token for the specified user and account.
	Generate(ctx context.Context, userID, accountID uint64, description string, expiresAt time.Time) (*DirectAccessToken, error)

	// Revoke invalidates direct access tokens matching the filter criteria.
	Revoke(ctx context.Context, opts ...QOption) error
}
