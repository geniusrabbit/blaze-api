package model

import (
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"gorm.io/gorm"
)

type OptionType string

const (
	UndefinedOptionType OptionType = "undefined"
	UserOptionType      OptionType = "user"
	AccountOptionType   OptionType = "account"
	SystemOptionType    OptionType = "system"
)

type Option struct {
	Type     OptionType              `json:"type"`
	TargetID uint64                  `json:"target_id"`
	Name     string                  `json:"name"`
	Value    gosql.NullableJSON[any] `json:"value" gorm:"type:jsonb"`

	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at"`
}

func (o *Option) TableName() string { return "option" }

func (o *Option) CreatorUserID() uint64 {
	if o != nil && o.Type == UserOptionType {
		return o.TargetID
	}
	return 0
}

func (o *Option) OwnerAccountID() uint64 {
	if o != nil && o.Type == AccountOptionType {
		return o.TargetID
	}
	return 0
}

// RBACResourceName returns the name of the resource for the RBAC
func (o *Option) RBACResourceName() string {
	return "option"
}
