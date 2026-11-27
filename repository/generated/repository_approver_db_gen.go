package generated

import (
	"context"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
)

// RepositoryApprover provides methods to approve or reject entities.
// It implements RepositoryApproveIface, so it can back UsecaseApprover;
// pass historylog.Message(...) to record a comment.
type RepositoryApprover[T any, TID comparable] struct {
	repository.Repository
	IDName     string
	StatusName string
}

var _ RepositoryApproveIface[uint64] = (*RepositoryApprover[struct{}, uint64])(nil)

// Approve approves an entity by ID
func (r *RepositoryApprover[T, TID]) Approve(ctx context.Context, id TID, opts ...Option) error {
	return r.setStatus(ctx, id, pkgModels.ApprovedApproveStatus, opts...)
}

// Reject rejects an entity by ID
func (r *RepositoryApprover[T, TID]) Reject(ctx context.Context, id TID, opts ...Option) error {
	return r.setStatus(ctx, id, pkgModels.DisapprovedApproveStatus, opts...)
}

func (r *RepositoryApprover[T, TID]) setStatus(ctx context.Context, id TID, status pkgModels.ApproveStatus, opts ...Option) error {
	// The model value has no PK, so the history log takes it from the context
	ctx = historylog.WithPK(ctx, id)
	res := Options(opts).PrepareQuery(r.Master(ctx)).Model((*T)(nil)).
		Where(r.IDName+"=?", id).
		Update(r.StatusName, status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
