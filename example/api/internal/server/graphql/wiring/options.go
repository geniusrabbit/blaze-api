package wiring

import (
	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	accountlogin "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql/account_login"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

type AccountQueryHandler = accountgraphql.AccountQueryHandler[*exmodels.Account, *exmodels.AccountPayload, *exmodels.AccountCreateInput, *exmodels.AccountUpdateInput, *exmodels.AccountListFilter, *exmodels.AccountListOrder]

// OptionsConfig carries all resolver overrides for the example/api GraphQL handler.
// Instantiate with concrete user/account types at the entry point (e.g. main.go).
type OptionsConfig struct {
	UserHandler    UserQueryResolver
	AuthHandler    accountgraphql.AuthQueryHandler
	LoginHandler   accountgraphql.AccountLoginHandler
	AccountHandler AccountQueryHandler
	MemberHandler  accountgraphql.MemberQueryHandler
}

// Option is a functional option applied to OptionsConfig.
type Option func(*OptionsConfig)

// Apply runs all options against cfg and returns it.
func Apply(opts []Option, cfg *OptionsConfig) *OptionsConfig {
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// WithUserAccountResolvers sets all four custom resolvers in one call.
// auth must be the concrete *AuthResolver so that the email+password login handler
// can be wired automatically via accountlogin.New(auth).
//
// Example:
//
//	graphql.WithUserAccountResolvers[*domain.User, *domain.Account](
//	    wiring.NewExampleUserQueryResolver(deps.UserModule),
//	    accountgraphql.NewAuthResolver(jwtProvider, login, deps.AccountRepo, deps.AccountUC, rbacrepo.New()),
//	    wiring.NewExampleAccountQueryResolver(cfg),
//	    wiring.NewExampleMemberQueryResolver(deps.AccountUC, deps.MemberUC, ...),
//	)
func WithUserAccountResolvers[TUser user.Model, TAccount account.Model](
	provider *jwt.Provider,
	user UserQueryResolver,
	auth *accountgraphql.AuthResolver[TUser, TAccount],
	accounts AccountQueryHandler,
	members accountgraphql.MemberQueryHandler,
) Option {
	return func(cfg *OptionsConfig) {
		cfg.UserHandler = user
		cfg.AuthHandler = auth
		cfg.AccountHandler = accounts
		cfg.MemberHandler = members
	}
}

// WithUserLoginHandler sets the email+password login handler for the example/api GraphQL handler.
// auth must be the concrete *AuthResolver so that the email+password login handler
// can be wired automatically via accountlogin.New(auth).
func WithUserLoginHandler[TUser user.Model, TAccount account.Model](
	provider *jwt.Provider,
	userLogin accountlogin.LoginPasswordAuth[TUser],
	sessionRepo account.SessionRepository[TUser, TAccount],
) Option {
	return func(cfg *OptionsConfig) {
		cfg.LoginHandler = accountlogin.New(provider, userLogin, sessionRepo)
	}
}
