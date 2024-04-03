package usecase

import (
	"context"

	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
)

// Usecase for social account
type Usecase struct {
	repo socialaccount.Repository
}

func New(repo socialaccount.Repository) *Usecase {
	return &Usecase{repo: repo}
}

// Get social account by ID
func (u *Usecase) Get(ctx context.Context, id uint64) (*model.AccountSocial, error) {
	obj, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, obj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view social account")
	}
	return obj, nil
}

// FetchList of social accounts
func (u *Usecase) FetchList(ctx context.Context, filter *socialaccount.Filter, order *socialaccount.Order, page *repository.Pagination) ([]*model.AccountSocial, error) {
	if !acl.HaveAccessList(ctx, &model.AccountSocial{}) {
		if !acl.HaveAccessList(ctx, &model.AccountSocial{UserID: session.User(ctx).ID}) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list social account")
		}
		if filter == nil {
			filter = &socialaccount.Filter{}
		}
		filter.UserID = append(filter.UserID[:0], session.User(ctx).ID)
	}
	return u.repo.FetchList(ctx, filter, order, page)
}

// Count social accounts
func (u *Usecase) Count(ctx context.Context, filter *socialaccount.Filter) (int64, error) {
	if !acl.HaveAccessCount(ctx, &model.AccountSocial{}) {
		if !acl.HaveAccessCount(ctx, &model.AccountSocial{UserID: session.User(ctx).ID}) {
			return 0, errors.Wrap(acl.ErrNoPermissions, "count social account")
		}
		if filter == nil {
			filter = &socialaccount.Filter{}
		}
		filter.UserID = append(filter.UserID[:0], session.User(ctx).ID)
	}
	return u.repo.Count(ctx, filter)
}

// Disconnect social account
func (u *Usecase) Disconnect(ctx context.Context, id uint64) (*model.AccountSocial, error) {
	obj, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessDelete(ctx, obj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "disconnect social account")
	}
	return obj, u.repo.Disconnect(ctx, id)
}

// FetchSessionList of social accounts
func (u *Usecase) FetchSessionList(ctx context.Context, socialAccountID []uint64) ([]*model.AccountSocialSession, error) {
	return u.repo.FetchSessionList(ctx, socialAccountID)
}
