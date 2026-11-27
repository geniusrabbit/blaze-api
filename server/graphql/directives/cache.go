package directives

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/demdxx/gocast/v2"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/server/graphql/context/gqlapicache"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

// cacheKeyPrefix versions the cache entries; entries of older formats are ignored.
const cacheKeyPrefix = "gqlcache:v2:"

// cacheEnvelope is the stored form of a cached value. A CacheEncodable value
// is kept as Type+Data and decoded into a new value of that type; any other
// value is gob-encoded as Value.
type cacheEnvelope struct {
	Type  string
	Data  []byte
	Value any
}

// decodableTypes maps the type names of CacheEncodable values seen by this
// process to their types, so they can be decoded on a cache hit.
var decodableTypes sync.Map // map[string]reflect.Type

func init() {
	gob.Register(&cacheEnvelope{})

	gob.Register(&types.DateTime{})
	gob.Register(&time.Time{})
}

var errNotDecodable = errors.New("cache: value implements CacheEncodable but not CacheDecodable")

type CacheEncodable interface {
	EncodableCacheValue() ([]byte, error)
}

type CacheDecodable interface {
	DecodableCacheValue([]byte) error
}

// CacheData is a GraphQL directive that caches the result of a field/method for a specified time to live (ttl) in seconds.
//
// Cached values are isolated per session user and account, because resolvers
// filter data by the session permissions. The directive only works when
// declared in the schema (FIELD_DEFINITION): when a client puts it on a field
// in a query, it is ignored.
func CacheData(ctx context.Context, obj any, next graphql.Resolver, ttl int, keyTmp *string, fields []string) (res any, err error) {
	opCtx := graphql.GetOperationContext(ctx)
	cacheCli := gqlapicache.GetGqlApiCache(ctx)

	if cacheCli == nil {
		ctxlogger.Get(ctx).Warn("No cache client available in context, skipping cache directive")
		return next(ctx)
	}

	fc := graphql.GetFieldContext(ctx)
	if isQueryDirective(fc, "cacheData") {
		ctxlogger.Get(ctx).Debug("Ignoring @cacheData set in the query")
		return next(ctx)
	}

	// Generate a cache key based on the session, operation context, object type, and specified fields
	cacheKey := sessionCacheKeyPrefix(ctx) + getCacheKey(opCtx, fc, obj, gocast.PtrAsValue(keyTmp, ""), fields)

	// Try to get the cached result
	if err = cacheGet(ctx, cacheCli, cacheKey, &res); err == nil {
		ctxlogger.Get(ctx).Debug("Cache hit",
			zap.String("key", cacheKey),
			zap.Int("ttl", ttl),
			zap.Any("type", reflect.TypeOf(res)),
		)
		return res, nil
	} else if !errors.Is(err, cache.ErrEntryNotFound) {
		ctxlogger.Get(ctx).Warn("Cache miss with error",
			zap.String("key", cacheKey),
			zap.Int("ttl", ttl),
			zap.Error(err),
			zap.Any("type", reflect.TypeOf(res)),
		)
	}

	// Call the next resolver in the chain and cache the result
	if res, err = next(ctx); err == nil {
		cErr := cacheSet(ctx, cacheCli, cacheKey, res, time.Duration(ttl)*time.Second)
		if cErr != nil {
			ctxlogger.Get(ctx).Warn("Failed to set cache",
				zap.String("key", cacheKey),
				zap.Int("ttl", ttl),
				zap.Error(cErr),
				zap.Any("type", reflect.TypeOf(res)),
			)
		} else {
			ctxlogger.Get(ctx).Debug("Cached result",
				zap.String("key", cacheKey),
				zap.Int("ttl", ttl),
				zap.Any("type", reflect.TypeOf(res)),
			)
		}
	}
	return res, err
}

// isQueryDirective reports whether the directive is set on the field in the
// query document (by the client), not in the schema.
func isQueryDirective(fc *graphql.FieldContext, name string) bool {
	return fc != nil && fc.Field.Field != nil && fc.Field.Directives.ForName(name) != nil
}

// sessionCacheKeyPrefix scopes cache entries to the session user and account.
func sessionCacheKeyPrefix(ctx context.Context) string {
	var userID, accountID uint64
	userObj, accountObj := session.UserAccount(ctx)
	if !gocast.IsNil(userObj) {
		userID = userObj.GetID()
	}
	if !gocast.IsNil(accountObj) {
		accountID = accountObj.GetID()
	}
	return fmt.Sprintf("%su%d:a%d:", cacheKeyPrefix, userID, accountID)
}

func getCacheKey(opCtx *graphql.OperationContext, fc *graphql.FieldContext, obj any, keyTmp string, fields []string) string {
	if keyTmp == "" {
		opName := opCtx.OperationName
		if opName == "" {
			opName = "op"
		}

		var objName string
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

	var objMap map[string]any

	// Replace placeholders using field args first (fallback to variables)
	for _, field := range fields {
		var v any
		if strings.HasPrefix(field, ".") || strings.HasPrefix(field, "this.") {
			if obj != nil {
				newField := strings.TrimPrefix(field, ".")
				newField = strings.TrimPrefix(newField, "this.")
				v, _ = gocast.StructFieldValue(obj, newField)
				if v == nil {
					if objMap == nil {
						objMap = gocast.Map[string, any](obj, `json`)
					}
					if v = objMap[newField]; v == nil {
						newField2 := strings.ToLower(newField)
						if newField2 != newField {
							v = objMap[newField2]
						}
					}
				}
			}
		} else {
			if fc != nil && fc.Args != nil {
				v = fc.Args[field]
			}
			if v == nil && opCtx != nil && opCtx.Variables != nil {
				if v = opCtx.Variables[field]; v == nil {
					v = opCtx.Variables[field]
				}
			}
		}
		keyTmp = strings.ReplaceAll(keyTmp, "{"+field+"}", asKeyValue(v))
	}
	return keyTmp
}

func cacheSet(ctx context.Context, cacheCli cache.Client, key string, value any, ttl time.Duration) error {
	var (
		env cacheEnvelope
		err error
	)
	if encodable, ok := value.(CacheEncodable); ok {
		if _, ok := value.(CacheDecodable); !ok {
			return errNotDecodable
		}
		typ := reflect.TypeOf(value)
		env.Type = typ.String()
		decodableTypes.Store(env.Type, typ)
		if env.Data, err = encodable.EncodableCacheValue(); err != nil {
			return err
		}
	} else {
		env.Value = value
	}
	data, err := marshalObject(env)
	if err != nil {
		return err
	}
	return cacheCli.Set(ctx, key, data, ttl)
}

// cacheGet loads a cached value. Entries that can't be decoded (an older
// format, or a type this process hasn't cached yet) are reported as not found.
func cacheGet(ctx context.Context, cacheCli cache.Client, key string, dest *any) error {
	data := []byte{}
	if err := cacheCli.Get(ctx, key, &data); err != nil {
		return err
	}

	var env cacheEnvelope
	if err := unmarshalObject(data, &env); err != nil {
		return cache.ErrEntryNotFound
	}
	if env.Type == "" {
		*dest = env.Value
		return nil
	}
	typ, ok := decodableTypes.Load(env.Type)
	if !ok {
		return cache.ErrEntryNotFound
	}
	rtyp := typ.(reflect.Type)
	var target any
	if rtyp.Kind() == reflect.Pointer {
		target = reflect.New(rtyp.Elem()).Interface()
	} else {
		target = reflect.New(rtyp).Interface()
	}
	decodable, ok := target.(CacheDecodable)
	if !ok {
		return cache.ErrEntryNotFound
	}
	if err := decodable.DecodableCacheValue(env.Data); err != nil {
		return err
	}
	if rtyp.Kind() == reflect.Pointer {
		*dest = target
	} else {
		*dest = reflect.ValueOf(target).Elem().Interface()
	}
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

func asKeyValue(obj any) string {
	if obj == nil {
		return "null"
	}
	type _valueHash interface {
		KeyHash() string
	}
	if v, ok := obj.(_valueHash); ok {
		return v.KeyHash()
	}
	return gocast.Str(obj)
}
