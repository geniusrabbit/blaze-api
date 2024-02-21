package graphql

import (
	"context"
	"errors"
	"time"

	lrbac "github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/jwt"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountrepo "github.com/geniusrabbit/blaze-api/repository/account/repository"
	accountusecase "github.com/geniusrabbit/blaze-api/repository/account/usecase"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var (
	errInvalidAccountTarget = errors.New(`invalid account target`)
	errUserIsNotAuthorized  = errors.New(`user is not authorized properly`)
)

// AuthResolver is the resolver for the Auth type
type AuthResolver struct {
	provider       *jwt.Provider
	userRepo       *userrepo.Repository
	accountRepo    *accountrepo.Repository
	accountUsecase account.Usecase
	roleRepo       rbac.Repository
}

// NewAuthResolver creates new resolver for the Auth type
func NewAuthResolver(provider *jwt.Provider, roleRepo rbac.Repository) *AuthResolver {
	return &AuthResolver{
		provider:       provider,
		userRepo:       userrepo.New(),
		accountRepo:    accountrepo.New(),
		accountUsecase: accountusecase.NewAccountUsecase(userrepo.New(), accountrepo.New()),
		roleRepo:       roleRepo,
	}
}

// Login is the resolver for the login field
func (r *AuthResolver) Login(ctx context.Context, login string, password string) (*models.SessionToken, error) {
	accountID := uint64(0)
	user, err := r.userRepo.GetByPassword(ctx, login, password)
	if err != nil {
		return nil, err
	}

	account, err := accountForUser(ctx, r.accountRepo, user, accountID)
	if err != nil {
		return nil, err
	}
	if account != nil {
		accountID = account.ID
	}

	token, err := r.provider.CreateToken(user.ID, accountID)
	if err != nil {
		return nil, err
	}

	roles := account.Permissions.ChildRoles()
	return &models.SessionToken{
		Token:     token,
		ExpiresAt: time.Now().Add(r.provider.TokenLifetime),
		IsAdmin:   account.IsAdminUser(user.GetID()), // Is current account admin
		Roles:     xtypes.SliceApply[lrbac.Role, string](roles, func(r lrbac.Role) string { return r.Name() }),
	}, nil
}

// Logout is the resolver for the logout field.
func (r *AuthResolver) Logout(ctx context.Context) (bool, error) {
	return true, nil
}

// CurrentSession is the resolver for the currentSession field
func (r *AuthResolver) CurrentSession(ctx context.Context) (*models.SessionToken, error) {
	user, account, token := session.User(ctx), session.Account(ctx), session.Token(ctx)
	roles := account.Permissions.ChildRoles()
	return &models.SessionToken{
		Token:     token,
		ExpiresAt: time.Now().Add(r.provider.TokenLifetime),
		IsAdmin:   account.IsAdminUser(user.GetID()), // Is current account admin
		Roles:     xtypes.SliceApply[lrbac.Role, string](roles, func(r lrbac.Role) string { return r.Name() }),
	}, nil
}

// ListRolesAndPermissions is the resolver for the listRolesAndPermissions field
func (r *AuthResolver) ListRolesAndPermissions(ctx context.Context, accountID uint64, order *models.RBACRoleListOrder) (*connectors.RBACRoleConnection, error) {
	var (
		err     error
		account *model.Account
		permIDs []uint64
	)
	if accountID != 0 {
		account, err = r.accountUsecase.Get(ctx, accountID)
		if err != nil {
			return nil, err
		}
	} else {
		account = session.Account(ctx)
		if account == nil {
			return nil, errUserIsNotAuthorized
		}
	}
	if account != nil && account.Permissions != nil {
		permIDs = xtypes.SliceApply[lrbac.Role](account.Permissions.ChildRoles(), func(r lrbac.Role) uint64 {
			switch ext := r.Ext().(type) {
			case *permissions.ExtData:
				return ext.ID
			default:
				return 0
			}
		}).Filter(func(id uint64) bool { return id != 0 })
	}
	return connectors.NewRBACRoleConnectionByIDs(ctx, r.roleRepo, permIDs, order), nil
}

func accountForUser(ctx context.Context, accountRepo account.Repository, user *model.User, accountID uint64) (*model.Account, error) {
	accounts, err := accountRepo.FetchList(ctx, &account.Filter{UserID: []uint64{user.ID}}, nil)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		if accountID != 0 {
			return nil, errInvalidAccountTarget
		}
		return nil, nil
	}
	// Load permissions for the account and check membership
	if err = accountRepo.LoadPermissions(ctx, accounts[0], user); err != nil {
		return nil, err
	}
	return accounts[0], nil
}
