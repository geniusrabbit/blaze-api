package option

import "github.com/geniusrabbit/blaze-api/repository/option/models"

type (
	// Option represents a single option in the system.
	Option = models.Option

	// OptionType defines the type of the option, such as user, project, etc.
	OptionType = models.OptionType

	// Order defines the sort direction.
	Order = models.Order
)

const (
	OrderUndefined = models.OrderUndefined
	OrderAsc       = models.OrderAsc
	OrderDesc      = models.OrderDesc
)

const (
	UndefinedOptionType = models.UndefinedOptionType
	UserOptionType      = models.UserOptionType
	AccountOptionType   = models.AccountOptionType
	SystemOptionType    = models.SystemOptionType
)
