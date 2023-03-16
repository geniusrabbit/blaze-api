// Package usecase user managing
package usecase

import (
	"context"
	"database/sql"

	"github.com/geniusrabbit/api-template-base/internal/acl"
	"github.com/geniusrabbit/api-template-base/internal/context/session"
	"github.com/geniusrabbit/api-template-base/internal/repository/user"
	"github.com/geniusrabbit/api-template-base/model"
)

const passwordUpdatePermission = `user:password.update`

// UserUsecase provides bussiness logic for user access
type UserUsecase struct {
	userRepo user.Repository
}

// NewUserUsecase user implementation
func NewUserUsecase(repo user.Repository) *UserUsecase {
	return &UserUsecase{userRepo: repo}
}

// Get returns the group by ID if have access
func (a *UserUsecase) Get(ctx context.Context, id uint64) (*model.User, error) {
	currentUser, _ := session.UserAccount(ctx)
	if currentUser.ID == id {
		if !acl.HaveAccessView(ctx, currentUser) {
			return nil, acl.ErrNoPermissions
		}
		return currentUser, nil
	}
	userObj, err := a.userRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, userObj) {
		return nil, acl.ErrNoPermissions
	}
	return userObj, nil
}

// GetByEmail returns the group by Email if have access
func (a *UserUsecase) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	currentUser, _ := session.UserAccount(ctx)
	if currentUser.Email == email {
		if !acl.HaveAccessView(ctx, currentUser) {
			return nil, acl.ErrNoPermissions
		}
		return currentUser, nil
	}
	userObj, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, userObj) {
		return nil, acl.ErrNoPermissions
	}
	return userObj, nil
}

// GetByPassword returns user by email + password
func (a *UserUsecase) GetByPassword(ctx context.Context, email, password string) (*model.User, error) {
	return a.userRepo.GetByPassword(ctx, email, password)
}

// GetByToken returns user + account by session token
func (a *UserUsecase) GetByToken(ctx context.Context, token string) (*model.User, *model.Account, error) {
	return a.userRepo.GetByToken(ctx, token)
}

// FetchList of users by filter
func (a *UserUsecase) FetchList(ctx context.Context, accountID uint64, page, num int) ([]*model.User, error) {
	if accountID < 1 {
		account := session.Account(ctx)
		accountID = account.ID
	}
	if !acl.HaveAccessList(ctx, &model.User{}) {
		return nil, acl.ErrNoPermissions
	}
	return a.userRepo.FetchList(ctx, &user.ListFilter{AccountID: []uint64{accountID}}, page, num)
}

// SetPassword for the exists user
func (a *UserUsecase) SetPassword(ctx context.Context, userObj *model.User, password string) error {
	if !acl.HaveAccessUpdate(ctx, userObj) || !acl.HavePermissions(ctx, passwordUpdatePermission) {
		return acl.ErrNoPermissions
	}
	return a.userRepo.SetPassword(ctx, userObj, password)
}

// Store new object into database
func (a *UserUsecase) Store(ctx context.Context, userObj *model.User, password string) (uint64, error) {
	var err error
	if userObj.ID == 0 && !acl.HaveAccessCreate(ctx, userObj) {
		return 0, acl.ErrNoPermissions
	}
	if userObj.ID != 0 && !acl.HaveAccessUpdate(ctx, userObj) {
		return 0, acl.ErrNoPermissions
	}
	if userObj.ID == 0 {
		userObj.ID, err = a.userRepo.Create(ctx, userObj, password)
	} else {
		err = a.userRepo.Update(ctx, userObj)
	}
	return userObj.ID, err
}

// Update existing object in database
func (a *UserUsecase) Update(ctx context.Context, userObj *model.User) error {
	if !acl.HaveAccessUpdate(ctx, userObj) {
		return acl.ErrNoPermissions
	}
	return a.userRepo.Update(ctx, userObj)
}

// Delete delites record by ID
func (a *UserUsecase) Delete(ctx context.Context, id uint64) error {
	userObj, err := a.getUserByID(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, userObj) {
		return acl.ErrNoPermissions
	}
	return a.userRepo.Delete(ctx, id)
}

func (a *UserUsecase) getUserByID(ctx context.Context, id uint64) (*model.User, error) {
	currentUser := session.User(ctx)
	if currentUser.ID == id {
		return currentUser, nil
	}
	return nil, sql.ErrNoRows
}
