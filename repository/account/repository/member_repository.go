package repository

import (
	"context"
	"time"

	"github.com/demdxx/xtypes"
	"github.com/guregu/null"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	prbac "github.com/geniusrabbit/blaze-api/repository/rbac"
	userbac "github.com/geniusrabbit/blaze-api/repository/rbac/usecase"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var (
	// ErrInvalidRoleList error in case of invalid role list
	ErrInvalidRoleList = errors.New(`invalid role list, check your permissions`)

	// ErrAccountHaveToHaveAdmin error in case of no any admin in account
	ErrAccountHaveToHaveAdmin = errors.New(`account must have at least one admin`)
)

// Repository DAO which provides functionality of working with accounts
type MemberRepository struct {
	repository.Repository
	rbacUse prbac.Usecase
}

// NewMemberRepository account repository
func NewMemberRepository() *MemberRepository {
	return &MemberRepository{
		rbacUse: userbac.NewDefault(),
	}
}

// FetchListMembers returns the list of members from account
func (r *MemberRepository) FetchListMembers(ctx context.Context, opts ...account.QOption) ([]*models.AccountMember, error) {
	var (
		list  []*models.AccountMember
		query = r.Slave(ctx).Model((*models.AccountMember)(nil))
	)
	query = account.ListOptions(opts).PrepareQuery(query)
	query = query.Preload(clause.Associations)
	err := query.Find(&list).Error
	return list, err
}

// CountMembers returns the count of members from account
func (r *MemberRepository) CountMembers(ctx context.Context, opts ...account.QOption) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*models.AccountMember)(nil))
	)
	query = account.ListOptions(opts).PrepareQuery(query)
	err := query.Count(&count).Error
	return count, err
}

// Member returns the member object by account and user
func (r *MemberRepository) Member(ctx context.Context, userID, accountID uint64) (*models.AccountMember, error) {
	return r.memberByQuery(ctx, `account_id=? AND user_id=?`, accountID, userID)
}

// MemberByID returns the member object by ID
func (r *MemberRepository) MemberByID(ctx context.Context, id uint64) (*models.AccountMember, error) {
	return r.memberByQuery(ctx, `id=?`, id)
}

func (r *MemberRepository) memberByQuery(ctx context.Context, query ...any) (*models.AccountMember, error) {
	var member models.AccountMember
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
func (r *MemberRepository) IsMember(ctx context.Context, userID, accountID uint64) bool {
	count, _ := r.CountMembers(ctx, &account.MemberFilter{
		UserID:    []uint64{userID},
		AccountID: []uint64{accountID},
	})
	return count > 0
}

// IsAdmin check the user if linked to account as admin
func (r *MemberRepository) IsAdmin(ctx context.Context, userID, accountID uint64) bool {
	if accountID == 0 || userID == 0 {
		return false
	}
	var member models.AccountMember
	err := r.Slave(ctx).
		Find(&member, `account_id=? AND user_id=?`, accountID, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || member.ID == 0 {
		return false
	}
	return err == nil && member.IsAdmin
}

// LinkMember into account
func (r *MemberRepository) LinkMember(ctx context.Context, accountObj *models.Account, isAdmin bool, members ...*user.User) error {
	return r.Master(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model((*models.AccountMember)(nil)).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "account_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"approve_status", "is_admin"}),
		})
		for _, userObj := range members {
			err := query.Create(&models.AccountMember{
				Approve:   pkgModels.ApprovedApproveStatus,
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
func (r *MemberRepository) UnlinkMember(ctx context.Context, accountObj *models.Account, users ...*user.User) error {
	ids := make([]uint64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return r.Master(ctx).Model((*models.AccountMember)(nil)).Delete(`id=ANY(?)`, ids).Error
}

// SetMemberRoles into account
func (r *MemberRepository) SetMemberRoles(ctx context.Context, accountObj *models.Account, user *user.User, roles ...string) error {
	var (
		listRoles   []*prbac.Role
		member, err = r.Member(ctx, user.ID, accountObj.ID)
	)
	if err != nil {
		return err
	}

	// Load roles for the member
	if len(roles) > 0 {
		if listRoles, err = r.rbacUse.FetchList(ctx, &prbac.Filter{Names: roles}); err != nil {
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
			Status:    []pkgModels.ApproveStatus{pkgModels.ApprovedApproveStatus},
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
		roleIDs := xtypes.SliceApply(listRoles, func(v *prbac.Role) uint64 { return v.ID })
		// Remove roles for the member
		err = tx.Model((*models.M2MAccountMemberRole)(nil)).
			Where(`member_id=?`, member.ID).
			Where(`role_id NOT IN (?)`, roleIDs).
			Delete(&models.M2MAccountMemberRole{}).Error
		if err != nil {
			return err
		}
		// Save roles for the member
		return tx.Save(xtypes.SliceApply(listRoles, func(v *prbac.Role) *models.M2MAccountMemberRole {
			return &models.M2MAccountMemberRole{
				MemberID:  member.ID,
				RoleID:    v.ID,
				CreatedAt: time.Now(),
			}
		})).Error
	})
}
