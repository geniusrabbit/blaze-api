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

func (o Option) TableName() string { return "option" }
