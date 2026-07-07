package graphql

import (
	"context"
	"errors"

	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
)

var (
	errInvalidIDOrEmail = errors.New("invalid id or email")
)

// QueryResolverEmail implements user lookup by email/username.
type QueryResolverEmail[
	TDomain user.EmailCapableModel,
	TGQLUser any,
	TGQLUserPayload any,
] struct {
	core       user.Usecase[TDomain]
	email      user.EmailUsecase[TDomain]
	toGraphQL  graphql.UserGraphQLConverter[TDomain, TGQLUser]
	newPayload graphql.UserPayloadFactory[TGQLUserPayload, TGQLUser]
}

// QueryResolverEmailConfig wires email lookup resolver.
type QueryResolverEmailConfig[
	TDomain user.EmailCapableModel,
	TGQLUser any,
	TGQLUserPayload any,
] struct {
	Core       user.Usecase[TDomain]
	Email      user.EmailUsecase[TDomain]
	ToGraphQL  graphql.UserGraphQLConverter[TDomain, TGQLUser]
	NewPayload graphql.UserPayloadFactory[TGQLUserPayload, TGQLUser]
}

// NewQueryResolverEmail returns email lookup resolver.
func NewQueryResolverEmail[
	TDomain user.EmailCapableModel,
	TGQLUser any,
	TGQLUserPayload any,
](cfg QueryResolverEmailConfig[TDomain, TGQLUser, TGQLUserPayload]) *QueryResolverEmail[TDomain, TGQLUser, TGQLUserPayload] {
	return &QueryResolverEmail[TDomain, TGQLUser, TGQLUserPayload]{
		core:       cfg.Core,
		email:      cfg.Email,
		toGraphQL:  cfg.ToGraphQL,
		newPayload: cfg.NewPayload,
	}
}

// User resolves user by ID or username (email).
func (r *QueryResolverEmail[TDomain, TGQLUser, TGQLUserPayload]) User(ctx context.Context, id uint64, email string) (TGQLUserPayload, error) {
	var (
		err     error
		userObj TDomain
	)
	switch {
	case id > 0:
		userObj, err = r.core.Get(ctx, id)
		if err == nil && email != "" && email != userObj.GetEmail() {
			err = errInvalidIDOrEmail
		}
	case email != "":
		userObj, err = r.email.GetByEmail(ctx, email)
		if err == nil && id > 0 && id != userObj.GetID() {
			err = errInvalidIDOrEmail
		}
	default:
		err = errInvalidIDOrEmail
	}
	var zero TGQLUserPayload
	if err != nil {
		return zero, err
	}
	return r.newPayload(requestid.Get(ctx), userObj.GetID(), r.toGraphQL(userObj)), nil
}
