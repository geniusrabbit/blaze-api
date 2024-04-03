// Package usecase user managing
package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var ErrInvalidPasswordResetCode = errors.New(`invalid password reset code`)

// UserUsecase provides bussiness logic for user access
type UserUsecase struct {
	userRepo user.Repository
}

// NewUserUsecase user implementation
func NewUserUsecase(repo user.Repository) *UserUsecase {
	return &UserUsecase{userRepo: repo}
}

// Get returns the group by ID if have access
func (a *UserUsecase) Get(ctx context.Context, id uint64) (*model.User, error) {
	currentUser, _ := session.UserAccount(ctx)
	if currentUser.ID == id {
		if !acl.HaveAccessView(ctx, currentUser) {
			return nil, acl.ErrNoPermissions
		}
		return currentUser, nil
	}
	userObj, err := a.userRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, userObj) {
		return nil, acl.ErrNoPermissions
	}
	return userObj, nil
}

// GetByEmail returns the group by Email if have access
func (a *UserUsecase) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	currentUser, _ := session.UserAccount(ctx)
	if currentUser.Email == email {
		if !acl.HaveAccessView(ctx, currentUser) {
			return nil, acl.ErrNoPermissions
		}
		return currentUser, nil
	}
	userObj, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, userObj) {
		return nil, acl.ErrNoPermissions
	}
	return userObj, nil
}

// GetByPassword returns user by email + password
func (a *UserUsecase) GetByPassword(ctx context.Context, email, password string) (*model.User, error) {
	return a.userRepo.GetByPassword(ctx, email, password)
}

// GetByToken returns user + account by session token
func (a *UserUsecase) GetByToken(ctx context.Context, token string) (*model.User, *model.Account, error) {
	return a.userRepo.GetByToken(ctx, token)
}

// FetchList of users by filter
func (a *UserUsecase) FetchList(ctx context.Context, filter *user.ListFilter, order *user.ListOrder, page *repository.Pagination) ([]*model.User, error) {
	if !acl.HaveAccessList(ctx, &model.User{}) {
		if !acl.HaveAccessList(ctx, session.User(ctx)) {
			return nil, acl.ErrNoPermissions
		}
		if filter == nil {
			filter = &user.ListFilter{}
		}
		filter.AccountID = []uint64{session.Account(ctx).ID}
	}
	return a.userRepo.FetchList(ctx, filter, order, page)
}

// Count of users by filter
func (a *UserUsecase) Count(ctx context.Context, filter *user.ListFilter) (int64, error) {
	if !acl.HaveAccessCount(ctx, &model.User{}) {
		if !acl.HaveAccessCount(ctx, session.User(ctx)) {
			return 0, acl.ErrNoPermissions
		}
		if filter == nil {
			filter = &user.ListFilter{}
		}
		filter.AccountID = []uint64{session.Account(ctx).ID}
	}
	return a.userRepo.Count(ctx, filter)
}

// SetPassword for the exists user
func (a *UserUsecase) SetPassword(ctx context.Context, userObj *model.User, password string) error {
	if !acl.HaveObjectPermissions(ctx, userObj, `password.set.*`) {
		return errors.Wrap(acl.ErrNoPermissions, `set password`)
	}
	return a.userRepo.SetPassword(ctx, userObj, password)
}

// ResetPassword for the exists user
func (a *UserUsecase) ResetPassword(ctx context.Context, email string) (*model.UserPasswordReset, *model.User, error) {
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows || err == gorm.ErrRecordNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	if !acl.HaveObjectPermissions(ctx, user, `password.reset.*`) {
		return nil, nil, errors.Wrap(acl.ErrNoPermissions, `reset password`)
	}
	reset, err := a.userRepo.CreateResetPassword(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	return reset, user, nil
}

// UpdatePassword for the exists user from reset token
func (a *UserUsecase) UpdatePassword(ctx context.Context, token, email, password string) error {
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows || err == gorm.ErrRecordNotFound {
			return ErrInvalidPasswordResetCode
		}
		return err
	}
	reset, err := a.userRepo.GetResetPassword(ctx, user.ID, token)
	if err != nil {
		if err == sql.ErrNoRows || err == gorm.ErrRecordNotFound {
			return ErrInvalidPasswordResetCode
		}
		return err
	}
	if reset.ExpiresAt.Before(time.Now()) {
		ctxlogger.Get(ctx).Info("Reset password token expired",
			zap.Uint64("user_id", reset.UserID),
			zap.Time("expires_at", reset.ExpiresAt),
			zap.String("token", token))
		return ErrInvalidPasswordResetCode
	}
	if err := a.userRepo.SetPassword(ctx, user, password); err != nil {
		return err
	}
	// Eliminate all reset password tokens for this user
	if err := a.userRepo.EliminateResetPassword(ctx, reset.UserID); err != nil {
		ctxlogger.Get(ctx).Error("Error eliminating reset password",
			zap.Uint64("user_id", reset.UserID), zap.Error(err))
	}
	return nil
}

// Store new object into database
func (a *UserUsecase) Store(ctx context.Context, userObj *model.User, password string) (uint64, error) {
	var err error
	if userObj.ID == 0 && !acl.HaveAccessCreate(ctx, userObj) {
		return 0, acl.ErrNoPermissions
	}
	if userObj.ID != 0 && !acl.HaveAccessUpdate(ctx, userObj) {
		return 0, acl.ErrNoPermissions
	}
	if userObj.ID == 0 {
		userObj.ID, err = a.userRepo.Create(ctx, userObj, password)
	} else {
		err = a.userRepo.Update(ctx, userObj)
	}
	return userObj.ID, err
}

// Update existing object in database
func (a *UserUsecase) Update(ctx context.Context, userObj *model.User) error {
	if !acl.HaveAccessUpdate(ctx, userObj) {
		return acl.ErrNoPermissions
	}
	return a.userRepo.Update(ctx, userObj)
}

// Delete delites record by ID
func (a *UserUsecase) Delete(ctx context.Context, id uint64) error {
	userObj, err := a.getUserByID(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, userObj) {
		return acl.ErrNoPermissions
	}
	return a.userRepo.Delete(ctx, id)
}

func (a *UserUsecase) getUserByID(ctx context.Context, id uint64) (*model.User, error) {
	currentUser := session.User(ctx)
	if currentUser.ID == id {
		return currentUser, nil
	}
	return nil, sql.ErrNoRows
}
