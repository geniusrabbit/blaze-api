package graphql

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/repository/option/models"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

// FromOptionType converts a models.OptionType to a GraphQL gqlmodels.OptionType.
func FromOptionType(tp models.OptionType) gqlmodels.OptionType {
	switch tp {
	case models.UserOptionType:
		return gqlmodels.OptionTypeUser
	case models.AccountOptionType:
		return gqlmodels.OptionTypeAccount
	case models.SystemOptionType:
		return gqlmodels.OptionTypeSystem
	}
	return gqlmodels.OptionTypeUndefined
}

// ModelOptionType converts a gqlmodels.OptionType to a models.OptionType.
func ModelOptionType(tp gqlmodels.OptionType) models.OptionType {
	switch tp {
	case gqlmodels.OptionTypeUser:
		return models.UserOptionType
	case gqlmodels.OptionTypeAccount:
		return models.AccountOptionType
	case gqlmodels.OptionTypeSystem:
		return models.SystemOptionType
	}
	return models.UndefinedOptionType
}

// FromOption converts a models.Option to a gqlmodels.Option.
// Returns nil if the input option is nil.
func FromOption(opt *models.Option) *gqlmodels.Option {
	if opt == nil {
		return nil
	}
	return &gqlmodels.Option{
		Name:     opt.Name,
		Type:     FromOptionType(opt.Type),
		TargetID: opt.TargetID,
		Value:    types.MustNullableJSONFrom(opt.Value.Data),
	}
}

// FromOptionModelList converts a slice of models.Option to a slice of gqlmodels.Option.
func FromOptionModelList(opts []*models.Option) []*gqlmodels.Option {
	return xtypes.SliceApply(opts, FromOption)
}
