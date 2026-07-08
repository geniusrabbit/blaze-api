package domain

import (
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
)

// Account is the bundled default for example/api: Base + consumer profile trait.
type Account struct {
	accountModels.AccountBase
	AccountProfile
}

// TableName is nil-safe.
func (a *Account) TableName() string {
	return "account_base"
}

// RBACResourceName is nil-safe.
func (a *Account) RBACResourceName() string {
	return "account"
}

// NewWithIDs returns a new account instance with ID and admin user IDs set.
func (a *Account) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	return &Account{AccountBase: accountModels.AccountBase{ID: id, Admins: adminUserIDs}}
}

// AccountMember is the bundled member type for example/api.
type AccountMember = account.Member[*User, *Account]

// AccountStub returns account with only ID set (ACL/GraphQL placeholders).
func AccountStub(id uint64) *Account {
	a := &Account{}
	a.ID = id
	return a
}

// AccountStubWithAdmins returns account stub with admin user IDs.
func AccountStubWithAdmins(id uint64, adminUserIDs ...uint64) *Account {
	a := AccountStub(id)
	a.Admins = adminUserIDs
	return a
}

// AccountFromProfile builds account from profile/base fields.
func AccountFromProfile(approve pkgModels.ApproveStatus, title, description, logoURI, policyURI, tosURI, clientURI string, contacts []string) *Account {
	a := &Account{}
	a.Approve = approve
	a.Title = title
	a.Description = description
	a.LogoURI = logoURI
	a.PolicyURI = policyURI
	a.TermsOfServiceURI = tosURI
	a.ClientURI = clientURI
	a.Contacts = contacts
	return a
}
