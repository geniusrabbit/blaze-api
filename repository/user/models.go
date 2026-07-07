package user

import "github.com/geniusrabbit/blaze-api/repository/user/models"

// UserPasswordReset is the model for password reset tokens.
type UserPasswordReset = models.UserPasswordReset

// Model trait type aliases — compose these in your consumer user struct.
type (
	// Base is the minimal user identity (ID, approve status, timestamps).
	Base = models.UserBase
	// Email is the optional email identity trait.
	Email = models.UserEmail
	// Password is the optional password authentication trait.
	Password = models.UserPassword
	// Username is the optional separate username trait.
	Username = models.UserUsername
)
