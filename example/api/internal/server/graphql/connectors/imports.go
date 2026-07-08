package connectors

import (
	authclientgraphql "github.com/geniusrabbit/blaze-api/repository/authclient/delivery/graphql"
	directaccesstokengraphql "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/delivery/graphql"
	historygraphql "github.com/geniusrabbit/blaze-api/repository/historylog/delivery/graphql"
	optiongraphql "github.com/geniusrabbit/blaze-api/repository/option/delivery/graphql"
	rbacgraphql "github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql"
	socaccgraphql "github.com/geniusrabbit/blaze-api/repository/socialaccount/delivery/graphql"
)

type (
	SocialAccountConnection     = socaccgraphql.SocialAccountConnection
	RBACRoleConnection          = rbacgraphql.RBACRoleConnection
	AuthClientConnection        = authclientgraphql.AuthClientConnection
	HistoryActionConnection     = historygraphql.HistoryActionConnection
	OptionConnection            = optiongraphql.OptionConnection
	DirectAccessTokenConnection = directaccesstokengraphql.DirectAccessTokenConnection
)
