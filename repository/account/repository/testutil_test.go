package repository_test

import (
	"github.com/geniusrabbit/gosql/v2"

	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
)

type testAccount struct {
	accountModels.AccountBase
	Title       string `json:"title"`
	Description string `json:"description"`

	LogoURI string `json:"logo_uri" gorm:"column:logo_uri"`

	PolicyURI         string `json:"policy_uri" gorm:"column:policy_uri"`
	TermsOfServiceURI string `json:"tos_uri" gorm:"column:tos_uri"`
	ClientURI         string `json:"client_uri" gorm:"column:client_uri"`

	Contacts gosql.NullableStringArray `json:"contacts" gorm:"column:contacts;type:text[]"`
}

func (a *testAccount) TableName() string { return "account_base" }

func (a *testAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	acc := &testAccount{}
	acc.ID = id
	acc.Admins = adminUserIDs
	return acc
}

func testAccountStub(id uint64) *testAccount {
	a := &testAccount{}
	a.ID = id
	return a
}

func testAccountFromProfile(title string) *testAccount {
	a := &testAccount{}
	a.Title = title
	return a
}
