package graphql

import (
	"context"
	"errors"

	"github.com/geniusrabbit/api-template-base/internal/context/session"
	"github.com/geniusrabbit/api-template-base/internal/repository/account"
	"github.com/geniusrabbit/api-template-base/internal/repository/account/repository"
	"github.com/geniusrabbit/api-template-base/internal/repository/account/usecase"
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
)

var (
	ErrAccountIDRequired = errors.New("account id is required")
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	accounts account.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		accounts: usecase.NewAccountUsecase(repository.New()),
	}
}

// CurrentAccount returns the current account info
func (r *QueryResolver) CurrentAccount(ctx context.Context) (*gqlmodels.AccountPayload, error) {
	account := session.Account(ctx)
	return &gqlmodels.AccountPayload{
		AccountID: int(account.ID),
		Account:   gqlmodels.FromAccountModel(account),
	}, nil
}

// CreateAccount creates a new account
func (r *QueryResolver) CreateAccount(ctx context.Context, input *gqlmodels.AccountInput) (*gqlmodels.AccountPayload, error) {
	return r.createUpdateAccount(ctx, 0, input)
}

// UpdateAccount is the resolver for the updateAccount field.
func (r *QueryResolver) UpdateAccount(ctx context.Context, id uint64, input *gqlmodels.AccountInput) (*gqlmodels.AccountPayload, error) {
	if id == 0 {
		return nil, ErrAccountIDRequired
	}
	return r.createUpdateAccount(ctx, id, input)
}

func (r *QueryResolver) createUpdateAccount(ctx context.Context, id uint64, input *gqlmodels.AccountInput) (*gqlmodels.AccountPayload, error) {
	id, err := r.accounts.Store(ctx, &model.Account{
		ID:                id,
		Approve:           input.Status.ModelStatus(),
		Title:             input.Title,
		Description:       input.Description,
		LogoURI:           input.LogoURI,
		PolicyURI:         input.PolicyURI,
		TermsOfServiceURI: input.TermsOfServiceURI,
		ClientURI:         input.ClientURI,
		Contacts:          append([]string{}, input.Contacts...),
	})
	if err != nil {
		return nil, err
	}
	acc, err := r.accounts.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.AccountPayload{
		AccountID: int(id),
		Account:   gqlmodels.FromAccountModel(acc),
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
	id, err = r.accounts.Store(ctx, acc)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.AccountPayload{
		AccountID: int(id),
		Account:   gqlmodels.FromAccountModel(acc),
	}, nil
}

// ListAccounts list by filter
func (r *QueryResolver) ListAccounts(ctx context.Context,
	filter *gqlmodels.AccountListFilter,
	order []*gqlmodels.AccountListOrder,
	page *gqlmodels.Page,
) (*connectors.AccountConnection, error) {
	return connectors.NewAccountConnection(ctx, r.accounts), nil
}
