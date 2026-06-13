package models

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

func FromOptionType(tp option.OptionType) OptionType {
	switch tp {
	case option.UserOptionType:
		return OptionTypeUser
	case option.AccountOptionType:
		return OptionTypeAccount
	case option.SystemOptionType:
		return OptionTypeSystem
	}
	return OptionTypeUndefined
}

func (tp OptionType) ModelType() option.OptionType {
	switch tp {
	case OptionTypeUser:
		return option.UserOptionType
	case OptionTypeAccount:
		return option.AccountOptionType
	case OptionTypeSystem:
		return option.SystemOptionType
	}
	return option.UndefinedOptionType
}

func (fl *OptionListFilter) Filter() *option.Filter {
	if fl == nil {
		return nil
	}
	return &option.Filter{
		Type:        xtypes.SliceApply(fl.Type, func(tp OptionType) option.OptionType { return tp.ModelType() }),
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

func FromOption(opt *option.Option) *Option {
	if opt == nil {
		return nil
	}
	return &Option{
		Name:     opt.Name,
		Type:     FromOptionType(opt.Type),
		TargetID: opt.TargetID,
		Value:    types.MustNullableJSONFrom(opt.Value.Data),
	}
}

func FromOptionModelList(opts []*option.Option) []*Option {
	return xtypes.SliceApply(opts, FromOption)
}
