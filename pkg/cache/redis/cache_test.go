package redis

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
)

func newTestRedis(mr *miniredis.Miniredis) redis.Cmdable {
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func TestRedisCache(t *testing.T) {
	var (
		key = "test"
		msg = struct{ s string }{s: `test`}
		trg any
		ctx = context.Background()
	)
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()
	marker := New(newTestRedis(mr), 0)
	assert.NoError(t, marker.TrySet(ctx, key, msg, time.Minute))
	assert.Error(t, marker.TrySet(ctx, key, msg, time.Minute))
	assert.NoError(t, marker.Del(ctx, key))
	assert.NoError(t, marker.TrySet(ctx, key, msg, time.Minute))
	assert.NoError(t, marker.Set(ctx, key, msg, time.Minute))
	assert.NoError(t, marker.Get(ctx, key, &trg))
	assert.EqualError(t, marker.Get(ctx, "undefined", &trg), cache.ErrEntryNotFound.Error())
}

func TestRedisCacheMaxLifetime(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()
	marker := New(newTestRedis(mr), time.Minute)
	ctx := context.Background()

	// Before the fix a lifetime of 0 bypassed the cap: the entry never expired
	assert.NoError(t, marker.Set(ctx, "no-ttl", "v", 0))
	assert.Equal(t, time.Minute, mr.TTL("no-ttl"))
	assert.NoError(t, marker.Set(ctx, "long", "v", time.Hour))
	assert.Equal(t, time.Minute, mr.TTL("long"))
	assert.NoError(t, marker.Set(ctx, "short", "v", time.Second))
	assert.Equal(t, time.Second, mr.TTL("short"))
}

func TestRedisCacheByURL(t *testing.T) {
	_, err := NewByURL(`redis://:password@localhost:6379/1`)
	assert.NoError(t, err)
	_, err = NewByURL(`x#:///%%@`)
	assert.Error(t, err)
}
