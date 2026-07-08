package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/demdxx/rbac"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	baseRepo "github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	authclientModels "github.com/geniusrabbit/blaze-api/repository/authclient/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

type sessionRepository[TUser user.Model, TAccount account.Model] struct {
	baseRepo.Repository
	core         account.Repository[TAccount]
	newUserModel func() TUser
	newMember    func() *account.Member[TUser, TAccount]
}

// NewSessionRepository creates account repository with session/auth helpers.
func NewSessionRepository[TUser user.Model, TAccount account.Model](
	newUser func() TUser,
	newAccount func() TAccount,
	newMember func() *account.Member[TUser, TAccount],
) account.SessionRepository[TUser, TAccount] {
	return &sessionRepository[TUser, TAccount]{
		core:         NewRepository(newAccount),
		newUserModel: newUser,
		newMember:    newMember,
	}
}

func (r *sessionRepository[TUser, TAccount]) EmptyObject() TAccount {
	return r.core.EmptyObject()
}

func (r *sessionRepository[TUser, TAccount]) Get(ctx context.Context, id uint64) (TAccount, error) {
	return r.core.Get(ctx, id)
}

func (r *sessionRepository[TUser, TAccount]) FetchList(ctx context.Context, opts ...account.QOption) ([]TAccount, error) {
	return r.core.FetchList(ctx, opts...)
}

func (r *sessionRepository[TUser, TAccount]) Count(ctx context.Context, opts ...account.QOption) (int64, error) {
	return r.core.Count(ctx, opts...)
}

func (r *sessionRepository[TUser, TAccount]) Create(ctx context.Context, accountObj TAccount) (uint64, error) {
	return r.core.Create(ctx, accountObj)
}

func (r *sessionRepository[TUser, TAccount]) Update(ctx context.Context, id uint64, accountObj TAccount) error {
	return r.core.Update(ctx, id, accountObj)
}

func (r *sessionRepository[TUser, TAccount]) Delete(ctx context.Context, id uint64) error {
	return r.core.Delete(ctx, id)
}

func (r *sessionRepository[TUser, TAccount]) LoadPermissions(ctx context.Context, accountObj TAccount, userObj TUser) error {
	var zeroAcc TAccount
	var zeroUser TUser
	if any(accountObj) == any(zeroAcc) || any(userObj) == any(zeroUser) {
		perm, err := r.PermissionManager(ctx).AsOneRole(ctx, false, nil)
		if err != nil {
			return err
		}
		if any(accountObj) != any(zeroAcc) {
			accountObj.SetPermissions(perm)
		}
		return nil
	}

	var (
		roles  []uint64
		member models.MemberBase
		query  = r.Slave(ctx).Model(&models.MemberBase{})
	)

	if err := query.Find(&member, `account_id=? AND user_id=?`, accountObj.GetID(), userObj.GetID()).Error; err != nil {
		return errors.WithStack(err)
	}

	err := r.Slave(ctx).Table((*models.M2MAccountMemberRole)(nil).TableName()).
		Where(`member_id=?`, member.ID).Select(`role_id`).Find(&roles).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, sql.ErrNoRows) {
		return errors.WithStack(err)
	}

	if member.IsAdmin {
		accountObj.ExtendAdminUsers(userObj.GetID())
	}

	userApprove := getApprove(userObj)
	accApprove := getApprove(accountObj)

	if !accApprove.IsRejected() && !userApprove.IsRejected() {
		perm, err := r.PermissionManager(ctx).AsOneRole(ctx, member.IsAdmin, nil, roles...)
		if err != nil {
			return err
		}
		accountObj.SetPermissions(perm)
		return nil
	}

	perm, err := r.PermissionManager(ctx).AsOneRole(ctx, false, func(_ context.Context, role rbac.Role) bool {
		return !strings.HasPrefix(role.Name(), "system:") || !strings.HasPrefix(role.Name(), "account:")
	}, roles...)
	if err != nil {
		return err
	}
	accountObj.SetPermissions(perm)
	return nil
}

func (r *sessionRepository[TUser, TAccount]) GetByToken(ctx context.Context, token string) (TUser, TAccount, error) {
	var (
		err           error
		roles         []uint64
		db            = r.Slave(ctx)
		userObj       = r.newUserModel()
		accObj        = r.EmptyObject()
		zeroUser      TUser
		zeroAcc       TAccount
		member        = r.newMember()
		memberTable   = models.MemberTableName()
		memberRequest = `WITH auth_client AS (` +
			`  SELECT user_id, account_id FROM ` + (*authclientModels.AuthClient)(nil).TableName() + ` WHERE id = (` +
			`    SELECT client_id FROM ` + (*authclientModels.AuthSession)(nil).TableName() + ` WHERE deleted_at IS NULL AND access_token=?` +
			`  )` +
			`)` +
			`SELECT am.* FROM ` + memberTable + ` AS am, auth_client AS ac` +
			` WHERE am.deleted_at IS NULL AND am.account_id=ac.account_id AND am.user_id=ac.user_id`
	)

	if err = db.Raw(memberRequest, token).Scan(member).Error; err != nil {
		return zeroUser, zeroAcc, errors.WithStack(err)
	}
	if err = db.First(userObj, member.UserID).Error; err != nil {
		return zeroUser, zeroAcc, errors.WithStack(err)
	}
	if err = db.First(accObj, member.AccountID).Error; err != nil {
		return zeroUser, zeroAcc, errors.WithStack(err)
	}

	err = db.Model(&models.M2MAccountMemberRole{}).
		Select("role_id").Where(`member_id=?`, member.ID).Scan(&roles).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return zeroUser, zeroAcc, errors.WithStack(err)
	}

	if len(roles) > 0 || member.IsAdmin {
		userApprove := getApprove(userObj)
		accApprove := getApprove(accObj)
		if accApprove.IsApproved() && userApprove.IsApproved() {
			perm, perr := r.PermissionManager(ctx).AsOneRole(ctx, member.IsAdmin, nil, roles...)
			if perr != nil {
				return zeroUser, zeroAcc, perr
			}
			accObj.SetPermissions(perm)
		} else {
			perm, perr := r.PermissionManager(ctx).AsOneRole(ctx, false,
				func(_ context.Context, role rbac.Role) bool {
					return !strings.HasPrefix(role.Name(), "system:")
				}, roles...)
			if perr != nil {
				return zeroUser, zeroAcc, perr
			}
			accObj.SetPermissions(perm)
		}
	}
	return userObj, accObj, nil
}

type approveGetter interface {
	GetApprove() pkgModels.ApproveStatus
}

func getApprove(obj any) pkgModels.ApproveStatus {
	if v, ok := obj.(approveGetter); ok {
		return v.GetApprove()
	}
	return pkgModels.UndefinedApproveStatus
}
