package models_test

import (
	"testing"

	"github.com/geniusrabbit/blaze-api/repository/account/models"
)

func TestAccountBaseDefaults(t *testing.T) {
	a := &models.AccountBase{}
	if got := a.TableName(); got != "account_base" {
		t.Fatalf("TableName() = %q, want account_base", got)
	}
	if got := a.RBACResourceName(); got != "account" {
		t.Fatalf("RBACResourceName() = %q, want account", got)
	}
}
