package resolvers

import (
	"github.com/geniusrabbit/blaze-api/jwt"
	account_graphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	authclient_graphql "github.com/geniusrabbit/blaze-api/repository/authclient/delivery/graphql"
	historylog_graphql "github.com/geniusrabbit/blaze-api/repository/historylog/delivery/graphql"
	option_graphql "github.com/geniusrabbit/blaze-api/repository/option/delivery/graphql"
	rbac_graphql "github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql"
	rbac "github.com/geniusrabbit/blaze-api/repository/rbac/repository"
	user_graphql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	users       *user_graphql.QueryResolver
	accAuth     *account_graphql.AuthResolver
	accounts    *account_graphql.QueryResolver
	roles       *rbac_graphql.QueryResolver
	authclients *authclient_graphql.QueryResolver
	historylogs *historylog_graphql.QueryResolver
	options     *option_graphql.QueryResolver
}

func NewResolver(provider *jwt.Provider) *Resolver {
	return &Resolver{
		users:       user_graphql.NewQueryResolver(),
		accAuth:     account_graphql.NewAuthResolver(provider, rbac.New()),
		accounts:    account_graphql.NewQueryResolver(),
		roles:       rbac_graphql.NewQueryResolver(),
		authclients: authclient_graphql.NewQueryResolver(),
		historylogs: historylog_graphql.NewQueryResolver(),
		options:     option_graphql.NewQueryResolver(),
	}
}