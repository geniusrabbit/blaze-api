package usecase

import (
	"context"

	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/models"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
)

// SessionUsecase provides business logic for AuthSession access
type SessionUsecase struct {
	repo authclient.SessionRepository
}

// NewSessionUsecase object controller
func NewSessionUsecase(repo authclient.SessionRepository) *SessionUsecase {
	return &SessionUsecase{repo: repo}
}

// Get returns the session by ID if the caller has view access
func (u *SessionUsecase) Get(ctx context.Context, id uint64, opts ...authclient.QOption) (*models.AuthSession, error) {
	sess, err := u.repo.Get(ctx, id, opts...)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, sess) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view auth session")
	}
	return sess, nil
}

// FetchList of sessions by filter
func (u *SessionUsecase) FetchList(ctx context.Context, opts ...authclient.QOption) ([]*models.AuthSession, error) {
	if !acl.HaveAccessList(ctx, &models.AuthSession{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list auth session")
	}
	list, err := u.repo.FetchList(ctx, opts...)
	for _, sess := range list {
		if !acl.HaveAccessList(ctx, sess) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list auth session")
		}
	}
	return list, err
}

// Count of sessions by filter
func (u *SessionUsecase) Count(ctx context.Context, opts ...authclient.QOption) (int64, error) {
	if !acl.HaveAccessCount(ctx, &models.AuthSession{}) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "count auth session")
	}
	return u.repo.Count(ctx, opts...)
}

// Delete removes a session by ID if the caller has delete access
func (u *SessionUsecase) Delete(ctx context.Context, id uint64, opts ...authclient.QOption) error {
	sess, err := u.Get(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, sess) {
		return errors.Wrap(acl.ErrNoPermissions, "delete auth session")
	}
	return u.repo.Delete(historylog.WithPK(ctx, id), id, opts...)
}
