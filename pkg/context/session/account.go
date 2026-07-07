package session

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountCtx "github.com/geniusrabbit/blaze-api/repository/account/context"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userContext "github.com/geniusrabbit/blaze-api/repository/user/context"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// Permission auth specific constants
const (
	PermAuthCross        = account.PermAuthCross
	AnonymousDefaultRole = `anonymous`
)

// placeholderAccount satisfies account.Model for anonymous/dev session bootstrap only.
type placeholderAccount struct {
	accountModels.AccountBase
}

func (placeholderAccount) TableName() string { return "account_base" }

func (a *placeholderAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	acc := &placeholderAccount{}
	acc.ID = id
	acc.Admins = adminUserIDs
	return acc
}

func newPlaceholderAccount(perm accountModels.PermissionChecker, adminUserID uint64) *placeholderAccount {
	acc := &placeholderAccount{}
	acc.SetPermissions(perm)
	if adminUserID > 0 {
		acc.Admins = []uint64{adminUserID}
	}
	return acc
}

func accountFromContext(ctx context.Context) account.Model {
	if acc, ok := accountCtx.SessionAccount(ctx).(account.Model); ok {
		return acc
	}
	return newPlaceholderAccount(nil, 0)
}

// anonymousSessionUser is a minimal user.Model for unauthenticated requests.
type anonymousSessionUser struct {
	userModels.UserBase
	userModels.UserEmail
}

func (anonymousSessionUser) TableName() string        { return "anonymous_user" }
func (anonymousSessionUser) RBACResourceName() string { return "anonymous.user" }

func (u *anonymousSessionUser) NewWithID(id uint64) user.Model {
	return &anonymousSessionUser{UserBase: userModels.UserBase{ID: id}}
}

var anonymousUser = &anonymousSessionUser{
	UserEmail: userModels.UserEmail{Email: "<anonymous>"},
	UserBase:  userModels.UserBase{Approve: models.ApprovedApproveStatus},
}

// WithUserAccount puts user and account models into context.
func WithUserAccount[TUser user.Model](ctx context.Context, userObj TUser, accountObj account.Model) context.Context {
	if accountObj == nil || (accountObj.GetID() == 0 && accountObj.IsAnonymous()) {
		pm := permissions.FromContext(ctx)
		role := pm.Role(ctx, AnonymousDefaultRole)
		adminID := uint64(0)
		if !userObj.IsNil() && !userObj.IsAnonymous() {
			adminID = userObj.GetID()
		}
		accountObj = newPlaceholderAccount(role, adminID)
	}
	ctx = userContext.WithSessionUser(ctx, userObj)
	ctx = accountCtx.WithSessionAccount(ctx, accountObj)
	return ctx
}

// WithAnonymousUserAccount puts anonymous user and account into context.
func WithAnonymousUserAccount(ctx context.Context) context.Context {
	pm := permissions.FromContext(ctx)
	role := pm.Role(ctx, AnonymousDefaultRole)
	return WithUserAccount(ctx, anonymousUser, newPlaceholderAccount(role, 0))
}

// WithUserAccountDevelop sets development objects into the context.
func WithUserAccountDevelop(ctx context.Context) context.Context {
	manager := permissions.NewTestManager(ctx)
	role := manager.Role(ctx, `test`)
	acc := newPlaceholderAccount(role, 1)
	acc.SetID(1)
	devUser := &anonymousSessionUser{UserBase: userModels.UserBase{ID: 1}}
	return WithUserAccount(ctx, devUser, acc)
}

// UserAccount returns user + account models from context.
func UserAccount(ctx context.Context) (user.Model, account.Model) {
	u := userContext.SessionUser(ctx)
	if u == nil {
		u = anonymousUser
	}
	return u, accountFromContext(ctx)
}

// UserModel returns current user as Model interface.
func UserModel(ctx context.Context) user.Model {
	u, _ := UserAccount(ctx)
	return u
}

// AccountModel returns current account as Model interface.
func AccountModel(ctx context.Context) account.Model {
	_, a := UserAccount(ctx)
	return a
}

// Account returns current account model.
func Account(ctx context.Context) account.Model {
	return accountFromContext(ctx)
}

// AccountID returns current account ID (0 when absent).
func AccountID(ctx context.Context) uint64 {
	if acc := Account(ctx); acc != nil {
		return acc.GetID()
	}
	return 0
}

// User returns current user model.
func User(ctx context.Context) user.Model {
	return UserModel(ctx)
}

// UserID returns current user ID (0 when absent).
func UserID(ctx context.Context) uint64 {
	if u := UserModel(ctx); u != nil {
		return u.GetID()
	}
	return 0
}
