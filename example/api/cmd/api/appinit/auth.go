package appinit

import (
	"context"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/example/api/cmd/api/appcontext"
	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/pkg/auth/oauth2/serverprovider"
	"github.com/geniusrabbit/blaze-api/pkg/cache"
	"github.com/geniusrabbit/blaze-api/pkg/cache/dummy"
	"github.com/geniusrabbit/blaze-api/pkg/cache/memory"
	"github.com/geniusrabbit/blaze-api/pkg/cache/redis"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

type oauthUserAccessor struct {
	email    user.EmailRepository[*domain.User]
	password user.PasswordRepository[*domain.User]
}

func (a oauthUserAccessor) GetByEmail(ctx context.Context, email string) (user.Model, error) {
	return a.email.GetByEmail(ctx, email)
}

func (a oauthUserAccessor) GetByPassword(ctx context.Context, userID uint64, password string) (user.Model, error) {
	return a.password.GetByPassword(ctx, userID, password)
}

// Auth new provider
func Auth(ctx context.Context, conf *appcontext.ConfigType, masterDatabase *gorm.DB, deps *Deps) (fosite.OAuth2Provider, *jwt.Provider) {
	oauth2config := &fosite.Config{
		AccessTokenLifespan:           conf.OAuth2.AccessTokenLifespan,
		RefreshTokenLifespan:          conf.OAuth2.RefreshTokenLifespan,
		AuthorizeCodeLifespan:         conf.OAuth2.AuthorizeCodeLifespan,
		HashCost:                      conf.OAuth2.HashCost,
		DisableRefreshTokenValidation: conf.OAuth2.DisableRefreshTokenValidation,
		SendDebugMessagesToClients:    conf.OAuth2.SendDebugMessagesToClients,
	}
	sessionCache := newCache(ctx, conf.OAuth2.CacheConnect, conf.OAuth2.CacheLifetime)
	oauth2storage := serverprovider.NewDatabaseStorage(
		masterDatabase,
		oauthUserAccessor{email: deps.UserModule.Repo, password: deps.UserModule.Repo},
		sessionCache,
		conf.OAuth2.CacheLifetime,
	)
	oauth2provider := serverprovider.NewProvider(
		oauth2config,
		oauth2storage,
		&compose.CommonStrategy{
			CoreStrategy: compose.NewOAuth2HMACStrategy(oauth2config),
		},
		nil,
	)
	jwtProvider := &jwt.Provider{
		TokenLifetime: conf.OAuth2.AccessTokenLifespan,
		Secret:        conf.OAuth2.Secret,
		MiddlewareOpts: &jwtmiddleware.Options{
			Debug: conf.IsDebug(),
			Extractor: jwtmiddleware.FromFirst(
				jwtmiddleware.FromAuthHeader,
				jwtmiddleware.FromParameter("access_token"),
			),
		},
	}
	return oauth2provider, jwtProvider
}

func newCache(ctx context.Context, connect string, lifetime time.Duration) cache.Client {
	switch {
	case connect == ":memory:":
		cacheObj, err := memory.NewTimeout(ctx, lifetime)
		fatalError(err, "memory cache")
		return cacheObj
	case connect == ":dummy:" || connect == "":
		return dummy.New()
	case strings.HasPrefix(connect, "redis://"):
		cli, err := redis.NewByURL(connect)
		if err != nil {
			panic(err)
		}
		return cli
	default:
		return dummy.New()
	}
}
