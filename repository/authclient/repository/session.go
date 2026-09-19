package repository

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/cache"
	"github.com/geniusrabbit/blaze-api/pkg/cache/dummy"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/models"
	"github.com/geniusrabbit/blaze-api/repository/generated"
)

var (
	defaultSessionCache   cache.Client
	defaultSessionCacheMu sync.RWMutex
)

// SetDefaultSessionCache sets the OAuth session cache used when a repository
// is constructed without an explicit client. Wire this to the same cache as
// oauth2 DatabaseStorage so revoked tokens die immediately.
func SetDefaultSessionCache(c cache.Client) {
	defaultSessionCacheMu.Lock()
	defer defaultSessionCacheMu.Unlock()
	defaultSessionCache = c
}

func sessionCacheOrDummy(c cache.Client) cache.Client {
	if c != nil {
		return c
	}
	defaultSessionCacheMu.RLock()
	def := defaultSessionCache
	defaultSessionCacheMu.RUnlock()
	if def != nil {
		return def
	}
	return dummy.New()
}

// SessionRepository DAO for AuthSession rows.
type SessionRepository struct {
	generated.Repository[models.AuthSession, uint64]
	cache cache.Client
}

// NewSessionRepository creates a session repository. A nil cache uses the
// default OAuth session cache when set, otherwise a dummy cache.
func NewSessionRepository(cacheClient ...cache.Client) *SessionRepository {
	var c cache.Client
	if len(cacheClient) > 0 {
		c = cacheClient[0]
	}
	return &SessionRepository{
		Repository: *generated.NewRepository[models.AuthSession, uint64](),
		cache:      sessionCacheOrDummy(c),
	}
}

// Delete soft-deletes the session and drops OAuth session cache keys so
// access/refresh tokens stop resolving immediately.
func (r *SessionRepository) Delete(ctx context.Context, id uint64, opts ...authclient.QOption) error {
	sess, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := r.Repository.Delete(ctx, id, opts...); err != nil {
		return err
	}
	r.dropSessionCache(ctx, sess)
	return nil
}

func (r *SessionRepository) dropSessionCache(ctx context.Context, sess *models.AuthSession) {
	if r.cache == nil || sess == nil {
		return
	}
	keys := make([]string, 0, 2)
	if sess.AccessToken != "" {
		keys = append(keys, "sess:"+sess.AccessToken)
	}
	if sess.RefreshToken.Valid && sess.RefreshToken.String != "" {
		keys = append(keys, "sess:"+sess.RefreshToken.String)
	}
	for _, key := range keys {
		if err := r.cache.Del(ctx, key); err != nil {
			ctxlogger.Get(ctx).Error("clear session cache",
				zap.String("cache_key", key), zap.Error(err))
		}
	}
}
