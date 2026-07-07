package custommodels

import (
	"time"

	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// Account Member represents a member of the account
type Member[TUser any, TAccount any] struct {
	// The primary key of the Member
	ID uint64 `json:"ID"`
	// Status of Member active
	Status gqlmodels.ApproveStatus `json:"status"`
	// User object accessor
	User TUser `json:"user"`
	// Account object accessor
	Account TAccount `json:"account"`
	// Is the user an admin of the account
	IsAdmin bool `json:"isAdmin"`
	// Roles of the member
	Roles     []*gqlmodels.RBACRole `json:"roles,omitempty"`
	CreatedAt time.Time             `json:"createdAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
	DeletedAt *time.Time            `json:"deletedAt,omitempty"`
}

type MemberEdge[TUser any, TAccount any] struct {
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
	// The item at the end of the edge.
	Node *Member[TUser, TAccount] `json:"node,omitempty"`
}
