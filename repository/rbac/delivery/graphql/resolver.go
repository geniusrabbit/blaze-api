package graphql

import (
	"context"
	"strings"

	"github.com/demdxx/gocast/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	rbacGen "github.com/geniusrabbit/blaze-api/repository/rbac"
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
	roles rbacGen.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		roles: usecase.New(repository.New()),
	}
}

// Role is the resolver for the Role field.
func (r *QueryResolver) Role(ctx context.Context, id uint64) (*gqlmodels.RBACRolePayload, error) {
	role, err := r.roles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.RBACRolePayload{
		ClientMutationID: requestid.Get(ctx),
		RoleID:           role.ID,
		Role:             gqlmodels.FromRBACRoleModel(ctx, role),
	}, nil
}

// Check the permission for the given role and target object
func (r *QueryResolver) Check(ctx context.Context, name string, key, targetID, idKey *string) (*string, error) {
	var obj any
	if key != nil {
		obj = permissions.FromContext(ctx).ObjectByName(*key)
		if obj == nil {
			ctxlogger.Get(ctx).Error("undefined permission key request", zap.String("key", *key))
			return nil, nil
		}
		obj = ownedObject(ctx, obj, session.User(ctx), session.Account(ctx))
		if targetID != nil && *targetID != "" {
			var err error
			key := "ID"
			if idKey != nil && *idKey != "" {
				key = *idKey
			}
			if id, _ := gocast.StructFieldValue(obj, key); id != nil {
				switch id.(type) {
				case uint64:
					err = gocast.SetStructFieldValue(ctx, obj, key, gocast.Uint64(*targetID))
				case int64:
					err = gocast.SetStructFieldValue(ctx, obj, key, gocast.Int64(*targetID))
				case string:
					err = gocast.SetStructFieldValue(ctx, obj, key, *targetID)
				case uuid.UUID:
					uid, uerr := uuid.Parse(*targetID)
					if uerr != nil {
						return nil, uerr
					}
					err = gocast.SetStructFieldValue(ctx, obj, key, uid)
				default:
					return nil, errors.Wrap(ErrInvalidTargetValue, *targetID)
				}
			} else {
				return nil, errors.Wrap(ErrInvalidTargetValue, *targetID)
			}
			if err != nil {
				return nil, err
			}
		}
	}

	perm := session.Account(ctx).CheckedPermissions(ctx, obj, name)
	if perm != nil {
		if strings.HasSuffix(perm.Name(), ".all") || strings.HasSuffix(perm.Name(), ".system") {
			return &[]string{"system"}[0], nil
		} else if strings.HasSuffix(perm.Name(), ".account") {
			return &[]string{"account"}[0], nil
		} else if strings.HasSuffix(perm.Name(), ".owner") {
			return &[]string{"owner"}[0], nil
		} else {
			return &[]string{"general"}[0], nil
		}
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
		Name:  gocast.PtrAsValue(input.Name, ""),
		Title: gocast.PtrAsValue(input.Title, ""),
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
		ClientMutationID: requestid.Get(ctx),
		RoleID:           id,
		Role:             gqlmodels.FromRBACRoleModel(ctx, roleObj),
	}, nil
}

// UpdateRole is the resolver for the updateRole field.
func (r *QueryResolver) UpdateRole(ctx context.Context, id uint64, input *gqlmodels.RBACRoleInput) (*gqlmodels.RBACRolePayload, error) {
	role, err := r.roles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// Update object fields
	role.Name = gocast.PtrAsValue(input.Name, role.Name)
	role.Title = gocast.PtrAsValue(input.Title, role.Title)
	if input.Context != nil {
		if err := role.Context.SetValue(input.Context.Data); err != nil {
			return nil, err
		}
	}
	if err := r.roles.Update(ctx, id, role); err != nil {
		return nil, err
	}
	return &gqlmodels.RBACRolePayload{
		ClientMutationID: requestid.Get(ctx),
		RoleID:           id,
		Role:             gqlmodels.FromRBACRoleModel(ctx, role),
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
		ClientMutationID: requestid.Get(ctx),
		RoleID:           id,
		Role:             gqlmodels.FromRBACRoleModel(ctx, role),
	}, nil
}

// ListPermissions is the resolver for the listPermissions field.
func (r *QueryResolver) ListPermissions(ctx context.Context, patterns []string) ([]*gqlmodels.RBACPermission, error) {
	list := permissions.FromContext(ctx).Permissions(patterns...)
	return gqlmodels.FromRBACPermissionModelList(list), nil
}

func (r *QueryResolver) ListMyPermissions(ctx context.Context, patterns []string) ([]*gqlmodels.RBACPermission, error) {
	list := session.Account(ctx).ListPermissions()
	return gqlmodels.FromRBACPermissionModelList(list), nil
}
