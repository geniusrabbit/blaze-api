package account

import (
	"github.com/guregu/null"
	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
)

// Filter of the objects list
type Filter struct {
	ID     []uint64
	UserID []uint64
	Title  []string
	Status []pkgModels.ApproveStatus
}

func (fl *Filter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	if len(fl.ID) > 0 {
		query = query.Where(`id IN (?)`, fl.ID)
	}
	if len(fl.UserID) > 0 {
		query = query.Where(`id IN (SELECT account_id FROM `+
			(*accountModels.AccountMember)(nil).TableName()+` WHERE user_id IN (?))`, fl.UserID)
	}
	if len(fl.Title) > 0 {
		query = query.Where(`title IN (?)`, fl.Title)
	}
	if len(fl.Status) > 0 {
		query = query.Where(`approve_status IN (?)`, fl.Status)
	}
	return query
}

// ListOrder of the objects list
type ListOrder struct {
	ID        pkgModels.Order
	Title     pkgModels.Order
	Status    pkgModels.Order
	CreatedAt pkgModels.Order
	UpdatedAt pkgModels.Order
}

func (ord *ListOrder) PrepareQuery(query *gorm.DB) *gorm.DB {
	if ord == nil {
		return query
	}
	query = ord.ID.PrepareQuery(query, `id`)
	query = ord.Title.PrepareQuery(query, `title`)
	query = ord.Status.PrepareQuery(query, `approve_status`)
	query = ord.CreatedAt.PrepareQuery(query, `created_at`)
	query = ord.UpdatedAt.PrepareQuery(query, `updated_at`)
	return query
}

// MemberFilter of the objects list
type MemberFilter struct {
	ID        []uint64
	AccountID []uint64
	UserID    []uint64
	NotUserID []uint64
	Status    []pkgModels.ApproveStatus
	IsAdmin   null.Bool
}

func (fl *MemberFilter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	if len(fl.ID) > 0 {
		query = query.Where(`id IN (?)`, fl.ID)
	}
	if len(fl.AccountID) > 0 {
		query = query.Where(`account_id IN (?)`, fl.AccountID)
	}
	if len(fl.UserID) > 0 {
		query = query.Where(`user_id IN (?)`, fl.UserID)
	}
	if len(fl.NotUserID) > 0 {
		query = query.Where(`user_id NOT IN (?)`, fl.NotUserID)
	}
	if len(fl.Status) > 0 {
		query = query.Where(`approve_status IN (?)`, fl.Status)
	}
	if fl.IsAdmin.Bool && fl.IsAdmin.Valid {
		query = query.Where(`is_admin = ?`, fl.IsAdmin.Bool)
	}
	return query
}

// MemberListOrder of the objects list
type MemberListOrder struct {
	ID        pkgModels.Order
	AccountID pkgModels.Order
	UserID    pkgModels.Order
	Status    pkgModels.Order
	IsAdmin   pkgModels.Order
	CreatedAt pkgModels.Order
	UpdatedAt pkgModels.Order
}

func (ord *MemberListOrder) PrepareQuery(query *gorm.DB) *gorm.DB {
	if ord == nil {
		return query
	}
	query = ord.ID.PrepareQuery(query, `id`)
	query = ord.AccountID.PrepareQuery(query, `account_id`)
	query = ord.UserID.PrepareQuery(query, `user_id`)
	query = ord.Status.PrepareQuery(query, `approve_status`)
	query = ord.IsAdmin.PrepareQuery(query, `is_admin`)
	query = ord.CreatedAt.PrepareQuery(query, `created_at`)
	query = ord.UpdatedAt.PrepareQuery(query, `updated_at`)
	return query
}

// Pagination of the objects list
type Pagination = repository.Pagination

type (
	// QOption is the query option interface
	QOption = repository.QOption
	// ListOptions is the list of query options
	ListOptions = repository.ListOptions
)
