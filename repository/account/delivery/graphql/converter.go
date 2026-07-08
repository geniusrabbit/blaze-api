package graphql

import (
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// AccountGraphQLConverter maps a consumer account model to the GraphQL Account type.
type AccountGraphQLConverter[TDomain account.Model, TGQL any] func(TDomain) TGQL

// AccountGraphQLListConverter converts account model slices.
func AccountGraphQLListConverter[TDomain account.Model, TGQL any](conv AccountGraphQLConverter[TDomain, TGQL]) func([]TDomain) []TGQL {
	return func(list []TDomain) []TGQL {
		out := make([]TGQL, 0, len(list))
		for _, item := range list {
			out = append(out, conv(item))
		}
		return out
	}
}

// AccountInputMapper fills a consumer account model from GraphQL account input.
type AccountInputMapper[TDomain account.Model, TGQLInput any] func(dest TDomain, input TGQLInput, appStatus ...pkgModels.ApproveStatus) TDomain

// UserGraphQLConverter maps a consumer user model to the GraphQL User type (used by account register payload).
type UserGraphQLConverter[TDomain user.Model, TGQL any] func(TDomain) TGQL

// UserInputMapper builds a domain user from GraphQL user input (registerAccount owner).
type UserInputMapper[TDomain user.Model, TGQLInput any] func(input TGQLInput, appStatus ...pkgModels.ApproveStatus) TDomain

// AccountPayloadFactory builds GraphQL AccountPayload from parts.
type AccountPayloadFactory[TGQLPayload any, TGQLAccount any] func(clientMutationID string, accountID uint64, account TGQLAccount) TGQLPayload

// AccountCreatePayloadFactory builds GraphQL AccountCreatePayload from parts.
type AccountCreatePayloadFactory[TGQLPayload any, TGQLAccount, TGQLUser any] func(clientMutationID string, account TGQLAccount, owner TGQLUser) TGQLPayload

// AccountCreateInputReader extracts registerAccount input fields.
type AccountCreateInputReader[TGQLCreateInput, TGQLAccountInput, TGQLUserInput any] func(input TGQLCreateInput) (ownerID *uint64, owner TGQLUserInput, account TGQLAccountInput, password string)

// AccountFilterMapperFnk converts GraphQL account list filter to domain filter.
type AccountFilterMapperFnk[T any] func(filter T) account.QOption

// AccountOrderMapperFnk converts GraphQL account list order to domain order.
type AccountOrderMapperFnk[T any] func(order T) account.QOption
