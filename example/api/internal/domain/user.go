package domain

import (
	"github.com/geniusrabbit/blaze-api/repository/user"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// User is the example consumer user model (Base + Email + Password traits).
type User struct {
	userModels.UserBase
	userModels.UserEmail
	userModels.UserPassword
	// userModels.UserUsername
}

func (u *User) NewWithID(id uint64) user.Model {
	return &User{UserBase: userModels.UserBase{ID: id}}
}

// Anonymous is the example anonymous session user placeholder.
var Anonymous = User{UserBase: userModels.UserBase{ID: 0}}
