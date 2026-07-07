package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AccountConnection implements collection accessor interface with pagination (base schema types).
type AccountConnection[TGQLAccount any] = connectors.CollectionConnection[TGQLAccount, any]

// NewAccountConnection based on query object.
func NewAccountConnection[
	TUser user.Model,
	TDomain account.Model,
	TGQLAccount any,
](
	ctx context.Context,
	accountsAccessor account.Usecase[TUser, TDomain],
	filter account.QOption,
	order []account.QOption,
	page *gqlmodels.Page,
	toGraphQL AccountGraphQLConverter[TDomain, TGQLAccount],
) *AccountConnection[TGQLAccount] {
	toList := AccountGraphQLListConverter(toGraphQL)
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[TGQLAccount, any]{
		FetchDataListFunc: func(ctx context.Context) ([]*TGQLAccount, error) {
			opts := []account.QOption{filter, page.Pagination()}
			opts = append(opts, order...)
			accounts, err := accountsAccessor.FetchList(ctx, opts...)
			if err != nil {
				return nil, err
			}
			return connectors.PtrSlice(toList(accounts)), nil
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return accountsAccessor.Count(ctx, filter)
		},
	}, page)
}

// MemberConnection implements collection accessor interface with pagination.
type MemberConnection = connectors.CollectionConnection[gqlmodels.Member, any]

// NewMemberConnection based on query object.
func NewMemberConnection[TUser user.Model, TDomain account.Model](
	ctx context.Context,
	membersAccessor account.MemberUsecase[TUser, TDomain],
	accounts account.Usecase[TUser, TDomain],
	users user.Repository[TUser],
	filter *gqlmodels.MemberListFilter,
	order []*gqlmodels.MemberListOrder,
	page *gqlmodels.Page,
) *MemberConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[gqlmodels.Member, any]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.Member, error) {
			opts := []account.QOption{FromMemberGQLFilter(filter), page.Pagination()}
			for _, o := range order {
				if ord := FromMemberGQLOrder(o); ord != nil {
					opts = append(opts, ord)
				}
			}
			members, err := membersAccessor.FetchListMembers(ctx, opts...)
			return FromMemberModelList(ctx, members, accounts, users), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return membersAccessor.CountMembers(ctx, FromMemberGQLFilter(filter))
		},
	}, page)
}
