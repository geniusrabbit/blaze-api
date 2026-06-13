package option

import (
	"context"
)

// Usecase defines operations for managing options in the system.
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	// Get retrieves a single option by name, type, and target ID.
	Get(ctx context.Context, name string, otype OptionType, targetID uint64) (*Option, error)

	// FetchList retrieves a list of options filtered, ordered, and paginated according to the parameters.
	FetchList(ctx context.Context, opts ...QOption) ([]*Option, error)

	// Count returns the total number of options matching the filter criteria.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Set stores or updates an option.
	Set(ctx context.Context, opt *Option) error

	// SetOption stores or updates an option by name, type, target ID, and value.
	SetOption(ctx context.Context, name string, otype OptionType, targetID uint64, value any) error

	// Delete removes an option by name, type, and target ID.
	Delete(ctx context.Context, name string, otype OptionType, targetID uint64) error
}
