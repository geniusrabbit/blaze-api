package graphql

import (
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	accountlogin "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql/account_login"
	"github.com/geniusrabbit/blaze-api/repository/user"

	"github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/wiring"
)

// Re-export wiring types so callers can use the graphql package directly.
type (
	OptionsConfig[TUser, TAccount any] = wiring.OptionsConfig[TUser, TAccount]
	Option[TUser, TAccount any]        = wiring.Option[TUser, TAccount]
	Options[TUser, TAccount any]       = []wiring.Option[TUser, TAccount]
)

// WithUserAccountResolvers sets all four custom resolvers in one call.
// auth must be the concrete *AuthResolver so the email+password login handler
// can be wired automatically.
func WithUserAccountResolvers[TUser user.Model, TAccount account.Model](
	provider *jwt.Provider,
	u wiring.UserQueryHandler,
	auth *accountgraphql.AuthResolver[TUser, TAccount],
	accounts wiring.ExampleAccountQueryHandler,
	members wiring.ExampleMemberQueryHandler,
) wiring.Option[TUser, TAccount] {
	return wiring.WithUserAccountResolvers(provider, u, auth, accounts, members)
}

// Apply applies opts to cfg and returns it.
func Apply[TUser, TAccount any](opts Options[TUser, TAccount], cfg *OptionsConfig[TUser, TAccount]) *OptionsConfig[TUser, TAccount] {
	return wiring.Apply(opts, cfg)
}

// WithUserLoginHandler re-exports wiring.WithUserLoginHandler.
func WithUserLoginHandler[TUser user.Model, TAccount account.Model](
	provider *jwt.Provider,
	userLogin accountlogin.LoginPasswordAuth[TUser],
	sessionRepo account.SessionRepository[TUser, TAccount],
) wiring.Option[TUser, TAccount] {
	return wiring.WithUserLoginHandler[TUser, TAccount](provider, userLogin, sessionRepo)
}
