package user

import (
	"github.com/geniusrabbit/blaze-api/pkg/auth"
)

// Model is the compile-time constraint for core user repository/usecase operations.
type Model interface {
	auth.IsNillable

	GetID() uint64
	IsAnonymous() bool
	NewWithID(id uint64) Model

	CreatorUserID() uint64
}

// EmailModel is the email identity trait (orthogonal to Model).
type EmailModel interface {
	GetEmail() string
	SetEmail(string)
}

// PasswordModel is the password authentication trait (orthogonal to Model and EmailModel).
type PasswordModel interface {
	GetPasswordHash() string
	SetPasswordHash(string)
	RequiredPasswordReset() bool
}

// EmailCapableModel combines core Model with email trait for email repository/usecase.
type EmailCapableModel interface {
	Model
	EmailModel
}

// PasswordCapableModel combines core Model with password trait for password repository/usecase.
type PasswordCapableModel interface {
	Model
	PasswordModel
}

// AuthCapableModel combines core user with email and password traits.
// Use only as constraint for PasswordResetQueryResolver (email → userID → password usecase).
type AuthCapableModel interface {
	Model
	EmailModel
	PasswordModel
}
