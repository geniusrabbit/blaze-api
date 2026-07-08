package graphql

import (
	"context"
	"errors"

	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	userusecase "github.com/geniusrabbit/blaze-api/repository/user/usecase"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// QueryResolverPassword implements password-related user GraphQL mutations.
type QueryResolverPassword[
	TDomain user.PasswordCapableModel,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
] struct {
	core          user.Usecase[TDomain]
	password      user.PasswordUsecase[TDomain]
	userFromInput graphql.UserInputMapper[TDomain, TGQLUserInput]
	newPayload    graphql.UserPayloadFactory[TGQLUserPayload, TGQLUser]
	toGraphQL     graphql.UserGraphQLConverter[TDomain, TGQLUser]
}

// QueryResolverPasswordConfig wires password resolver.
type QueryResolverPasswordConfig[
	TDomain user.PasswordCapableModel,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
] struct {
	Core          user.Usecase[TDomain]
	Password      user.PasswordUsecase[TDomain]
	UserFromInput graphql.UserInputMapper[TDomain, TGQLUserInput]
	NewPayload    graphql.UserPayloadFactory[TGQLUserPayload, TGQLUser]
	ToGraphQL     graphql.UserGraphQLConverter[TDomain, TGQLUser]
}

// NewQueryResolverPassword returns password resolver.
func NewQueryResolverPassword[
	TDomain user.PasswordCapableModel,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
](cfg QueryResolverPasswordConfig[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) *QueryResolverPassword[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder] {
	return &QueryResolverPassword[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]{
		core:          cfg.Core,
		password:      cfg.Password,
		userFromInput: cfg.UserFromInput,
		newPayload:    cfg.NewPayload,
		toGraphQL:     cfg.ToGraphQL,
	}
}

// CreateUser creates user with password via PasswordRepository.
func (r *QueryResolverPassword[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) CreateUser(ctx context.Context, input TGQLUserInput) (TGQLUserPayload, error) {
	var zero TGQLUserPayload
	userObj := r.userFromInput(input)
	uid, err := r.password.Repo().CreateWithPassword(ctx, userObj, "GQL create user")
	if err != nil {
		return zero, err
	}
	userObj, err = r.core.Get(ctx, uid)
	if err != nil {
		return zero, err
	}
	return r.newPayload(requestid.Get(ctx), userObj.GetID(), r.toGraphQL(userObj)), nil
}

// ChangeUserPassword is the resolver for the changeUserPassword field.
func (r *QueryResolverPassword[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) ChangeUserPassword(ctx context.Context, currentPassword, newPassword string) (*gqlmodels.StatusResponse, error) {
	err := r.password.ChangePassword(ctx, currentPassword, newPassword)
	if err != nil {
		switch {
		case errors.Is(err, userusecase.ErrInvalidCurrentPassword):
			return nil, userusecase.ErrInvalidCurrentPassword
		case errors.Is(err, userusecase.ErrPasswordTooShort):
			return nil, userusecase.ErrPasswordTooShort
		default:
			return nil, err
		}
	}
	return &gqlmodels.StatusResponse{
		ClientMutationID: requestid.Get(ctx),
		Status:           gqlmodels.ResponseStatusSuccess,
		Message:          &[]string{"Password changed successfully"}[0],
	}, nil
}
