// Package account present full API functionality of the specific object
package account

import (
	"context"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Filter of the objects list
type Filter struct {
	ID     []uint64
	UserID []uint64
	Title  []string
	Status []model.ApproveStatus
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
			(*model.AccountMember)(nil).TableName()+` WHERE user_id IN (?))`, fl.UserID)
	}
	if len(fl.Title) > 0 {
		query = query.Where(`title IN (?)`, fl.Title)
	}
	if len(fl.Status) > 0 {
		query = query.Where(`approve_status IN (?)`, fl.Status)
	}
	return query
}

type ListOrder struct {
	ID        model.Order
	Title     model.Order
	Status    model.Order
	CreatedAt model.Order
	UpdatedAt model.Order
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

// Repository of access to the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*model.Account, error)
	GetByTitle(ctx context.Context, title string) (*model.Account, error)
	LoadPermissions(ctx context.Context, account *model.Account, user *model.User) error
	FetchList(ctx context.Context, filter *Filter, order *ListOrder, pagination *repository.Pagination) ([]*model.Account, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Create(ctx context.Context, account *model.Account) (uint64, error)
	Update(ctx context.Context, id uint64, account *model.Account) error
	Delete(ctx context.Context, id uint64) error
	FetchMembers(ctx context.Context, account *model.Account) ([]*model.AccountMember, error)
	FetchMemberUsers(ctx context.Context, account *model.Account) ([]*model.AccountMember, []*model.User, error)
	Member(ctx context.Context, userID, accountID uint64) (*model.AccountMember, error)
	IsMember(ctx context.Context, userID, accountID uint64) bool
	IsAdmin(ctx context.Context, userID, accountID uint64) bool
	LinkMember(ctx context.Context, account *model.Account, isAdmin bool, members ...*model.User) error
	UnlinkMember(ctx context.Context, account *model.Account, members ...*model.User) error
}
