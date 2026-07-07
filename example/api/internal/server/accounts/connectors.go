package accounts

import (
	accgql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	usrgql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

type (
	UserConnection    = usrgql.UserConnection[gqlmodels.User, gqlmodels.UserEdge]
	AccountConnection = accgql.AccountConnection[gqlmodels.Account, gqlmodels.AccountEdge]
)
