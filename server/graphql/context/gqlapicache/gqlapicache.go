package gqlapicache

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
)

var (
	ctxGqlApiCacheKey = struct{}{}
)

func WithGqlApiCache(ctx context.Context, cache cache.Client) context.Context {
	return context.WithValue(ctx, ctxGqlApiCacheKey, cache)
}

func GetGqlApiCache(ctx context.Context) cache.Client {
	if cache, ok := ctx.Value(ctxGqlApiCacheKey).(cache.Client); ok {
		return cache
	}
	return nil
}
