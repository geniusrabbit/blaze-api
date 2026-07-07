package graphql

import (
	"github.com/geniusrabbit/blaze-api/repository/user"
)

type UserGraphQLMappersBase[T user.Model, TGQLUser any, TGQLCreateInput any, TGQLUpdateInput any] interface {
	// New creates a new empty domain user model.
	New() T
	// ToGQL converts a domain user to the GraphQL representation.
	ToGQL(T) TGQLUser
	// FromCreateInput builds a new domain user from a create mutation input.
	FromCreateInput(TGQLCreateInput) T
	// FromUpdateInput merges an update mutation input into an existing domain user.
	// dest is the existing model loaded from the repository; must return dest.
	FromUpdateInput(TGQLUpdateInput, T) T
}

// FuncUserMapperBase implements UserGraphQLMappersBase using plain function fields.
// Useful for wiring the base mapper inline without creating a named struct.
// Nil function fields return zero values instead of panicking.
type FuncUserMapperBase[T user.Model, TGQLUser any, TGQLCreateInput any, TGQLUpdateInput any] struct {
	NewFn        func() T
	ToGQLFn      func(T) TGQLUser
	FromCreateFn func(TGQLCreateInput) T
	FromUpdateFn func(TGQLUpdateInput, T) T
}

func (m FuncUserMapperBase[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput]) New() T {
	if m.NewFn != nil {
		return m.NewFn()
	}
	var zero T
	return zero
}

func (m FuncUserMapperBase[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput]) ToGQL(u T) TGQLUser {
	if m.ToGQLFn != nil {
		return m.ToGQLFn(u)
	}
	var zero TGQLUser
	return zero
}

func (m FuncUserMapperBase[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput]) FromCreateInput(inp TGQLCreateInput) T {
	if m.FromCreateFn != nil {
		return m.FromCreateFn(inp)
	}
	var zero T
	return zero
}

func (m FuncUserMapperBase[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput]) FromUpdateInput(inp TGQLUpdateInput, dest T) T {
	if m.FromUpdateFn != nil {
		return m.FromUpdateFn(inp, dest)
	}
	return dest
}

// UserGraphQLMappers is the contract a consumer implements to map between
// a domain user model and the GraphQL protocol types.
//
// Type parameters:
//   - T               — consumer domain user model (implements user.Model)
//   - TGQLUser        — GraphQL User type (base or extended via extend type)
//   - TGQLCreateInput — GraphQL input type for create mutations
//   - TGQLUpdateInput — GraphQL input type for update mutations (may equal TGQLCreateInput)
//   - TGQLPayload     — GraphQL payload type wrapping TGQLUser (e.g. *gqlmodels.UserPayload)
//   - TFilter         — GraphQL filter type for list queries
//   - TOrder          — GraphQL order type for list queries
//
// Consumers implement this as a plain struct:
//
//	type UserMapper struct{ CDNBase string }
//	func (m UserMapper) ToGQL(u *domain.User) gqlmodels.User { ... }
//	func (m UserMapper) FromCreateInput(inp *gqlmodels.UserInput) *domain.User { ... }
//	func (m UserMapper) FromUpdateInput(inp *gqlmodels.UserInput, dest *domain.User) *domain.User { ... }
//	func (m UserMapper) NewPayload(id string, uid uint64, u gqlmodels.User) *gqlmodels.UserPayload { ... }{ ... }
//	func (m UserMapper) FromFilter(f *exmodels.UserListFilter) user.QOption { ... }
//	func (m UserMapper) FromOrder(o *exmodels.UserListOrder) user.QOption { ... }
//
// NOTE: Go generics do not allow storing a parameterised interface as a plain
// variable.  Pass the mapper as a concrete type parameter C constrained to
// UserGraphQLMappers[T, ...] on the containing struct or function.
type UserGraphQLMappers[
	T user.Model,
	TGQLUser any,
	TGQLCreateInput any,
	TGQLUpdateInput any,
	TGQLPayload any,
	TFilter any,
	TOrder any,
] interface {
	UserGraphQLMappersBase[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput]
	// NewPayload wraps a GQL user in the mutation payload type.
	NewPayload(clientMutationID string, userID uint64, u TGQLUser) TGQLPayload
	// FromFilter converts a GraphQL list filter to a domain QOption.
	FromFilter(TFilter) user.QOption
	// FromOrder converts a GraphQL list order to a domain QOption.
	FromOrder(TOrder) user.QOption
}

// FuncUserMapper implements UserGraphQLMappers using plain function fields.
// Useful when building a mapper inline without creating a named struct type.
// Nil function fields return zero values instead of panicking.
type FuncUserMapper[
	T user.Model,
	TGQLUser any,
	TGQLCreateInput any,
	TGQLUpdateInput any,
	TGQLPayload any,
	TFilter any,
	TOrder any,
] struct {
	NewFn        func() T
	ToGQLFn      func(T) TGQLUser
	FromCreateFn func(TGQLCreateInput) T
	FromUpdateFn func(TGQLUpdateInput, T) T
	NewPayloadFn func(clientMutationID string, userID uint64, u TGQLUser) TGQLPayload
	FromFilterFn func(TFilter) user.QOption
	FromOrderFn  func(TOrder) user.QOption
}

func (m FuncUserMapper[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput, TGQLPayload, TFilter, TOrder]) New() T {
	if m.NewFn != nil {
		return m.NewFn()
	}
	var zero T
	return zero
}

func (m FuncUserMapper[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput, TGQLPayload, TFilter, TOrder]) ToGQL(u T) TGQLUser {
	if m.ToGQLFn != nil {
		return m.ToGQLFn(u)
	}
	var zero TGQLUser
	return zero
}

func (m FuncUserMapper[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput, TGQLPayload, TFilter, TOrder]) FromCreateInput(inp TGQLCreateInput) T {
	if m.FromCreateFn != nil {
		return m.FromCreateFn(inp)
	}
	var zero T
	return zero
}

func (m FuncUserMapper[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput, TGQLPayload, TFilter, TOrder]) FromUpdateInput(inp TGQLUpdateInput, dest T) T {
	if m.FromUpdateFn != nil {
		return m.FromUpdateFn(inp, dest)
	}
	return dest
}

func (m FuncUserMapper[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput, TGQLPayload, TFilter, TOrder]) NewPayload(clientMutationID string, userID uint64, u TGQLUser) TGQLPayload {
	if m.NewPayloadFn != nil {
		return m.NewPayloadFn(clientMutationID, userID, u)
	}
	var zero TGQLPayload
	return zero
}

func (m FuncUserMapper[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput, TGQLPayload, TFilter, TOrder]) FromFilter(f TFilter) user.QOption {
	if m.FromFilterFn != nil {
		return m.FromFilterFn(f)
	}
	return nil
}

func (m FuncUserMapper[T, TGQLUser, TGQLCreateInput, TGQLUpdateInput, TGQLPayload, TFilter, TOrder]) FromOrder(o TOrder) user.QOption {
	if m.FromOrderFn != nil {
		return m.FromOrderFn(o)
	}
	return nil
}
