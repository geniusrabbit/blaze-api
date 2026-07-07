package wiring

import (
	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
)

// ExampleAccountQueryResolver is account QueryResolver wired to example GraphQL models.
type ExampleAccountQueryResolver = accountgraphql.QueryResolver[
	*domain.User,
	*domain.Account,
	*exmodels.Account,
	*exmodels.AccountPayload,
	*exmodels.AccountCreateInput,
	*exmodels.AccountUpdateInput,
	*exmodels.AccountListFilter,
	*exmodels.AccountListOrder,
	*exmodels.User,
	*exmodels.UserCreateInput,
	*exmodels.UserUpdateInput,
]

// ExampleAccountQueryResolverConfig wires extended account GraphQL models.
type ExampleAccountQueryResolverConfig = accountgraphql.QueryResolverConfig[
	*domain.User,
	*domain.Account,
	*exmodels.Account,
	*exmodels.AccountPayload,
	*exmodels.AccountCreateInput,
	*exmodels.AccountUpdateInput,
	*exmodels.AccountListFilter,
	*exmodels.AccountListOrder,
	*exmodels.User,
	*exmodels.UserCreateInput,
	*exmodels.UserUpdateInput,
]

// NewExampleAccountQueryResolver wires account resolvers with extended GraphQL models.
func NewExampleAccountQueryResolver(cfg ExampleAccountQueryResolverConfig) *ExampleAccountQueryResolver {
	if cfg.AccountsMapper == nil {
		cfg.AccountsMapper = domain.AccountGraphQLMappersImpl{}
	}
	if cfg.UsersMapper == nil {
		cfg.UsersMapper = domain.UserGraphQLMappersImpl{}
	}
	return accountgraphql.NewQueryResolver(cfg)
}
