// Package accountlogin provides the email+password login mutation as a schema extension.
// It wraps AuthResolver so consumers include it only when they support this auth method.
//
// Usage in consumer gqlgen config:
//
//	schema:
//	  - "repository/account/delivery/graphql/account_base.graphql"
//	  - "repository/account/delivery/graphql/account_login/login.graphql"
//	  - ...
package accountlogin

import (
	"context"
	"errors"
	"time"

	"github.com/demdxx/gocast/v2"
	lrbac "github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var errInvalidAccountTarget = errors.New(`invalid account target`)

// LoginPasswordAuth is the interface for username/password login.
type LoginPasswordAuth[T user.Model] interface {
	Login(ctx context.Context, login, password string) (T, error)
}

// emailPasswordLogin adapts separate email + password repositories into LoginPasswordAuth.
type emailPasswordLogin[T user.AuthCapableModel] struct {
	email    user.EmailRepository[T]
	password user.PasswordRepository[T]
}

// NewEmailPasswordLogin creates a LoginPasswordAuth from separate email and password repositories.
func NewEmailPasswordLogin[T user.AuthCapableModel](
	email user.EmailRepository[T],
	password user.PasswordRepository[T],
) LoginPasswordAuth[T] {
	return &emailPasswordLogin[T]{email: email, password: password}
}

func (l *emailPasswordLogin[T]) Login(ctx context.Context, login, pass string) (T, error) {
	userObj, err := l.email.GetByEmail(ctx, login)
	if err != nil {
		var zero T
		return zero, err
	}
	return l.password.GetByPassword(ctx, userObj.GetID(), pass)
}

// Resolver exposes the login mutation for consumers that extend the schema
// with account_login/login.graphql.
type Resolver[TUser user.Model, TAccount account.Model] struct {
	provider    *jwt.Provider
	userLogin   LoginPasswordAuth[TUser]
	accountRepo account.SessionRepository[TUser, TAccount]
}

// New wraps an existing AuthResolver to serve the login mutation.
func New[TUser user.Model, TAccount account.Model](
	provider *jwt.Provider,
	userLogin LoginPasswordAuth[TUser],
	accountRepo account.SessionRepository[TUser, TAccount],
) *Resolver[TUser, TAccount] {
	return &Resolver[TUser, TAccount]{provider: provider, userLogin: userLogin, accountRepo: accountRepo}
}

// Login resolves mutation { login(email, password, accountID) }.
// accountID is optional — nil means use the user's default account.
func (r *Resolver[TUser, TAccount]) Login(ctx context.Context, login, password string, accountID ...uint64) (*gqlmodels.SessionToken, error) {
	var accID uint64
	if len(accountID) > 0 {
		accID = accountID[0]
	}

	user, err := r.userLogin.Login(ctx, login, password)
	if err != nil {
		return nil, err
	}

	acc, err := r.accountForUser(ctx, user, accID)
	if err != nil {
		return nil, err
	}
	if !acc.IsNil() && !acc.IsAnonymous() {
		accID = acc.GetID()
	}

	token, expiresAt, err := r.provider.CreateToken(user.GetID(), accID, 0)
	if err != nil {
		return nil, err
	}

	return r.sessionTokenFromAccount(user, acc, token, expiresAt)
}

func (r *Resolver[TUser, TAccount]) sessionTokenFromAccount(
	user TUser,
	acc TAccount,
	token string,
	expiresAt time.Time,
) (*gqlmodels.SessionToken, error) {
	isAdmin := false
	roles := []lrbac.Role{}
	if !gocast.IsEmpty(acc) {
		if checker := acc.PermissionsChecker(); checker != nil {
			roles = append(roles, checker.ChildRoles()...)
			if role, ok := checker.(lrbac.Role); ok {
				roles = append(roles, role)
			}
		}
		if !gocast.IsEmpty(user) {
			isAdmin = acc.IsAdminUser(user.GetID())
		}
	}
	return &gqlmodels.SessionToken{
		Token:     token,
		ExpiresAt: expiresAt.UTC(),
		IsAdmin:   isAdmin,
		Roles:     xtypes.SliceApply(roles, func(r lrbac.Role) string { return r.Name() }),
	}, nil
}

func (r *Resolver[TUser, TAccount]) accountForUser(ctx context.Context, user TUser, accountID uint64) (TAccount, error) {
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
