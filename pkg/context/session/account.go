package session

import (
	"context"

	// "github.com/geniusrabbit/blaze-api/repository/user"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	// "github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	// "google.golang.org/grpc/metadata"
)

var (
	ctxUserKey    = &struct{ s string }{"account:user"}
	ctxAccountKey = &struct{ s string }{"account:account"}
)

// Permission auth specific constants
const (
	PermAuthCross        = `auth.cross`
	AnonymousDefaultRole = `anonymous`
)

// WithUserAccount puts to the context user and account models
func WithUserAccount(ctx context.Context, userObj *model.User, accountObj *model.Account) context.Context {
	if accountObj == nil {
		pm := permissions.FromContext(ctx)
		role := pm.Role(ctx, AnonymousDefaultRole)
		accountObj = &model.Account{Title: "<anonymous>", Permissions: role, Admins: []uint64{userObj.ID}}
	}
	ctx = context.WithValue(ctx, ctxUserKey, userObj)
	ctx = context.WithValue(ctx, ctxAccountKey, accountObj)
	return ctx
}

// WithAnonymousUserAccount puts to context user and account with anonym permissions
func WithAnonymousUserAccount(ctx context.Context) context.Context {
	pm := permissions.FromContext(ctx)
	role := pm.Role(ctx, AnonymousDefaultRole)
	return WithUserAccount(ctx,
		&model.User{Email: "<anonymous>", Approve: model.ApprovedApproveStatus},
		&model.Account{Title: "<anonymous>", Permissions: role})
}

// WithUserAccountDevelop sets development objects into the context
// nolint:unused // ...
func WithUserAccountDevelop(ctx context.Context) context.Context {
	manager := permissions.NewTestManager(ctx)
	role := manager.Role(ctx, `test`) // INFO: Assume that there is no error because of this is the test manager
	ctx = WithUserAccount(ctx,
		&model.User{ID: 1},
		&model.Account{ID: 1, Permissions: role, Admins: []uint64{1}},
	)
	// if changelog.MessageQueue(ctx) == nil {
	// 	ctx = changelog.WithMessageQueue(ctx)
	// }
	return ctx
}

// UserAccount returns user + account models
func UserAccount(ctx context.Context) (u *model.User, a *model.Account) {
	if u, _ = ctx.Value(ctxUserKey).(*model.User); u == nil {
		u = &model.Anonymous
	}
	if a, _ = ctx.Value(ctxAccountKey).(*model.Account); a == nil {
		a = &model.Account{}
	}
	return u, a
}

// Account returns current account model
func Account(ctx context.Context) *model.Account {
	return ctx.Value(ctxAccountKey).(*model.Account)
}

// User returns current user model
// nolint:unused // temporary
func User(ctx context.Context) *model.User {
	return ctx.Value(ctxUserKey).(*model.User)
}
