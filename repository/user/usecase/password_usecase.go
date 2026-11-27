package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"errors"
	"github.com/demdxx/gocast/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var (
	ErrInvalidPasswordResetCode = errors.New(`invalid password reset code`)
	ErrInvalidCurrentPassword   = errors.New(`current password is incorrect`)
	ErrPasswordTooShort         = errors.New(`password must be at least 8 characters`)
)

// PasswordUsecase provides password business logic.
type PasswordUsecase[T user.PasswordCapableModel] struct {
	core     user.Usecase[T]
	passRepo user.PasswordRepository[T]
}

// NewPasswordUsecase creates password usecase.
func NewPasswordUsecase[T user.PasswordCapableModel](core user.Usecase[T], passRepo user.PasswordRepository[T]) user.PasswordUsecase[T] {
	return &PasswordUsecase[T]{
		core:     core,
		passRepo: passRepo,
	}
}

// Repo returns the underlying password repository.
func (a *PasswordUsecase[T]) Repo() user.PasswordRepository[T] {
	return a.passRepo
}

// SetPassword sets a new password for the user with access control check.
func (a *PasswordUsecase[T]) SetPassword(ctx context.Context, userObj T, password string) error {
	if !acl.HaveObjectPermissions(ctx, userObj, `password.set.*`) {
		return fmt.Errorf("%s: %w", `set password`, acl.ErrNoPermissions)
	}
	return a.passRepo.SetPassword(ctx, userObj, password)
}

// ChangePassword changes the password for the current user with access control check.
func (a *PasswordUsecase[T]) ChangePassword(ctx context.Context, currentPassword, newPassword string) error {
	userObj := sessionUserTyped[T](ctx)
	if gocast.IsNil(userObj) || userObj.GetID() == 0 {
		return acl.ErrNoPermissions
	}

	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	if _, err := a.passRepo.GetByPassword(ctx, userObj.GetID(), currentPassword); err != nil {
		return ErrInvalidCurrentPassword
	}

	return a.SetPassword(ctx, userObj, newPassword)
}

// ResetPassword generates a password reset token for the user with access control check.
func (a *PasswordUsecase[T]) ResetPassword(ctx context.Context, userID uint64) (*user.UserPasswordReset, T, error) {
	var zero T
	userObj, err := a.core.Get(ctx, userID)
	if err != nil {
		return nil, zero, err
	}

	if gocast.IsNil(userObj) || userObj.GetID() == 0 {
		return nil, zero, nil
	}

	if !acl.HaveObjectPermissions(ctx, userObj, `password.reset.*`) {
		return nil, zero, fmt.Errorf("%s: %w", `reset password`, acl.ErrNoPermissions)
	}

	reset, err := a.passRepo.CreateResetPassword(ctx, userObj.GetID())
	if err != nil {
		return nil, zero, err
	}

	return reset, userObj, nil
}

// UpdatePassword updates the password for the user using a reset token with access control check.
func (a *PasswordUsecase[T]) UpdatePassword(ctx context.Context, userID uint64, token, password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	userObj, err := a.core.Get(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows || err == gorm.ErrRecordNotFound {
			return ErrInvalidPasswordResetCode
		}
		return err
	}

	if gocast.IsNil(userObj) || userObj.GetID() == 0 {
		return ErrInvalidPasswordResetCode
	}

	reset, err := a.passRepo.GetResetPassword(ctx, userObj.GetID(), token)
	if err != nil {
		if err == sql.ErrNoRows || err == gorm.ErrRecordNotFound {
			return ErrInvalidPasswordResetCode
		}
		return err
	}

	if reset.ExpiresAt.Before(time.Now()) {
		// Never log the token: it is a credential
		ctxlogger.Get(ctx).Info("Reset password token expired",
			zap.Uint64("user_id", reset.UserID),
			zap.Time("expires_at", reset.ExpiresAt))
		return ErrInvalidPasswordResetCode
	}

	if err := a.passRepo.SetPassword(ctx, userObj, password); err != nil {
		return err
	}

	if err := a.passRepo.EliminateResetPassword(ctx, reset.UserID); err != nil {
		ctxlogger.Get(ctx).Error("Error eliminating reset password",
			zap.Uint64("user_id", reset.UserID), zap.Error(err))
	}

	return nil
}
