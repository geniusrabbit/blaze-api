package testutil

import (
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/models"
)

// User is a minimal auth-capable user type for framework tests.
type User struct {
	models.UserBase
	models.UserEmail
	models.UserPassword
}

func (u *User) TableName() string { return "test_user" }

func (u *User) RBACResourceName() string { return "test.user" }

func (u *User) NewWithID(id uint64) user.Model {
	return &User{UserBase: models.UserBase{ID: id}}
}

// Stub returns a test user with the given ID.
func Stub(id uint64) *User {
	u := &User{}
	u.SetID(id)
	return u
}

// StubWithEmail returns a test user with email and approval status.
func StubWithEmail(email string, approve pkgModels.ApproveStatus) *User {
	u := &User{}
	u.SetEmail(email)
	u.Approve = approve
	return u
}
