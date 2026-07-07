package account

import (
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// Member links a user to an account with optional preloaded Account and User refs.
type Member[TUser user.Model, TAccount Model] struct {
	models.MemberBase
	Account TAccount `db:"-" gorm:"foreignKey:AccountID;references:ID"`
	User    TUser    `db:"-" gorm:"foreignKey:UserID;references:ID"`
}

// TableName is nil-safe (used in query builders via (*Member[...])(nil).TableName()).
func (m *Member[TUser, TAccount]) TableName() string {
	return models.MemberTableName()
}

// RBACResourceName is nil-safe.
func (m *Member[TUser, TAccount]) RBACResourceName() string {
	return "account.member"
}

// MemberStub returns member with ID and optional account/user keys (ACL/GraphQL placeholders).
func MemberStub[TUser user.Model, TAccount Model](id, accountID, userID uint64) *Member[TUser, TAccount] {
	m := new(Member[TUser, TAccount])
	m.ID = id
	m.AccountID = accountID
	m.UserID = userID
	return m
}

// MemberStubFromUser builds a member ACL placeholder from account and user models.
func MemberStubFromUser[TUser user.Model, TAccount Model](accountObj TAccount, userObj TUser) *Member[TUser, TAccount] {
	var zeroAcc TAccount
	var zeroUser TUser
	m := new(Member[TUser, TAccount])
	if any(accountObj) != any(zeroAcc) {
		m.AccountID = accountObj.GetID()
		m.Account = accountObj
	}
	if any(userObj) != any(zeroUser) {
		m.UserID = userObj.GetID()
		m.User = userObj
	}
	return m
}
