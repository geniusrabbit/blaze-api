package connectors

import (
	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	usergraphql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
)

type (
	UserConnection    = usergraphql.UserConnection[exmodels.User]
	AccountConnection = accountgraphql.AccountConnection[exmodels.Account]
	MemberConnection  = accountgraphql.MemberConnection
)
