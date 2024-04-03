package redis

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/elliotchance/redismock/v8"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
)

// newTestRedis returns a redis.Cmdable.
func newTestRedis(mr *miniredis.Miniredis) *redismock.ClientMock {
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return redismock.NewNiceMock(client)
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
	marker := New(newTestRedis(mr))
	assert.NoError(t, marker.TrySet(ctx, key, msg, time.Minute))
	assert.Error(t, marker.TrySet(ctx, key, msg, time.Minute))
	assert.NoError(t, marker.Del(ctx, key))
	assert.NoError(t, marker.TrySet(ctx, key, msg, time.Minute))
	assert.NoError(t, marker.Set(ctx, key, msg, time.Minute))
	assert.NoError(t, marker.Get(ctx, key, &trg))
	assert.EqualError(t, marker.Get(ctx, "undefined", &trg), cache.ErrEntryNotFound.Error())
}

func TestRedisCacheByURL(t *testing.T) {
	_, err := NewByURL(`redis://:password@localhost:6379/1`)
	assert.NoError(t, err)
	_, err = NewByURL(`x#:///%%@`)
	assert.Error(t, err)
}
