package wiring

import (
	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// ExampleAccountQueryResolver is account QueryResolver wired to example GraphQL models.
type ExampleAccountQueryResolver[TUser user.AuthCapableModel, TDomain account.Model] = accountgraphql.QueryResolver[
	TUser,
	TDomain,
	*exmodels.Account,
	*exmodels.AccountPayload,
	*exmodels.AccountCreateInput,
	*exmodels.AccountCreatePayload,
	*exmodels.AccountUpdateInput,
	*exmodels.AccountListFilter,
	*exmodels.AccountListOrder,
	*exmodels.User,
	*exmodels.UserCreateInput,
	*exmodels.UserUpdateInput,
]

// ExampleAccountQueryResolverConfig wires extended account GraphQL models.
type ExampleAccountQueryResolverConfig[TUser user.AuthCapableModel, TDomain account.Model] = accountgraphql.QueryResolverConfig[
	TUser,
	TDomain,
	*exmodels.Account,
	*exmodels.AccountPayload,
	*exmodels.AccountCreateInput,
	*exmodels.AccountCreatePayload,
	*exmodels.AccountUpdateInput,
	*exmodels.AccountListFilter,
	*exmodels.AccountListOrder,
	*exmodels.User,
	*exmodels.UserCreateInput,
	*exmodels.UserUpdateInput,
]

// exampleAccountResolverMapper implements AccountGraphQLMappers for the example wiring context
// where TGQLCreateInput = TGQLUpdateInput = *exmodels.AccountUpdateInput.
// It delegates to domain.AccountGraphQLMappersImpl via type assertions.
type exampleAccountResolverMapper struct{}

func (exampleAccountResolverMapper) New() *domain.Account {
	return &domain.Account{}
}

func (exampleAccountResolverMapper) ToGQL(a *domain.Account) *exmodels.Account {
	if typed, ok := any(a).(*domain.Account); ok {
		return domain.AccountToGraphQL(typed)
	}
	return nil
}

func (exampleAccountResolverMapper) NewPayload(clientMutationID string, accountID uint64, account exmodels.Account) *exmodels.AccountPayload {
	acc := account
	return &exmodels.AccountPayload{
		ClientMutationID: clientMutationID,
		AccountID:        accountID,
		Account:          &acc,
	}
}

func (exampleAccountResolverMapper) NewCreatePayload(clientMutationID string, account *exmodels.Account, owner *gqlmodels.User) *exmodels.AccountCreatePayload {
	acc := account
	return &exmodels.AccountCreatePayload{
		ClientMutationID: clientMutationID,
		Account:          acc,
		Owner:            owner,
	}
}

func (exampleAccountResolverMapper) FromCreateInput(inp *exmodels.AccountCreateInput) *domain.Account {
	updated := domain.AccountGraphQLMappersImpl{}.FromCreateInput(inp)
	return updated
}

func (exampleAccountResolverMapper) FromUpdateInput(inp *exmodels.AccountUpdateInput, dest *domain.Account) *domain.Account {
	return domain.AccountGraphQLMappersImpl{}.FromUpdateInput(inp, dest)
}

func (exampleAccountResolverMapper) FromFilter(f *exmodels.AccountListFilter) account.QOption {
	return domain.AccountGraphQLMappersImpl{}.FromFilter(f)
}

func (exampleAccountResolverMapper) FromOrder(o *exmodels.AccountListOrder) account.QOption {
	return domain.AccountGraphQLMappersImpl{}.FromOrder(o)
}

// NewExampleAccountQueryResolver wires account resolvers with extended GraphQL models.
func NewExampleAccountQueryResolver[TUser user.AuthCapableModel, TDomain account.Model](
	cfg ExampleAccountQueryResolverConfig[*domain.User, *domain.Account],
) *ExampleAccountQueryResolver[TUser, TDomain] {
	if cfg.AccountsMapper == nil {
		cfg.AccountsMapper = exampleAccountResolverMapper{}
	}
	if cfg.UsersMapper == nil {
		cfg.UsersMapper = domain.UserGraphQLMappersImpl{}
	}
	return accountgraphql.NewQueryResolver(cfg)
}
