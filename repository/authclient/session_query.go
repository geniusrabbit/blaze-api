package authclient

import (
	"strings"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/models"
)

// SessionFilter of the auth_session list
type SessionFilter struct {
	ID       []uint64
	ClientID []string
	Username []string
	Subject  []string
	Active   *bool
	Query    string
}

// PrepareQuery applies filter conditions to a GORM query
func (fl *SessionFilter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	if len(fl.ID) > 0 {
		query = query.Where(`id IN (?)`, fl.ID)
	}
	if len(fl.ClientID) > 0 {
		query = query.Where(`client_id IN (?)`, fl.ClientID)
	}
	if len(fl.Username) > 0 {
		query = query.Where(`username IN (?)`, fl.Username)
	}
	if len(fl.Subject) > 0 {
		query = query.Where(`subject IN (?)`, fl.Subject)
	}
	if fl.Active != nil {
		query = query.Where(`active = ?`, *fl.Active)
	}
	if q := strings.TrimSpace(fl.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where(
			`LOWER(username) LIKE ? OR LOWER(client_id) LIKE ? OR LOWER(subject) LIKE ? OR CAST(id AS TEXT) LIKE ?`,
			like, like, like, like,
		)
	}
	return query
}

// SessionListOrder defines sort order for session query results
type SessionListOrder struct {
	ID                   models.Order
	ClientID             models.Order
	CreatedAt            models.Order
	AccessTokenExpiresAt models.Order
}

// PrepareQuery applies sort order to a GORM query
func (ord *SessionListOrder) PrepareQuery(query *gorm.DB) *gorm.DB {
	if ord == nil {
		return query
	}
	query = ord.ID.PrepareQuery(query, "id")
	query = ord.ClientID.PrepareQuery(query, "client_id")
	query = ord.CreatedAt.PrepareQuery(query, "created_at")
	query = ord.AccessTokenExpiresAt.PrepareQuery(query, "access_token_expires_at")
	return query
}
