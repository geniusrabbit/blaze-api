package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// UserQueryHandler defines the interface for user GraphQL operations.
type UserBaseQueryResolver[
	TDomain user.Model,
	TGQLUser any,
	TGQLUserCreateInput any,
	TGQLUserUpdateInput any,
	TGQLUserPayload any,
	TGQLUserListFilter any,
	TGQLUserListOrder any,
] interface {
	CurrentUser(ctx context.Context) (TGQLUserPayload, error)
	UpdateUser(ctx context.Context, id uint64, input TGQLUserUpdateInput) (TGQLUserPayload, error)
	ApproveUser(ctx context.Context, id uint64, msg *string) (TGQLUserPayload, error)
	RejectUser(ctx context.Context, id uint64, msg *string) (TGQLUserPayload, error)
	ListUsers(ctx context.Context, filter TGQLUserListFilter, order []TGQLUserListOrder, page *gqlmodels.Page) (*UserConnection[TGQLUser], error)
	UserFromInput(input TGQLUserCreateInput) TDomain
	ToGraphQL(userObj TDomain) TGQLUser
	NewUserPayload(ctx context.Context, userID uint64, userObj TDomain) TGQLUserPayload
	Core() user.Usecase[TDomain]
}

// UserEmailQueryResolver defines the interface for user email operations.
type UserEmailQueryResolver[
	TDomain user.EmailCapableModel,
	TGQLUser any,
	TGQLUserPayload any,
] interface {
	User(ctx context.Context, id uint64, email string) (TGQLUserPayload, error)
}

// UserPasswordQueryResolver defines the interface for user password operations.
type UserPasswordQueryResolver[
	TDomain user.PasswordCapableModel,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
] interface {
	CreateUser(ctx context.Context, input TGQLUserInput) (TGQLUserPayload, error)
	ChangeUserPassword(ctx context.Context, currentPassword, newPassword string) (*gqlmodels.StatusResponse, error)
}

// UserUsernameQueryResolver resolves the username field on User (Username trait).
type UserPasswordResetQueryResolver[
	TDomain user.PasswordCapableModel,
	TGQLUser any,
	TGQLUserPayload any,
] interface {
	ResetUserPassword(ctx context.Context, email string) (*gqlmodels.StatusResponse, error)
	UpdateUserPassword(ctx context.Context, token, email, password string) (*gqlmodels.StatusResponse, error)
	UpdateResetedUserPassword(ctx context.Context, token, email, password string) (*gqlmodels.StatusResponse, error)
}

// UsernameModel is the constraint for user models that carry a separate username.
type UsernameModel interface {
	user.Model
	GetUsername() string
}

// UserUsernameQueryResolver resolves the username field on User (Username trait).
type UserUsernameQueryResolver[
	TDomain UsernameModel,
	TGQLUser any,
	TGQLUserInput any,
	TGQLUserPayload any,
] interface {
	UpdateUserUsername(ctx context.Context, id uint64, input TGQLUserInput) error
}
