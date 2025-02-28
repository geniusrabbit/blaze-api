package graphql

import (
	"context"

	"github.com/geniusrabbit/gosql/v2"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	uc option.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver(uc option.Usecase) *QueryResolver {
	return &QueryResolver{uc: uc}
}

// Set Option is the resolver for the setOption field.
func (r *QueryResolver) Set(ctx context.Context, name string, value *types.NullableJSON, typeArg models.OptionType, targetID uint64) (*models.OptionPayload, error) {
	opt := model.Option{
		Name:     name,
		Type:     typeArg.ModelType(),
		TargetID: targetID,
	}
	if value != nil {
		opt.Value = gosql.NullableJSON[any](*value)
	}
	err := r.uc.Set(ctx, &opt)
	if err != nil {
		return nil, err
	}
	return &models.OptionPayload{
		ClientMutationID: requestid.Get(ctx),
		OptionName:       name,
		Option:           models.FromOption(&opt),
	}, nil
}

// Get Option is the resolver for the option field.
func (r *QueryResolver) Get(ctx context.Context, name string, otype models.OptionType, targetID uint64) (*models.OptionPayload, error) {
	opt, err := r.uc.Get(ctx, name, otype.ModelType(), targetID)
	if err != nil {
		return nil, err
	}
	return &models.OptionPayload{
		ClientMutationID: requestid.Get(ctx),
		OptionName:       name,
		Option:           models.FromOption(opt),
	}, nil
}

// List Options is the resolver for the listOptions field.
func (r *QueryResolver) List(ctx context.Context, filter *models.OptionListFilter, order *models.OptionListOrder, page *models.Page) (*connectors.OptionConnection, error) {
	return connectors.NewOptionConnection(ctx, r.uc, filter, order, page), nil
}
