package graphql

import (
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// UserGraphQLConverter maps a consumer user model to the GraphQL User type.
type UserGraphQLConverter[TDomain user.Model, TGQL any] func(TDomain) TGQL

// UserGraphQLListConverter converts user model slices.
func UserGraphQLListConverter[TDomain user.Model, TGQL any](conv UserGraphQLConverter[TDomain, TGQL]) func([]TDomain) []TGQL {
	return func(list []TDomain) []TGQL {
		out := make([]TGQL, 0, len(list))
		for _, item := range list {
			out = append(out, conv(item))
		}
		return out
	}
}

// UserInputMapper fills a consumer user model from GraphQL user input.
type UserInputMapper[TDomain user.Model, TGQLInput any] func(input TGQLInput, appStatus ...pkgModels.ApproveStatus) TDomain

// UserPayloadFactory builds GraphQL UserPayload from parts.
type UserPayloadFactory[TGQLPayload any, TGQLUser any] func(clientMutationID string, userID uint64, user TGQLUser) TGQLPayload

// UserListFilterMapper converts GraphQL user list filter to domain filter.
type UserListFilterMapper interface {
	Filter() user.QOption
}

// UserListOrderMapper converts GraphQL user list order to domain order.
type UserListOrderMapper interface {
	Order() user.QOption
}

// UserEdgeBuilder builds a GraphQL UserEdge from a cursor and node.
type UserEdgeBuilder[TGQLUser any, TGQLUserEdge any] func(node TGQLUser) *TGQLUserEdge
