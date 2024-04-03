package redis

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/go-redis/redis/v8"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
)

var errTrySetValue = errors.New(`try set value failed`)

// Cache containse memory cache storage
type Cache struct {
	client redis.Cmdable
}

// New marker wraps redis client
func New(client redis.Cmdable) *Cache {
	return &Cache{client: client}
}

// NewByURL marker
func NewByURL(host string) (*Cache, error) {
	var (
		urlHost, err = url.Parse(host)
		password     string
	)
	if err != nil {
		return nil, err
	}
	if urlHost.User != nil {
		password, _ = urlHost.User.Password()
	}
	query := urlHost.Query()
	return New(redis.NewClient(&redis.Options{
		DB:           gocast.Int(strings.Trim(urlHost.Path, `/`)),
		Addr:         urlHost.Host,
		Password:     password,
		PoolSize:     gocast.Int(query.Get(`pool`)),
		MaxRetries:   gocast.Int(query.Get(`max_retries`)),
		MinIdleConns: gocast.Int(query.Get(`idle_cons`)),
	})), nil
}

// Set cache item
func (c *Cache) Set(ctx context.Context, key string, value any, lifetime time.Duration) error {
	data, err := json.Marshal(value)
	if err == nil {
		err = c.client.Set(ctx, key, data, lifetime).Err()
	}
	return err
}

// TrySet only if not exists
func (c *Cache) TrySet(ctx context.Context, key string, value any, lifetime time.Duration) error {
	res := false
	data, err := json.Marshal(value)
	if err == nil {
		res, err = c.client.SetNX(ctx, key, data, lifetime).Result()
		if err == nil && !res {
			err = errTrySetValue
		}
	}
	return err
}

// Get cached item
func (c *Cache) Get(ctx context.Context, key string, target any) error {
	data, err := c.client.Get(ctx, key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(data), target)
	} else if err == redis.Nil {
		err = cache.ErrEntryNotFound
	}
	return err
}

// Del removes cache item by key
func (c *Cache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
