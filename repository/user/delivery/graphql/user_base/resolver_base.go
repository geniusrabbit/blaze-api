package graphql

import (
	"context"
	"errors"
	"strings"

	"github.com/demdxx/sendmsg"
	"github.com/demdxx/xtypes"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/messanger"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

type FilterMapperFnk[T any] func(filter T) user.QOption
type OrderMapperFnk[T any] func(order T) user.QOption

func userFromSession[T user.Model](ctx context.Context) T {
	var zero T
	if u, ok := any(session.UserModel(ctx)).(T); ok {
		return u
	}
	return zero
}

// QueryResolverBase implements core user GraphQL operations (Model trait only).
type QueryResolverBase[
	TDomain user.Model,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
] struct {
	core   user.Usecase[TDomain]
	mapper graphql.UserGraphQLMappers[TDomain, TGQLUser, TGQLUserInput, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]
}

// QueryResolverBaseConfig wires core user GraphQL resolver.
type QueryResolverBaseConfig[
	TDomain user.Model,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
] struct {
	Core   user.Usecase[TDomain]
	Mapper graphql.UserGraphQLMappers[TDomain, TGQLUser, TGQLUserInput, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]
}

// NewQueryResolverBase returns core user API resolver.
func NewQueryResolverBase[
	TDomain user.Model,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
](cfg QueryResolverBaseConfig[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder] {
	return &QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]{
		core:   cfg.Core,
		mapper: cfg.Mapper,
	}
}

// CurrentUser returns the current user info.
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) CurrentUser(ctx context.Context) (TGQLUserPayload, error) {
	userObj := userFromSession[TDomain](ctx)
	return r.mapper.NewPayload(requestid.Get(ctx), userObj.GetID(), r.mapper.ToGQL(userObj)), nil
}

// UpdateUser is the resolver for the updateUser field.
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) UpdateUser(ctx context.Context, id uint64, input TGQLUserInput) (TGQLUserPayload, error) {
	var zero TGQLUserPayload
	userObj, err := r.core.Get(ctx, id)
	if err != nil {
		return zero, err
	}
	userObj = r.mapper.FromUpdateInput(input, userObj)
	if err := r.core.Update(ctx, userObj); err != nil {
		return zero, err
	}
	return r.mapper.NewPayload(requestid.Get(ctx), userObj.GetID(), r.mapper.ToGQL(userObj)), nil
}

// ApproveUser is the resolver for the approveUser field.
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) ApproveUser(ctx context.Context, id uint64, msg *string) (TGQLUserPayload, error) {
	return r.updateApproveStatus(ctx, id, pkgModels.ApprovedApproveStatus, msg)
}

// RejectUser is the resolver for the rejectUser field.
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) RejectUser(ctx context.Context, id uint64, msg *string) (TGQLUserPayload, error) {
	return r.updateApproveStatus(ctx, id, pkgModels.DisapprovedApproveStatus, msg)
}

func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) updateApproveStatus(ctx context.Context, id uint64, status pkgModels.ApproveStatus, msg *string) (TGQLUserPayload, error) {
	var zero TGQLUserPayload
	userObj, err := r.core.Get(ctx, id)
	if err != nil {
		return zero, err
	}
	if setter, ok := any(userObj).(interface{ SetApprove(pkgModels.ApproveStatus) }); ok {
		setter.SetApprove(status)
	}
	if msg != nil {
		ctx = historylog.WithMessage(ctx, *msg)
	}
	if err = r.core.Update(ctx, userObj); err != nil {
		return zero, err
	}
	msgName := "user." + strings.ToLower(status.String())
	err = messanger.Get(ctx).Send(ctx, msgName, []string{}, map[string]any{})
	if err != nil && !errors.Is(err, sendmsg.ErrTemplateNotFound) {
		ctxlogger.Get(ctx).Error("User status update",
			zap.String("msgname", msgName),
			zap.Error(err))
	}
	return r.mapper.NewPayload(requestid.Get(ctx), id, r.mapper.ToGQL(userObj)), nil
}

// ListUsers list by filter.
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) ListUsers(
	ctx context.Context,
	filter TGQLUserListFilter,
	order []TGQLUserListOrder,
	page *gqlmodels.Page,
) (*graphql.UserConnection[TGQLUser], error) {
	return graphql.NewUserConnection(
		ctx,
		r.core,
		r.mapper.FromFilter(filter),
		xtypes.SliceApply(order, r.mapper.FromOrder),
		page,
		r.mapper.ToGQL,
	), nil
}

// UserFromInput builds domain user from GraphQL input (used by password resolver).
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) UserFromInput(input TGQLUserInput) TDomain {
	return r.mapper.FromCreateInput(input)
}

// ToGraphQL converts domain user to GraphQL type.
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) ToGraphQL(userObj TDomain) TGQLUser {
	return r.mapper.ToGQL(userObj)
}

// NewUserPayload builds GraphQL payload.
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) NewUserPayload(ctx context.Context, userID uint64, userObj TDomain) TGQLUserPayload {
	return r.mapper.NewPayload(requestid.Get(ctx), userID, r.mapper.ToGQL(userObj))
}

// Core returns core usecase (for email resolver ID lookup).
func (r *QueryResolverBase[TDomain, TGQLUser, TGQLUserInput, TGQLUserPayload, TGQLUserListFilter, TGQLUserListOrder]) Core() user.Usecase[TDomain] {
	return r.core
}
