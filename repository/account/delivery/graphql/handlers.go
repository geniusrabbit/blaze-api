package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/account"
	rbacgql "github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql"
	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AuthQueryHandler is the method set required for account auth GraphQL resolvers.
// Login is intentionally excluded — it lives in consumer schema extensions
// (e.g. account_login for email+password, socialauth for OAuth2).
type AuthQueryHandler interface {
	Logout(ctx context.Context) (bool, error)
	SwitchAccount(ctx context.Context, id uint64) (*gqlmodels.SessionToken, error)
	CurrentSession(ctx context.Context) (*gqlmodels.SessionToken, error)
	ListRolesAndPermissions(ctx context.Context, accountID uint64, order []*gqlmodels.RBACRoleListOrder) (*rbacgql.RBACRoleConnection, error)
}

// AccountQueryHandler is the method set required for account GraphQL resolvers.
type AccountQueryHandler[TGQLAccount, TPayload, TCreateInput, TUpdateInput any] interface {
	CurrentAccount(ctx context.Context) (TPayload, error)
	Account(ctx context.Context, id uint64) (TPayload, error)
	RegisterAccount(ctx context.Context, ownerID uint64, input TCreateInput) (TPayload, error)
	UpdateAccount(ctx context.Context, id uint64, input TUpdateInput) (TPayload, error)
	ApproveAccount(ctx context.Context, id uint64, msg string) (TPayload, error)
	RejectAccount(ctx context.Context, id uint64, msg string) (TPayload, error)
	ListAccounts(ctx context.Context, filter account.QOption, order []*account.QOption, page *gqlmodels.Page) (*AccountConnection[TGQLAccount], error)
}

// MemberQueryHandler is the method set required for account member GraphQL resolvers.
type MemberQueryHandler interface {
	Invite(ctx context.Context, accountID uint64, member gqlmodels.InviteMemberInput) (*gqlmodels.MemberPayload, error)
	Update(ctx context.Context, memberID uint64, member gqlmodels.MemberInput) (*gqlmodels.MemberPayload, error)
	Remove(ctx context.Context, memberID uint64) (*gqlmodels.MemberPayload, error)
	Approve(ctx context.Context, memberID uint64, msg string) (*gqlmodels.MemberPayload, error)
	Reject(ctx context.Context, memberID uint64, msg string) (*gqlmodels.MemberPayload, error)
	List(ctx context.Context, filter *gqlmodels.MemberListFilter, order []*gqlmodels.MemberListOrder, page *gqlmodels.Page) (*MemberConnection, error)
}

// ModelWithID returns a new model instance with primary key set (fallback when repo lookup is unavailable).
func ModelWithID[T any](newModel func() T, id uint64) T {
	m := newModel()
	if setter, ok := any(m).(interface{ SetID(uint64) }); ok {
		setter.SetID(id)
	}
	return m
}

func userEmail[T user.Model](u T) string {
	if em, ok := any(u).(user.EmailModel); ok {
		return em.GetEmail()
	}
	return ""
}
