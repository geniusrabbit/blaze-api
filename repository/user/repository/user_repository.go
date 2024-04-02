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
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// Errors list...
var (
	ErrInvalidPassword   = errors.New(`invalid password`)
	ErrInvalidUserObject = errors.New(`invalid object`)
)

// Repository DAO which provides functionality of working with users and authorization
type Repository struct {
	repository.Repository
}

// New repository accessor to work with users and profiles
func New() *Repository {
	return &Repository{}
}

// Get one object by ID
func (r *Repository) Get(ctx context.Context, id uint64) (*model.User, error) {
	object := new(model.User)
	if err := r.Slave(ctx).First(object, id).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// GetByEmail one object by Email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	object := new(model.User)
	if err := r.Slave(ctx).First(object, `lower(email)=?`, strings.ToLower(email)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return object, nil
}

// GetByPassword user returns user object by password
func (r *Repository) GetByPassword(ctx context.Context, email, password string) (*model.User, error) {
	object, err := r.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if object.Password == "" || !r.comparePasswords(object.Password, []byte(password)) {
		return nil, ErrInvalidPassword
	}
	return object, nil
}

// GetByToken returns the user object linked to the token (external session ID)
func (r *Repository) GetByToken(ctx context.Context, token string) (*model.User, *model.Account, error) {
	var (
		err           error
		roles         []uint64
		db            = r.Slave(ctx)
		userObj       = new(model.User)
		account       = new(model.Account)
		memeber       = new(model.AccountMember)
		memberRequest = `WITH auth_client AS (` +
			`  SELECT user_id, account_id FROM ` + (*model.AuthClient)(nil).TableName() + ` WHERE id = (` +
			`    SELECT client_id FROM ` + (*model.AuthSession)(nil).TableName() + ` WHERE deleted_at IS NULL AND access_token=?` +
			`  )` +
			`)` +
			`SELECT am.* FROM ` + (*model.AccountMember)(nil).TableName() + ` AS am, auth_client AS ac` +
			` WHERE am.deleted_at IS NULL AND am.account_id=ac.account_id AND am.user_id=ac.user_id`
	)
	if err = db.Raw(memberRequest, token).Scan(memeber).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if err = db.First(userObj, memeber.UserID).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if err = db.First(account, memeber.AccountID).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	err = db.Model(&model.M2MAccountMemberRole{}).
		Select("role_id").Where(`member_id=?`, memeber.ID).Scan(&roles).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		// `sql.ErrNoRows` in case of no any linked permissions
		return nil, nil, errors.WithStack(err)
	}
	if len(roles) > 0 || memeber.IsAdmin {
		if account.Approve.IsApproved() && userObj.Approve.IsApproved() {
			account.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, memeber.IsAdmin, nil, roles...)
		} else {
			account.Permissions, err = r.PermissionManager(ctx).AsOneRole(ctx, false, func(_ context.Context, r rbac.Role) bool {
				return !strings.HasPrefix(r.Name(), "system:")
			}, roles...)
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return userObj, account, nil
}

// FetchList of users by filter
func (r *Repository) FetchList(ctx context.Context, filter *user.ListFilter, order *user.ListOrder, page *repository.Pagination) ([]*model.User, error) {
	var (
		list  []*model.User
		query = r.Slave(ctx).Model((*model.User)(nil))
	)
	query = filter.PrepareQuery(query)
	query = order.PrepareQuery(query)
	query = page.PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

// Count of users by filter
func (r *Repository) Count(ctx context.Context, filter *user.ListFilter) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*model.User)(nil))
	)
	query = filter.PrepareQuery(query)
	err := query.Count(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return count, err
}

// SetPassword to the user
func (r *Repository) SetPassword(ctx context.Context, userObj *model.User, password string) error {
	userObj.Password = r.hashAndSalt([]byte(password))
	return r.Update(ctx, userObj)
}

// CreateResetPassword creates new reset password token
func (r *Repository) CreateResetPassword(ctx context.Context, userID uint64) (*model.UserPasswordReset, error) {
	var (
		token   = generateResetToken(128)
		expires = time.Now().Add(time.Hour * 1)
		reset   = &model.UserPasswordReset{
			UserID:    userID,
			Token:     token,
			CreatedAt: time.Now(),
			ExpiresAt: expires,
		}
	)
	if err := r.Master(ctx).Create(reset).Error; err != nil {
		return nil, err
	}
	return reset, nil
}

// GetResetPassword returns reset password token
func (r *Repository) GetResetPassword(ctx context.Context, userID uint64, token string) (*model.UserPasswordReset, error) {
	reset := new(model.UserPasswordReset)
	if err := r.Slave(ctx).First(reset, `token=? AND user_id=?`, token, userID).Error; err != nil {
		return nil, err
	}
	return reset, nil
}

// EliminateResetPassword removes reset password token
func (r *Repository) EliminateResetPassword(ctx context.Context, userID uint64) error {
	return r.Master(ctx).Delete(&model.UserPasswordReset{}, `user_id=?`, userID).Error
}

// Create new user object to database
func (r *Repository) Create(ctx context.Context, userObj *model.User, password string) (uint64, error) {
	if password != "" {
		userObj.Password = r.hashAndSalt([]byte(password))
	} else {
		userObj.Password = "" // If password is empty then user can reset it
	}
	userObj.CreatedAt = time.Now()
	userObj.UpdatedAt = userObj.CreatedAt
	userObj.Approve = model.UndefinedApproveStatus
	err := r.Master(ctx).Create(userObj).Error
	return userObj.ID, err
}

// Update existing object in database
func (r *Repository) Update(ctx context.Context, userObj *model.User) error {
	if userObj.ID == 0 {
		return ErrInvalidUserObject
	}
	return r.Master(ctx).Select("*").Updates(userObj).Error
}

// Delete delites record by ID
func (r *Repository) Delete(ctx context.Context, id uint64) error {
	res := r.Master(ctx).Delete(&model.User{}, id)
	if res.Error != nil {
		return res.Error
	} else if res.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
