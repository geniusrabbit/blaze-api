package memory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/allegro/bigcache/v3"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
)

var errTrySetValue = errors.New(`try set value failed`)

// Cache containse memory cache storage
type Cache struct {
	big *bigcache.BigCache
}

// New memory cache
func New(ctx context.Context, cfg bigcache.Config) (*Cache, error) {
	big, err := bigcache.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Cache{big: big}, nil
}

// NewTimeout returns cache with timout value
func NewTimeout(ctx context.Context, timeout time.Duration) (*Cache, error) {
	return New(ctx, bigcache.DefaultConfig(timeout))
}

// Set cache item
// NOTE: timeout is not used, it can be defined globaly
func (c *Cache) Set(ctx context.Context, key string, value any, _ time.Duration) error {
	data, err := json.Marshal(value)
	if err == nil {
		err = c.big.Set(key, data)
	}
	return err
}

// TrySet only if not exists
func (c *Cache) TrySet(ctx context.Context, key string, value any, lifetime time.Duration) error {
	_, err := c.big.Get(key)
	if err != bigcache.ErrEntryNotFound {
		return errTrySetValue
	}
	return c.Set(ctx, key, value, lifetime)
}

// Get cached item
func (c *Cache) Get(ctx context.Context, key string, target any) error {
	data, err := c.big.Get(key)
	if err == nil {
		return json.Unmarshal(data, target)
	} else if err == bigcache.ErrEntryNotFound {
		err = cache.ErrEntryNotFound
	}
	return err
}

// Del removes cache item by key
func (c *Cache) Del(ctx context.Context, key string) error {
	return c.big.Delete(key)
}
