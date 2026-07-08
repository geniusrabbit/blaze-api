package repository

import (
	"context"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/pkg/errors"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	baseRepo "github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/models"
	"github.com/geniusrabbit/blaze-api/repository/user/password"
)

var ErrInvalidPassword = errors.New(`invalid password`)

type passwordRepository[T user.PasswordCapableModel] struct {
	baseRepo.Repository
	core     user.Repository[T]
	newModel func() T
}

// NewPasswordRepository creates password auth repository.
func NewPasswordRepository[T user.PasswordCapableModel](core user.Repository[T], newModel func() T) user.PasswordRepository[T] {
	return &passwordRepository[T]{core: core, newModel: newModel}
}

// GetByPassword retrieves user by ID and checks password.
// Returns ErrInvalidPassword if password is incorrect.
func (r *passwordRepository[T]) GetByPassword(ctx context.Context, userID uint64, plainPassword string) (T, error) {
	var zero T
	object, err := r.core.Get(ctx, userID)
	if err != nil {
		return zero, err
	}
	if gocast.IsNil(object) {
		return zero, ErrInvalidPassword
	}
	if object.GetPasswordHash() == "" || !comparePasswords(object.GetPasswordHash(), []byte(plainPassword)) {
		return zero, ErrInvalidPassword
	}
	return object, nil
}

// CreateWithPassword creates a new user with the given password.
// If the password is empty, the user will be created with an empty password hash.
func (r *passwordRepository[T]) CreateWithPassword(ctx context.Context, userObj T, pwd string) (uint64, error) {
	if pwd != "" {
		userObj.SetPasswordHash(hashAndSalt([]byte(pwd)))
	} else {
		userObj.SetPasswordHash("")
	}
	setApproveOnModel(userObj, pkgModels.UndefinedApproveStatus)
	setTimestamps(userObj)
	err := r.Master(ctx).Create(userObj).Error
	if err != nil {
		return 0, err
	}
	return userObj.GetID(), nil
}

// SetPassword sets a new password for the user and updates the user in the database.
func (r *passwordRepository[T]) SetPassword(ctx context.Context, userObj T, pwd string) error {
	userObj.SetPasswordHash(hashAndSalt([]byte(pwd)))
	return r.core.Update(ctx, userObj)
}

// SetPasswordHash sets a new password hash for the user and updates the user in the database.
func (r *passwordRepository[T]) CreateResetPassword(ctx context.Context, userID uint64) (*models.UserPasswordReset, error) {
	token := password.GenerateResetToken(128)
	reset := &models.UserPasswordReset{
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := r.Master(ctx).Create(reset).Error; err != nil {
		return nil, err
	}
	return reset, nil
}

// GetResetPassword retrieves a password reset record by user ID and token.
func (r *passwordRepository[T]) GetResetPassword(ctx context.Context, userID uint64, token string) (*models.UserPasswordReset, error) {
	reset := new(models.UserPasswordReset)
	if err := r.Slave(ctx).First(reset, `token=? AND user_id=?`, token, userID).Error; err != nil {
		return nil, err
	}
	return reset, nil
}

// EliminateResetPassword deletes a password reset record by user ID.
func (r *passwordRepository[T]) EliminateResetPassword(ctx context.Context, userID uint64) error {
	return r.Master(ctx).Delete(&models.UserPasswordReset{}, `user_id=?`, userID).Error
}

type timestampModel interface {
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

func setTimestamps(obj any) {
	now := time.Now()
	if v, ok := obj.(timestampModel); ok {
		v.SetCreatedAt(now)
		v.SetUpdatedAt(now)
	}
}
