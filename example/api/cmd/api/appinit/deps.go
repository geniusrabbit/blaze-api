package appinit

import (
	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	userstack "github.com/geniusrabbit/blaze-api/example/api/internal/user"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	accountrepo "github.com/geniusrabbit/blaze-api/repository/account/repository"
	accountuc "github.com/geniusrabbit/blaze-api/repository/account/usecase"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// UserType is the example consumer user model.
type UserType = domain.User

// AccountType is the example domain account model.
type AccountType = domain.Account

// AccountMemberType is the example domain member model.
type AccountMemberType = domain.AccountMember

// GraphQLDeps holds account GraphQL wiring dependencies for example/api.
type GraphQLDeps struct {
	UserRepo    user.Repository[*UserType]
	AccountRepo account.SessionRepository[*UserType, *AccountType]
	MemberRepo  account.MemberRepository[*UserType, *AccountType]
	AccountUC   account.Usecase[*UserType, *AccountType]
	MemberUC    account.MemberUsecase[*UserType, *AccountType]
	ToGraphQL   accountgraphql.AccountGraphQLConverter[*AccountType, *exmodels.Account]
	FromInput   accountgraphql.AccountInputMapper[*AccountType, *exmodels.AccountUpdateInput]
}

// Deps holds typed dependencies for example/api.
type Deps struct {
	UserModule  userstack.Module[*UserType]
	AccountRepo account.SessionRepository[*UserType, *AccountType]
	MemberRepo  account.MemberRepository[*UserType, *AccountType]
	AccountUC   account.Usecase[*UserType, *AccountType]
	MemberUC    account.MemberUsecase[*UserType, *AccountType]
	AuthLoader  *accauth.Loader[*UserType, *AccountType]
	GraphQL     GraphQLDeps
}

// NewDeps wires example User/Account repository and usecase stack.
func NewDeps() *Deps {
	newUser := func() *UserType { return new(UserType) }
	newAccount := func() *AccountType { return new(AccountType) }
	newMember := func() *AccountMemberType { return new(AccountMemberType) }
	userModule := userstack.NewModule(newUser)
	accountRepo := accountrepo.NewSessionRepository(newUser, newAccount, newMember)
	memberRepo := accountrepo.NewMemberRepositoryFor(newMember)
	accountUC := accountuc.NewAccountUsecase(userModule.Repo, accountRepo, memberRepo)
	memberUC := accountuc.NewMemberUsecase(userModule.Repo, accountRepo, memberRepo)
	authLoader := accauth.NewLoader(userModule.Repo, accountRepo, memberRepo)
	graphqlDeps := GraphQLDeps{
		UserRepo:    userModule.Repo,
		AccountRepo: accountRepo,
		MemberRepo:  memberRepo,
		AccountUC:   accountUC,
		MemberUC:    memberUC,
		ToGraphQL:   domain.AccountToGraphQL,
		FromInput:   domain.FillAccountFromUpdateInput,
	}
	return &Deps{
		UserModule:  userModule,
		AccountRepo: accountRepo,
		MemberRepo:  memberRepo,
		AccountUC:   accountUC,
		MemberUC:    memberUC,
		AuthLoader:  authLoader,
		GraphQL:     graphqlDeps,
	}
}
