package resolvers

import (
	"github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/wiring"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	authclientgraphql "github.com/geniusrabbit/blaze-api/repository/authclient/delivery/graphql"
	datokengraphql "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/delivery/graphql"
	historyloggraphql "github.com/geniusrabbit/blaze-api/repository/historylog/delivery/graphql"
	"github.com/geniusrabbit/blaze-api/repository/option"
	optiongraphql "github.com/geniusrabbit/blaze-api/repository/option/delivery/graphql"
	rbacgraphql "github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql"
	socialaccountgraphql "github.com/geniusrabbit/blaze-api/repository/socialaccount/delivery/graphql"
)

// Resolver is example/api GraphQL root with extended Account schema types.
type Resolver struct {
	users             wiring.UserQueryHandler
	accAuth           accountgraphql.AuthQueryHandler
	loginHandler      wiring.EmailPasswordLoginHandler
	accounts          wiring.ExampleAccountQueryHandler
	members           wiring.ExampleMemberQueryHandler
	socAccounts       *socialaccountgraphql.QueryResolver
	roles             *rbacgraphql.QueryResolver
	authclients       *authclientgraphql.QueryResolver
	historylogs       *historyloggraphql.QueryResolver
	options           *optiongraphql.QueryResolver
	directaccesstoken *datokengraphql.QueryResolver
}

// NewResolver wires the example/api GraphQL handler from explicit resolver handles.
// All custom logic (user, auth, accounts, members) must be provided by the caller.
// Standard resolvers (rbac, authclient, historylog, option, DAT) are initialized internally.
func NewResolver(
	provider *jwt.Provider,
	options option.Usecase,
	userHandler wiring.UserQueryHandler,
	authHandler accountgraphql.AuthQueryHandler,
	loginHandler wiring.EmailPasswordLoginHandler,
	accountHandler wiring.ExampleAccountQueryHandler,
	memberHandler wiring.ExampleMemberQueryHandler,
) *Resolver {
	return &Resolver{
		users:             userHandler,
		accAuth:           authHandler,
		loginHandler:      loginHandler,
		accounts:          accountHandler,
		members:           memberHandler,
		socAccounts:       socialaccountgraphql.NewDefaultQueryResolver(),
		roles:             rbacgraphql.NewDefaultQueryResolver(),
		authclients:       authclientgraphql.NewDefaultQueryResolver(),
		historylogs:       historyloggraphql.NewDefaultQueryResolver(),
		options:           optiongraphql.NewQueryResolver(options),
		directaccesstoken: datokengraphql.NewDefaultQueryResolver(),
	}
}
