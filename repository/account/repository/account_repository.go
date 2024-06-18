// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"
	"github.com/guregu/null"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	prbac "github.com/geniusrabbit/blaze-api/repository/rbac"
	userbac "github.com/geniusrabbit/blaze-api/repository/rbac/usecase"
)

var (
	// ErrInvalidRoleList error in case of invalid role list
	ErrInvalidRoleList = errors.New(`invalid role list, check your permissions`)

	// ErrAccountHaveToHaveAdmin error in case of no any admin in account
	ErrAccountHaveToHaveAdmin = errors.New(`account must have at least one admin`)
)

// Repository DAO which provides functionality of working with accounts
type Repository struct {
	repository.Repository
	rbacUse prbac.Usecase
}

// New account repository
func New() *Repository {
	return &Repository{
		rbacUse: userbac.NewDefault(),
	}
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
func (r *Repository) LoadPermissions(ctx context.Context, accountObj *model.Account, userObj *model.User) (err error) {
	if accountObj == nil || userObj == nil {
		accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, false, nil)
		return err
	}
	var (
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
	if !accountObj.Approve.IsRejected() && !userObj.Approve.IsRejected() {
		accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, memeber.IsAdmin, nil, roles...)
	} else {
		accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, false, func(_ context.Context, r rbac.Role) bool {
			// Skip system or account roles for not approved accounts
			return !strings.HasPrefix(r.Name(), "system:") || !strings.HasPrefix(r.Name(), "account:")
		}, roles...)
	}
	return err
}

// FetchList returns list of accounts by filter
func (r *Repository) FetchList(ctx context.Context, filter *account.Filter, order *account.ListOrder, pagination *repository.Pagination) ([]*model.Account, error) {
	var (
		list  []*model.Account
		query = r.Slave(ctx).Model((*model.Account)(nil))
	)
	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
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

// FetchListMembers returns the list of members from account
func (r *Repository) FetchListMembers(ctx context.Context, filter *account.MemberFilter, order *account.MemberListOrder, pagination *repository.Pagination) ([]*model.AccountMember, error) {
	var (
		list  []*model.AccountMember
		query = r.Slave(ctx).Model((*model.AccountMember)(nil))
	)
	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
	query = pagination.PrepareQuery(query)
	query = query.Preload(clause.Associations)
	err := query.Find(&list).Error
	return list, err
}

// CountMembers returns the count of members from account
func (r *Repository) CountMembers(ctx context.Context, filter *account.MemberFilter) (int64, error) {
	var (
		count int64
		err   = filter.PrepareQuery(
			r.Slave(ctx).Model((*model.AccountMember)(nil)),
		).Count(&count).Error
	)
	return count, err
}

// Member returns the member object by account and user
func (r *Repository) Member(ctx context.Context, userID, accountID uint64) (*model.AccountMember, error) {
	return r.memberByQuery(ctx, `account_id=? AND user_id=?`, accountID, userID)
}

// MemberByID returns the member object by ID
func (r *Repository) MemberByID(ctx context.Context, id uint64) (*model.AccountMember, error) {
	return r.memberByQuery(ctx, `id=?`, id)
}

func (r *Repository) memberByQuery(ctx context.Context, query ...any) (*model.AccountMember, error) {
	var member model.AccountMember
	err := r.Slave(ctx).
		Preload(clause.Associations).
		Find(&member, query...).Error
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
	count, _ := r.CountMembers(ctx, &account.MemberFilter{
		UserID:    []uint64{userID},
		AccountID: []uint64{accountID},
	})
	return count > 0
}

// IsAdmin check the user if linked to account as admin
func (r *Repository) IsAdmin(ctx context.Context, userID, accountID uint64) bool {
	if accountID == 0 || userID == 0 {
		return false
	}
	var member model.AccountMember
	err := r.Slave(ctx).
		Find(&member, `account_id=? AND user_id=?`, accountID, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || member.ID == 0 {
		return false
	}
	return err == nil && member.IsAdmin
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

// SetMemberRoles into account
func (r *Repository) SetMemberRoles(ctx context.Context, accountObj *model.Account, user *model.User, roles ...string) error {
	var (
		listRoles   []*model.Role
		member, err = r.Member(ctx, user.ID, accountObj.ID)
	)
	if err != nil {
		return err
	}

	// Load roles for the member
	if len(roles) > 0 {
		if listRoles, err = r.rbacUse.FetchList(ctx, &prbac.Filter{Names: roles}, nil, nil); err != nil {
			return err
		}
		if len(listRoles) != len(roles) {
			return ErrInvalidRoleList
		}
	}

	// Prepare member roles
	wasAdmin := member.IsAdmin
	member.Roles = listRoles
	member.IsAdmin = xtypes.Slice[string](roles).Has(func(v string) bool { return v == "admin" || v == "account:admin" })

	// Check if we have at least one admin in account
	if wasAdmin != member.IsAdmin && !member.IsAdmin {
		cnt, err := r.CountMembers(ctx, &account.MemberFilter{
			AccountID: []uint64{accountObj.ID},
			NotUserID: []uint64{user.ID},
			IsAdmin:   null.BoolFrom(true),
			Status:    []model.ApproveStatus{model.ApprovedApproveStatus},
		})
		if err != nil {
			return err
		}
		if cnt == 0 {
			return ErrAccountHaveToHaveAdmin
		}
	}

	// Transaction for updating member roles
	return r.TransactionExec(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Save member object state
		err := tx.Omit(clause.Associations).Save(member).Error
		if err != nil {
			return err
		}
		roleIDs := xtypes.SliceApply(listRoles, func(v *model.Role) uint64 { return v.ID })
		// Remove roles for the member
		err = tx.Model((*model.M2MAccountMemberRole)(nil)).
			Where(`member_id=?`, member.ID).
			Where(`role_id NOT IN (?)`, roleIDs).
			Delete(&model.M2MAccountMemberRole{}).Error
		if err != nil {
			return err
		}
		// Save roles for the member
		return tx.Save(xtypes.SliceApply(listRoles, func(v *model.Role) *model.M2MAccountMemberRole {
			return &model.M2MAccountMemberRole{
				MemberID:  member.ID,
				RoleID:    v.ID,
				CreatedAt: time.Now(),
			}
		})).Error
	})
}
