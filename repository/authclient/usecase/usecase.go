// Package usecase account implementation
package usecase

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
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

// Create new object in database
func (a *AuthclientUsecase) Create(ctx context.Context, authclientObj *models.AuthClient, opts ...authclient.QOption) (string, error) {
	var err error
	if !acl.HaveAccessCreate(ctx, authclientObj) {
		return "", errors.Wrap(acl.ErrNoPermissions, "create authclient")
	}
	authclientObj.ID, err = a.authclientRepo.Create(ctx, authclientObj, opts...)
	return authclientObj.ID, err
}

// Update object in database
func (a *AuthclientUsecase) Update(ctx context.Context, id string, authclientObj *models.AuthClient, opts ...authclient.QOption) error {
	if !acl.HaveAccessUpdate(ctx, authclientObj) {
		return errors.Wrap(acl.ErrNoPermissions, "update authclient")
	}
	return a.authclientRepo.Update(historylog.WithPK(ctx, id), id, authclientObj, opts...)
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
