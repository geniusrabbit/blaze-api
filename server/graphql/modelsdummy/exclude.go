package modelsdummy

import (
	"time"

	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// Account is a company account that can be used to login to the system.
// Core fields only — extend in consumer schema via `extend type Account`.
type Account struct {
	// The primary key of the Account
	ID uint64 `json:"ID"`
	// Status of Account active
	Status models.ApproveStatus `json:"status"`
	// Message which defined during user approve/rejection process
	StatusMessage *string   `json:"statusMessage,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// AccountConnection implements collection accessor interface with pagination.
type AccountConnection struct {
	// The total number of campaigns
	TotalCount int `json:"totalCount"`
	// A list of the accounts, as a convenience when edges are not needed.
	List []*Account `json:"list,omitempty"`
	// Information for paginating this connection
	PageInfo *models.PageInfo `json:"pageInfo"`
}

type AccountCreateInput struct {
	Status *models.ApproveStatus `json:"status,omitempty"`
}

type AccountListFilter struct {
	ID     []uint64               `json:"ID,omitempty"`
	UserID []uint64               `json:"UserID,omitempty"`
	Status []models.ApproveStatus `json:"status,omitempty"`
}

type AccountListOrder struct {
	ID        *models.Ordering `json:"ID,omitempty"`
	Status    *models.Ordering `json:"status,omitempty"`
	CreatedAt *models.Ordering `json:"createdAt,omitempty"`
	UpdatedAt *models.Ordering `json:"updatedAt,omitempty"`
}

// AccountPayload wrapper to access of Account oprtation results
type AccountPayload struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationID string `json:"clientMutationID"`
	// Account ID operation result
	AccountID uint64 `json:"accountID"`
	// Account object accessor
	Account *Account `json:"account,omitempty"`
}

type AccountUpdateInput struct {
	Status *models.ApproveStatus `json:"status,omitempty"`
}

// User represents a user object of the system.
// Core fields only — extend in consumer schema via `extend type User`.
type User struct {
	// The primary key of the user
	ID uint64 `json:"ID"`
	// Status of user active
	Status models.ApproveStatus `json:"status"`
	// Message which defined during user approve/rejection process
	StatusMessage *string   `json:"statusMessage,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	// Email address (optional trait — present only when user.Email is embedded).
	Email string `json:"email"`
	// Unique username (separate from email, optional trait).
	Username string `json:"username"`
}

// UserConnection implements collection accessor interface with pagination.
type UserConnection struct {
	// The total number of campaigns
	TotalCount int `json:"totalCount"`
	// A list of the users, as a convenience when edges are not needed.
	List []*User `json:"list,omitempty"`
	// Information for paginating this connection
	PageInfo *models.PageInfo `json:"pageInfo"`
}

type UserCreateInput struct {
	Status   *models.ApproveStatus `json:"status,omitempty"`
	Email    string                `json:"email"`
	Username *string               `json:"username,omitempty"`
}

// UserListFilter implements filter for user list query
type UserListFilter struct {
	ID     []uint64 `json:"ID,omitempty"`
	Emails []string `json:"emails,omitempty"`
}

// UserListOrder implements order for user list query
type UserListOrder struct {
	ID        *models.Ordering `json:"ID,omitempty"`
	Status    *models.Ordering `json:"status,omitempty"`
	CreatedAt *models.Ordering `json:"createdAt,omitempty"`
	UpdatedAt *models.Ordering `json:"updatedAt,omitempty"`
	Email     *models.Ordering `json:"email,omitempty"`
	Username  *models.Ordering `json:"username,omitempty"`
}

// UserPayload wrapper to access of user oprtation results
type UserPayload struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationID string `json:"clientMutationID"`
	// User ID operation result
	UserID uint64 `json:"userID"`
	// User object accessor
	User *User `json:"user,omitempty"`
}

type UserUpdateInput struct {
	Status *models.ApproveStatus `json:"status,omitempty"`
	Email  *string               `json:"email,omitempty"`
}
