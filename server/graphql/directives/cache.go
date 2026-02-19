package directives

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/demdxx/gocast/v2"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/server/graphql/context/gqlapicache"
)

// CacheData is a GraphQL directive that caches the result of a field/method for a specified time to live (ttl) in seconds.
func CacheData(ctx context.Context, obj any, next graphql.Resolver, ttl int, keyTmp *string, fields []string) (res any, err error) {
	cacheCli := gqlapicache.GetGqlApiCache(ctx)
	if cacheCli == nil {
		return next(ctx)
	}

	// Implement caching logic here
	if gocast.IsNil(obj) {
		obj = reflect.Zero(reflect.TypeOf(obj).Elem()).Interface()
	}

	// Generate a cache key based on the operation context, object type, and specified fields
	opCtx := graphql.GetOperationContext(ctx)
	cacheKey := getCacheKey(opCtx, obj, gocast.PtrAsValue(keyTmp, ""), fields)

	// Try to get the cached result
	if err := cacheCli.Get(ctx, cacheKey, obj); err == nil {
		return obj, nil
	} else if !errors.Is(err, cache.ErrEntryNotFound) {
		ctxlogger.Get(ctx).Warn("Cache miss for key:",
			zap.String("key", cacheKey), zap.Error(err))
	}

	// Call the next resolver in the chain and cache the result
	if res, err = next(ctx); err == nil {
		cErr := cacheCli.Set(ctx, cacheKey, res, time.Duration(ttl)*time.Second)
		if cErr != nil {
			ctxlogger.Get(ctx).Warn("Failed to set cache for key:",
				zap.String("key", cacheKey), zap.Error(cErr), zap.Int("ttl", ttl))
		}
	}
	return res, err
}

func getCacheKey(opCtx *graphql.OperationContext, obj any, keyTmp string, fields []string) string {
	if keyTmp == "" {
		keyTmp = fmt.Sprintf("%s.%s", opCtx.OperationName, reflect.TypeOf(obj).String())
		if len(fields) > 0 {
			keyTmp += "@{" + strings.Join(fields, "}:{") + "}"
		} else {
			keyTmp += "@{ID}"
			fields = []string{"ID"}
		}
	}
	for _, field := range fields {
		value := gocast.Str(opCtx.Variables[field])
		keyTmp = strings.ReplaceAll(keyTmp, "{"+field+"}", value)
	}
	return keyTmp
}
