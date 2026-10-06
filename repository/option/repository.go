// Package option presents full API functionality of the specific object.
package option

import (
	"context"
)

type TargetPair struct {
	TargetID   uint64
	OptionType OptionType
}

// Repository provides access to options with CRUD operations.
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	// Get retrieves a single option by name, type, and target ID.
	Get(ctx context.Context, name string, otype OptionType, targetID uint64) (*Option, error)

	// GetOneOfType retrieves a single option by name, target ID, and one of the given types.
	GetOneOfType(ctx context.Context, name string, targetPairs []TargetPair) (*Option, error)

	// FetchList retrieves a list of options with filtering, ordering, and pagination.
	FetchList(ctx context.Context, opts ...QOption) ([]*Option, error)

	// Count returns the total number of options matching the filter.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Set creates or updates an option.
	Set(ctx context.Context, opt *Option) error

	// Delete removes an option by name, type, and target ID.
	Delete(ctx context.Context, name string, otype OptionType, targetID uint64) error
}
