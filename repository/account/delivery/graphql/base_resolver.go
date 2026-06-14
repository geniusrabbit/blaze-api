package graphql

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/messanger"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user"
	usergql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var (
	ErrAccountIDRequired = errors.New("account id is required")
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	userRepo user.Repository
	accounts account.Usecase
	members  account.MemberUsecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver(accounts account.Usecase, members account.MemberUsecase, userRepo user.Repository) *QueryResolver {
	return &QueryResolver{
		userRepo: userRepo,
		accounts: accounts,
		members:  members,
	}
}

// CurrentAccount returns the current account info
func (r *QueryResolver) CurrentAccount(ctx context.Context) (*gqlmodels.AccountPayload, error) {
	account := session.Account(ctx)
	return &gqlmodels.AccountPayload{
		ClientMutationID: requestid.Get(ctx),
		AccountID:        account.ID,
		Account:          FromAccountModel(account),
	}, nil
}

// Account returns the account info
func (r *QueryResolver) Account(ctx context.Context, id uint64) (*gqlmodels.AccountPayload, error) {
	acc, err := r.accounts.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.AccountPayload{
		ClientMutationID: requestid.Get(ctx),
		AccountID:        id,
		Account:          FromAccountModel(acc),
	}, nil
}

// RegisterAccount creates a new account
func (r *QueryResolver) RegisterAccount(ctx context.Context, input *gqlmodels.AccountCreateInput) (*gqlmodels.AccountCreatePayload, error) {
	if (input.OwnerID == nil || *input.OwnerID == 0) && input.Owner == nil {
		return nil, errors.New("owner is required")
	}

	if input.Owner != nil && input.Password == "" {
		return nil, errors.New("password is required")
	}

	var (
		userObj = input.Owner.Model(pkgModels.UndefinedApproveStatus)
		accObj  = input.Account.Model(pkgModels.UndefinedApproveStatus)
	)

	if input.OwnerID != nil && *input.OwnerID > 0 {
		if userObj != nil {
			userObj.ID = *input.OwnerID
		} else {
			userObj = &userModels.User{ID: *input.OwnerID}
		}
	}

	if _, err := r.accounts.Register(ctx, userObj, accObj, input.Password); err != nil {
		return nil, err
	} else {
		userObj, _ = r.userRepo.Get(ctx, userObj.ID)
	}

	// Send message to the account owner about the account creation (welcome message)
	err := messanger.Get(ctx).Send(ctx, "account.register",
		[]string{userObj.Email}, map[string]any{
			"id":      accObj.ID,
			"account": accObj,
			"owner":   userObj,
		})
	if err != nil {
		// Log error if message sending failed but do not return error to the client
		ctxlogger.Get(ctx).Error("Failed to send message",
			zap.String("template", "account.register"),
			zap.Error(err))
	}

	return &gqlmodels.AccountCreatePayload{
		ClientMutationID: requestid.Get(ctx),
		Account:          FromAccountModel(accObj),
		Owner:            usergql.FromUserModel(userObj),
	}, nil
}

// UpdateAccount is the resolver for the updateAccount field.
func (r *QueryResolver) UpdateAccount(ctx context.Context, id uint64, input *gqlmodels.AccountInput) (*gqlmodels.AccountPayload, error) {
	if id == 0 {
		return nil, ErrAccountIDRequired
	}
	return r.createUpdateAccount(ctx, id, input)
}

func (r *QueryResolver) createUpdateAccount(ctx context.Context, id uint64, input *gqlmodels.AccountInput) (*gqlmodels.AccountPayload, error) {
	accModel := input.Model()
	accModel.ID = id
	id, err := r.accounts.Store(ctx, accModel)
	if err != nil {
		return nil, err
	}
	acc, err := r.accounts.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.AccountPayload{
		ClientMutationID: requestid.Get(ctx),
		AccountID:        id,
		Account:          FromAccountModel(acc),
	}, nil
}

// ApproveAccount is the resolver for the approveAccount field.
func (r *QueryResolver) ApproveAccount(ctx context.Context, id uint64, msg string) (*gqlmodels.AccountPayload, error) {
	return r.updateApproveStatus(ctx, id, pkgModels.ApprovedApproveStatus, msg)
}

// RejectAccount is the resolver for the rejectAccount field.
func (r *QueryResolver) RejectAccount(ctx context.Context, id uint64, msg string) (*gqlmodels.AccountPayload, error) {
	return r.updateApproveStatus(ctx, id, pkgModels.DisapprovedApproveStatus, msg)
}

func (r *QueryResolver) updateApproveStatus(ctx context.Context, id uint64, status pkgModels.ApproveStatus, msg string) (*gqlmodels.AccountPayload, error) {
	acc, err := r.accounts.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	acc.Approve = status
	saveCtx := historylog.WithMessage(ctx, msg)
	saveCtx = historylog.WithAction(saveCtx, strings.ToLower(status.String()))

	// Store the updated account
	if _, err = r.accounts.Store(saveCtx, acc); err != nil {
		return nil, err
	}

	// Get account owner
	members, err := r.members.FetchListMembers(ctx,
		&account.MemberFilter{AccountID: []uint64{acc.ID}}, nil, nil)
	if err != nil {
		return nil, err
	}

	recipients := make([]string, 0, len(members))
	for _, member := range members {
		if member.IsAdmin {
			recipients = append(recipients, member.User.Email)
		}
	}

	// Send message to the account owner about the account creation (welcome message)
	tmplName := "account." + strings.ToLower(status.String())
	err = messanger.Get(ctx).Send(ctx, tmplName, recipients, map[string]any{
		"id":      id,
		"account": acc,
		"status":  status,
	})
	if err != nil {
		ctxlogger.Get(ctx).Error("Failed to send message",
			zap.String("template", tmplName),
			zap.Error(err))
		return nil, err
	}

	return &gqlmodels.AccountPayload{
		ClientMutationID: requestid.Get(ctx),
		AccountID:        id,
		Account:          gqlmodels.FromAccountModel(acc),
	}, nil
}

// ListAccounts list by filter
func (r *QueryResolver) ListAccounts(ctx context.Context,
	filter *gqlmodels.AccountListFilter,
	order *gqlmodels.AccountListOrder,
	page *gqlmodels.Page,
) (*AccountConnection, error) {
	return NewAccountConnection(ctx, r.accounts, filter, order, page), nil
}
