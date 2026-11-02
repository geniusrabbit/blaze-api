package sysops

import (
	"context"
	"sync"
	"time"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
)

type cacheEntry struct {
	value   any
	ok      bool
	expires time.Time
}

type cacheKey struct {
	userID    uint64
	accountID uint64
	name      string
}

type cachedReader struct {
	inner readOnlyReader
	ttl   time.Duration

	mu    sync.Mutex
	items map[cacheKey]cacheEntry
}

// NewCachedReader wraps inner with a TTL cache.
// ttl <= 0 disables the cache and returns inner as-is.
// The cache key is the session user, the session account, and the option name.
// A request with no session uses 0 and 0.
func NewCachedReader(inner readOnlyReader, ttl time.Duration) readOnlyReader {
	if inner == nil || ttl <= 0 {
		return inner
	}
	return &cachedReader{
		inner: inner,
		ttl:   ttl,
		items: make(map[cacheKey]cacheEntry),
	}
}

func (r *cachedReader) Has(ctx context.Context, key string) bool {
	ck := readerCacheKey(ctx, key)
	if entry, ok := r.lookup(ck); ok {
		return entry.ok
	}
	exists := r.inner.Has(ctx, key)
	if !exists {
		r.store(ck, nil, false)
	}
	return exists
}

func (r *cachedReader) Get(ctx context.Context, key string) (any, bool) {
	ck := readerCacheKey(ctx, key)
	if entry, ok := r.lookup(ck); ok {
		return entry.value, entry.ok
	}
	value, exists := r.inner.Get(ctx, key)
	r.store(ck, value, exists)
	return value, exists
}

func readerCacheKey(ctx context.Context, name string) cacheKey {
	return cacheKey{
		userID:    session.UserID(ctx),
		accountID: session.AccountID(ctx),
		name:      name,
	}
}

func (r *cachedReader) lookup(key cacheKey) (cacheEntry, bool) {
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

func (r *cachedReader) store(key cacheKey, value any, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[key] = cacheEntry{
		value:   value,
		ok:      ok,
		expires: time.Now().Add(r.ttl),
	}
}
