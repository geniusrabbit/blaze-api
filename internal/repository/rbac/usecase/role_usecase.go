// Package usecase account implementation
package usecase

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/acl"
	"github.com/geniusrabbit/api-template-base/internal/repository/rbac"
	"github.com/geniusrabbit/api-template-base/model"
	"github.com/pkg/errors"
)

// RoleUsecase provides bussiness logic for account access
type RoleUsecase struct {
	roleRepo rbac.Repository
}

// NewRoleUsecase object controller
func NewRoleUsecase(repo rbac.Repository) *RoleUsecase {
	return &RoleUsecase{
		roleRepo: repo,
	}
}

// Get returns the group by ID if have access
func (a *RoleUsecase) Get(ctx context.Context, id uint64) (*model.Role, error) {
	roleObj, err := a.roleRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, roleObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view role/permission")
	}
	return roleObj, nil
}

// GetByName returns the role by name if have access
func (a *RoleUsecase) GetByName(ctx context.Context, name string) (*model.Role, error) {
	roleObj, err := a.roleRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, roleObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view role/permission")
	}
	return roleObj, nil
}

// FetchList of accounts by filter
func (a *RoleUsecase) FetchList(ctx context.Context, filter *rbac.Filter) ([]*model.Role, error) {
	if filter == nil {
		filter = &rbac.Filter{}
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if !acl.HaveAccessList(ctx, &model.Role{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list role/permission")
	}
	list, err := a.roleRepo.FetchList(ctx, filter)
	for _, link := range list {
		if !acl.HaveAccessList(ctx, link) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list role/permission")
		}
	}
	return list, err
}

// Create new object in database
func (a *RoleUsecase) Create(ctx context.Context, roleObj *model.Role) (uint64, error) {
	var err error
	if !acl.HaveAccessCreate(ctx, roleObj) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "create role/permission")
	}
	roleObj.ID, err = a.roleRepo.Create(ctx, roleObj)
	return roleObj.ID, err
}

// Update object in database
func (a *RoleUsecase) Update(ctx context.Context, id uint64, roleObj *model.Role) error {
	upRoleObj := *roleObj
	upRoleObj.ID = id
	if !acl.HaveAccessUpdate(ctx, upRoleObj) {
		return errors.Wrap(acl.ErrNoPermissions, "update role/permission")
	}
	return a.roleRepo.Update(ctx, id, &upRoleObj)
}

// Delete delites record by ID
func (a *RoleUsecase) Delete(ctx context.Context, id uint64) error {
	roleObj, err := a.roleRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, roleObj) {
		return errors.Wrap(acl.ErrNoPermissions, "delete role/permission")
	}
	return a.roleRepo.Delete(ctx, id)
}
