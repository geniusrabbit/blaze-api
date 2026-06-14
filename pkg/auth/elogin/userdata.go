package elogin

import (
	"golang.org/x/oauth2"
)

// UserData represents the user data retrieved from an OAuth2 provider.
type UserData struct {
	// ID is the unique identifier of the user
	ID string `json:"id"`

	// Email is the user's email address
	Email string `json:"email"`

	// FirstName is the user's first name
	FirstName string `json:"first_name"`

	// LastName is the user's last name
	LastName string `json:"last_name"`

	// Username is the user's username or handle
	Username string `json:"username"`

	// AvatarURL is the URL to the user's avatar image
	AvatarURL string `json:"avatar_url"`

	// Link is the user's profile link or homepage URL
	Link string `json:"link"`

	// Ext contains additional provider-specific user data
	Ext map[string]any `json:"ext,omitempty"`

	// OAuth2conf is the OAuth2 configuration (excluded from JSON serialization)
	OAuth2conf *oauth2.Config `json:"-"`
}
