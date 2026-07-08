package account

import "github.com/geniusrabbit/blaze-api/repository/account/models"

// ACLAccount is a minimal account.Model for permission checks (ID + admins only).
type ACLAccount struct {
	models.AccountBase
}

// ACLAccountStub returns an ACL placeholder account for permission checks.
func ACLAccountStub(id uint64, adminUserIDs ...uint64) *ACLAccount {
	a := &ACLAccount{}
	a.ID = id
	a.Admins = adminUserIDs
	return a
}
