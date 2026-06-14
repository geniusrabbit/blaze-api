// Package repository implements methods of working with the repository objects
package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/models"
	"github.com/geniusrabbit/blaze-api/repository/user/password"
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

// NewUserRepository repository accessor to work with users and profiles
func NewUserRepository() *Repository {
	return &Repository{}
}

// Get one object by ID
func (r *Repository) Get(ctx context.Context, id uint64) (*models.User, error) {
	object := new(models.User)
	if err := r.Slave(ctx).First(object, id).Error; err != nil {
		return nil, err
	}
	return object, nil
}

// GetByEmail one object by Email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	object := new(models.User)
	if err := r.Slave(ctx).First(object, `lower(email)=?`, strings.ToLower(email)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return object, nil
}

// GetByPassword user returns user object by password
func (r *Repository) GetByPassword(ctx context.Context, email, password string) (*models.User, error) {
	object, err := r.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if object.Password == "" || !r.comparePasswords(object.Password, []byte(password)) {
		return nil, ErrInvalidPassword
	}
	return object, nil
}

// FetchList of users by filter
func (r *Repository) FetchList(ctx context.Context, opts ...user.QOption) ([]*models.User, error) {
	var (
		list  []*models.User
		query = r.Slave(ctx).Model((*models.User)(nil))
	)
	query = user.ListOptions(opts).PrepareQuery(query)
	err := query.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return list, err
}

// Count of users by filter
func (r *Repository) Count(ctx context.Context, opts ...user.QOption) (int64, error) {
	var (
		count int64
		query = r.Slave(ctx).Model((*models.User)(nil))
	)
	query = user.ListOptions(opts).PrepareQuery(query)
	err := query.Count(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return count, err
}

// SetPassword to the user
func (r *Repository) SetPassword(ctx context.Context, userObj *models.User, password string) error {
	userObj.Password = r.hashAndSalt([]byte(password))
	return r.Update(ctx, userObj)
}

// CreateResetPassword creates new reset password token
func (r *Repository) CreateResetPassword(ctx context.Context, userID uint64) (*models.UserPasswordReset, error) {
	var (
		token   = password.GenerateResetToken(128)
		expires = time.Now().Add(time.Hour * 1)
		reset   = &models.UserPasswordReset{
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
func (r *Repository) GetResetPassword(ctx context.Context, userID uint64, token string) (*models.UserPasswordReset, error) {
	reset := new(models.UserPasswordReset)
	if err := r.Slave(ctx).First(reset, `token=? AND user_id=?`, token, userID).Error; err != nil {
		return nil, err
	}
	return reset, nil
}

// EliminateResetPassword removes reset password token
func (r *Repository) EliminateResetPassword(ctx context.Context, userID uint64) error {
	return r.Master(ctx).Delete(&models.UserPasswordReset{}, `user_id=?`, userID).Error
}

// Create new user object to database
func (r *Repository) Create(ctx context.Context, userObj *models.User, password string) (uint64, error) {
	if password != "" {
		userObj.Password = r.hashAndSalt([]byte(password))
	} else {
		userObj.Password = "" // If password is empty then user can reset it
	}
	userObj.CreatedAt = time.Now()
	userObj.UpdatedAt = userObj.CreatedAt
	userObj.Approve = pkgModels.UndefinedApproveStatus
	err := r.Master(ctx).Create(userObj).Error
	return userObj.ID, err
}

// Update existing object in database
func (r *Repository) Update(ctx context.Context, userObj *models.User) error {
	if userObj.ID == 0 {
		return ErrInvalidUserObject
	}
	return r.Master(ctx).Select("*").Updates(userObj).Error
}

// Delete delites record by ID
func (r *Repository) Delete(ctx context.Context, id uint64) error {
	res := r.Master(ctx).Delete(&models.User{}, id)
	if res.Error != nil {
		return res.Error
	} else if res.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
