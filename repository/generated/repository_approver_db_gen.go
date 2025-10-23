package generated

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
)

// RepositoryApprover provides methods to approve or reject entities
type RepositoryApprover[T any, TID any] struct {
	repository.Repository
	IDName     string
	StatusName string
}

// Approve approves an entity by ID
func (r *RepositoryApprover[T, TID]) Approve(ctx context.Context, id TID, message string) error {
	return r.Master(
		historylog.WithMessage(ctx, message),
	).Model((*T)(nil)).
		Where(r.IDName+"=?", id).
		Update(r.StatusName, model.ApprovedApproveStatus).Error
}

// Reject rejects an entity by ID
func (r *RepositoryApprover[T, TID]) Reject(ctx context.Context, id TID, message string) error {
	return r.Master(
		historylog.WithMessage(ctx, message),
	).Model((*T)(nil)).
		Where(r.IDName+"=?", id).
		Update(r.StatusName, model.DisapprovedApproveStatus).Error
}
