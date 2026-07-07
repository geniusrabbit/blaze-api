package usecase

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// EmailUsecase provides email lookup business logic.
type EmailUsecase[T user.EmailCapableModel] struct {
	emailRepo user.EmailRepository[T]
}

// NewEmailUsecase creates email usecase.
func NewEmailUsecase[T user.EmailCapableModel](emailRepo user.EmailRepository[T], _ func() T) user.EmailUsecase[T] {
	return &EmailUsecase[T]{emailRepo: emailRepo}
}

// GetByEmail retrieves user by email (case-insensitive) with access control check.
func (a *EmailUsecase[T]) GetByEmail(ctx context.Context, email string) (T, error) {
	var zero T
	currentEmail, ok := sessionUserEmail(ctx)
	if ok && currentEmail == email {
		if u := sessionUserTyped[T](ctx); any(u) != any(zero) {
			if !acl.HaveAccessView(ctx, u) {
				return zero, acl.ErrNoPermissions
			}
			return u, nil
		}
	}
	userObj, err := a.emailRepo.GetByEmail(ctx, email)
	if err != nil {
		return zero, err
	}
	if any(userObj) != any(zero) && !acl.HaveAccessView(ctx, userObj) {
		return zero, acl.ErrNoPermissions
	}
	return userObj, nil
}

func sessionUserEmail(ctx context.Context) (string, bool) {
	u := sessionUserModel(ctx)
	if u == nil {
		return "", false
	}
	if e, ok := u.(user.EmailModel); ok {
		return e.GetEmail(), true
	}
	return "", false
}

func sessionUserTyped[T user.Model](ctx context.Context) T {
	var zero T
	u := sessionUserModel(ctx)
	if u == nil {
		return zero
	}
	if typed, ok := u.(T); ok {
		return typed
	}
	return zero
}
