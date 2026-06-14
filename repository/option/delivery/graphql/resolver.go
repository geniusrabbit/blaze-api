package graphql

import (
	"context"

	"github.com/geniusrabbit/gosql/v2"

	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/repository/option/models"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
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
func (r *QueryResolver) Set(ctx context.Context, name string, value *types.NullableJSON, typeArg gqlmodels.OptionType, targetID uint64) (*gqlmodels.OptionPayload, error) {
	opt := models.Option{
		Name:     name,
		Type:     ModelOptionType(typeArg),
		TargetID: targetID,
	}
	if value != nil {
		opt.Value = gosql.NullableJSON[any](*value)
	}
	err := r.uc.Set(ctx, &opt)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.OptionPayload{
		ClientMutationID: requestid.Get(ctx),
		Name:             name,
		Option:           FromOption(&opt),
	}, nil
}

// Get Option is the resolver for the option field.
func (r *QueryResolver) Get(ctx context.Context, name string, otype gqlmodels.OptionType, targetID uint64) (*gqlmodels.OptionPayload, error) {
	opt, err := r.uc.Get(ctx, name, ModelOptionType(otype), targetID)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.OptionPayload{
		ClientMutationID: requestid.Get(ctx),
		Name:             name,
		Option:           FromOption(opt),
	}, nil
}

// List Options is the resolver for the listOptions field.
func (r *QueryResolver) List(ctx context.Context, filter *gqlmodels.OptionListFilter, order []*gqlmodels.OptionListOrder, page *gqlmodels.Page) (*OptionConnection, error) {
	return NewOptionConnection(ctx, r.uc, filter, order, page), nil
}
