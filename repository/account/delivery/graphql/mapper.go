package graphql

import "github.com/geniusrabbit/blaze-api/repository/account"

// AccountGraphQLMappers is the contract a consumer implements to map between
// a domain account model and the GraphQL protocol types.
//
// Type parameters:
//   - T               — consumer domain account model (implements account.Model)
//   - TGQLAccount     — GraphQL Account type (base or extended via extend type)
//   - TGQLCreateInput — GraphQL input type for create mutations
//   - TGQLUpdateInput — GraphQL input type for update mutations (may equal TGQLCreateInput)
//   - TGQLFilter      — GraphQL filter type for list queries
//   - TGQLOrder       — GraphQL order type for list queries
//
// Usage — consumer implements as a plain struct:
//
//	type AccountMapper struct{}
//	func (m AccountMapper) ToGQL(a *domain.Account) exmodels.Account { ... }
//	func (m AccountMapper) FromCreateInput(inp *exmodels.AccountCreateInput) *domain.Account { ... }
//	func (m AccountMapper) FromUpdateInput(inp *exmodels.AccountInput, dest *domain.Account) *domain.Account { ... }
//	func (m AccountMapper) FromFilter(f *exmodels.AccountListFilter) account.QOption { ... }
//	func (m AccountMapper) FromOrder(o *exmodels.AccountListOrder) account.QOption { ... }
//
// NOTE: Go generics do not allow storing a parameterised interface as a plain
// variable.  Pass the mapper as a concrete type parameter C constrained to
// AccountGraphQLMappers[T, ...] on the containing struct or function.
type AccountGraphQLMappers[
	T account.Model,
	TGQLAccount any,
	TGQLPayload any,
	TGQLCreateInput any,
	TGQLUpdateInput any,
	TGQLFilter any,
	TGQLOrder any,
	TGQLUser any,
] interface {
	// New creates a new empty domain account model.
	New() T
	// ToGQL converts a domain account to the GraphQL representation.
	ToGQL(T) TGQLAccount
	// NewPayload wraps a GQL account in the mutation payload type.
	NewPayload(clientMutationID string, accountID uint64, account TGQLAccount) TGQLPayload
	// FromCreateInput builds a new domain account from a create mutation input.
	FromCreateInput(TGQLCreateInput) T
	// FromUpdateInput merges an update mutation input into an existing domain account.
	// dest is the existing model loaded from the repository; must return dest.
	FromUpdateInput(TGQLUpdateInput, T) T
	// FromFilter converts a GraphQL list filter to a domain QOption.
	FromFilter(TGQLFilter) account.QOption
	// FromOrder converts a GraphQL list order to a domain QOption.
	FromOrder(TGQLOrder) account.QOption
}

// FuncAccountMapper implements AccountGraphQLMappers using plain function fields.
// Useful when building a mapper inline without creating a named struct type.
// Nil function fields return zero values instead of panicking.
type FuncAccountMapper[
	T account.Model,
	TGQLAccount any,
	TGQLPayload any,
	TGQLCreateInput any,
	TGQLUpdateInput any,
	TGQLFilter any,
	TGQLOrder any,
	TGQLUser any,
] struct {
	NewFn        func() T
	ToGQLFn      func(T) TGQLAccount
	NewPayloadFn func(clientMutationID string, accountID uint64, account TGQLAccount) TGQLPayload
	FromCreateFn func(TGQLCreateInput) T
	FromUpdateFn func(TGQLUpdateInput, T) T
	FromFilterFn func(TGQLFilter) account.QOption
	FromOrderFn  func(TGQLOrder) account.QOption
}

func (m FuncAccountMapper[T, TGQLAccount, TGQLPayload, TGQLCreateInput, TGQLUpdateInput, TGQLFilter, TGQLOrder, TGQLUser]) New() T {
	return m.NewFn()
}

func (m FuncAccountMapper[T, TGQLAccount, TGQLPayload, TGQLCreateInput, TGQLUpdateInput, TGQLFilter, TGQLOrder, TGQLUser]) ToGQL(a T) TGQLAccount {
	return m.ToGQLFn(a)
}

func (m FuncAccountMapper[T, TGQLAccount, TGQLPayload, TGQLCreateInput, TGQLUpdateInput, TGQLFilter, TGQLOrder, TGQLUser]) NewPayload(clientMutationID string, accountID uint64, account TGQLAccount) TGQLPayload {
	return m.NewPayloadFn(clientMutationID, accountID, account)
}

func (m FuncAccountMapper[T, TGQLAccount, TGQLPayload, TGQLCreateInput, TGQLUpdateInput, TGQLFilter, TGQLOrder, TGQLUser]) FromCreateInput(inp TGQLCreateInput) T {
	return m.FromCreateFn(inp)
}

func (m FuncAccountMapper[T, TGQLAccount, TGQLPayload, TGQLCreateInput, TGQLUpdateInput, TGQLFilter, TGQLOrder, TGQLUser]) FromUpdateInput(inp TGQLUpdateInput, dest T) T {
	return m.FromUpdateFn(inp, dest)
}

func (m FuncAccountMapper[T, TGQLAccount, TGQLPayload, TGQLCreateInput, TGQLUpdateInput, TGQLFilter, TGQLOrder, TGQLUser]) FromFilter(f TGQLFilter) account.QOption {
	if m.FromFilterFn != nil {
		return m.FromFilterFn(f)
	}
	type convertable interface {
		ConvertFilter() account.QOption
	}
	if conv, ok := any(f).(convertable); ok {
		return conv.ConvertFilter()
	}
	panic("FromFilterFn is nil and TGQLFilter does not implement ConvertFilter()")
}

func (m FuncAccountMapper[T, TGQLAccount, TGQLPayload, TGQLCreateInput, TGQLUpdateInput, TGQLFilter, TGQLOrder, TGQLUser]) FromOrder(o TGQLOrder) account.QOption {
	if m.FromOrderFn != nil {
		return m.FromOrderFn(o)
	}
	type convertable interface {
		ConvertOrder() account.QOption
	}
	if conv, ok := any(o).(convertable); ok {
		return conv.ConvertOrder()
	}
	panic("FromOrderFn is nil and TGQLOrder does not implement ConvertOrder()")
}
