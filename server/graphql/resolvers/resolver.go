package resolvers

import (
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	account_graphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	accountrepo "github.com/geniusrabbit/blaze-api/repository/account/repository"
	accountusecase "github.com/geniusrabbit/blaze-api/repository/account/usecase"
	authclient_graphql "github.com/geniusrabbit/blaze-api/repository/authclient/delivery/graphql"
	authclientrepo "github.com/geniusrabbit/blaze-api/repository/authclient/repository"
	authclientusecase "github.com/geniusrabbit/blaze-api/repository/authclient/usecase"
	directaccesstoken_graphql "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/delivery/graphql"
	datokenrepo "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/repository"
	datokenusecase "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/usecase"
	historylog_graphql "github.com/geniusrabbit/blaze-api/repository/historylog/delivery/graphql"
	historylogrepo "github.com/geniusrabbit/blaze-api/repository/historylog/repository"
	historylogusecase "github.com/geniusrabbit/blaze-api/repository/historylog/usecase"
	"github.com/geniusrabbit/blaze-api/repository/option"
	option_graphql "github.com/geniusrabbit/blaze-api/repository/option/delivery/graphql"
	rbac_graphql "github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql"
	rbacrepo "github.com/geniusrabbit/blaze-api/repository/rbac/repository"
	rbacusecase "github.com/geniusrabbit/blaze-api/repository/rbac/usecase"
	socialaccount_graphql "github.com/geniusrabbit/blaze-api/repository/socialaccount/delivery/graphql"
	socaccrepo "github.com/geniusrabbit/blaze-api/repository/socialaccount/repository"
	socaccusecase "github.com/geniusrabbit/blaze-api/repository/socialaccount/usecase"
	user_graphql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
	userusecase "github.com/geniusrabbit/blaze-api/repository/user/usecase"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	users             *user_graphql.QueryResolver
	accAuth           *account_graphql.AuthResolver
	accounts          *account_graphql.QueryResolver
	members           *account_graphql.MemberQueryResolver
	socAccounts       *socialaccount_graphql.QueryResolver
	roles             *rbac_graphql.QueryResolver
	authclients       *authclient_graphql.QueryResolver
	historylogs       *historylog_graphql.QueryResolver
	options           *option_graphql.QueryResolver
	directaccesstoken *directaccesstoken_graphql.QueryResolver
}

func NewResolver(provider *jwt.Provider, options option.Usecase) *Resolver {
	userRepoInst := userrepo.NewUserRepository()
	accountRepoInst := accountrepo.NewAccountRepository()
	memberRepoInst := accountrepo.NewMemberRepository()
	rbacRepoInst := rbacrepo.New()

	accountUsecaseInst := accountusecase.NewAccountUsecase(userRepoInst, accountRepoInst, memberRepoInst)
	memberUsecaseInst := accountusecase.NewMemberUsecase(userRepoInst, accountRepoInst, memberRepoInst)

	return &Resolver{
		users:             user_graphql.NewQueryResolver(userusecase.NewUserUsecase(userRepoInst)),
		accAuth:           account_graphql.NewAuthResolver(provider, userRepoInst, accountRepoInst, accountUsecaseInst, rbacRepoInst),
		accounts:          account_graphql.NewQueryResolver(accountUsecaseInst, memberUsecaseInst, userRepoInst),
		members:           account_graphql.NewMemberQueryResolver(accountUsecaseInst, memberUsecaseInst),
		socAccounts:       socialaccount_graphql.NewQueryResolver(socaccusecase.NewSocaccUsecase(socaccrepo.NewSocaccRepository())),
		roles:             rbac_graphql.NewQueryResolver(rbacusecase.New(rbacRepoInst)),
		authclients:       authclient_graphql.NewQueryResolver(authclientusecase.NewAuthclientUsecase(authclientrepo.NewAuthclientRepository())),
		historylogs:       historylog_graphql.NewQueryResolver(historylogusecase.NewUsecase(historylogrepo.New())),
		options:           option_graphql.NewQueryResolver(options),
		directaccesstoken: directaccesstoken_graphql.NewQueryResolver(datokenusecase.New(datokenrepo.NewDirectAccessTokenRepository())),
	}
}
