// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/demdxx/rbac"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/geniusrabbit/api-template-base/internal/repository"
	"github.com/geniusrabbit/api-template-base/internal/repository/account"
	"github.com/geniusrabbit/api-template-base/model"
)

// Repository DAO which provides functionality of working with accounts
type Repository struct {
	repository.Repository
}

// New account repository
func New() *Repository {
	return &Repository{}
}

// Get returns account model by ID
func (r *Repository) Get(ctx context.Context, id uint64) (*model.Account, error) {
	object := new(model.Account)
	if err := r.Slave(ctx).Find(object, id).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// GetByTitle returns account model by title
func (r *Repository) GetByTitle(ctx context.Context, title string) (*model.Account, error) {
	object := new(model.Account)
	if err := r.Slave(ctx).Find(object, `title=?`, title).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// LoadPermissions into account object
func (r *Repository) LoadPermissions(ctx context.Context, accountObj *model.Account, userObj *model.User) error {
	var (
		err     error
		roles   []uint64
		memeber = new(model.AccountMember)
		query   = r.Slave(ctx)
	)
	if err = query.Find(memeber, `account_id=? AND user_id=?`, accountObj.ID, userObj.ID).Error; err != nil {
		return errors.WithStack(err)
	}
	err = query.Table((*model.M2MAccountMemberRole)(nil).TableName()).
		Where(`member_id=?`, memeber.ID).Select(`role_id`).Find(&roles).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, sql.ErrNoRows) {
		// `sql.ErrNoRows` in case of no any linked permissions
		return errors.WithStack(err)
	}
	if len(roles) > 0 || memeber.IsAdmin {
		if accountObj.Approve.IsApproved() && userObj.Approve.IsApproved() {
			accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, memeber.IsAdmin, nil, roles...)
		} else {
			accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, false, func(r rbac.Role) bool {
				return !strings.HasPrefix(r.Name(), "system:")
			}, roles...)
		}
	}
	return err
}

// FetchList returns list of accounts by filter
func (r *Repository) FetchList(ctx context.Context, filter *account.Filter, pagination *repository.Pagination) ([]*model.Account, error) {
	var (
		list  []*model.Account
		query = r.Slave(ctx).Model((*model.Account)(nil))
	)
	query = filter.PrepareQuery(query)
	query = pagination.PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

// Count returns count of accounts by filter
func (r *Repository) Count(ctx context.Context, filter *account.Filter) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*model.Account)(nil))
		err   = filter.PrepareQuery(query).Count(&count).Error
	)
	return count, err
}

// Create new object into database
func (r *Repository) Create(ctx context.Context, accountObj *model.Account) (uint64, error) {
	accountObj.CreatedAt = time.Now()
	accountObj.UpdatedAt = accountObj.CreatedAt
	err := r.Master(ctx).Create(accountObj).Error
	return accountObj.ID, err
}

// Update existing object in database
func (r *Repository) Update(ctx context.Context, id uint64, accountObj *model.Account) error {
	obj := *accountObj
	obj.ID = id
	return r.Master(ctx).Updates(&obj).Error
}

// Delete delites record by ID
func (r *Repository) Delete(ctx context.Context, id uint64) error {
	return r.Master(ctx).Model((*model.Account)(nil)).Delete(`id=?`, id).Error
}

// FetchMembers returns the list of members from account
func (r *Repository) FetchMembers(ctx context.Context, accountObj *model.Account) ([]*model.AccountMember, error) {
	var (
		list  []*model.AccountMember
		query = r.Slave(ctx).Model((*model.AccountMember)(nil))
	)
	if accountObj != nil {
		query = query.Where(`id=?`, accountObj.ID)
	}
	err := query.Find(&list).Error
	return list, err
}

// IsMember check the user if linked to account
func (r *Repository) IsMember(ctx context.Context, userObj *model.User, accountObj *model.Account) bool {
	var id []uint64
	r.Slave(ctx).Model((*model.AccountMember)(nil)).
		Where(`account_id=? AND user_id=?`, accountObj.ID, userObj.ID).Select(`id`).Limit(1).Find(&id)
	return len(id) > 0
}

// LinkMember into account
func (r *Repository) LinkMember(ctx context.Context, accountObj *model.Account, isAdmin bool, members ...*model.User) error {
	return r.Master(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model((*model.AccountMember)(nil)).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "account_id"}, {Name: "user_id"}},
			Where:     clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: `deleted_at IS NULL`}}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "is_admin"}),
		})
		for _, userObj := range members {
			err := query.Create(&model.AccountMember{
				Approve:   model.ApprovedApproveStatus,
				AccountID: accountObj.ID,
				UserID:    userObj.ID,
				IsAdmin:   isAdmin,
			}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// UnlinkMember from the account
func (r *Repository) UnlinkMember(ctx context.Context, accountObj *model.Account, users ...*model.User) error {
	ids := make([]uint64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return r.Master(ctx).Model((*model.AccountMember)(nil)).Delete(`id=ANY(?)`, ids).Error
}
