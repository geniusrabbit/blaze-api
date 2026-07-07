package graphql

import (
	"context"
	"errors"
	"strings"

	"github.com/demdxx/xtypes"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/messanger"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user"
	user_graphql "github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var (
	ErrAccountIDRequired = errors.New("account id is required")
)

// QueryResolver implements account GQL API methods.
// TGQL* type parameters are consumer GraphQL schema types (base or extended via extend type).
type QueryResolver[
	TUser user.Model,
	TDomain account.Model,
	TGQLAccount any,
	TGQLAccountPayload any,
	TGQLAccountCreateInput any,
	TGQLAccountUpdateInput any,
	TGQLAccountListFilter any,
	TGQLAccountListOrder any,
	TGQLUser any,
	TGQLUserCreateInput any,
	TGQLUserUpdateInput any,
] struct {
	users          user.Usecase[TUser]
	usersMapper    user_graphql.UserGraphQLMappersBase[TUser, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]
	accounts       account.Usecase[TUser, TDomain]
	accountsMapper AccountGraphQLMappers[TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser]
	members        account.MemberUsecase[TUser, TDomain]
}

// QueryResolverConfig wires generic account GraphQL resolvers.
type QueryResolverConfig[
	TUser user.Model,
	TDomain account.Model,
	TGQLAccount any,
	TGQLAccountPayload any,
	TGQLAccountCreateInput any,
	TGQLAccountUpdateInput any,
	TGQLAccountListFilter any,
	TGQLAccountListOrder any,
	TGQLUser any,
	TGQLUserCreateInput any,
	TGQLUserUpdateInput any,
] struct {
	Users          user.Usecase[TUser]
	UsersMapper    user_graphql.UserGraphQLMappersBase[TUser, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]
	Accounts       account.Usecase[TUser, TDomain]
	AccountsMapper AccountGraphQLMappers[TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser]
	Members        account.MemberUsecase[TUser, TDomain]
}

// NewQueryResolver returns new API resolver.
func NewQueryResolver[
	TUser user.Model,
	TDomain account.Model,
	TGQLAccount any,
	TGQLAccountPayload any,
	TGQLAccountCreateInput any,
	TGQLAccountUpdateInput any,
	TGQLAccountListFilter any,
	TGQLAccountListOrder any,
	TGQLUser any,
	TGQLUserCreateInput any,
	TGQLUserUpdateInput any,
](
	cfg QueryResolverConfig[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput],
) *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput] {
	return &QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]{
		users:          cfg.Users,
		usersMapper:    cfg.UsersMapper,
		accounts:       cfg.Accounts,
		accountsMapper: cfg.AccountsMapper,
		members:        cfg.Members,
	}
}

// CurrentAccount returns the current account info.
func (r *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]) CurrentAccount(ctx context.Context) (TGQLAccountPayload, error) {
	var zero TDomain
	acc := session.Account(ctx)
	typed, _ := acc.(TDomain)
	if any(typed) == any(zero) {
		typed = zero
	}
	return r.accountsMapper.NewPayload(
		requestid.Get(ctx), typed.GetID(), r.accountsMapper.ToGQL(typed)), nil
}

// Account returns the account info.
func (r *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]) Account(ctx context.Context, id uint64) (TGQLAccountPayload, error) {
	acc, err := r.accounts.Get(ctx, id)
	if err != nil {
		var zero TGQLAccountPayload
		return zero, err
	}
	return r.accountsMapper.NewPayload(requestid.Get(ctx), id, r.accountsMapper.ToGQL(acc)), nil
}

// RegisterAccount creates a new account.
func (r *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]) RegisterAccount(ctx context.Context, ownerID uint64, input TGQLAccountCreateInput) (TGQLAccountPayload, error) {
	var zero TGQLAccountPayload
	accObj := r.accountsMapper.FromCreateInput(input)
	userObj, err := r.users.Get(ctx, ownerID)
	if err != nil {
		return zero, err
	}

	if _, err := r.accounts.Register(ctx, userObj, accObj); err != nil {
		return zero, err
	}

	err = messanger.Get(ctx).Send(ctx, "account.register",
		[]string{userEmail(userObj)}, map[string]any{
			"id":      accObj.GetID(),
			"account": accObj,
			"owner":   userObj,
		})
	if err != nil {
		ctxlogger.Get(ctx).Error("Failed to send message",
			zap.String("template", "account.register"),
			zap.Error(err))
	}

	return r.accountsMapper.NewPayload(
		requestid.Get(ctx), accObj.GetID(), r.accountsMapper.ToGQL(accObj)), nil
}

// UpdateAccount is the resolver for the updateAccount field.
func (r *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]) UpdateAccount(ctx context.Context, id uint64, input TGQLAccountUpdateInput) (TGQLAccountPayload, error) {
	var zero TGQLAccountPayload
	accModel := r.accountsMapper.FromUpdateInput(input, r.accountsMapper.New())
	if setter, ok := any(accModel).(interface{ SetID(uint64) }); ok {
		setter.SetID(id)
	}
	id, err := r.accounts.Update(ctx, accModel)
	if err != nil {
		return zero, err
	}
	acc, err := r.accounts.Get(ctx, id)
	if err != nil {
		return zero, err
	}
	return r.accountsMapper.NewPayload(requestid.Get(ctx), id, r.accountsMapper.ToGQL(acc)), nil
}

// ApproveAccount is the resolver for the approveAccount field.
func (r *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]) ApproveAccount(ctx context.Context, id uint64, msg string) (TGQLAccountPayload, error) {
	return r.updateApproveStatus(ctx, id, pkgModels.ApprovedApproveStatus, msg)
}

// RejectAccount is the resolver for the rejectAccount field.
func (r *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]) RejectAccount(ctx context.Context, id uint64, msg string) (TGQLAccountPayload, error) {
	return r.updateApproveStatus(ctx, id, pkgModels.DisapprovedApproveStatus, msg)
}

func (r *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]) updateApproveStatus(ctx context.Context, id uint64, status pkgModels.ApproveStatus, msg string) (TGQLAccountPayload, error) {
	var zero TGQLAccountPayload
	type approvable interface {
		SetApprove(pkgModels.ApproveStatus)
	}

	// Get account and check permissions
	acc, err := r.accounts.Get(ctx, id)
	if err != nil {
		return zero, err
	}

	// Set approval status and save account
	if setter, ok := any(acc).(approvable); ok {
		setter.SetApprove(status)
	}

	// Save account with history log
	saveCtx := historylog.WithMessageAndPK(ctx, msg, id)
	saveCtx = historylog.WithAction(saveCtx, strings.ToLower(status.String()))

	if _, err = r.accounts.Update(saveCtx, acc); err != nil {
		return zero, err
	}

	// Notify account admins about approval status change
	members, err := r.members.FetchListMembers(ctx,
		&account.MemberFilter{AccountID: []uint64{acc.GetID()}}, nil, nil)
	if err != nil {
		return zero, err
	}

	recipients := make([]string, 0, len(members))
	for _, member := range members {
		if member.IsAdmin {
			if email := userEmail(member.User); email != "" {
				recipients = append(recipients, email)
			}
		}
	}

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
		return zero, err
	}

	return r.accountsMapper.NewPayload(requestid.Get(ctx),
		id, r.accountsMapper.ToGQL(acc)), nil
}

// ListAccounts list by filter.
func (r *QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountPayload, TGQLAccountCreateInput, TGQLAccountUpdateInput, TGQLAccountListFilter, TGQLAccountListOrder, TGQLUser, TGQLUserCreateInput, TGQLUserUpdateInput]) ListAccounts(
	ctx context.Context,
	filter TGQLAccountListFilter,
	order []TGQLAccountListOrder,
	page *gqlmodels.Page,
) (*AccountConnection[TGQLAccount], error) {
	return NewAccountConnection(
		ctx,
		r.accounts,
		r.accountsMapper.FromFilter(filter),
		xtypes.SliceApply(order, r.accountsMapper.FromOrder),
		page,
		r.accountsMapper.ToGQL,
	), nil
}
