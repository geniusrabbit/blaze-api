package graphql

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/messanger"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userusecase "github.com/geniusrabbit/blaze-api/repository/user/usecase"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// PasswordResetQueryResolver handles password reset GraphQL mutations (email → userID).
type PasswordResetQueryResolver[T user.AuthCapableModel] struct {
	email    user.EmailUsecase[T]
	password user.PasswordUsecase[T]
}

// PasswordResetQueryResolverConfig wires password reset resolver.
type PasswordResetQueryResolverConfig[T user.AuthCapableModel] struct {
	Email    user.EmailUsecase[T]
	Password user.PasswordUsecase[T]
}

// NewPasswordResetQueryResolver returns password reset resolver.
func NewPasswordResetQueryResolver[T user.AuthCapableModel](cfg PasswordResetQueryResolverConfig[T]) *PasswordResetQueryResolver[T] {
	return &PasswordResetQueryResolver[T]{
		email:    cfg.Email,
		password: cfg.Password,
	}
}

// ResetUserPassword is the resolver for the resetUserPassword field.
func (r *PasswordResetQueryResolver[T]) ResetUserPassword(ctx context.Context, email string) (*gqlmodels.StatusResponse, error) {
	if !messanger.Get(ctx).IsEnabled() {
		ctxlogger.Get(ctx).Error("Email service not configured")
		return &gqlmodels.StatusResponse{
			ClientMutationID: requestid.Get(ctx),
			Status:           gqlmodels.ResponseStatusError,
			Message:          &[]string{"Internal service problem. Request again later."}[0],
		}, nil
	}

	// The answer is the same whether the email is registered or not, so the
	// mutation can't be used to find out which emails have accounts.
	sentResponse := &gqlmodels.StatusResponse{
		ClientMutationID: requestid.Get(ctx),
		Status:           gqlmodels.ResponseStatusSuccess,
		Message:          &[]string{"If the email is registered, a password reset link has been sent"}[0],
	}

	// The lookup itself isn't access checked: the caller (usually anonymous)
	// can't view users, the user isn't returned, and ResetPassword checks the
	// `password.reset.*` permission on it.
	userObj, err := r.email.GetByEmail(acl.WithNoPermCheck(ctx), email)
	if errors.Is(err, acl.ErrNoPermissions) {
		ctxlogger.Get(ctx).Info("Reset password: user is not accessible")
		return sentResponse, nil
	}
	if err != nil {
		return nil, err
	}
	var zero T
	if any(userObj) == any(zero) || userObj.GetID() == 0 {
		ctxlogger.Get(ctx).Info("Reset password: user not found")
		return sentResponse, nil
	}

	userID := userObj.GetID()
	pswReset, userObj, err := r.password.ResetPassword(ctx, userID)
	if errors.Is(err, acl.ErrNoPermissions) {
		ctxlogger.Get(ctx).Info("Reset password: not allowed for the user", zap.Uint64("user_id", userID))
		return sentResponse, nil
	}
	if err != nil {
		return nil, err
	}

	if pswReset != nil && pswReset.UserID > 0 {
		const msgName = "user.reset-password"
		err = messanger.Get(ctx).Send(ctx, msgName, []string{email}, map[string]any{
			"user":        userObj,
			"email":       email,
			"reset":       pswReset,
			"reset_token": pswReset.Token,
		})
		if err != nil {
			ctxlogger.Get(ctx).Error("Error sending reset password email",
				zap.String("msgname", msgName),
				zap.Error(err))
			return &gqlmodels.StatusResponse{
				ClientMutationID: requestid.Get(ctx),
				Status:           gqlmodels.ResponseStatusError,
				Message:          &[]string{"Error sending reset password email"}[0],
			}, nil
		}
	}

	return sentResponse, nil
}

// UpdateUserPassword is the resolver for the updateUserPassword field.
func (r *PasswordResetQueryResolver[T]) UpdateUserPassword(ctx context.Context, token, email, password string) (*gqlmodels.StatusResponse, error) {
	return r.UpdateResetedUserPassword(ctx, token, email, password)
}

// UpdateResetedUserPassword completes password reset with token.
func (r *PasswordResetQueryResolver[T]) UpdateResetedUserPassword(ctx context.Context, token, email, password string) (*gqlmodels.StatusResponse, error) {
	// Not access checked: the reset token is the authorization
	userObj, err := r.email.GetByEmail(acl.WithNoPermCheck(ctx), email)
	if errors.Is(err, acl.ErrNoPermissions) {
		// Same answer as for an unknown email or a wrong token
		return nil, userusecase.ErrInvalidPasswordResetCode
	}
	if err != nil {
		return nil, err
	}
	var zero T
	if any(userObj) == any(zero) || userObj.GetID() == 0 {
		return nil, userusecase.ErrInvalidPasswordResetCode
	}
	if err := r.password.UpdatePassword(ctx, userObj.GetID(), token, password); err != nil {
		return nil, err
	}
	return &gqlmodels.StatusResponse{
		ClientMutationID: requestid.Get(ctx),
		Status:           gqlmodels.ResponseStatusSuccess,
		Message:          &[]string{"Password updated"}[0],
	}, nil
}
