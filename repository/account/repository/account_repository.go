// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
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
	if memeber.IsAdmin {
		accountObj.ExtendAdminUsers(userObj.ID)
	}
	if len(roles) > 0 || memeber.IsAdmin {
		if !accountObj.Approve.IsRejected() && !userObj.Approve.IsRejected() {
			accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, memeber.IsAdmin, nil, roles...)
		} else {
			accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, false, func(_ context.Context, r rbac.Role) bool {
				// Skip system or account roles for not approved accounts
				return !strings.HasPrefix(r.Name(), "system:") || !strings.HasPrefix(r.Name(), "account:")
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
	accountObj.Approve = model.UndefinedApproveStatus
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

// FetchMemberUsers returns the list of members from account
func (r *Repository) FetchMemberUsers(ctx context.Context, accountObj *model.Account) ([]*model.AccountMember, []*model.User, error) {
	var (
		userIDs      []uint64
		users        []*model.User
		members, err = r.FetchMembers(ctx, accountObj)
	)
	if err != nil {
		return nil, nil, err
	}

	userIDs = xtypes.SliceApply(members, func(m *model.AccountMember) uint64 { return m.UserID })
	err = r.Slave(ctx).Model((*model.User)(nil)).Where(`id IN (?)`, userIDs).Find(&users).Error

	return members, users, err
}

// Member returns the member object by account and user
func (r *Repository) Member(ctx context.Context, userID, accountID uint64) (*model.AccountMember, error) {
	if accountID == 0 || userID == 0 {
		return nil, nil
	}
	var member model.AccountMember
	err := r.Slave(ctx).Find(&member, `account_id=? AND user_id=?`, accountID, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || member.ID == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &member, err
}

// IsMember check the user if linked to account
func (r *Repository) IsMember(ctx context.Context, userID, accountID uint64) bool {
	member, err := r.Member(ctx, userID, accountID)
	return err == nil && member != nil
}

// IsAdmin check the user if linked to account as admin
func (r *Repository) IsAdmin(ctx context.Context, userID, accountID uint64) bool {
	member, err := r.Member(ctx, userID, accountID)
	return err == nil && member != nil && member.IsAdmin
}

// LinkMember into account
func (r *Repository) LinkMember(ctx context.Context, accountObj *model.Account, isAdmin bool, members ...*model.User) error {
	return r.Master(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model((*model.AccountMember)(nil)).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "account_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"approve_status", "is_admin"}),
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
