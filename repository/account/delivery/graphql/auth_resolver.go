package graphql

import (
	"context"
	"errors"
	"time"

	"github.com/demdxx/gocast/v2"
	lrbac "github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	rbacgql "github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql"
	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var (
	errInvalidAccountTarget = errors.New(`invalid account target`)
	errUserIsNotAuthorized  = errors.New(`user is not authorized properly`)
)

// AuthResolver is the resolver for the Auth type.
type AuthResolver[TUser user.Model, TAccount account.Model] struct {
	provider       *jwt.Provider
	accountRepo    account.SessionRepository[TUser, TAccount]
	accountUsecase account.Usecase[TUser, TAccount]
	roleRepo       rbac.Repository
}

// NewAuthResolver creates new resolver for the Auth type.
func NewAuthResolver[TUser user.Model, TAccount account.Model](
	provider *jwt.Provider,
	accountRepo account.SessionRepository[TUser, TAccount],
	accountUsecase account.Usecase[TUser, TAccount],
	roleRepo rbac.Repository,
) *AuthResolver[TUser, TAccount] {
	return &AuthResolver[TUser, TAccount]{
		provider:       provider,
		accountRepo:    accountRepo,
		accountUsecase: accountUsecase,
		roleRepo:       roleRepo,
	}
}

// Logout is the resolver for the logout field.
func (r *AuthResolver[TUser, TAccount]) Logout(ctx context.Context) (bool, error) {
	return true, nil
}

// SwitchAccount is the resolver for the switchAccount field.
func (r *AuthResolver[TUser, TAccount]) SwitchAccount(ctx context.Context, id uint64) (*gqlmodels.SessionToken, error) {
	userObj := session.User(ctx)
	if userObj == nil || userObj.IsAnonymous() {
		return nil, errUserIsNotAuthorized
	}

	var typedUser TUser
	if u, ok := any(userObj).(TUser); ok {
		typedUser = u
	}

	acc, err := r.accountForUser(ctx, typedUser, id)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := r.provider.CreateToken(userObj.GetID(), acc.GetID(), 0)
	if err != nil {
		return nil, err
	}

	return r.sessionTokenFromAccount(typedUser, acc, token, expiresAt)
}

// CurrentSession is the resolver for the currentSession field.
func (r *AuthResolver[TUser, TAccount]) CurrentSession(ctx context.Context) (*gqlmodels.SessionToken, error) {
	userObj := session.User(ctx)
	accModel := session.Account(ctx)
	token := session.Token(ctx)
	var zero TAccount
	acc, _ := accModel.(TAccount)
	if any(acc) == any(zero) || acc.IsNil() {
		acc = zero
	}
	var typedUser TUser
	if u, ok := any(userObj).(TUser); ok {
		typedUser = u
	}
	return r.sessionTokenFromAccount(typedUser, acc, token, time.Now().Add(r.provider.TokenLifetime))
}

// ListRolesAndPermissions is the resolver for the listRolesAndPermissions field.
func (r *AuthResolver[TUser, TAccount]) ListRolesAndPermissions(ctx context.Context, accountID uint64, order []*gqlmodels.RBACRoleListOrder) (*rbacgql.RBACRoleConnection, error) {
	var (
		err error
		acc TAccount
	)

	if accountID != 0 {
		if acc, err = r.accountUsecase.Get(ctx, accountID); err != nil {
			return nil, err
		}
	} else if sessionAcc := session.Account(ctx); sessionAcc != nil {
		if typed, ok := sessionAcc.(TAccount); ok {
			acc = typed
		}
	}
	if acc.IsNil() {
		return nil, errUserIsNotAuthorized
	}

	var permIDs []uint64
	if checker := acc.PermissionsChecker(); checker != nil {
		childRoles := append([]lrbac.Role{}, checker.ChildRoles()...)
		if role, ok := checker.(lrbac.Role); ok {
			childRoles = append(childRoles, role)
		}
		permIDs = xtypes.SliceApply(childRoles, func(r lrbac.Role) uint64 {
			switch ext := r.Ext().(type) {
			case *permissions.ExtData:
				return ext.ID
			default:
				return 0
			}
		}).Filter(func(id uint64) bool { return id != 0 })
	}
	return rbacgql.NewRBACRoleConnectionByIDs(ctx, r.roleRepo, permIDs, order), nil
}

func (r *AuthResolver[TUser, TAccount]) sessionTokenFromAccount(
	user TUser,
	acc TAccount,
	token string,
	expiresAt time.Time,
) (*gqlmodels.SessionToken, error) {
	var zeroAcc TAccount
	roles := []lrbac.Role{}
	if any(acc) != any(zeroAcc) {
		if checker := acc.PermissionsChecker(); checker != nil {
			roles = append(roles, checker.ChildRoles()...)
			if role, ok := checker.(lrbac.Role); ok {
				roles = append(roles, role)
			}
		}
	}
	var zeroUser TUser
	isAdmin := false
	if any(acc) != any(zeroAcc) && any(user) != any(zeroUser) {
		isAdmin = acc.IsAdminUser(user.GetID())
	}
	return &gqlmodels.SessionToken{
		Token:     token,
		ExpiresAt: expiresAt.UTC(),
		IsAdmin:   isAdmin,
		Roles:     xtypes.SliceApply(roles, func(r lrbac.Role) string { return r.Name() }),
	}, nil
}

func (r *AuthResolver[TUser, TAccount]) accountForUser(
	ctx context.Context,
	user TUser,
	accountID uint64,
) (TAccount, error) {
	var zero TAccount
	accounts, err := r.accountRepo.FetchList(ctx,
		&account.Filter{
			ID:     gocast.IfThen(accountID > 0, []uint64{accountID}, nil),
			UserID: []uint64{user.GetID()},
		},
		&repository.Pagination{Size: 1},
	)
	if err != nil {
		return zero, err
	}
	if len(accounts) == 0 {
		if accountID != 0 {
			return zero, errInvalidAccountTarget
		}
		return zero, nil
	}
	if err = r.accountRepo.LoadPermissions(ctx, accounts[0], user); err != nil {
		return zero, err
	}
	return accounts[0], nil
}
