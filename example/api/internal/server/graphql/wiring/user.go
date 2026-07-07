package wiring

import (
	"context"

	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	userstack "github.com/geniusrabbit/blaze-api/example/api/internal/user"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	"github.com/geniusrabbit/blaze-api/repository/user"
	usergraphql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	userbase "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql/user_base"
	useremail "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql/user_email"
	userpassword "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql/user_password"
	userpasswordreset "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql/user_password_reset"
	basemodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// ExampleUserQueryResolver composes user_base with email/password/reset extensions.
type ExampleUserQueryResolver[T user.AuthCapableModel] struct {
	*userbase.QueryResolverBase[
		T,
		exmodels.User,
		*exmodels.UserInput,
		*exmodels.UserPayload,
		*exmodels.UserListFilter,
		*exmodels.UserListOrder,
	]
	*useremail.QueryResolverEmail[T, exmodels.User, *exmodels.UserPayload]
	*userpassword.QueryResolverPassword[
		T,
		exmodels.User,
		*exmodels.UserInput,
		*exmodels.UserPayload,
		*exmodels.UserListFilter,
		*exmodels.UserListOrder,
	]
	*userpasswordreset.PasswordResetQueryResolver[T]
}

// exampleUserMapper adapts domain.UserGraphQLMappersImpl to the generic T via type assertions.
type exampleUserMapper[T user.AuthCapableModel] struct{}

func (exampleUserMapper[T]) New() T {
	return any(new(domain.User)).(T)
}

func (exampleUserMapper[T]) ToGQL(u T) exmodels.User {
	if typed, ok := any(u).(*domain.User); ok {
		return domain.UserGraphQLMappersImpl{}.ToGQL(typed)
	}
	return exmodels.User{}
}

func (exampleUserMapper[T]) FromCreateInput(inp *exmodels.UserInput) T {
	return any(domain.UserGraphQLMappersImpl{}.FromCreateInput(inp)).(T)
}

func (exampleUserMapper[T]) FromUpdateInput(inp *exmodels.UserInput, dest T) T {
	if typed, ok := any(dest).(*domain.User); ok {
		return any(domain.UserGraphQLMappersImpl{}.FromUpdateInput(inp, typed)).(T)
	}
	return dest
}

func (exampleUserMapper[T]) NewPayload(clientMutationID string, userID uint64, u exmodels.User) *exmodels.UserPayload {
	return domain.UserGraphQLMappersImpl{}.NewPayload(clientMutationID, userID, u)
}

func (exampleUserMapper[T]) ToAccountGQL(u T) accountgraphql.BaseAccountGQLUser {
	if typed, ok := any(u).(*domain.User); ok {
		return domain.UserGraphQLMappersImpl{}.ToAccountGQL(typed)
	}
	return nil
}

func (exampleUserMapper[T]) FromAccountInput(inp accountgraphql.BaseAccountGQLUserInput, appStatus ...pkgModels.ApproveStatus) T {
	return any(domain.UserGraphQLMappersImpl{}.FromAccountInput(inp, appStatus...)).(T)
}

func (exampleUserMapper[T]) FromFilter(f *exmodels.UserListFilter) user.QOption {
	return domain.UserGraphQLMappersImpl{}.FromFilter(f)
}

func (exampleUserMapper[T]) FromOrder(o *exmodels.UserListOrder) user.QOption {
	return domain.UserGraphQLMappersImpl{}.FromOrder(o)
}

// NewExampleUserQueryResolver wires full user GraphQL with consumer mappers from example/domain.
// Mapping logic is delegated to [domain.UserGraphQLMappersImpl] — stateless, add fields to it
// to carry configuration (e.g. CDN base URL) without touching the resolver code.
func NewExampleUserQueryResolver[T user.AuthCapableModel](mod userstack.Module[T]) *ExampleUserQueryResolver[T] {
	m := exampleUserMapper[T]{}

	base := userbase.NewQueryResolverBase(userbase.QueryResolverBaseConfig[
		T,
		exmodels.User,
		*exmodels.UserInput,
		*exmodels.UserPayload,
		*exmodels.UserListFilter,
		*exmodels.UserListOrder,
	]{
		Core: mod.Core,
		Mapper: usergraphql.FuncUserMapper[
			T,
			exmodels.User,
			*exmodels.UserInput,
			*exmodels.UserInput,
			*exmodels.UserPayload,
			*exmodels.UserListFilter,
			*exmodels.UserListOrder,
		]{
			NewFn:        m.New,
			ToGQLFn:      m.ToGQL,
			FromCreateFn: m.FromCreateInput,
			FromUpdateFn: m.FromUpdateInput,
			NewPayloadFn: m.NewPayload,
			FromFilterFn: m.FromFilter,
			FromOrderFn:  m.FromOrder,
		},
	})

	toGraphQL := m.ToGQL
	newPayload := m.NewPayload
	userFromInput := func(input *exmodels.UserInput, appStatus ...pkgModels.ApproveStatus) T {
		u := m.FromCreateInput(input)
		if len(appStatus) > 0 {
			if setter, ok := any(u).(interface{ SetApprove(pkgModels.ApproveStatus) }); ok {
				setter.SetApprove(appStatus[0])
			}
		}
		return u
	}

	return &ExampleUserQueryResolver[T]{
		QueryResolverBase: base,
		QueryResolverEmail: useremail.NewQueryResolverEmail(useremail.QueryResolverEmailConfig[T, exmodels.User, *exmodels.UserPayload]{
			Core:       mod.Core,
			Email:      mod.Email,
			ToGraphQL:  toGraphQL,
			NewPayload: newPayload,
		}),
		QueryResolverPassword: userpassword.NewQueryResolverPassword(userpassword.QueryResolverPasswordConfig[
			T,
			exmodels.User,
			*exmodels.UserInput,
			*exmodels.UserPayload,
			*exmodels.UserListFilter,
			*exmodels.UserListOrder,
		]{
			Core:          mod.Core,
			PassRepo:      mod.Repo,
			Password:      mod.Password,
			UserFromInput: userFromInput,
			NewPayload:    newPayload,
			ToGraphQL:     toGraphQL,
		}),
		PasswordResetQueryResolver: userpasswordreset.NewPasswordResetQueryResolver(userpasswordreset.PasswordResetQueryResolverConfig[T]{
			Email:    mod.Email,
			Password: mod.Password,
		}),
	}
}

// UpdateResetedUserPassword aliases password reset resolver naming for gqlgen.
func (r *ExampleUserQueryResolver[T]) UpdateResetedUserPassword(ctx context.Context, token, email, password string) (*basemodels.StatusResponse, error) {
	return r.PasswordResetQueryResolver.UpdateResetedUserPassword(ctx, token, email, password)
}

// CreateUser adapts the extended UserCreateInput to the embedded QueryResolverPassword.
func (r *ExampleUserQueryResolver[T]) CreateUser(ctx context.Context, input *exmodels.UserCreateInput) (*exmodels.UserPayload, error) {
	var userInput *exmodels.UserInput
	if input != nil && input.Status != nil {
		_ = input.Status // Status applied via ApproveUser after creation
	}
	return r.QueryResolverPassword.CreateUser(ctx, userInput)
}

// UpdateUser adapts the extended UserUpdateInput to the embedded QueryResolverBase.
func (r *ExampleUserQueryResolver[T]) UpdateUser(ctx context.Context, id uint64, input *exmodels.UserUpdateInput) (*exmodels.UserPayload, error) {
	return r.QueryResolverBase.UpdateUser(ctx, id, nil)
}
