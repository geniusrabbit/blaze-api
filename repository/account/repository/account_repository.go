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

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
)

// Repository is a DAO which provides functionality for working with accounts.
type Repository struct {
	repository.Repository
}

// NewAccountRepository creates and returns a new account repository instance.
func NewAccountRepository() *Repository {
	return &Repository{}
}

// Get retrieves an account model by its ID.
func (r *Repository) Get(ctx context.Context, id uint64) (*account.Account, error) {
	object := new(account.Account)
	if err := r.Slave(ctx).Find(object, id).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// LoadPermissions loads and assigns permissions into the account object based on user roles.
// If userObj is nil, returns public permissions; otherwise loads member-specific permissions.
func (r *Repository) LoadPermissions(ctx context.Context, accountObj *account.Account, userObj *account.User) error {
	if accountObj == nil || userObj == nil {
		var err error
		accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, false, nil)
		return err
	}

	var (
		roles  []uint64
		member = new(account.AccountMember)
		query  = r.Slave(ctx)
	)

	// Fetch the account member record for the given user and account.
	if err := query.Find(member, `account_id=? AND user_id=?`, accountObj.ID, userObj.ID).Error; err != nil {
		return errors.WithStack(err)
	}

	// Retrieve all roles assigned to this member.
	err := query.Table((*account.M2MAccountMemberRole)(nil).TableName()).
		Where(`member_id=?`, member.ID).Select(`role_id`).Find(&roles).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, sql.ErrNoRows) {
		return errors.WithStack(err)
	}

	// Mark admin users and load permissions based on approval status.
	if member.IsAdmin {
		accountObj.ExtendAdminUsers(userObj.ID)
	}

	if !accountObj.Approve.IsRejected() && !userObj.Approve.IsRejected() {
		accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, member.IsAdmin, nil, roles...)
	} else {
		// For unapproved accounts, skip system and account roles.
		accountObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, false, func(_ context.Context, r rbac.Role) bool {
			return !strings.HasPrefix(r.Name(), "system:") || !strings.HasPrefix(r.Name(), "account:")
		}, roles...)
	}
	return err
}

// FetchList retrieves a list of accounts filtered, ordered, and paginated according to parameters.
func (r *Repository) FetchList(ctx context.Context, filter *account.Filter, order *account.ListOrder, pagination *account.Pagination) ([]*account.Account, error) {
	var (
		list  []*account.Account
		query = r.Slave(ctx).Model((*account.Account)(nil))
	)

	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
	query = pagination.PrepareQuery(query)
	err := query.Find(&list).Error

	// Treat "no records found" as success with empty list.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

// Count returns the number of accounts matching the filter criteria.
func (r *Repository) Count(ctx context.Context, filter *account.Filter) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*account.Account)(nil))
		err   = filter.PrepareQuery(query).Count(&count).Error
	)
	return count, err
}

// Create inserts a new account record into the database.
func (r *Repository) Create(ctx context.Context, accountObj *account.Account) (uint64, error) {
	accountObj.CreatedAt = time.Now()
	accountObj.UpdatedAt = accountObj.CreatedAt
	accountObj.Approve = model.UndefinedApproveStatus
	err := r.Master(ctx).Create(accountObj).Error
	return accountObj.ID, err
}

// Update modifies an existing account record in the database.
func (r *Repository) Update(ctx context.Context, id uint64, accountObj *account.Account) error {
	obj := *accountObj
	obj.ID = id
	return r.Master(ctx).Updates(&obj).Error
}

// Delete removes an account record by its ID.
func (r *Repository) Delete(ctx context.Context, id uint64) error {
	return r.Master(ctx).Model((*account.Account)(nil)).Delete(`id=?`, id).Error
}

// GetByToken retrieves the user and account objects associated with an authentication token.
func (r *Repository) GetByToken(ctx context.Context, token string) (*account.User, *account.Account, error) {
	var (
		err     error
		roles   []uint64
		db      = r.Slave(ctx)
		userObj = new(account.User)
		accObj  = new(account.Account)
		member  = new(account.AccountMember)
		// Query to find the account member linked to the given token via auth session.
		memberRequest = `WITH auth_client AS (` +
			`  SELECT user_id, account_id FROM ` + (*model.AuthClient)(nil).TableName() + ` WHERE id = (` +
			`    SELECT client_id FROM ` + (*model.AuthSession)(nil).TableName() + ` WHERE deleted_at IS NULL AND access_token=?` +
			`  )` +
			`)` +
			`SELECT am.* FROM ` + (*account.AccountMember)(nil).TableName() + ` AS am, auth_client AS ac` +
			` WHERE am.deleted_at IS NULL AND am.account_id=ac.account_id AND am.user_id=ac.user_id`
	)

	// Fetch member record and associated user and account objects.
	if err = db.Raw(memberRequest, token).Scan(member).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if err = db.First(userObj, member.UserID).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if err = db.First(accObj, member.AccountID).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Fetch all roles assigned to this member.
	err = db.Model(&account.M2MAccountMemberRole{}).
		Select("role_id").Where(`member_id=?`, member.ID).Scan(&roles).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, errors.WithStack(err)
	}

	// Load permissions if member has roles or is an admin.
	if len(roles) > 0 || member.IsAdmin {
		if accObj.Approve.IsApproved() && userObj.Approve.IsApproved() {
			accObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, member.IsAdmin, nil, roles...)
		} else {
			// Skip system roles for unapproved accounts.
			accObj.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, false,
				func(_ context.Context, r rbac.Role) bool {
					return !strings.HasPrefix(r.Name(), "system:")
				}, roles...)
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return userObj, accObj, nil
}
