package sysops

import (
	"context"

	"github.com/demdxx/xtypes"
)

var _globalOpts options

// SetExtraReader sets an extra reader for the global options
func SetExtraReader(reader readOnlyReader) {
	_globalOpts.reader = reader
}

// Has checks if the option with the given key exists
func Has(ctx context.Context, key string) bool {
	return _globalOpts.Has(ctx, key)
}

// Get option value by key
// If the key does not exist, it returns the default value if provided
// Otherwise, it returns nil
func Get(ctx context.Context, key string, def ...any) *xtypes.Any {
	return _globalOpts.Get(ctx, key, def...)
}

// First returns the first option value for the given keys
func First(ctx context.Context, keys ...string) *xtypes.Any {
	for _, key := range keys {
		if value := _globalOpts.Get(ctx, key); value != nil {
			return value
		}
	}
	return nil
}

// Set option value by key
func Set(ctx context.Context, key string, value any) {
	_globalOpts.Set(ctx, key, value)
}

// Delete option by key
func Delete(ctx context.Context, keys ...string) {
	_globalOpts.Delete(ctx, keys...)
}
