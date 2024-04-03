package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
)

func Test_MemoryCache(t *testing.T) {
	var (
		target = 1
		key    = "test"
		msg    = struct{ s string }{s: `test`}
		ctx    = context.Background()
	)

	cacheObj, err := NewTimeout(ctx, time.Second)
	assert.NoError(t, err, "create new cache")

	assert.NoError(t, cacheObj.TrySet(ctx, key, msg, time.Minute))
	assert.Error(t, cacheObj.TrySet(ctx, key, msg, time.Minute))

	err = cacheObj.Set(ctx, key, target, 0)
	assert.NoError(t, err, "set new item")

	target = 100
	err = cacheObj.Get(ctx, key, &target)
	assert.NoError(t, err, "get the item")
	assert.Equal(t, 1, target)

	err = cacheObj.Del(ctx, key)
	assert.NoError(t, err, "del the item")

	err = cacheObj.Get(ctx, key, &target)
	assert.EqualError(t, err, cache.ErrEntryNotFound.Error())
}
