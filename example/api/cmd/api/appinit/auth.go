package appinit

import (
	"context"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/cache"
	"github.com/geniusrabbit/blaze-api/cache/dummy"
	"github.com/geniusrabbit/blaze-api/cache/memory"
	"github.com/geniusrabbit/blaze-api/cache/redis"
	"github.com/geniusrabbit/blaze-api/example/api/cmd/api/appcontext"
	"github.com/geniusrabbit/blaze-api/jwt"
	"github.com/geniusrabbit/blaze-api/oauth2srvprovider"
	user_repository "github.com/geniusrabbit/blaze-api/repository/user/repository"
)

// Auth new provider
func Auth(ctx context.Context, conf *appcontext.ConfigType, masterDatabase *gorm.DB) (fosite.OAuth2Provider, *jwt.Provider) {
	oauth2config := &fosite.Config{
		AccessTokenLifespan:           conf.OAuth2.AccessTokenLifespan,
		RefreshTokenLifespan:          conf.OAuth2.RefreshTokenLifespan,
		AuthorizeCodeLifespan:         conf.OAuth2.AuthorizeCodeLifespan,
		HashCost:                      conf.OAuth2.HashCost,
		DisableRefreshTokenValidation: conf.OAuth2.DisableRefreshTokenValidation,
		SendDebugMessagesToClients:    conf.OAuth2.SendDebugMessagesToClients,
	}
	sessionCache := newCache(ctx, conf.OAuth2.CacheConnect, conf.OAuth2.CacheLifetime)
	userRepository := user_repository.New()
	oauth2storage := oauth2srvprovider.NewDatabaseStorage(
		masterDatabase,
		userRepository,
		sessionCache,
		conf.OAuth2.CacheLifetime,
	)
	oauth2provider := oauth2srvprovider.NewProvider(
		oauth2config,
		oauth2storage,
		&compose.CommonStrategy{
			CoreStrategy: compose.NewOAuth2HMACStrategy(oauth2config),
		},
		nil,
	)
	jwtProvider := &jwt.Provider{
		TokenLifetime:  conf.OAuth2.AccessTokenLifespan,
		Secret:         conf.OAuth2.Secret,
		MiddlewareOpts: &jwtmiddleware.Options{Debug: conf.IsDebug()},
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
