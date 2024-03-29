// Package user present full API functionality of the specific object
package user

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/demdxx/xtypes"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// ListFilter object with filtered values which is not NULL
type ListFilter struct {
	AccountID []uint64
	UserID    []uint64
	Emails    []string
	Roles     []uint64
}

// PrepareQuery returns the query with applied filters
func (fl *ListFilter) PrepareQuery(q *gorm.DB) *gorm.DB {
	if fl == nil {
		return q
	}
	if len(fl.Roles) > 0 {
		qstr := `SELECT member_id FROM ` +
			(*model.M2MAccountMemberRole)(nil).TableName() + ` WHERE role_id IN (?)`
		if len(fl.AccountID) > 0 {
			q = q.Where(`id IN (SELECT user_id FROM `+(*model.AccountMember)(nil).TableName()+
				` WHERE account_id IN (?) OR id IN (`+qstr+`))`, fl.AccountID, fl.Roles)
		} else {
			q = q.Where(`id IN (SELECT user_id FROM `+(*model.AccountMember)(nil).TableName()+
				` WHERE id IN (`+qstr+`))`, fl.Roles)
		}
	} else if len(fl.AccountID) > 0 {
		q = q.Where(`id IN (SELECT user_id FROM `+
			(*model.AccountMember)(nil).TableName()+` WHERE account_id IN (?))`, fl.AccountID)
	}
	if len(fl.UserID) > 0 {
		q = q.Where(`id IN (?)`, fl.UserID)
	}
	if len(fl.Emails) > 0 {
		q = q.Where(`lower(email) IN (?)`, xtypes.SliceApply(fl.Emails, func(v string) string {
			return strings.ToLower(v)
		}))
	}
	return q
}

// ListOrder object with order values which is not NULL
type ListOrder struct {
	ID        model.Order
	Email     model.Order
	Status    model.Order
	CreatedAt model.Order
	UpdatedAt model.Order
}

// PrepareQuery returns the query with applied order
func (ord *ListOrder) PrepareQuery(q *gorm.DB) *gorm.DB {
	if ord != nil {
		q = ord.ID.PrepareQuery(q, "id")
		q = ord.Email.PrepareQuery(q, "email")
		q = ord.Status.PrepareQuery(q, "approve_status")
		q = ord.CreatedAt.PrepareQuery(q, "created_at")
		q = ord.UpdatedAt.PrepareQuery(q, "updated_at")
	}
	return q
}

// Repository describes basic user methods
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByPassword(ctx context.Context, email, password string) (*model.User, error)
	GetByToken(ctx context.Context, token string) (*model.User, *model.Account, error)
	FetchList(ctx context.Context, filter *ListFilter, order *ListOrder, page *repository.Pagination) ([]*model.User, error)
	Count(ctx context.Context, filter *ListFilter) (int64, error)
	Create(ctx context.Context, user *model.User, password string) (uint64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint64) error

	SetPassword(ctx context.Context, user *model.User, password string) error
	CreateResetPassword(ctx context.Context, userID uint64) (*model.UserPasswordReset, error)
	GetResetPassword(ctx context.Context, userID uint64, token string) (*model.UserPasswordReset, error)
	EliminateResetPassword(ctx context.Context, userID uint64) error
}
