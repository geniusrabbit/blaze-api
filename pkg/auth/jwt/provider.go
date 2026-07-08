package jwt

import (
	"errors"
	"net/http"
	"time"

	// "github.com/golang-jwt/jwt/v4"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/demdxx/gocast/v2"
	"github.com/form3tech-oss/jwt-go"

	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
)

var (
	errJWTTokenIsExpired = errors.New(`JWT token is expired`)
	errJWTInvalidToken   = errors.New(`JWT invalid token`)
)

const (
	claimUserID          = "uid"
	claimAccountID       = "acc"
	claimExpiredAt       = "exp"
	claimSocialAccountID = "sid"
)

type (
	// Middleware object type
	Middleware = jwtmiddleware.JWTMiddleware

	// Token of JWT session
	Token = jwt.Token

	// MapClaims describes the Claims type that uses map[string]interface{} for JSON decoding
	MapClaims = jwt.MapClaims
)

// TokenData contains extracted token information
type TokenData struct {
	UserID          uint64
	AccountID       uint64
	SocialAccountID uint64
	ExpireAt        int64
}

// Provider manages JWT token creation and validation
type Provider struct {
	TokenLifetime  time.Duration          // Valid time period for tokens
	Secret         string                 // Secret key for signing
	MiddlewareOpts *jwtmiddleware.Options // Middleware configuration
}

// NewDefaultProvider creates a new JWT provider with default settings
func NewDefaultProvider(secret string, tokenLifetime time.Duration, isDebug bool) *Provider {
	return &Provider{
		TokenLifetime: tokenLifetime,
		Secret:        secret,
		MiddlewareOpts: &jwtmiddleware.Options{
			Debug:               isDebug,
			CredentialsOptional: true,
			Extractor:           tokenextractor.DefaultExtractor,
		},
	}
}

// CreateToken generates a new signed JWT token for the given user
func (provider *Provider) CreateToken(userID, accountID, socialAccountID uint64) (string, time.Time, error) {
	lifetime := gocast.IfThen(provider.TokenLifetime > time.Minute, provider.TokenLifetime, time.Hour)
	expireAt := time.Now().Add(lifetime)

	// Build token claims
	atClaims := jwt.MapClaims{
		claimUserID:    userID,
		claimExpiredAt: expireAt.Unix(),
	}

	if accountID > 0 {
		atClaims[claimAccountID] = accountID
	}
	if socialAccountID > 0 {
		atClaims[claimSocialAccountID] = socialAccountID
	}

	// Sign and return token
	opt := provider.MiddlewareOptions()
	at := jwt.NewWithClaims(opt.SigningMethod, atClaims)
	token, err := at.SignedString([]byte(provider.Secret))
	if err != nil {
		return "", expireAt, err
	}

	return token, expireAt, nil
}

// MiddlewareOptions returns configured middleware options with defaults
func (provider *Provider) MiddlewareOptions() *jwtmiddleware.Options {
	if provider.MiddlewareOpts == nil {
		provider.MiddlewareOpts = &jwtmiddleware.Options{}
	}

	if provider.MiddlewareOpts.ValidationKeyGetter == nil {
		provider.MiddlewareOpts.ValidationKeyGetter = provider.validationKeyGetter
	}

	if provider.MiddlewareOpts.SigningMethod == nil {
		provider.MiddlewareOpts.SigningMethod = jwt.SigningMethodHS256
	}

	if provider.MiddlewareOpts.ErrorHandler == nil {
		provider.MiddlewareOpts.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err string) {}
	}

	return provider.MiddlewareOpts
}

// Middleware returns a configured JWT middleware handler
func (provider *Provider) Middleware() *jwtmiddleware.JWTMiddleware {
	return jwtmiddleware.New(*provider.MiddlewareOptions())
}

// ExtractTokenData extracts and validates claims from a token
func (provider *Provider) ExtractTokenData(token *Token) (*TokenData, error) {
	if token == nil || token.Claims == nil {
		return nil, errJWTInvalidToken
	}

	claims := token.Claims.(MapClaims)
	data := &TokenData{
		UserID:          gocast.Uint64(claims[claimUserID]),
		AccountID:       gocast.Uint64(claims[claimAccountID]),
		SocialAccountID: gocast.Uint64(claims[claimSocialAccountID]),
		ExpireAt:        gocast.Int64(claims[claimExpiredAt]),
	}

	// Validate expiration and user ID
	if data.ExpireAt < time.Now().Unix() {
		return nil, errJWTTokenIsExpired
	}
	if data.UserID <= 0 {
		return nil, jwt.ErrInvalidKey
	}

	return data, nil
}

// validationKeyGetter retrieves and validates the secret key for token verification
func (provider *Provider) validationKeyGetter(token *Token) (any, error) {
	if token.Claims == nil {
		return nil, jwt.ErrInvalidKey
	}

	claims := token.Claims.(MapClaims)
	exp := gocast.Int64(claims[claimExpiredAt])
	uid := gocast.Int64(claims[claimUserID])

	// Validate token claims
	if exp < time.Now().Unix() {
		return nil, errJWTTokenIsExpired
	}
	if uid <= 0 {
		return nil, jwt.ErrInvalidKey
	}

	return []byte(provider.Secret), nil
}
