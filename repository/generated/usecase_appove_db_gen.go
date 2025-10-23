package generated

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
)

// UsecaseApprover provides approve/reject usecase methods
type UsecaseApprover[T any, TID any] struct {
	Repo RepositoryIfaceWithApprove[T, TID]
}

// Approve approves an entity by ID with ACL permission check
func (u *UsecaseApprover[T, TID]) Approve(ctx context.Context, id TID, message string) error {
	// Fetch existing entity to check permissions
	existingObj, err := u.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Check if user has approve permissions for the existing entity
	if !acl.HaveAccessApprove(ctx, existingObj) {
		return acl.ErrNoPermissions.WithMessage("approve")
	}
	return u.Repo.Approve(ctx, id, message)
}

// Reject rejects an entity by ID with ACL permission check
func (u *UsecaseApprover[T, TID]) Reject(ctx context.Context, id TID, message string) error {
	// Fetch existing entity to check permissions
	existingObj, err := u.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Check if user has reject permissions for the existing entity
	if !acl.HaveAccessApprove(ctx, existingObj) {
		return acl.ErrNoPermissions.WithMessage("reject")
	}
	return u.Repo.Reject(ctx, id, message)
}
