package models

import (
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/models"
)

// Order is an alias for the Order type from the models package
type Order = models.Order

// Order constants define the sort direction
const (
	OrderUndefined = models.OrderUndefined
	OrderAsc       = models.OrderAsc
	OrderDesc      = models.OrderDesc
)

// OptionType defines the type of option
type OptionType string

// OptionType constants represent different option categories
const (
	UndefinedOptionType OptionType = "undefined"
	UserOptionType      OptionType = "user"
	AccountOptionType   OptionType = "account"
	SystemOptionType    OptionType = "system"
)

// Option represents a configuration option with associated metadata
type Option struct {
	Type     OptionType              `json:"type"`                    // Type of option
	TargetID uint64                  `json:"target_id"`               // ID of the target entity
	Name     string                  `json:"name"`                    // Option name
	Value    gosql.NullableJSON[any] `json:"value" gorm:"type:jsonb"` // JSON value

	CreatedAt time.Time      `db:"created_at"` // Creation timestamp
	UpdatedAt time.Time      `db:"updated_at"` // Last update timestamp
	DeletedAt gorm.DeletedAt `db:"deleted_at"` // Soft delete timestamp
}

// TableName specifies the database table name for Option
func (o *Option) TableName() string { return "option" }

// CreatorUserID returns the user ID if this option is a user option
func (o *Option) CreatorUserID() uint64 {
	if o != nil && o.Type == UserOptionType {
		return o.TargetID
	}
	return 0
}

// OwnerAccountID returns the account ID if this option is an account option
func (o *Option) OwnerAccountID() uint64 {
	if o != nil && o.Type == AccountOptionType {
		return o.TargetID
	}
	return 0
}

// RBACResourceName returns the RBAC resource name for authorization
func (o *Option) RBACResourceName() string {
	return "option"
}
