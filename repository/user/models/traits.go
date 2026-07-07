package models

// UserEmail optional trait: email identity.
type UserEmail struct {
	Email string `json:"email" gorm:"column:email;not null;default:'';"`
}

// GetEmail returns the user email.
func (u *UserEmail) GetEmail() string {
	if u == nil {
		return ""
	}
	return u.Email
}

// SetEmail sets the user email.
func (u *UserEmail) SetEmail(email string) {
	if u != nil {
		u.Email = email
	}
}

// EmailColumn returns the database column name.
func (u *UserEmail) EmailColumn() string {
	return "email"
}

// UserUsername optional trait: separate username (GraphQL compatibility).
type UserUsername struct {
	Username string `json:"username" gorm:"column:username"`
}

// GetUsername returns the username.
func (u *UserUsername) GetUsername() string {
	if u == nil {
		return ""
	}
	return u.Username
}

// UserPassword optional trait: password authentication.
type UserPassword struct {
	Password          string `json:"password" gorm:"column:password;not null;default:'';"`
	MustResetPassword bool   `json:"required_password_reset" gorm:"column:required_password_reset;not null;default:false"`
}

// GetPasswordHash returns the stored password hash.
func (u *UserPassword) GetPasswordHash() string {
	if u == nil {
		return ""
	}
	return u.Password
}

// SetPasswordHash sets the stored password hash.
func (u *UserPassword) SetPasswordHash(hash string) {
	if u != nil {
		u.Password = hash
	}
}

// RequiredPasswordReset reports whether the user must reset password on next login.
func (u *UserPassword) RequiredPasswordReset() bool {
	if u == nil {
		return false
	}
	return u.MustResetPassword
}

// PasswordColumn returns the database column name.
func (u *UserPassword) PasswordColumn() string {
	return "password"
}
