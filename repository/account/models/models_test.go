package models_test

import (
	"testing"

	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type testAccount struct {
	models.AccountBase
}

func (a *testAccount) TableName() string { return "account_base" }

func (a *testAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	return &testAccount{AccountBase: models.AccountBase{ID: id, Admins: adminUserIDs}}
}

func TestAccountBaseDefaults(t *testing.T) {
	a := &models.AccountBase{}
	if got := a.TableName(); got != "account_base" {
		t.Fatalf("TableName() = %q, want account_base", got)
	}
	if got := a.RBACResourceName(); got != "account" {
		t.Fatalf("RBACResourceName() = %q, want account", got)
	}
}

func TestMemberGenericTableName(t *testing.T) {
	m := new(account.Member[*testutil.User, *testAccount])
	if got := m.TableName(); got != models.MemberTableName() {
		t.Fatalf("TableName() = %q, want %q", got, models.MemberTableName())
	}
	if got := m.RBACResourceName(); got != "account.member" {
		t.Fatalf("RBACResourceName() = %q, want account.member", got)
	}
}

func TestMemberTableNameHelper(t *testing.T) {
	if got := models.MemberTableName(); got != "account_member" {
		t.Fatalf("MemberTableName() = %q, want account_member", got)
	}
}
