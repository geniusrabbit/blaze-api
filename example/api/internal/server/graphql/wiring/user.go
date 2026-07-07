package wiring

import (
	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	userstack "github.com/geniusrabbit/blaze-api/example/api/internal/user"
	"github.com/geniusrabbit/blaze-api/repository/user"
	usergraphql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	userbase "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql/user_base"
	useremail "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql/user_email"
	userpassword "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql/user_password"
	userpassreset "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql/user_password_reset"
)

type UserQueryResolver interface {
	usergraphql.UserBaseQueryResolver[
		*domain.User,
		*exmodels.User,
		*exmodels.UserCreateInput,
		*exmodels.UserUpdateInput,
		*exmodels.UserPayload,
		*exmodels.UserListFilter,
		*exmodels.UserListOrder,
	]
	usergraphql.UserEmailQueryResolver[
		*domain.User,
		*exmodels.User,
		*exmodels.UserPayload,
	]
	// usergraphql.UserUsernameQueryResolver[
	// 	*domain.User,
	// 	*exmodels.User,
	// 	*exmodels.UserCreateInput,
	// 	*exmodels.UserPayload,
	// ]
	usergraphql.UserPasswordQueryResolver[
		*domain.User,
		*exmodels.User,
		*exmodels.UserCreateInput,
		*exmodels.UserPayload,
	]
	usergraphql.UserPasswordResetQueryResolver[
		*domain.User,
		*exmodels.User,
		*exmodels.UserPayload,
	]
}

type userQueryResolver struct {
	userbase.QueryResolverBase[
		*domain.User,
		*exmodels.User,
		*exmodels.UserCreateInput,
		*exmodels.UserUpdateInput,
		*exmodels.UserPayload,
		*exmodels.UserListFilter,
		*exmodels.UserListOrder,
	]
	useremail.QueryResolverEmail[
		*domain.User,
		*exmodels.User,
		*exmodels.UserPayload,
	]
	// userusername.QueryResolverUsername[
	// 	*domain.User,
	// 	*exmodels.User,
	// 	*exmodels.UserCreateInput,
	// 	*exmodels.UserPayload,
	// 	*exmodels.UserListFilter,
	// 	*exmodels.UserListOrder,
	// ]
	userpassword.QueryResolverPassword[
		*domain.User,
		*exmodels.User,
		*exmodels.UserCreateInput,
		*exmodels.UserPayload,
		*exmodels.UserListFilter,
		*exmodels.UserListOrder,
	]
	userpassreset.PasswordResetQueryResolver[*domain.User]
}

// NewExampleUserQueryResolver wires a full user resolver from a userstack Module.
func NewExampleUserQueryResolver(module userstack.Module[*domain.User]) UserQueryResolver {
	return NewUserQueryResolver(module.Core, module.Email, module.Password)
}

func NewUserQueryResolver(
	core user.Usecase[*domain.User],
	emailUsecase user.EmailUsecase[*domain.User],
	passwordUsecase user.PasswordUsecase[*domain.User],
) UserQueryResolver {
	mapper := &domain.UserGraphQLMappersImpl{}
	return &userQueryResolver{
		QueryResolverBase: *userbase.NewQueryResolverBase(userbase.QueryResolverBaseConfig[
			*domain.User,
			*exmodels.User,
			*exmodels.UserCreateInput,
			*exmodels.UserUpdateInput,
			*exmodels.UserPayload,
			*exmodels.UserListFilter,
			*exmodels.UserListOrder,
		]{
			Core:   core,
			Mapper: mapper,
		}),
		QueryResolverEmail: *useremail.NewQueryResolverEmail(useremail.QueryResolverEmailConfig[
			*domain.User,
			*exmodels.User,
			*exmodels.UserPayload,
		]{
			Core:       core,
			Email:      emailUsecase,
			ToGraphQL:  mapper.ToGQL,
			NewPayload: mapper.NewPayload,
		}),
		// QueryResolverUsername: *userusername.NewQueryResolverUsername(userusername.QueryResolverUsernameConfig[
		// 	*domain.User,
		// 	*exmodels.User,
		// 	*exmodels.UserCreateInput,
		// 	*exmodels.UserPayload,
		// 	*exmodels.UserListFilter,
		// 	*exmodels.UserListOrder,
		// ]{
		// 	Core:      core,
		// 	ToGraphQL: mapper.ToGQL,
		// }),
		QueryResolverPassword: *userpassword.NewQueryResolverPassword(userpassword.QueryResolverPasswordConfig[
			*domain.User,
			*exmodels.User,
			*exmodels.UserCreateInput,
			*exmodels.UserPayload,
			*exmodels.UserListFilter,
			*exmodels.UserListOrder,
		]{
			Core:      core,
			ToGraphQL: mapper.ToGQL,
		}),
		PasswordResetQueryResolver: *userpassreset.NewPasswordResetQueryResolver(userpassreset.PasswordResetQueryResolverConfig[*domain.User]{
			Email:    emailUsecase,
			Password: passwordUsecase,
		}),
	}
}
