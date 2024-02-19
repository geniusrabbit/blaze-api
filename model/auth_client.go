package model

import (
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"gorm.io/gorm"
)

// AuthClient object represents an OAuth 2.0 client
type AuthClient struct {
	// ClientID is the client ID which represents unique connection indentificator
	ID string `db:"id"`

	// Owner and creator of the auth client
	AccountID uint64 `db:"account_id"`
	UserID    uint64 `db:"user_id"`

	// Title of the AuthClient as himan readable name
	Title string `db:"title"`

	// Secret is the client's secret. The secret will be included in the create request as cleartext, and then
	// never again. The secret is stored using BCrypt so it is impossible to recover it. Tell your users
	// that they need to write the secret down as it will not be made available again.
	Secret string `db:"secret"`

	// RedirectURIs is an array of allowed redirect urls for the client, for example http://mydomain/oauth/callback .
	RedirectURIs gosql.NullableStringArray `db:"redirect_uris" gorm:"type:text[]"`

	// GrantTypes is an array of grant types the client is allowed to use.
	//
	// Pattern: client_credentials|authorization_code|implicit|refresh_token
	GrantTypes gosql.NullableStringArray `db:"grant_types" gorm:"type:text[]"`

	// ResponseTypes is an array of the OAuth 2.0 response type strings that the client can
	// use at the authorization endpoint.
	//
	// Pattern: id_token|code|token
	ResponseTypes gosql.NullableStringArray `db:"response_types" gorm:"type:text[]"`

	// Scope is a string containing a space-separated list of scope values (as
	// described in Section 3.3 of OAuth 2.0 [RFC6749]) that the client
	// can use when requesting access tokens.
	//
	// Pattern: ([a-zA-Z0-9\.\*]+\s?)+
	Scope string `db:"scope"`

	// Audience is a whitelist defining the audiences this client is allowed to request tokens for. An audience limits
	// the applicability of an OAuth 2.0 Access Token to, for example, certain API endpoints. The value is a list
	// of URLs. URLs MUST NOT contain whitespaces.
	Audience gosql.NullableStringArray `json:"audience" gorm:"type:text[]"`

	// SubjectType requested for responses to this Client. The subject_types_supported Discovery parameter contains a
	// list of the supported subject_type values for this server. Valid types include `pairwise` and `public`.
	SubjectType string `db:"subject_type"`

	// AllowedCORSOrigins are one or more URLs (scheme://host[:port]) which are allowed to make CORS requests
	// to the /oauth/token endpoint. If this array is empty, the sever's CORS origin configuration (`CORS_ALLOWED_ORIGINS`)
	// will be used instead. If this array is set, the allowed origins are appended to the server's CORS origin configuration.
	// Be aware that environment variable `CORS_ENABLED` MUST be set to `true` for this to work.
	AllowedCORSOrigins gosql.NullableStringArray `db:"allowed_cors_origins" gorm:"type:text[]"`

	// Public flag tells that the client is public
	Public bool `db:"public"`

	// ExpiresAt contins the time of expiration of the client
	ExpiresAt time.Time `db:"expires_at"`

	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at"`
}

// TableName in database
func (m *AuthClient) TableName() string {
	return `auth_client`
}

// OwnerAccountID returns the account ID which belongs the object
func (m *AuthClient) OwnerAccountID() uint64 {
	return m.AccountID
}

// RBACResourceName returns the name of the resource for the RBAC
func (m *AuthClient) RBACResourceName() string {
	return `auth_client`
}
