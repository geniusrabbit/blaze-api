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
	baseRepo "github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	prbac "github.com/geniusrabbit/blaze-api/repository/rbac"
	userbac "github.com/geniusrabbit/blaze-api/repository/rbac/usecase"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var (
	ErrInvalidRoleList        = errors.New(`invalid role list, check your permissions`)
	ErrAccountHaveToHaveAdmin = errors.New(`account must have at least one admin`)
	ErrMemberNotFound         = errors.New(`member not found`)
)

type memberRepository[TUser user.Model, TAccount account.Model] struct {
	baseRepo.Repository
	rbacUse   prbac.Usecase
	newMember func() *account.Member[TUser, TAccount]
}

// NewMemberRepositoryFor creates generic member repository.
func NewMemberRepositoryFor[TUser user.Model, TAccount account.Model](newMember func() *account.Member[TUser, TAccount]) account.MemberRepository[TUser, TAccount] {
	if newMember == nil {
		panic(`newMember function must be provided`)
	}
	return &memberRepository[TUser, TAccount]{rbacUse: userbac.NewDefault(), newMember: newMember}
}

// EmptyObject returns a new empty member object of type Member[TUser, TAccount].
func (r *memberRepository[TUser, TAccount]) EmptyObject() *account.Member[TUser, TAccount] {
	return r.newMember()
}

func (r *memberRepository[TUser, TAccount]) FetchListMembers(ctx context.Context, opts ...account.QOption) ([]*account.Member[TUser, TAccount], error) {
	var (
		bases []models.MemberBase
		query = r.Slave(ctx).Model(&models.MemberBase{})
	)
	query = account.ListOptions(opts).PrepareQuery(query).Preload("Roles")
	if err := query.Find(&bases).Error; err != nil {
		return nil, err
	}
	list := make([]*account.Member[TUser, TAccount], len(bases))
	for i, base := range bases {
		m := r.newMember()
		m.MemberBase = base
		list[i] = m
	}
	return list, nil
}

func (r *memberRepository[TUser, TAccount]) CountMembers(ctx context.Context, opts ...account.QOption) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model(&models.MemberBase{})
	)
	query = account.ListOptions(opts).PrepareQuery(query)
	err := query.Count(&count).Error
	return count, err
}

func (r *memberRepository[TUser, TAccount]) Member(ctx context.Context, userID, accountID uint64) (*account.Member[TUser, TAccount], error) {
	return r.memberByQuery(ctx, `account_id=? AND user_id=?`, accountID, userID)
}

func (r *memberRepository[TUser, TAccount]) MemberByID(ctx context.Context, id uint64) (*account.Member[TUser, TAccount], error) {
	return r.memberByQuery(ctx, `id=?`, id)
}

func (r *memberRepository[TUser, TAccount]) memberByQuery(ctx context.Context, query ...any) (*account.Member[TUser, TAccount], error) {
	var base models.MemberBase
	err := r.Master(ctx).
		Model(&models.MemberBase{}).
		Preload("Roles").
		Where(query[0], query[1:]...).
		First(&base).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || base.ID == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	member := r.newMember()
	member.MemberBase = base
	return member, nil
}

func (r *memberRepository[TUser, TAccount]) IsMember(ctx context.Context, userID, accountID uint64) bool {
	count, _ := r.CountMembers(ctx, &account.MemberFilter{
		UserID:    []uint64{userID},
		AccountID: []uint64{accountID},
	})
	return count > 0
}

func (r *memberRepository[TUser, TAccount]) IsAdmin(ctx context.Context, userID, accountID uint64) bool {
	if accountID == 0 || userID == 0 {
		return false
	}
	// Scan (not First): missing membership is the common staff-ACL case and
	// must not log gorm.ErrRecordNotFound.
	var isAdmin bool
	err := r.Slave(ctx).
		Model(&models.MemberBase{}).
		Select("is_admin").
		Where(`account_id=? AND user_id=?`, accountID, userID).
		Limit(1).
		Scan(&isAdmin).Error
	return err == nil && isAdmin
}

func (r *memberRepository[TUser, TAccount]) LinkMember(ctx context.Context, accountObj TAccount, isAdmin bool, members ...TUser) error {
	return r.Master(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&models.MemberBase{}).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "account_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"approve_status", "is_admin", "deleted_at"}),
		})
		for _, userObj := range members {
			err := query.Create(&models.MemberBase{
				Approve:   pkgModels.ApprovedApproveStatus,
				AccountID: accountObj.GetID(),
				UserID:    userObj.GetID(),
				IsAdmin:   isAdmin,
			}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *memberRepository[TUser, TAccount]) UnlinkMember(ctx context.Context, accountObj TAccount, users ...TUser) error {
	ids := make([]uint64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.GetID())
	}
	return r.Master(ctx).
		Where(`account_id=? AND user_id IN ?`, accountObj.GetID(), ids).
		Delete(&models.MemberBase{}).Error
}

func (r *memberRepository[TUser, TAccount]) SetMemberRoles(ctx context.Context, accountObj TAccount, userObj TUser, roles ...string) error {
	var (
		listRoles   []*prbac.Role
		member, err = r.Member(ctx, userObj.GetID(), accountObj.GetID())
	)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrMemberNotFound
	}

	if len(roles) > 0 {
		if listRoles, err = r.rbacUse.FetchList(ctx, &prbac.Filter{Names: roles}); err != nil {
			return err
		}
		if len(listRoles) != len(roles) {
			return ErrInvalidRoleList
		}
	}

	wasAdmin := member.IsAdmin
	member.Roles = listRoles
	member.IsAdmin = xtypes.Slice[string](roles).Has(
		func(v string) bool { return v == "admin" || v == "account:admin" || v == "system:admin" })

	if wasAdmin != member.IsAdmin && !member.IsAdmin {
		cnt, err := r.CountMembers(ctx, &account.MemberFilter{
			AccountID: []uint64{accountObj.GetID()},
			NotUserID: []uint64{userObj.GetID()},
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

	return r.TransactionExec(ctx, func(ctx context.Context, tx *gorm.DB) error {
		err := tx.Omit(clause.Associations).Save(&member.MemberBase).Error
		if err != nil {
			return err
		}
		del := tx.Model((*models.M2MAccountMemberRole)(nil)).
			Where(`member_id=?`, member.ID)
		if len(listRoles) > 0 {
			roleIDs := xtypes.SliceApply(listRoles, func(v *prbac.Role) uint64 { return v.ID })
			del = del.Where(`role_id NOT IN (?)`, roleIDs)
		}
		if err = del.Delete(&models.M2MAccountMemberRole{}).Error; err != nil {
			return err
		}
		if len(listRoles) == 0 {
			return nil
		}
		return tx.Save(xtypes.SliceApply(listRoles, func(v *prbac.Role) *models.M2MAccountMemberRole {
			return &models.M2MAccountMemberRole{
				MemberID:  member.ID,
				RoleID:    v.ID,
				CreatedAt: time.Now(),
			}
		})).Error
	})
}
