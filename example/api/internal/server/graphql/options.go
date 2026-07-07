package graphql

import (
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	accountlogin "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql/account_login"
	"github.com/geniusrabbit/blaze-api/repository/user"

	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/wiring"
)

// Re-export wiring types so callers can use the graphql package directly.
type (
	OptionsConfig = wiring.OptionsConfig
	Option        = wiring.Option
	Options       = []wiring.Option
)

// WithUserAccountResolvers sets all four custom resolvers in one call.
// auth must be the concrete *AuthResolver so the email+password login handler
// can be wired automatically.
func WithUserAccountResolvers[TUser user.Model, TAccount account.Model](
	provider *jwt.Provider,
	users wiring.UserQueryResolver,
	auth *accountgraphql.AuthResolver[TUser, TAccount],
	accounts accountgraphql.AccountQueryHandler[*exmodels.Account, *exmodels.AccountPayload, *exmodels.AccountCreateInput, *exmodels.AccountUpdateInput, *exmodels.AccountListFilter, *exmodels.AccountListOrder],
	members accountgraphql.MemberQueryHandler,
) wiring.Option {
	return wiring.WithUserAccountResolvers(provider, users, auth, accounts, members)
}

// Apply applies opts to cfg and returns it.
func Apply(opts Options, cfg *OptionsConfig) *OptionsConfig {
	return wiring.Apply(opts, cfg)
}

// WithUserLoginHandler re-exports wiring.WithUserLoginHandler.
func WithUserLoginHandler[TUser user.Model, TAccount account.Model](
	provider *jwt.Provider,
	userLogin accountlogin.LoginPasswordAuth[TUser],
	sessionRepo account.SessionRepository[TUser, TAccount],
) wiring.Option {
	return wiring.WithUserLoginHandler(provider, userLogin, sessionRepo)
}
