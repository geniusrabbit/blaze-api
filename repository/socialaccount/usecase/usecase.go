package usecase

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
)

// Usecase for social account
type Usecase struct {
	repo socialaccount.Repository
}

func NewSocaccUsecase(repo socialaccount.Repository) *Usecase {
	return &Usecase{repo: repo}
}

// Get social account by ID
func (u *Usecase) Get(ctx context.Context, id uint64) (*models.AccountSocial, error) {
	obj, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, obj) {
		return nil, acl.ErrNoPermissions.WithMessage("view social account")
	}
	return obj, nil
}

// FetchList of social accounts
func (u *Usecase) FetchList(ctx context.Context, opts ...socialaccount.QOption) ([]*models.AccountSocial, error) {
	if !acl.HaveAccessList(ctx, &models.AccountSocial{}) {
		if !acl.HaveAccessList(ctx, &models.AccountSocial{UserID: session.User(ctx).GetID()}) {
			return nil, acl.ErrNoPermissions.WithMessage("list social account")
		}
		opts = append(opts, &socialaccount.Filter{
			UserID: []uint64{session.User(ctx).GetID()},
		})
	}
	return u.repo.FetchList(ctx, opts...)
}

// Count social accounts
func (u *Usecase) Count(ctx context.Context, opts ...socialaccount.QOption) (int64, error) {
	if !acl.HaveAccessCount(ctx, &models.AccountSocial{}) {
		if !acl.HaveAccessCount(ctx, &models.AccountSocial{UserID: session.User(ctx).GetID()}) {
			return 0, acl.ErrNoPermissions.WithMessage("count social account")
		}
		opts = append(opts, &socialaccount.Filter{
			UserID: []uint64{session.User(ctx).GetID()},
		})
	}
	return u.repo.Count(ctx, opts...)
}

// Disconnect social account
func (u *Usecase) Disconnect(ctx context.Context, id uint64) (*models.AccountSocial, error) {
	obj, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessDelete(ctx, obj) {
		return nil, acl.ErrNoPermissions.WithMessage("disconnect social account")
	}
	return obj, u.repo.Disconnect(ctx, id)
}

// FetchSessionList of social accounts
func (u *Usecase) FetchSessionList(ctx context.Context, socialAccountID []uint64) ([]*models.AccountSocialSession, error) {
	return u.repo.FetchSessionList(ctx, socialAccountID)
}
