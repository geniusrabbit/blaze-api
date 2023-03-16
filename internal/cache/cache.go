package cache

import (
	"context"
	"errors"
	"time"
)

// Errors list
var (
	ErrEntryNotFound = errors.New("[cache] entry is not found")
)

// Client data accessor
type Client interface {
	Set(ctx context.Context, key string, value any, lifetime time.Duration) error
	TrySet(ctx context.Context, key string, value any, lifetime time.Duration) error
	Get(ctx context.Context, key string, target any) error
	Del(ctx context.Context, key string) error
}
