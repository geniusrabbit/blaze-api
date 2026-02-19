package directives

import (
	"context"
	"encoding/gob"
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
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

type gobEnvelope struct {
	Value any
}

func init() {
	gob.Register(&gobEnvelope{})

	gob.Register(&types.DateTime{})
	gob.Register(&time.Time{})
}

// CacheData is a GraphQL directive that caches the result of a field/method for a specified time to live (ttl) in seconds.
func CacheData(ctx context.Context, obj any, next graphql.Resolver, ttl int, keyTmp *string, fields []string) (res any, err error) {
	opCtx := graphql.GetOperationContext(ctx)
	cacheCli := gqlapicache.GetGqlApiCache(ctx)

	if cacheCli == nil {
		return next(ctx)
	}

	fc := graphql.GetFieldContext(ctx)

	// Generate a cache key based on the operation context, object type, and specified fields
	cacheKey := getCacheKey(opCtx, fc, obj, gocast.PtrAsValue(keyTmp, ""), fields)

	// Try to get the cached result
	if err = cacheGet(ctx, cacheCli, cacheKey, &res); err == nil {
		return res, nil
	} else if !errors.Is(err, cache.ErrEntryNotFound) {
		ctxlogger.Get(ctx).Warn("Cache miss for key:",
			zap.String("key", cacheKey), zap.Error(err))
	}

	// Call the next resolver in the chain and cache the result
	if res, err = next(ctx); err == nil {
		cErr := cacheSet(ctx, cacheCli, cacheKey, res, time.Duration(ttl)*time.Second)
		if cErr != nil {
			ctxlogger.Get(ctx).Warn("Failed to set cache for key:",
				zap.String("key", cacheKey), zap.Error(cErr), zap.Int("ttl", ttl),
				zap.Any("type", reflect.TypeOf(res)))
		}
	}
	return res, err
}

func getCacheKey(opCtx *graphql.OperationContext, fc *graphql.FieldContext, obj any, keyTmp string, fields []string) string {
	if keyTmp == "" {
		opName := opCtx.OperationName
		if opName == "" {
			opName = "op"
		}

		objName := "unknown"
		if fc != nil {
			// fc.Object is like "Query", "Video", etc.
			objName = fc.Object + "." + fc.Field.Name
		} else {
			objName = fmt.Sprintf("%T", obj)
		}

		if len(fields) > 0 {
			keyTmp = fmt.Sprintf("%s.%s@{%s}", opName, objName, strings.Join(fields, "}:{"))
		} else {
			// safer default than "ID"
			fields = []string{"id"}
			keyTmp = fmt.Sprintf("%s.%s@{id}", opName, objName)
		}
	}

	// Replace placeholders using field args first (fallback to variables)
	for _, field := range fields {
		var v any
		if fc != nil && fc.Args != nil {
			v = fc.Args[field]
		}
		if v == nil && opCtx != nil && opCtx.Variables != nil {
			if v = opCtx.Variables[field]; v == nil {
				v = opCtx.Variables[field]
			}
		}
		keyTmp = strings.ReplaceAll(keyTmp, "{"+field+"}", gocast.Str(v))
	}
	return keyTmp
}

func cacheSet(ctx context.Context, cacheCli cache.Client, key string, value any, ttl time.Duration) error {
	data, err := marshalObject(gobEnvelope{Value: value})
	if err != nil {
		return err
	}
	return cacheCli.Set(ctx, key, data, ttl)
}

func cacheGet(ctx context.Context, cacheCli cache.Client, key string, dest *any) error {
	data := []byte{}
	if err := cacheCli.Get(ctx, key, &data); err != nil {
		return err
	}

	var env gobEnvelope
	if err := unmarshalObject(data, &env); err != nil {
		return err
	}
	*dest = env.Value
	return nil
}

func marshalObject(obj any) ([]byte, error) {
	var buf strings.Builder
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(obj); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func unmarshalObject(data []byte, obj any) error {
	buf := strings.NewReader(string(data))
	dec := gob.NewDecoder(buf)
	return dec.Decode(obj)
}
