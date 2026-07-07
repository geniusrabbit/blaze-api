package wiring

import (
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	accountlogin "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql/account_login"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// OptionsConfig carries all resolver overrides for the example/api GraphQL handler.
// Instantiate with concrete user/account types at the entry point (e.g. main.go).
type OptionsConfig[TUser, TAccount any] struct {
	UserHandler    UserQueryHandler
	AuthHandler    accountgraphql.AuthQueryHandler
	LoginHandler   EmailPasswordLoginHandler
	AccountHandler ExampleAccountQueryHandler
	MemberHandler  ExampleMemberQueryHandler
}

// Option is a functional option applied to OptionsConfig[TUser, TAccount].
type Option[TUser, TAccount any] func(*OptionsConfig[TUser, TAccount])

// Apply runs all options against cfg and returns it.
func Apply[TUser, TAccount any](opts []Option[TUser, TAccount], cfg *OptionsConfig[TUser, TAccount]) *OptionsConfig[TUser, TAccount] {
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
	user UserQueryHandler,
	auth *accountgraphql.AuthResolver[TUser, TAccount],
	accounts ExampleAccountQueryHandler,
	members ExampleMemberQueryHandler,
) Option[TUser, TAccount] {
	return func(cfg *OptionsConfig[TUser, TAccount]) {
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
) Option[TUser, TAccount] {
	return func(cfg *OptionsConfig[TUser, TAccount]) {
		cfg.LoginHandler = accountlogin.New[TUser, TAccount](provider, userLogin, sessionRepo)
	}
}
