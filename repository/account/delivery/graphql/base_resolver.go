package graphql

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/messanger"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/repository"
	"github.com/geniusrabbit/blaze-api/repository/account/usecase"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var (
	ErrAccountIDRequired = errors.New("account id is required")
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	userRepo *userrepo.Repository
	accounts account.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		userRepo: userrepo.New(),
		accounts: usecase.NewAccountUsecase(userrepo.New(), repository.New()),
	}
}

// CurrentAccount returns the current account info
func (r *QueryResolver) CurrentAccount(ctx context.Context) (*gqlmodels.AccountPayload, error) {
	account := session.Account(ctx)
	return &gqlmodels.AccountPayload{
		ClientMutationID: requestid.Get(ctx),
		AccountID:        account.ID,
		Account:          gqlmodels.FromAccountModel(account),
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
		Account:          gqlmodels.FromAccountModel(acc),
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
		userObj = input.Owner.Model(model.UndefinedApproveStatus)
		accObj  = input.Account.Model(model.UndefinedApproveStatus)
	)

	if input.OwnerID != nil && *input.OwnerID > 0 {
		if userObj != nil {
			userObj.ID = *input.OwnerID
		} else {
			userObj = &model.User{ID: *input.OwnerID}
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
		Account:          gqlmodels.FromAccountModel(accObj),
		Owner:            gqlmodels.FromUserModel(userObj),
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
		Account:          gqlmodels.FromAccountModel(acc),
	}, nil
}

// ApproveAccount is the resolver for the approveAccount field.
func (r *QueryResolver) ApproveAccount(ctx context.Context, id uint64, msg string) (*gqlmodels.AccountPayload, error) {
	return r.updateApproveStatus(ctx, id, model.ApprovedApproveStatus, msg)
}

// RejectAccount is the resolver for the rejectAccount field.
func (r *QueryResolver) RejectAccount(ctx context.Context, id uint64, msg string) (*gqlmodels.AccountPayload, error) {
	return r.updateApproveStatus(ctx, id, model.DisapprovedApproveStatus, msg)
}

func (r *QueryResolver) updateApproveStatus(ctx context.Context, id uint64, status model.ApproveStatus, msg string) (*gqlmodels.AccountPayload, error) {
	acc, err := r.accounts.Get(ctx, uint64(id))
	if err != nil {
		return nil, err
	}
	acc.Approve = status
	saveCtx := historylog.WithMessage(ctx, msg)
	saveCtx = historylog.WithAction(saveCtx, strings.ToLower(status.String()))
	if _, err = r.accounts.Store(saveCtx, acc); err != nil {
		return nil, err
	}

	// Send message to the account owner
	if true {
		// Get account owner
		members, err := r.accounts.FetchListMembers(ctx,
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
) (*connectors.AccountConnection, error) {
	return connectors.NewAccountConnection(ctx, r.accounts, filter, order, page), nil
}
