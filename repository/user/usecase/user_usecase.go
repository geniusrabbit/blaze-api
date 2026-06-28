// Package usecase user managing
package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var (
	ErrInvalidPasswordResetCode = errors.New(`invalid password reset code`)
	ErrInvalidCurrentPassword   = errors.New(`current password is incorrect`)
	ErrPasswordTooShort         = errors.New(`password must be at least 8 characters`)
)

// UserUsecase provides bussiness logic for user access
type UserUsecase struct {
	userRepo user.Repository
}

// NewUserUsecase user implementation
func NewUserUsecase(repo user.Repository) *UserUsecase {
	return &UserUsecase{userRepo: repo}
}

// Get returns the group by ID if have access
func (a *UserUsecase) Get(ctx context.Context, id uint64) (*user.User, error) {
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
func (a *UserUsecase) GetByEmail(ctx context.Context, email string) (*user.User, error) {
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
func (a *UserUsecase) GetByPassword(ctx context.Context, email, password string) (*user.User, error) {
	return a.userRepo.GetByPassword(ctx, email, password)
}

// FetchList of users by filter
func (a *UserUsecase) FetchList(ctx context.Context, opts ...user.QOption) ([]*user.User, error) {
	if !acl.HaveAccessList(ctx, &user.User{}) {
		if !acl.HaveAccessList(ctx, session.User(ctx)) {
			return nil, acl.ErrNoPermissions
		}
		if err := adjustFetchListPermissions(ctx, opts...); err != nil {
			return nil, err
		}
	}
	return a.userRepo.FetchList(ctx, opts...)
}

// Count of users by filter
func (a *UserUsecase) Count(ctx context.Context, opts ...user.QOption) (int64, error) {
	if !acl.HaveAccessCount(ctx, &user.User{}) {
		if !acl.HaveAccessCount(ctx, session.User(ctx)) {
			return 0, acl.ErrNoPermissions
		}
		if err := adjustFetchListPermissions(ctx, opts...); err != nil {
			return 0, err
		}
	}
	return a.userRepo.Count(ctx, opts...)
}

// SetPassword for the exists user
func (a *UserUsecase) SetPassword(ctx context.Context, userObj *user.User, password string) error {
	if !acl.HaveObjectPermissions(ctx, userObj, `password.set.*`) {
		return errors.Wrap(acl.ErrNoPermissions, `set password`)
	}
	return a.userRepo.SetPassword(ctx, userObj, password)
}

// ChangePassword for the current session user
func (a *UserUsecase) ChangePassword(ctx context.Context, currentPassword, newPassword string) error {
	userObj := session.User(ctx)
	if userObj == nil || userObj.ID == 0 {
		return acl.ErrNoPermissions
	}
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}
	if _, err := a.userRepo.GetByPassword(ctx, userObj.Email, currentPassword); err != nil {
		return ErrInvalidCurrentPassword
	}
	return a.SetPassword(ctx, userObj, newPassword)
}

// ResetPassword for the exists user
func (a *UserUsecase) ResetPassword(ctx context.Context, email string) (*user.UserPasswordReset, *user.User, error) {
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

// Create new object into database
func (a *UserUsecase) Create(ctx context.Context, userObj *user.User, password string) (uint64, error) {
	if !acl.HaveAccessCreate(ctx, userObj) {
		return 0, acl.ErrNoPermissions
	}
	return a.userRepo.Create(ctx, userObj, password)
}

// Update existing object in database
func (a *UserUsecase) Update(ctx context.Context, userObj *user.User) error {
	if !acl.HaveAccessUpdate(ctx, userObj) {
		return acl.ErrNoPermissions
	}
	return a.userRepo.Update(historylog.WithPK(ctx, userObj.ID), userObj)
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
	return a.userRepo.Delete(historylog.WithPK(ctx, id), id)
}

func (a *UserUsecase) getUserByID(ctx context.Context, id uint64) (*user.User, error) {
	currentUser := session.User(ctx)
	if currentUser.ID == id {
		return currentUser, nil
	}
	return nil, sql.ErrNoRows
}

func adjustFetchListPermissions(ctx context.Context, opts ...user.QOption) error {
	adjusted := false
	for _, opt := range opts {
		if adjuster, ok := opt.(repository.QueryPermissionAdjuster); ok {
			if err := adjuster.AdjustPermissions(ctx); err != nil {
				return err
			}
			adjusted = true
		}
	}
	if !adjusted {
		return acl.ErrNoPermissions
	}
	return nil
}
