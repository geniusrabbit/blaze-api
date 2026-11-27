package jwt

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
)

// ErrRevocationDisabled is returned by revocation methods when the provider has no store.
var ErrRevocationDisabled = errors.New("jwt: token revocation is not configured")

// RevocationStore keeps revoked tokens until they expire.
type RevocationStore interface {
	// RevokeToken marks the token ID as revoked until the time.
	RevokeToken(ctx context.Context, id string, until time.Time) error
	// IsTokenRevoked reports whether the token ID is revoked.
	IsTokenRevoked(ctx context.Context, id string) (bool, error)
	// RevokeUserTokens revokes the user's tokens issued before issuedBefore;
	// the record may be dropped after until (when all those tokens expired).
	RevokeUserTokens(ctx context.Context, userID uint64, issuedBefore, until time.Time) error
	// UserTokensRevokedBefore returns the time set by RevokeUserTokens, or zero.
	UserTokensRevokedBefore(ctx context.Context, userID uint64) (time.Time, error)
}

// CacheRevocationStore keeps revocations in a cache. Use a shared cache
// (e.g. Redis) when the API runs on several instances.
type CacheRevocationStore struct {
	cache cache.Client
}

// NewCacheRevocationStore creates a revocation store over the cache.
func NewCacheRevocationStore(c cache.Client) *CacheRevocationStore {
	return &CacheRevocationStore{cache: c}
}

// RevokeToken implements RevocationStore.
func (s *CacheRevocationStore) RevokeToken(ctx context.Context, id string, until time.Time) error {
	return s.cache.Set(ctx, "jwt:revoked:"+id, []byte("1"), time.Until(until)+time.Minute)
}

// IsTokenRevoked implements RevocationStore.
func (s *CacheRevocationStore) IsTokenRevoked(ctx context.Context, id string) (bool, error) {
	var v []byte
	err := s.cache.Get(ctx, "jwt:revoked:"+id, &v)
	if errors.Is(err, cache.ErrEntryNotFound) {
		return false, nil
	}
	return err == nil, err
}

// RevokeUserTokens implements RevocationStore.
func (s *CacheRevocationStore) RevokeUserTokens(ctx context.Context, userID uint64, issuedBefore, until time.Time) error {
	return s.cache.Set(ctx, userRevocationKey(userID),
		[]byte(strconv.FormatInt(issuedBefore.Unix(), 10)), time.Until(until)+time.Minute)
}

// UserTokensRevokedBefore implements RevocationStore.
func (s *CacheRevocationStore) UserTokensRevokedBefore(ctx context.Context, userID uint64) (time.Time, error) {
	var v []byte
	err := s.cache.Get(ctx, userRevocationKey(userID), &v)
	if errors.Is(err, cache.ErrEntryNotFound) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	sec, err := strconv.ParseInt(string(v), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}

func userRevocationKey(userID uint64) string {
	return "jwt:revoked:user:" + strconv.FormatUint(userID, 10)
}
