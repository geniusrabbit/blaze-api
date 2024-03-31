package model

import (
	"context"
	"time"

	"github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"
	"github.com/geniusrabbit/gosql/v2"
	"gorm.io/gorm"
)

// Account provides the information about the account
type Account struct {
	ID      uint64        `json:"id" gorm:"primaryKey"`
	Approve ApproveStatus `json:"approved" db:"approve_status" gorm:"column:approve_status" `

	Title       string `json:"title"`
	Description string `json:"description"`

	// LogoURI is an URL string that references a logo for the client.
	LogoURI string `json:"logo_uri" gorm:"column:logo_uri"`

	// PolicyURI is a URL string that points to a human-readable privacy policy document
	// that describes how the deployment organization collects, uses,
	// retains, and discloses personal data.
	PolicyURI string `json:"policy_uri" gorm:"column:policy_uri"`

	// TermsOfServiceURI is a URL string that points to a human-readable terms of service
	// document for the client that describes a contractual relationship
	// between the end-user and the client that the end-user accepts when
	// authorizing the client.
	TermsOfServiceURI string `json:"tos_uri" gorm:"column:tos_uri"`

	// ClientURI is an URL string of a web page providing information about the client.
	// If present, the server SHOULD display this URL to the end-user in
	// a clickable fashion.
	ClientURI string `json:"client_uri" gorm:"column:client_uri"`

	// Contacts is a array of strings representing ways to contact people responsible
	// for this client, typically email addresses.
	Contacts gosql.NullableStringArray `json:"contacts" gorm:"column:contacts;type:text[]"`

	Permissions permissionChecker `json:"-" gorm:"-"`
	Admins      []uint64          `json:"-" gorm:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

// TableName of the model in the database
func (acc *Account) TableName() string {
	return `account_base`
}

// IsAnonymous account
func (acc *Account) IsAnonymous() bool {
	return acc == nil || acc.ID == 0
}

// IsApproved account
func (acc *Account) IsAdminUser(userID uint64) bool {
	if acc == nil || len(acc.Admins) == 0 {
		return false
	}
	return xtypes.Slice[uint64](acc.Admins).Has(func(id uint64) bool { return id == userID })
}

// RBACResourceName returns the name of the resource for the RBAC
func (acc *Account) RBACResourceName() string {
	return "account"
}

// ExtendAdminUsers to the account
func (acc *Account) ExtendAdminUsers(ids ...uint64) {
	if acc == nil {
		return
	}
	acc.Admins = xtypes.SliceUnique[uint64](append(acc.Admins, ids...))
}

// CheckPermissions for some specific resource
func (acc *Account) CheckPermissions(ctx context.Context, resource any, patterns ...string) bool {
	if acc == nil || acc.Permissions == nil {
		return false
	}
	ctx = context.WithValue(ctx, ctxPermissionCheckAccount, acc)
	return acc.Permissions.CheckPermissions(ctx, resource, patterns...)
}

// CheckedPermissions for some specific resource
func (acc *Account) CheckedPermissions(ctx context.Context, resource any, patterns ...string) rbac.Permission {
	if acc == nil || acc.Permissions == nil {
		return nil
	}
	ctx = context.WithValue(ctx, ctxPermissionCheckAccount, acc)
	return acc.Permissions.CheckedPermissions(ctx, resource, patterns...)
}

// ListPermissions for the account
func (acc *Account) ListPermissions(patterns ...string) []rbac.Permission {
	if acc == nil || acc.Permissions == nil {
		return nil
	}
	return acc.Permissions.Permissions(patterns...)
}

// HasPermission for the account
func (acc *Account) HasPermission(patterns ...string) bool {
	return acc.Permissions.HasPermission(patterns...)
}

// OwnerAccountID returns the account ID which belongs the object
func (acc *Account) OwnerAccountID() uint64 {
	return acc.ID
}

// IsOwnerUser of the account
func (acc *Account) IsOwnerUser(userID uint64) bool {
	return acc.IsAdminUser(userID)
}

// ExtendPermissions of the account for the user
func (acc *Account) ExtendPermissions(perm permissionChecker) {
	if perm == nil {
		return
	}
	switch prev := acc.Permissions.(type) {
	case groupPermissionChecker:
		prev = append(prev, perm)
		acc.Permissions = prev
	case nil:
		acc.Permissions = perm
	default:
		acc.Permissions = groupPermissionChecker{prev, perm}
	}
}
