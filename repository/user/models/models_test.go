package models_test

import (
	"testing"

	"github.com/geniusrabbit/blaze-api/repository/user/models"
)

// MinimalUser is UserBase-only (no email/password traits).
type MinimalUser struct {
	models.UserBase
}

func (u *MinimalUser) TableName() string        { return "minimal_user" }
func (u *MinimalUser) RBACResourceName() string { return "minimal.user" }

func TestCustomUserOverrides(t *testing.T) {
	u := &MinimalUser{}
	if got := u.TableName(); got != "minimal_user" {
		t.Fatalf("TableName() = %q, want minimal_user", got)
	}
	if got := u.RBACResourceName(); got != "minimal.user" {
		t.Fatalf("RBACResourceName() = %q, want minimal.user", got)
	}
}

// ExtendedUser adds custom fields on top of auth-capable traits.
type ExtendedUser struct {
	models.UserBase
	models.UserEmail
	models.UserPassword
	Country string
}

func (u *ExtendedUser) TableName() string { return "extended_user" }

func TestExtendedUserCustomTable(t *testing.T) {
	u := &ExtendedUser{Country: "US"}
	u.Email = "a@b.c"
	if u.TableName() != "extended_user" {
		t.Fatalf("TableName() = %q", u.TableName())
	}
	if u.RBACResourceName() != "user" {
		t.Fatalf("RBACResourceName() = %q, want promoted user", u.RBACResourceName())
	}
	if u.GetEmail() != "a@b.c" {
		t.Fatalf("GetEmail() = %q", u.GetEmail())
	}
}
