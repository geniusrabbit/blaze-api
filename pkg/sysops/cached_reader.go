package sysops

import (
	"context"
	"sync"
	"time"
)

type cacheEntry struct {
	value   any
	ok      bool
	expires time.Time
}

type cachedReader struct {
	inner readOnlyReader
	ttl   time.Duration

	mu    sync.Mutex
	items map[string]cacheEntry
}

// NewCachedReader wraps inner with a TTL cache.
// ttl <= 0 disables the cache and returns inner as-is.
func NewCachedReader(inner readOnlyReader, ttl time.Duration) readOnlyReader {
	if inner == nil || ttl <= 0 {
		return inner
	}
	return &cachedReader{
		inner: inner,
		ttl:   ttl,
		items: make(map[string]cacheEntry),
	}
}

func (r *cachedReader) Has(ctx context.Context, key string) bool {
	if entry, ok := r.lookup(key); ok {
		return entry.ok
	}
	exists := r.inner.Has(ctx, key)
	if !exists {
		r.store(key, nil, false)
	}
	return exists
}

func (r *cachedReader) Get(ctx context.Context, key string) (any, bool) {
	if entry, ok := r.lookup(key); ok {
		return entry.value, entry.ok
	}
	value, exists := r.inner.Get(ctx, key)
	r.store(key, value, exists)
	return value, exists
}

func (r *cachedReader) lookup(key string) (cacheEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.items[key]
	if !ok {
		return cacheEntry{}, false
	}
	if !entry.expires.After(time.Now()) {
		delete(r.items, key)
		return cacheEntry{}, false
	}
	return entry, true
}

func (r *cachedReader) store(key string, value any, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[key] = cacheEntry{
		value:   value,
		ok:      ok,
		expires: time.Now().Add(r.ttl),
	}
}
