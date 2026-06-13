package session

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountCtx "github.com/geniusrabbit/blaze-api/repository/account/context"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userContext "github.com/geniusrabbit/blaze-api/repository/user/context"
	// "github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	// "google.golang.org/grpc/metadata"
)

// Permission auth specific constants
const (
	PermAuthCross        = account.PermAuthCross
	AnonymousDefaultRole = `anonymous`
)

// WithUserAccount puts to the context user and account models
func WithUserAccount(ctx context.Context, userObj *user.User, accountObj *account.Account) context.Context {
	if accountObj == nil {
		pm := permissions.FromContext(ctx)
		role := pm.Role(ctx, AnonymousDefaultRole)
		accountObj = &account.Account{
			Title:       "<anonymous>",
			Permissions: role,
			Admins:      []uint64{userObj.ID},
		}
	}
	ctx = userContext.WithSessionUser(ctx, userObj)
	ctx = accountCtx.WithSessionAccount(ctx, accountObj)
	return ctx
}

// WithAnonymousUserAccount puts to context user and account with anonym permissions
func WithAnonymousUserAccount(ctx context.Context) context.Context {
	pm := permissions.FromContext(ctx)
	role := pm.Role(ctx, AnonymousDefaultRole)
	return WithUserAccount(ctx,
		&user.User{Email: "<anonymous>", Approve: models.ApprovedApproveStatus},
		&account.Account{Title: "<anonymous>", Permissions: role})
}

// WithUserAccountDevelop sets development objects into the context
// nolint:unused // ...
func WithUserAccountDevelop(ctx context.Context) context.Context {
	manager := permissions.NewTestManager(ctx)
	role := manager.Role(ctx, `test`) // INFO: Assume that there is no error because of this is the test manager
	ctx = WithUserAccount(ctx,
		&user.User{ID: 1},
		&account.Account{ID: 1, Permissions: role, Admins: []uint64{1}},
	)
	// if changelog.MessageQueue(ctx) == nil {
	// 	ctx = changelog.WithMessageQueue(ctx)
	// }
	return ctx
}

// UserAccount returns user + account models
func UserAccount(ctx context.Context) (u *user.User, a *account.Account) {
	if u = userContext.SessionUser(ctx); u == nil {
		u = &user.Anonymous
	}
	if a = accountCtx.SessionAccount(ctx); a == nil {
		a = &account.Account{}
	}
	return u, a
}

// Account returns current account model
func Account(ctx context.Context) *account.Account {
	return accountCtx.SessionAccount(ctx)
}

// User returns current user model
// nolint:unused // temporary
func User(ctx context.Context) *user.User {
	return userContext.SessionUser(ctx)
}
