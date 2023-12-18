package graphql

import (
	"context"

	"github.com/geniusrabbit/gosql/v2"

	"github.com/geniusrabbit/api-template-base/internal/repository/option"
	"github.com/geniusrabbit/api-template-base/internal/repository/option/repository"
	"github.com/geniusrabbit/api-template-base/internal/repository/option/usecase"
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/connectors"
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	uc option.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		uc: usecase.NewUsecase(repository.New()),
	}
}

// Set Option is the resolver for the setOption field.
func (r *QueryResolver) Set(ctx context.Context, name string, input *models.OptionInput) (*models.OptionPayload, error) {
	opt := model.Option{
		Name:     name,
		Type:     input.OptionType.ModelType(),
		TargetID: input.TargetID,
	}
	if input.Value != nil {
		opt.Value = gosql.NullableJSON[any](*input.Value)
	}
	err := r.uc.Set(ctx, &opt)
	if err != nil {
		return nil, err
	}
	return &models.OptionPayload{
		OptionName: name,
		Option:     models.FromOption(&opt),
	}, nil
}

// Get Option is the resolver for the option field.
func (r *QueryResolver) Get(ctx context.Context, name string, otype models.OptionType, targetID uint64) (*models.OptionPayload, error) {
	opt, err := r.uc.Get(ctx, name, otype.ModelType(), targetID)
	if err != nil {
		return nil, err
	}
	return &models.OptionPayload{
		OptionName: name,
		Option:     models.FromOption(opt),
	}, nil
}

// List Options is the resolver for the listOptions field.
func (r *QueryResolver) List(ctx context.Context, filter *models.OptionListFilter, order *models.OptionListOrder, page *models.Page) (*connectors.OptionConnection, error) {
	return connectors.NewOptionConnection(ctx, r.uc, filter, order, page), nil
}
