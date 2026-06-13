package model

import optionModels "github.com/geniusrabbit/blaze-api/repository/option/models"

// Option represents a configuration option with associated metadata
type Option = optionModels.Option

// OptionType defines the type of option
type OptionType = optionModels.OptionType

// OptionType constants
const (
	UndefinedOptionType = optionModels.UndefinedOptionType
	UserOptionType      = optionModels.UserOptionType
	AccountOptionType   = optionModels.AccountOptionType
	SystemOptionType    = optionModels.SystemOptionType
)

