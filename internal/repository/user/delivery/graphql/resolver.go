package graphql

import (
	"context"
	"errors"
	"fmt"

	"github.com/geniusrabbit/api-template-base/internal/context/session"
	"github.com/geniusrabbit/api-template-base/internal/repository/user"
	"github.com/geniusrabbit/api-template-base/internal/repository/user/repository"
	"github.com/geniusrabbit/api-template-base/internal/repository/user/usecase"
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
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
		UserID: user.ID,
		User:   gqlmodels.FromUserModel(user),
	}, nil
}

// CreateUser is the resolver for the createUser field.
func (r *QueryResolver) CreateUser(ctx context.Context, input *gqlmodels.UserInput) (*gqlmodels.UserPayload, error) {
	uid, err := r.users.Store(ctx, &model.User{
		Email:   *input.Username,
		Approve: gqlmodels.ApproveStatus(*input.Status).ModelStatus(),
	}, "")
	if err != nil {
		return nil, err
	}
	user, err := r.users.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.UserPayload{
		UserID: user.ID,
		User:   gqlmodels.FromUserModel(user),
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
		user.Approve = gqlmodels.ApproveStatus(*input.Status).ModelStatus()
	}
	if err := r.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return &gqlmodels.UserPayload{
		UserID: user.ID,
		User:   gqlmodels.FromUserModel(user),
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
	// user.ApproveMessage = msg
	err = r.users.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.UserPayload{
		UserID: id,
		User:   gqlmodels.FromUserModel(user),
	}, nil
}

// ResetUserPassword is the resolver for the resetUserPassword field.
func (r *QueryResolver) ResetUserPassword(ctx context.Context, id uint64) (*gqlmodels.UserPayload, error) {
	panic(fmt.Errorf("not implemented: ResetUserPassword - resetUserPassword"))
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
		UserID: user.ID,
		User:   gqlmodels.FromUserModel(user),
	}, nil
}

// ListUsers list by filter
func (r *QueryResolver) ListUsers(ctx context.Context,
	filter *gqlmodels.UserListFilter, order *gqlmodels.UserListOrder,
	page *gqlmodels.Page) (*connectors.UserConnection, error) {
	return connectors.NewUserConnection(ctx, r.users, filter, order, page), nil
}
