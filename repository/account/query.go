package account

import (
	"context"
	"errors"

	"github.com/guregu/null"
	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
	accountCtx "github.com/geniusrabbit/blaze-api/repository/account/context"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	userCtx "github.com/geniusrabbit/blaze-api/repository/user/context"
)

// Sentinel errors returned by AdjustPermissions.
// Callers in the usecase layer wrap these with acl.ErrNoPermissions.
var (
	errListFilterTooWide   = errors.New("list account (too wide filter)")
	errMemberFilterTooWide = errors.New("member account for that account")
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
			accountModels.MemberTableName()+` WHERE user_id IN (?))`, fl.UserID)
	}
	if len(fl.Title) > 0 {
		query = query.Where(`title IN (?)`, fl.Title)
	}
	if len(fl.Status) > 0 {
		query = query.Where(`approve_status IN (?)`, fl.Status)
	}
	return query
}

// AdjustPermissions narrows the filter to the current user's accounts.
// Cannot import pkg/acl or pkg/context/session due to an import cycle;
// uses the lower-level context sub-packages instead.
func (fl *Filter) AdjustPermissions(ctx context.Context) error {
	usr := userCtx.SessionUser(ctx)
	if usr == nil {
		return errListFilterTooWide
	}
	userID := usr.GetID()
	if len(fl.UserID) == 0 {
		fl.UserID = []uint64{userID}
	}
	if len(fl.UserID) != 1 || fl.UserID[0] != userID {
		return errListFilterTooWide
	}
	return nil
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

// AdjustPermissions narrows the member filter to the current session account.
// Cannot import pkg/acl or pkg/context/session due to an import cycle;
// uses the lower-level context sub-packages instead.
func (fl *MemberFilter) AdjustPermissions(ctx context.Context) error {
	accID := accountCtx.SessionAccount(ctx).GetID()
	if l := len(fl.AccountID); l > 1 || (l == 1 && fl.AccountID[0] != accID) {
		return errMemberFilterTooWide
	}
	fl.AccountID = []uint64{accID}
	return nil
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
