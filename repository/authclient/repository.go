// Package authclient provides repository access for authentication client management.
package authclient

import (
	"context"
)

// Repository defines the interface for AuthClient data access operations.
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	// Get retrieves an AuthClient by ID.
	Get(ctx context.Context, id string) (*AuthClient, error)

	// FetchList retrieves a list of AuthClients with optional query parameters.
	FetchList(ctx context.Context, opts ...QOption) ([]*AuthClient, error)

	// Count returns the total number of AuthClients matching the query options.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Create adds a new AuthClient and returns its ID.
	Create(ctx context.Context, authClient *AuthClient, message string) (string, error)

	// Update modifies an existing AuthClient by ID.
	Update(ctx context.Context, id string, authClient *AuthClient, message string) error

	// Delete removes an AuthClient by ID.
	Delete(ctx context.Context, id, message string) error
}
