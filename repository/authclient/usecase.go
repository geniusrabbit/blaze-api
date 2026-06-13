package authclient

import (
	"context"
)

// Usecase defines the business logic operations for AuthClient management.
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	// Get retrieves a single AuthClient by ID.
	Get(ctx context.Context, id string) (*AuthClient, error)

	// FetchList retrieves multiple AuthClients with optional query parameters.
	FetchList(ctx context.Context, opts ...QOption) ([]*AuthClient, error)

	// Count returns the total number of AuthClients matching the query options.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Create adds a new AuthClient and records the change message.
	Create(ctx context.Context, authClient *AuthClient, message string) (string, error)

	// Update modifies an existing AuthClient by ID with a change message.
	Update(ctx context.Context, id string, authClient *AuthClient, message string) error

	// Delete removes an AuthClient by ID with a change message.
	Delete(ctx context.Context, id, message string) error
}
