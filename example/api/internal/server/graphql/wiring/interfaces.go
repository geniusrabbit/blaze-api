package wiring

import (
	"context"

	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	accgql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	usergql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// EmailPasswordLoginHandler handles the login(email, password, accountID) mutation.
type EmailPasswordLoginHandler interface {
	Login(ctx context.Context, email, password string, accountID ...uint64) (*gqlmodels.SessionToken, error)
}

// UserQueryHandler is satisfied by ExampleUserQueryResolver and by
// *frameworkResolvers.UserResolverDelegate typed to exmodels filter/order.
type UserQueryHandler interface {
	CurrentUser(ctx context.Context) (*exmodels.UserPayload, error)
	User(ctx context.Context, id uint64, username string) (*exmodels.UserPayload, error)
	CreateUser(ctx context.Context, input *exmodels.UserCreateInput) (*exmodels.UserPayload, error)
	UpdateUser(ctx context.Context, id uint64, input *exmodels.UserUpdateInput) (*exmodels.UserPayload, error)
	ApproveUser(ctx context.Context, id uint64, msg *string) (*exmodels.UserPayload, error)
	RejectUser(ctx context.Context, id uint64, msg *string) (*exmodels.UserPayload, error)
	ResetUserPassword(ctx context.Context, email string) (*gqlmodels.StatusResponse, error)
	UpdateUserPassword(ctx context.Context, token, email, password string) (*gqlmodels.StatusResponse, error)
	ChangeUserPassword(ctx context.Context, currentPassword, newPassword string) (*gqlmodels.StatusResponse, error)
	ListUsers(ctx context.Context, filter *exmodels.UserListFilter, order []*exmodels.UserListOrder, page *gqlmodels.Page) (*usergql.UserConnection[exmodels.User], error)
}

// ExampleAccountQueryHandler covers account CRUD methods as called by generated resolvers
// using extended exmodels types.
type ExampleAccountQueryHandler interface {
	CurrentAccount(ctx context.Context) (*exmodels.AccountPayload, error)
	Account(ctx context.Context, id uint64) (*exmodels.AccountPayload, error)
	RegisterAccount(ctx context.Context, ownerID uint64, input *exmodels.AccountCreateInput) (*exmodels.AccountCreatePayload, error)
	UpdateAccount(ctx context.Context, id uint64, input *exmodels.AccountUpdateInput) (*exmodels.AccountPayload, error)
	ApproveAccount(ctx context.Context, id uint64, msg string) (*exmodels.AccountPayload, error)
	RejectAccount(ctx context.Context, id uint64, msg string) (*exmodels.AccountPayload, error)
	ListAccounts(ctx context.Context, filter *exmodels.AccountListFilter, order []*exmodels.AccountListOrder, page *gqlmodels.Page) (*accgql.AccountConnection[exmodels.Account], error)
}

// ExampleMemberQueryHandler covers member CRUD methods as called by generated resolvers
// using extended exmodels types.
type ExampleMemberQueryHandler interface {
	Invite(ctx context.Context, accountID uint64, member *gqlmodels.InviteMemberInput) (*gqlmodels.MemberPayload, error)
	Update(ctx context.Context, memberID uint64, member *gqlmodels.MemberInput) (*gqlmodels.MemberPayload, error)
	Remove(ctx context.Context, memberID uint64) (*gqlmodels.MemberPayload, error)
	Approve(ctx context.Context, memberID uint64, msg string) (*gqlmodels.MemberPayload, error)
	Reject(ctx context.Context, memberID uint64, msg string) (*gqlmodels.MemberPayload, error)
	List(ctx context.Context, filter *gqlmodels.MemberListFilter, order []*gqlmodels.MemberListOrder, page *gqlmodels.Page) (*accgql.MemberConnection, error)
}
