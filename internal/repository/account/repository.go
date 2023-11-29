// Package account present full API functionality of the specific object
package account

import (
	"context"

	"gorm.io/gorm"

	"github.com/geniusrabbit/api-template-base/internal/repository"
	"github.com/geniusrabbit/api-template-base/model"
)

// Filter of the objects list
type Filter struct {
	ID     []uint64
	UserID []uint64
	Title  []string
	Status []model.ApproveStatus
}

func (fl *Filter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl != nil {
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
	}
	return query
}

// Repository of access to the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*model.Account, error)
	GetByTitle(ctx context.Context, title string) (*model.Account, error)
	LoadPermissions(ctx context.Context, account *model.Account, user *model.User) error
	FetchList(ctx context.Context, filter *Filter, pagination *repository.Pagination) ([]*model.Account, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Create(ctx context.Context, account *model.Account) (uint64, error)
	Update(ctx context.Context, id uint64, account *model.Account) error
	Delete(ctx context.Context, id uint64) error
	FetchMembers(ctx context.Context, account *model.Account) ([]*model.AccountMember, error)
	IsMember(ctx context.Context, user *model.User, account *model.Account) bool
	LinkMember(ctx context.Context, account *model.Account, isAdmin bool, members ...*model.User) error
	UnlinkMember(ctx context.Context, account *model.Account, members ...*model.User) error
}
