package dummy

import (
	"context"
	"time"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
)

// Cache containse memory cache storage
type Cache struct{}

// New memory cache
func New() *Cache {
	return &Cache{}
}

// Set cache item
func (c *Cache) Set(ctx context.Context, key string, value any, _ time.Duration) error {
	return nil
}

// TrySet only if not exists
func (c *Cache) TrySet(ctx context.Context, key string, value any, _ time.Duration) error {
	return nil
}

// Get cached item
func (c *Cache) Get(ctx context.Context, key string, target any) error {
	return cache.ErrEntryNotFound
}

// Del removes cache item by key
func (c *Cache) Del(ctx context.Context, key string) error {
	return nil
}
