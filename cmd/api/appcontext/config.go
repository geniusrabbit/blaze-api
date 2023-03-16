// Package appcontext provides config options
package appcontext

/**
 ██████╗ ██████╗ ███╗   ██╗███████╗██╗ ██████╗
██╔════╝██╔═══██╗████╗  ██║██╔════╝██║██╔════╝
██║     ██║   ██║██╔██╗ ██║█████╗  ██║██║  ███╗
██║     ██║   ██║██║╚██╗██║██╔══╝  ██║██║   ██║
╚██████╗╚██████╔╝██║ ╚████║██║     ██║╚██████╔╝
 ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝     ╚═╝ ╚═════╝
*/

import (
	"encoding/json"
	"strings"
	"time"
)

type serverConfig struct {
	HTTP struct {
		Listen       string        `default:":8080" field:"listen" json:"listen" yaml:"listen" cli:"http-listen" env:"SERVER_HTTP_LISTEN"`
		ReadTimeout  time.Duration `default:"120s" field:"read_timeout" json:"read_timeout" yaml:"read_timeout" env:"SERVER_HTTP_READ_TIMEOUT"`
		WriteTimeout time.Duration `default:"120s" field:"write_timeout" json:"write_timeout" yaml:"write_timeout" env:"SERVER_HTTP_WRITE_TIMEOUT"`
	}
	Profile struct {
		Mode   string `json:"mode" yaml:"mode" default:"net" env:"SERVER_PROFILE_MODE"`
		Listen string `json:"listen" yaml:"listen" default:"" env:"SERVER_PROFILE_LISTEN"`
	} `json:"profile" yaml:"profile"`
}

type sessionConfig struct {
	CookieName string        `json:"cookie_name" yaml:"cookie_name" default:"sessid" env:"SESSION_COOKIE_NAME"`
	Lifetime   time.Duration `json:"lifetime" yaml:"lifetime" default:"1h" env:"SESSION_LIFETIME"`
	// DevToken is the permanent token which can be used to API access in develop mode
	DevToken     string `json:"dev_token" yaml:"dev_token" env:"SESSION_DEV_TOKEN"`
	DevUserID    uint64 `json:"dev_user_id" yaml:"dev_user_id" env:"SESSION_DEV_USER_ID"`
	DevAccountID uint64 `json:"dev_account_id" yaml:"dev_account_id" env:"SESSION_DEV_ACCOUNT_ID"`
}

type storageConfig struct {
	MasterConnect string `json:"master_connect" yaml:"master_connect" env:"SYSTEM_STORAGE_DATABASE_MASTER_CONNECT"`
	SlaveConnect  string `json:"slave_connect" yaml:"slave_connect" env:"SYSTEM_STORAGE_DATABASE_SLAVE_CONNECT"`
}

type oauth2Config struct {
	// Secret used by server to preprocess the secrets. Minimal size is 32 symbols
	Secret string `json:"secret" yaml:"secret" env:"OAUTH2_SECRET"`

	// AccessTokenLifespan sets how long an access token is going to be valid. Defaults to one hour.
	AccessTokenLifespan time.Duration `json:"access_token_lifespan" yaml:"access_token_lifespan" env:"OAUTH2_ACCESS_TOKEN_LIFESPAN" default:"1h"`

	// RefreshTokenLifespan sets how long a refresh token is going to be valid. Defaults to 30 days. Set to -1 for
	// refresh tokens that never expire.
	RefreshTokenLifespan time.Duration `json:"refresh_token_lifespan" yaml:"refresh_token_lifespan" env:"OAUTH2_REFRESH_TOKEN_LIFESPAN" default:"720h"`

	// AuthorizeCodeLifespan sets how long an authorize code is going to be valid. Defaults to fifteen minutes.
	AuthorizeCodeLifespan time.Duration `json:"authorize_code_lifespan" yaml:"authorize_code_lifespan" env:"OAUTH2_AUTHORIZE_CODE_LIFESPAN" default:"15m"`

	// HashCost sets the cost of the password hashing cost. Defaults to 12.
	HashCost int `json:"hash_cost" yaml:"hash_cost" env:"OAUTH2_HASH_COST"`

	// DisableRefreshTokenValidation sets the introspection endpoint to disable refresh token validation.
	DisableRefreshTokenValidation bool `json:"disable_refresh_token_validation" yaml:"disable_refresh_token_validation" env:"OAUTH2_DISABLE_REFRESH_TOKEN_VALIDATION"`

	// SendDebugMessagesToClients if set to true, includes error debug messages in response payloads. Be aware that sensitive
	// data may be exposed, depending on your implementation of Fosite. Such sensitive data might include database error
	// codes or other information. Proceed with caution!
	SendDebugMessagesToClients bool `json:"send_debug_messages_to_clients" yaml:"send_debug_messages_to_clients" env:"OAUTH2_SEND_DEBUG_MESSAGES_TO_CLIENTS"`

	// CacheConnect provides functionality of session cache to reduce amount of requests to the database
	// Supports: redis://host:port/dbNum, :memory:, :dummy:
	CacheConnect string `json:"cache_connect" yaml:"cache_connect" env:"OAUTH2_CACHE_CONNECT"`

	// CacheLifetime define the lifetime of elements in the cache
	CacheLifetime time.Duration `json:"cache_lifetime" yaml:"cache_lifetime" env:"OAUTH2_CACHE_LIFETIME"`
}

type permissionConfig struct {
	RoleCacheLifetime time.Duration `json:"role_cache_lifetime" yaml:"role_cache_lifetime" env:"PERMISSIONS_CACHE_LIFETIME" default:"10s"`
}

type systemConfig struct {
	Storage storageConfig `json:"storage" yaml:"storage"`
}

// ConfigType contains all application options
type ConfigType struct {
	ServiceName    string `json:"service_name" yaml:"service_name" env:"SERVICE_NAME" default:"websource.api"`
	DatacenterName string `json:"datacenter_name" yaml:"datacenter_name" env:"DC_NAME" default:"??"`
	Hostname       string `json:"hostname" yaml:"hostname" env:"HOSTNAME"`
	Hostcode       string `json:"hostcode" yaml:"hostcode" env:"HOSTCODE"`

	LogAddr    string `json:"log_addr" default:"" env:"LOG_ADDR"`
	LogLevel   string `json:"log_level" default:"debug" env:"LOG_LEVEL"`
	LogEncoder string `json:"log_encoder" env:"LOG_ENCODER"`

	Server      serverConfig     `json:"server" yaml:"server"`
	Session     sessionConfig    `json:"session" yaml:"session"`
	System      systemConfig     `json:"system" yaml:"system"`
	OAuth2      oauth2Config     `json:"oauth2" yaml:"oauth2"`
	Permissions permissionConfig `json:"permissions" yaml:"permissions"`
}

// String implementation of Stringer interface
func (cfg *ConfigType) String() (res string) {
	if data, err := json.MarshalIndent(cfg, "", "  "); err != nil {
		res = `{"error":"` + err.Error() + `"}`
	} else {
		res = string(data)
	}
	return res
}

// IsDebug mode
func (cfg *ConfigType) IsDebug() bool {
	return strings.EqualFold(cfg.LogLevel, "debug")
}

// Config global value
var Config ConfigType
