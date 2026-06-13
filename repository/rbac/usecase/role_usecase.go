// Package usecase provides business logic for RBAC role management
package usecase

import (
	"context"

	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/generated"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	rbacrepo "github.com/geniusrabbit/blaze-api/repository/rbac/repository"
)

// RoleUsecase provides business logic for role access control
type RoleUsecase struct {
	generated.Usecase[rbac.Role, uint64]
}

// New creates a new RoleUsecase with the provided repository
func New(repo rbac.Repository) *RoleUsecase {
	return &RoleUsecase{
		Usecase: generated.Usecase[rbac.Role, uint64]{Repo: repo},
	}
}

// NewDefault creates a new RoleUsecase with default repository
func NewDefault() *RoleUsecase {
	return New(rbacrepo.New())
}

// GetByName retrieves a role by name with access control validation
func (a *RoleUsecase) GetByName(ctx context.Context, name string) (*rbac.Role, error) {
	roleObj, err := a.Usecase.Repo.(rbac.Repository).GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, roleObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view role/permission")
	}
	return roleObj, nil
}

// FetchList retrieves a filtered list of roles with access control validation
func (a *RoleUsecase) FetchList(ctx context.Context, qops ...rbac.QOption) ([]*rbac.Role, error) {
	if !acl.HaveAccessList(ctx, &rbac.Role{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list role/permission")
	}
	list, err := a.Usecase.Repo.(rbac.Repository).
		FetchList(ctx, prepareQueryOptions(ctx, qops, `list`)...)
	for _, link := range list {
		if !acl.HaveAccessList(ctx, link) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list role/permission")
		}
	}
	return list, err
}

// Count returns the count of roles matching the filter with access control
func (a *RoleUsecase) Count(ctx context.Context, qops ...rbac.QOption) (int64, error) {
	if !acl.HaveAccessList(ctx, &rbac.Role{}) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list role/permission")
	}
	return a.Usecase.Repo.(rbac.Repository).Count(ctx, prepareQueryOptions(ctx, qops, `count`)...)
}

func prepareQueryOptions(ctx context.Context, qops []rbac.QOption, accessName string) []rbac.QOption {
	var filter *rbac.Filter
	for _, ops := range qops {
		if ops != nil {
			if f, ok := ops.(*rbac.Filter); ok {
				filter = f
				break
			}
		}
	}
	if filter == nil {
		filter = prepareFilter(ctx, filter, accessName)
		qops = append(qops, filter)
	} else {
		prepareFilter(ctx, filter, accessName)
	}
	return qops
}

// prepareFilter applies access-based filters to role queries based on user permissions
func prepareFilter(ctx context.Context, filter *rbac.Filter, accessName string) *rbac.Filter {
	if acl.HasPermission(ctx, "role."+accessName+".all") {
		return filter
	}
	if filter == nil {
		filter = &rbac.Filter{}
	}
	if acl.HasPermission(ctx, "role."+accessName+".account") {
		filter.MaxAccessLevel = rbac.AccessLevelAccount
	} else if session.User(ctx).IsAnonymous() {
		filter.MaxAccessLevel = rbac.AccessLevelBasic
	} else {
		filter.MaxAccessLevel = rbac.AccessLevelNoAnonymous
	}
	return filter
}
