// Package usecase account implementation
package usecase

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
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
func (a *AuthclientUsecase) Get(ctx context.Context, id string) (*model.AuthClient, error) {
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
func (a *AuthclientUsecase) FetchList(ctx context.Context, filter *authclient.Filter) ([]*model.AuthClient, error) {
	if filter == nil {
		filter = &authclient.Filter{}
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if !acl.HaveAccessList(ctx, &model.AuthClient{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list authclient")
	}
	list, err := a.authclientRepo.FetchList(ctx, filter)
	for _, link := range list {
		if !acl.HaveAccessList(ctx, link) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list authclient")
		}
	}
	return list, err
}

// Count of accounts by filter
func (a *AuthclientUsecase) Count(ctx context.Context, filter *authclient.Filter) (int64, error) {
	if filter == nil {
		filter = &authclient.Filter{}
	}
	if !acl.HaveAccessList(ctx, &model.AuthClient{}) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list authclient")
	}
	return a.authclientRepo.Count(ctx, filter)
}

// Create new object in database
func (a *AuthclientUsecase) Create(ctx context.Context, authclientObj *model.AuthClient) (string, error) {
	var err error
	if !acl.HaveAccessCreate(ctx, authclientObj) {
		return "", errors.Wrap(acl.ErrNoPermissions, "create authclient")
	}
	authclientObj.ID, err = a.authclientRepo.Create(ctx, authclientObj)
	return authclientObj.ID, err
}

// Update object in database
func (a *AuthclientUsecase) Update(ctx context.Context, id string, authclientObj *model.AuthClient) error {
	if !acl.HaveAccessUpdate(ctx, authclientObj) {
		return errors.Wrap(acl.ErrNoPermissions, "update authclient")
	}
	return a.authclientRepo.Update(ctx, id, authclientObj)
}

// Delete delites record by ID
func (a *AuthclientUsecase) Delete(ctx context.Context, id string) error {
	authclientObj, err := a.Get(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, authclientObj) {
		return errors.Wrap(acl.ErrNoPermissions, "delete authclient")
	}
	return a.authclientRepo.Delete(ctx, id)
}
