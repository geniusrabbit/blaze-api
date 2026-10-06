package sysops

import (
	"context"
	"sync"

	"github.com/demdxx/xtypes"
)

type readOnlyReader interface {
	Has(ctx context.Context, key string) bool
	Get(ctx context.Context, key string) (any, bool)
}

type options struct {
	values sync.Map
	reader readOnlyReader
}

// Has checks if the option with the given key exists
func (o *options) Has(ctx context.Context, key string) bool {
	_, ok := o.values.Load(key)
	if !ok && o.reader != nil {
		ok = o.reader.Has(ctx, key)
	}
	return ok
}

// Get option value by key
// If the key does not exist, it returns the default value if provided
// Otherwise, it returns nil
func (o *options) Get(ctx context.Context, key string, def ...any) *xtypes.Any {
	if v, ok := o.values.Load(key); ok && v != nil {
		return &xtypes.Any{Val: v}
	}
	if o.reader != nil {
		if v, ok := o.reader.Get(ctx, key); ok {
			return &xtypes.Any{Val: v}
		}
	}
	if len(def) > 0 {
		return &xtypes.Any{Val: def[0]}
	}
	return nil
}

// Set option value by key
func (o *options) Set(ctx context.Context, key string, value any) {
	o.values.Store(key, value)
}

// Delete option by key
func (o *options) Delete(ctx context.Context, keys ...string) {
	for _, key := range keys {
		o.values.Delete(key)
	}
}
