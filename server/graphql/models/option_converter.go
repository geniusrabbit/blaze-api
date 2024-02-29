package models

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

func FromOptionType(tp model.OptionType) OptionType {
	switch tp {
	case model.UserOptionType:
		return OptionTypeUser
	case model.AccountOptionType:
		return OptionTypeAccount
	case model.SystemOptionType:
		return OptionTypeSystem
	}
	return OptionTypeUndefined
}

func (tp OptionType) ModelType() model.OptionType {
	switch tp {
	case OptionTypeUser:
		return model.UserOptionType
	case OptionTypeAccount:
		return model.AccountOptionType
	case OptionTypeSystem:
		return model.SystemOptionType
	}
	return model.UndefinedOptionType
}

func (fl *OptionListFilter) Filter() *option.Filter {
	if fl == nil {
		return nil
	}
	return &option.Filter{
		Type:        xtypes.SliceApply(fl.OptionType, func(tp OptionType) model.OptionType { return tp.ModelType() }),
		TargetID:    fl.TargetID,
		Name:        fl.Name,
		NamePattern: fl.NamePattern,
	}
}

func (ol *OptionListOrder) Order() *option.ListOrder {
	if ol == nil {
		return nil
	}
	return &option.ListOrder{
		Name:     ol.Name.AsOrder(),
		TargetID: ol.TargetID.AsOrder(),
	}
}

func FromOption(opt *model.Option) *Option {
	if opt == nil {
		return nil
	}
	return &Option{
		Name:       opt.Name,
		OptionType: FromOptionType(opt.Type),
		TargetID:   opt.TargetID,
		Value:      types.MustNullableJSONFrom(opt.Value.Data),
	}
}

func FromOptionModelList(opts []*model.Option) []*Option {
	return xtypes.SliceApply(opts, FromOption)
}
