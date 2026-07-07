package domain

import (
	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	accountrepo "github.com/geniusrabbit/blaze-api/repository/account"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
)

// AccountGraphQLMappers is the 2-param consumer alias for example/api.
// It pins the 4 concrete types (create input, update input, filter, order)
// so callers need only supply T (domain account) and TGQLAccount (GraphQL account type).
type AccountGraphQLMappers[T accountrepo.Model, TGQLAccount any] = accountgraphql.AccountGraphQLMappers[
	T,
	TGQLAccount,
	*exmodels.AccountPayload,     // TGQLPayload
	*exmodels.AccountCreateInput, // TGQLCreateInput
	*exmodels.AccountUpdateInput, // TGQLUpdateInput (status-only update in base schema)
	*exmodels.AccountListFilter,  // TFilter
	*exmodels.AccountListOrder,   // TOrder
	*exmodels.User,               // TGQLUser
]

// AccountGraphQLMappersImpl implements AccountGraphQLMappers for example/api.
// All methods delegate to the free functions in domain/graphql.go.
// Instantiate with AccountGraphQLMappersImpl{} — stateless.
type AccountGraphQLMappersImpl struct{}

// Compile-time assertion: AccountGraphQLMappersImpl satisfies the 2-param alias when T = *Account.
var _ AccountGraphQLMappers[*Account, *exmodels.Account] = AccountGraphQLMappersImpl{}

// New creates a new empty domain Account.
func (AccountGraphQLMappersImpl) New() *Account {
	return new(Account)
}

// ToGQL maps a domain Account to the extended GraphQL Account model.
func (AccountGraphQLMappersImpl) ToGQL(a *Account) *exmodels.Account {
	return AccountToGraphQL(a)
}

// NewPayload builds account payload from parts.
func (AccountGraphQLMappersImpl) NewPayload(clientMutationID string, accountID uint64, account *exmodels.Account) *exmodels.AccountPayload {
	acc := account
	return &exmodels.AccountPayload{
		ClientMutationID: clientMutationID,
		AccountID:        accountID,
		Account:          acc,
	}
}

// FromCreateInput builds a new domain Account from a create-account account input.
func (AccountGraphQLMappersImpl) FromCreateInput(inp *exmodels.AccountCreateInput) *Account {
	if inp == nil {
		return new(Account)
	}
	return FillAccountFromCreateInput(new(Account), inp)
}

// FromUpdateInput merges an update mutation input into an existing domain Account.
// AccountUpdateInput carries only the approval status; profile edits use a separate mutation.
func (AccountGraphQLMappersImpl) FromUpdateInput(inp *exmodels.AccountUpdateInput, dest *Account) *Account {
	if inp == nil || dest == nil {
		return dest
	}
	if inp.Status != nil {
		dest.SetApprove(inp.Status.ModelStatus())
	}
	return dest
}

// FromFilter converts the extended GraphQL account list filter to a domain QOption.
func (AccountGraphQLMappersImpl) FromFilter(f *exmodels.AccountListFilter) accountrepo.QOption {
	if f == nil {
		return nil
	}
	return f.Filter()
}

// FromOrder converts the extended GraphQL account list order to a domain QOption.
func (AccountGraphQLMappersImpl) FromOrder(o *exmodels.AccountListOrder) accountrepo.QOption {
	if o == nil {
		return nil
	}
	return o.Order()
}

// FillAccountFromInputWithStatus is a helper for wiring account repos that need an approve-status override.
// Used internally by account graphql wiring.
func FillAccountFromInputWithStatus(dest *Account, inp *exmodels.AccountUpdateInput, appStatus ...pkgModels.ApproveStatus) *Account {
	return FillAccountFromUpdateInput(dest, inp, appStatus...)
}
