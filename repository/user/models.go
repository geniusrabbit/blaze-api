package user

import "github.com/geniusrabbit/blaze-api/repository/user/models"

type (
	// User is the main user model
	User = models.User

	// UserPasswordReset is the model for password reset tokens
	UserPasswordReset = models.UserPasswordReset
)

// Anonymous user object
var Anonymous = User{ID: 0}
