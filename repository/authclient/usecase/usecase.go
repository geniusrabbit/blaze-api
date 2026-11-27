// Package usecase account implementation
package usecase

import (
	"context"
	"github.com/geniusrabbit/blaze-api/pkg/context/database"

	"github.com/demdxx/gocast/v2"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/models"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/pkg/errors"
)

// AuthclientUsecase provides bussiness logic for account access
type AuthclientUsecase struct {
	authclientRepo authclient.Repository
}

// NewAuthclientUsecase object controller
func NewAuthclientUsecase(repo authclient.Repository) *AuthclientUsecase {
	return &AuthclientUsecase{
		authclientRepo: repo,
	}
}

// Get returns the group by ID if have access
func (a *AuthclientUsecase) Get(ctx context.Context, id string) (*models.AuthClient, error) {
	authclientObj, err := a.authclientRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, authclientObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view authclient")
	}
	return authclientObj, nil
}

// FetchList of accounts by filter
func (a *AuthclientUsecase) FetchList(ctx context.Context, opts ...authclient.QOption) ([]*models.AuthClient, error) {
	if !acl.HaveAccessList(ctx, &models.AuthClient{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list authclient")
	}
	list, err := a.authclientRepo.FetchList(ctx, opts...)
	for _, link := range list {
		if !acl.HaveAccessList(ctx, link) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list authclient")
		}
	}
	return list, err
}

// Count of accounts by filter
func (a *AuthclientUsecase) Count(ctx context.Context, opts ...authclient.QOption) (int64, error) {
	if !acl.HaveAccessList(ctx, &models.AuthClient{}) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list authclient")
	}
	return a.authclientRepo.Count(ctx, opts...)
}

// Create new object in database.
//
// An OAuth2 client acts as its owner (user_id, account_id): a token issued to
// it resolves to that user and account. So the owner defaults to the session
// user and account, and only `auth_client.create.all|system` may set another one.
func (a *AuthclientUsecase) Create(ctx context.Context, authclientObj *models.AuthClient, opts ...authclient.QOption) (string, error) {
	var err error
	if !acl.HaveAccessCreate(ctx, authclientObj) {
		return "", errors.Wrap(acl.ErrNoPermissions, "create authclient")
	}
	if err = assignOwner(ctx, authclientObj, acl.PermCreate); err != nil {
		return "", err
	}
	authclientObj.ID, err = a.authclientRepo.Create(ctx, authclientObj, opts...)
	return authclientObj.ID, err
}

// Update object in database.
// Access is checked on the stored object and on the new values; changing the
// owner requires `auth_client.update.all|system`.
func (a *AuthclientUsecase) Update(ctx context.Context, id string, authclientObj *models.AuthClient, opts ...authclient.QOption) error {
	current, err := a.authclientRepo.Get(database.WithPrimaryRead(ctx), id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessUpdate(ctx, current) || !acl.HaveAccessUpdate(ctx, authclientObj) {
		return errors.Wrap(acl.ErrNoPermissions, "update authclient")
	}
	if (current.UserID != authclientObj.UserID || current.AccountID != authclientObj.AccountID) &&
		!haveGlobalAccess(ctx, authclientObj, acl.PermUpdate) {
		return errors.Wrap(acl.ErrNoPermissions, "change authclient owner")
	}
	return a.authclientRepo.Update(historylog.WithPK(ctx, id), id, authclientObj, opts...)
}

// haveGlobalAccess reports whether the session has the action on any owner.
func haveGlobalAccess(ctx context.Context, obj *models.AuthClient, action string) bool {
	return acl.HaveObjectPermissions(ctx, obj, action+".all", action+".system")
}

// assignOwner sets the session user and account as the client owner. Another
// owner is accepted only with the global permission for the action.
func assignOwner(ctx context.Context, obj *models.AuthClient, action string) error {
	userObj, accountObj := session.UserAccount(ctx)
	var userID, accountID uint64
	if !gocast.IsNil(userObj) {
		userID = userObj.GetID()
	}
	if !gocast.IsNil(accountObj) {
		accountID = accountObj.GetID()
	}
	if (obj.UserID == 0 || obj.UserID == userID) && (obj.AccountID == 0 || obj.AccountID == accountID) {
		obj.UserID, obj.AccountID = userID, accountID
		return nil
	}
	if haveGlobalAccess(ctx, obj, action) {
		return nil
	}
	return errors.Wrap(acl.ErrNoPermissions, "authclient owner must be the session user and account")
}

// Delete delites record by ID
func (a *AuthclientUsecase) Delete(ctx context.Context, id string, opts ...authclient.QOption) error {
	authclientObj, err := a.Get(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, authclientObj) {
		return errors.Wrap(acl.ErrNoPermissions, "delete authclient")
	}
	return a.authclientRepo.Delete(historylog.WithPK(ctx, id), id, opts...)
}
