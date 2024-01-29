package graphql

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/context/permissionmanager"
	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/permissions"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	"github.com/geniusrabbit/blaze-api/repository/rbac/repository"
	"github.com/geniusrabbit/blaze-api/repository/rbac/usecase"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var (
	ErrUndefinedPermissionKey = errors.New("undefined permission key")
	ErrInvalidTargetValue     = errors.New("invalid target value")
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	roles rbac.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		roles: usecase.NewRoleUsecase(repository.New()),
	}
}

// Role is the resolver for the Role field.
func (r *QueryResolver) Role(ctx context.Context, id uint64) (*gqlmodels.RBACRolePayload, error) {
	role, err := r.roles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.RBACRolePayload{
		RoleID: role.ID,
		Role:   gqlmodels.FromRBACRoleModel(role),
	}, nil
}

// Check the permission for the given role and target object
func (r *QueryResolver) Check(ctx context.Context, name string, key *string, targetID *string) (*string, error) {
	var obj any
	if key != nil {
		obj = permissionmanager.Get(ctx).ObjectByName(*key)
		if obj == nil {
			return nil, errors.Wrap(ErrUndefinedPermissionKey, *key)
		}
		if targetID != nil && *targetID != "" {
			if id, _ := gocast.StructFieldValue(obj, "ID"); id != nil {
				var err error
				switch id.(type) {
				case uint64:
					err = gocast.SetStructFieldValue(ctx, obj, "ID", gocast.Uint64(*targetID))
				case int64:
					err = gocast.SetStructFieldValue(ctx, obj, "ID", gocast.Int64(*targetID))
				case string:
					err = gocast.SetStructFieldValue(ctx, obj, "ID", *targetID)
				case uuid.UUID:
					uid, uerr := uuid.Parse(*targetID)
					if uerr != nil {
						return nil, uerr
					}
					err = gocast.SetStructFieldValue(ctx, obj, "ID", uid)
				default:
					return nil, errors.Wrap(ErrInvalidTargetValue, *targetID)
				}
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.Wrap(ErrInvalidTargetValue, *targetID)
			}
		}
	}

	perm := session.Account(ctx).CheckedPermissions(ctx, obj, name)
	if perm != nil {
		switch ext := perm.Ext().(type) {
		case nil:
		case *permissions.ExtData:
			return &[]string{ext.Cover}[0], nil
		}
		return &[]string{"user"}[0], nil
	}
	return nil, nil
}

// ListRoles is the resolver for the listRoles field.
func (r *QueryResolver) ListRoles(ctx context.Context, filter *gqlmodels.RBACRoleListFilter, order *gqlmodels.RBACRoleListOrder, page *gqlmodels.Page) (*connectors.RBACRoleConnection, error) {
	return connectors.NewRBACRoleConnection(ctx, r.roles, filter, order, page), nil
}

// CreateRole is the resolver for the createRole field.
func (r *QueryResolver) CreateRole(ctx context.Context, input *gqlmodels.RBACRoleInput) (*gqlmodels.RBACRolePayload, error) {
	roleObj := &model.Role{
		ParentID: null.IntFromPtr(int64Ptr(input.ParentID)),
		Name:     valOrDef(input.Name, ""),
		Title:    valOrDef(input.Title, ""),
		Type:     model.RoleType(input.Type),
	}
	if input.Context != nil {
		if err := roleObj.Context.SetValue(input.Context.Data); err != nil {
			return nil, err
		}
	}
	id, err := r.roles.Create(ctx, roleObj)
	if err != nil {
		return nil, err
	}
	// role, err := r.roles.Get(ctx, id)
	// if err != nil {
	// 	return nil, err
	// }
	return &gqlmodels.RBACRolePayload{
		RoleID: id,
		Role:   gqlmodels.FromRBACRoleModel(roleObj),
	}, nil
}

// UpdateRole is the resolver for the updateRole field.
func (r *QueryResolver) UpdateRole(ctx context.Context, id uint64, input *gqlmodels.RBACRoleInput) (*gqlmodels.RBACRolePayload, error) {
	role, err := r.roles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// Update object fields
	role.ParentID = null.IntFrom(valOrDef(int64Ptr(input.ParentID), role.ParentID.Int64))
	role.Name = valOrDef(input.Name, role.Name)
	role.Title = valOrDef(input.Title, role.Title)
	role.Type = model.RoleType(input.Type)
	if input.Context != nil {
		if err := role.Context.SetValue(input.Context.Data); err != nil {
			return nil, err
		}
	}
	if err := r.roles.Update(ctx, id, role); err != nil {
		return nil, err
	}
	return &gqlmodels.RBACRolePayload{
		RoleID: id,
		Role:   gqlmodels.FromRBACRoleModel(role),
	}, nil
}

// DeleteRole is the resolver for the deleteRole field.
func (r *QueryResolver) DeleteRole(ctx context.Context, id uint64, msg *string) (*gqlmodels.RBACRolePayload, error) {
	err := r.roles.Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	role, err := r.roles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.RBACRolePayload{
		RoleID: id,
		Role:   gqlmodels.FromRBACRoleModel(role),
	}, nil
}

func valOrDef[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}

func int64Ptr(v *uint64) *int64 {
	if v == nil {
		return nil
	}
	u := int64(*v)
	return &u
}
