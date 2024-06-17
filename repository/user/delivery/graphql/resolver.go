package graphql

import (
	"context"
	"errors"
	"strings"

	"github.com/demdxx/sendmsg"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/messanger"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/repository"
	"github.com/geniusrabbit/blaze-api/repository/user/usecase"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var (
	errInvalidIDOrUsername = errors.New("invalid ID or USERNAME parameter")
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	users user.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		users: usecase.NewUserUsecase(repository.New()),
	}
}

// CurrentUser returns the current user info
func (r *QueryResolver) CurrentUser(ctx context.Context) (*gqlmodels.UserPayload, error) {
	user := session.User(ctx)
	return &gqlmodels.UserPayload{
		ClientMutationID: requestid.Get(ctx),
		UserID:           user.ID,
		User:             gqlmodels.FromUserModel(user),
	}, nil
}

// CreateUser is the resolver for the createUser field.
func (r *QueryResolver) CreateUser(ctx context.Context, input *gqlmodels.UserInput) (*gqlmodels.UserPayload, error) {
	uid, err := r.users.Store(ctx, &model.User{
		Email:   *input.Username,
		Approve: input.Status.ModelStatus(),
	}, "")
	if err != nil {
		return nil, err
	}
	user, err := r.users.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.UserPayload{
		ClientMutationID: requestid.Get(ctx),
		UserID:           user.ID,
		User:             gqlmodels.FromUserModel(user),
	}, nil
}

// UpdateUser is the resolver for the updateUser field.
func (r *QueryResolver) UpdateUser(ctx context.Context, id uint64, input *gqlmodels.UserInput) (*gqlmodels.UserPayload, error) {
	user, err := r.users.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Username != nil {
		user.Email = *input.Username
	}
	if input.Status != nil {
		user.Approve = (*gqlmodels.ApproveStatus)(input.Status).ModelStatus()
	}
	if err := r.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return &gqlmodels.UserPayload{
		ClientMutationID: requestid.Get(ctx),
		UserID:           user.ID,
		User:             gqlmodels.FromUserModel(user),
	}, nil
}

// ApproveUser is the resolver for the approveUser field.
func (r *QueryResolver) ApproveUser(ctx context.Context, id uint64, msg *string) (*gqlmodels.UserPayload, error) {
	return r.updateApproveStatus(ctx, id, model.ApprovedApproveStatus, msg)
}

// RejectUser is the resolver for the rejectUser field.
func (r *QueryResolver) RejectUser(ctx context.Context, id uint64, msg *string) (*gqlmodels.UserPayload, error) {
	return r.updateApproveStatus(ctx, id, model.DisapprovedApproveStatus, msg)
}

func (r *QueryResolver) updateApproveStatus(ctx context.Context, id uint64, status model.ApproveStatus, msg *string) (*gqlmodels.UserPayload, error) {
	user, err := r.users.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Approve = status
	if msg != nil {
		ctx = historylog.WithMessage(ctx, *msg)
	}
	err = r.users.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	if true {
		msgName := "user." + strings.ToLower(status.String())
		err = messanger.Get(ctx).Send(ctx, msgName, []string{}, map[string]any{})
		if !errors.Is(err, sendmsg.ErrTemplateNotFound) {
			ctxlogger.Get(ctx).Error("User status update",
				zap.String("msgname", msgName),
				zap.Error(err))
		} else {
			ctxlogger.Get(ctx).Info("User status update template not found",
				zap.String("msgname", msgName),
				zap.Error(err))
		}
	}
	return &gqlmodels.UserPayload{
		ClientMutationID: requestid.Get(ctx),
		UserID:           id,
		User:             gqlmodels.FromUserModel(user),
	}, nil
}

// ResetUserPassword is the resolver for the resetUserPassword field.
func (r *QueryResolver) ResetUserPassword(ctx context.Context, email string) (*gqlmodels.StatusResponse, error) {
	if !messanger.Get(ctx).IsEnabled() {
		// Messanger is required for sending emails otherwise no any sense to reset password
		ctxlogger.Get(ctx).Error("Email service not configured")
		return &gqlmodels.StatusResponse{
			ClientMutationID: requestid.Get(ctx),
			Status:           gqlmodels.ResponseStatusError,
			Message:          &[]string{"Internal service problem. Request again later."}[0],
		}, nil
	}

	pswReset, user, err := r.users.ResetPassword(ctx, email)
	if err != nil {
		return nil, err
	}

	// Send message to user about the password reset
	if pswReset != nil && pswReset.UserID > 0 {
		const msgName = "user.reset-password"
		err = messanger.Get(ctx).Send(ctx, msgName, []string{email}, map[string]any{
			"user":        user,
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
	} else {
		ctxlogger.Get(ctx).Info("User not found for reset password", zap.String("email", email))
	}

	return &gqlmodels.StatusResponse{
		ClientMutationID: requestid.Get(ctx),
		Status:           gqlmodels.ResponseStatusSuccess,
		Message:          &[]string{"Password reset link sent to " + email}[0],
	}, nil
}

// UpdateUserPassword is the resolver for the updateUserPassword field.
func (r *QueryResolver) UpdateUserPassword(ctx context.Context, token, email, password string) (*gqlmodels.StatusResponse, error) {
	return r.UpdateResetedUserPassword(ctx, token, email, password)
}

// UpdateResetedUserPassword is the resolver for the updateResetedUserPassword field
func (r *QueryResolver) UpdateResetedUserPassword(ctx context.Context, token, email, password string) (*gqlmodels.StatusResponse, error) {
	err := r.users.UpdatePassword(ctx, token, email, password)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.StatusResponse{
		ClientMutationID: requestid.Get(ctx),
		Status:           gqlmodels.ResponseStatusSuccess,
		Message:          &[]string{"Password updated"}[0],
	}, nil
}

// User user by ID or username
func (r *QueryResolver) User(ctx context.Context, id uint64, username string) (*gqlmodels.UserPayload, error) {
	var (
		err  error
		user *model.User
	)
	switch {
	case id > 0:
		user, err = r.users.Get(ctx, id)
		if err == nil && username != "" && username != user.Email {
			err = errInvalidIDOrUsername
		}
	case username != "":
		user, err = r.users.GetByEmail(ctx, username)
		if err == nil && id > 0 && id != user.ID {
			err = errInvalidIDOrUsername
		}
	default:
		err = errInvalidIDOrUsername
	}
	if err != nil {
		return nil, err
	}
	return &gqlmodels.UserPayload{
		ClientMutationID: requestid.Get(ctx),
		UserID:           user.ID,
		User:             gqlmodels.FromUserModel(user),
	}, nil
}

// ListUsers list by filter
func (r *QueryResolver) ListUsers(ctx context.Context, filter *gqlmodels.UserListFilter, order *gqlmodels.UserListOrder, page *gqlmodels.Page) (*connectors.UserConnection, error) {
	return connectors.NewUserConnection(ctx, r.users, filter, order, page), nil
}
