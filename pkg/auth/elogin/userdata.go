package elogin

import (
	"golang.org/x/oauth2"
)

// UserData represents the user data
type UserData struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Link      string `json:"link"`

	// Ext is the extra data
	Ext map[string]any `json:"ext,omitempty"`

	OAuth2conf *oauth2.Config `json:"-"`
}
