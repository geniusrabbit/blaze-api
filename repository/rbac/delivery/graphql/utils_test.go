package graphql

import (
	"context"
	"testing"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

type testUser struct {
	userModels.UserBase
}

func (u *testUser) NewWithID(id uint64) user.Model {
	return &testUser{UserBase: userModels.UserBase{ID: id}}
}

type testAccount struct {
	accountModels.AccountBase
}

func (a *testAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	return &testAccount{AccountBase: accountModels.AccountBase{ID: id, Admins: adminUserIDs}}
}

func TestOwnedObjectUserMatchesSessionType(t *testing.T) {
	usr := &testUser{UserBase: userModels.UserBase{ID: 7}}
	acc := &testAccount{}
	acc.ID = 3

	got := ownedObject(context.Background(), &testUser{}, usr, acc)
	want := usr.NewWithID(usr.GetID())

	if rbac.GetResType(got) != rbac.GetResType(want) {
		t.Fatalf("user subject type = %v, want %v", rbac.GetResType(got), rbac.GetResType(want))
	}
	gotUser, ok := got.(*testUser)
	if !ok {
		t.Fatalf("user subject = %T, want *testUser (not a local stub)", got)
	}
	if gotUser.GetID() != usr.GetID() {
		t.Fatalf("user subject ID = %d, want %d", gotUser.GetID(), usr.GetID())
	}
}

func TestOwnedObjectAccountMatchesSessionType(t *testing.T) {
	usr := &testUser{UserBase: userModels.UserBase{ID: 7}}
	acc := &testAccount{}
	acc.ID = 3

	got := ownedObject(context.Background(), &testAccount{}, usr, acc)
	want := acc.NewWithIDs(acc.GetID(), usr.GetID())

	if rbac.GetResType(got) != rbac.GetResType(want) {
		t.Fatalf("account subject type = %v, want %v", rbac.GetResType(got), rbac.GetResType(want))
	}
	if _, isStub := got.(*account.ACLAccount); isStub {
		t.Fatalf("account subject is *account.ACLAccount; want consumer *testAccount")
	}
	gotAcc, ok := got.(*testAccount)
	if !ok {
		t.Fatalf("account subject = %T, want *testAccount", got)
	}
	if gotAcc.GetID() != acc.GetID() {
		t.Fatalf("account subject ID = %d, want %d", gotAcc.GetID(), acc.GetID())
	}
	if !gotAcc.IsAdminUser(usr.GetID()) {
		t.Fatalf("account subject missing session user %d in Admins", usr.GetID())
	}
}
