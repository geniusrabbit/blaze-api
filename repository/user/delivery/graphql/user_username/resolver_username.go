package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/user"
	usergraphql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// UsernameModel is the constraint for user models that carry a separate username.
type UsernameModel interface {
	user.Model
	GetUsername() string
}

// QueryResolverUsername resolves the username field on User (Username trait).
type QueryResolverUsername[
	TDomain UsernameModel,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
] struct {
	core      user.Usecase[TDomain]
	toGraphQL usergraphql.UserGraphQLConverter[TDomain, TGQLUser]
}

// QueryResolverUsernameConfig wires the username resolver.
type QueryResolverUsernameConfig[
	TDomain UsernameModel,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
] struct {
	Core      user.Usecase[TDomain]
	ToGraphQL usergraphql.UserGraphQLConverter[TDomain, TGQLUser]
}

// NewQueryResolverUsername returns the username trait resolver.
func NewQueryResolverUsername[
	TDomain UsernameModel,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
](cfg QueryResolverUsernameConfig[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) *QueryResolverUsername[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder] {
	return &QueryResolverUsername[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]{
		core:      cfg.Core,
		toGraphQL: cfg.ToGraphQL,
	}
}

// UpdateUserUsername sets the username on an existing user and persists it.
func (r *QueryResolverUsername[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) UpdateUserUsername(ctx context.Context, id uint64, input TGQLUserInput) error {
	userObj, err := r.core.Get(ctx, id)
	if err != nil {
		return err
	}
	if inp, ok := any(input).(*gqlmodels.UserInput); ok && inp.Username != nil {
		if setter, ok := any(userObj).(interface{ SetUsername(string) }); ok {
			setter.SetUsername(*inp.Username)
		}
	}
	return r.core.Update(ctx, userObj)
}
