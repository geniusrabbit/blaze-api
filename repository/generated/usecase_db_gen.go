package generated

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/go-faster/errors"
)

// Usecase provides a generic business logic layer with ACL (Access Control List) support
// for CRUD operations on entities of type T with ID type TID.
type Usecase[T any, TID any] struct {
	Repo RepositoryIface[T, TID] // Repository interface for data access operations
}

// Get retrieves a single entity by ID with ACL permission check.
// Returns the entity if found and user has read permissions, otherwise returns an error.
func (u *Usecase[T, TID]) Get(ctx context.Context, id TID, qops ...Option) (*T, error) {
	// Fetch the entity from repository
	targetObj, err := u.Repo.Get(ctx, id, qops...)
	if err != nil {
		return nil, err
	}

	// Check if user has read permissions for this specific object
	if !acl.HaveObjectPermissions(ctx, targetObj, acl.PermGet+`.*`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "get")
	}
	return targetObj, nil
}

// FetchList retrieves a list of entities with ACL permission checks.
// Validates both general list access and individual object access permissions.
func (u *Usecase[T, TID]) FetchList(ctx context.Context, qops ...Option) ([]*T, error) {
	// Check if user has general list access permission for this entity type
	if !acl.HaveAccessList(ctx, new(T)) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list")
	}

	// Fetch the list from repository
	list, err := u.Repo.FetchList(ctx, qops...)

	// Verify access permissions for each individual object in the list
	for _, obj := range list {
		if !acl.HaveAccessList(ctx, obj) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list")
		}
	}
	return list, err
}

// Count returns the total number of entities with ACL permission check.
// Only counts entities the user has permission to list.
func (u *Usecase[T, TID]) Count(ctx context.Context, qops ...Option) (int64, error) {
	// Check if user has list access permission for this entity type
	if !acl.HaveAccessList(ctx, new(T)) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list")
	}
	return u.Repo.Count(ctx, qops...)
}

// Create creates a new entity with ACL permission check.
// Returns the ID of the created entity if successful.
func (u *Usecase[T, TID]) Create(ctx context.Context, obj *T, message string) (id TID, err error) {
	// Check if user has create permissions for this entity
	if !acl.HaveAccessCreate(ctx, obj) {
		return id, acl.ErrNoPermissions.WithMessage("create")
	}
	return u.Repo.Create(ctx, obj, message)
}

// Update modifies an existing entity with ACL permission check.
// Fetches the existing entity first to verify update permissions.
func (u *Usecase[T, TID]) Update(ctx context.Context, id TID, obj *T, message string) error {
	// Fetch existing entity to check permissions
	existingObj, err := u.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Check if user has update permissions for the existing entity
	if !acl.HaveAccessUpdate(ctx, existingObj) {
		return acl.ErrNoPermissions.WithMessage("update")
	}
	return u.Repo.Update(ctx, id, obj, message)
}

// Delete removes an entity with ACL permission check.
// Fetches the existing entity first to verify delete permissions.
func (u *Usecase[T, TID]) Delete(ctx context.Context, id TID, message string) error {
	// Fetch existing entity to check permissions
	existingObj, err := u.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Check if user has delete permissions for the existing entity
	if !acl.HaveAccessDelete(ctx, existingObj) {
		return acl.ErrNoPermissions.WithMessage("delete")
	}
	return u.Repo.Delete(ctx, id, message)
}
