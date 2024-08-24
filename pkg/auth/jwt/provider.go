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

	// MapClaims describes the Claims type that uses the map[string]interface{} for JSON decoding
	// This is the default claims type if you don't supply one
	MapClaims = jwt.MapClaims
)

// TokenData extracted from token
type TokenData struct {
	UserID          uint64
	AccountID       uint64
	SocealAccountID uint64
	ExpireAt        int64
}

// Provider to JWT constructions
type Provider struct {
	// TokenLifetime defineds the valid time-period of token
	TokenLifetime time.Duration

	// Secret of session generation
	Secret string

	// MiddlewareOpts to get middelware procedure
	MiddlewareOpts *jwtmiddleware.Options
}

// NewDefaultProvider returns new provider
func NewDefaultProvider(secret string, tokenLifetime time.Duration, isDebug bool) *Provider {
	return &Provider{
		TokenLifetime: tokenLifetime,
		Secret:        secret,
		MiddlewareOpts: &jwtmiddleware.Options{
			Debug:     isDebug,
			Extractor: tokenextractor.DefaultExtractor,
		},
	}
}

// CreateToken new token for user ID
func (provider *Provider) CreateToken(userID, accountID, socialAccountID uint64) (string, time.Time, error) {
	var (
		err      error
		lifetime = gocast.IfThen(provider.TokenLifetime > time.Minute, provider.TokenLifetime, time.Hour)
		expireAt = time.Now().Add(lifetime)
	)
	//Creating Access Token
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
	opt := provider.MiddlewareOptions()
	at := jwt.NewWithClaims(opt.SigningMethod, atClaims)
	token, err := at.SignedString([]byte(provider.Secret))
	if err != nil {
		return "", expireAt, err
	}
	return token, expireAt, nil
}

// MiddlewareOptions returns the options of middelware
func (provider *Provider) MiddlewareOptions() *jwtmiddleware.Options {
	if provider.MiddlewareOpts == nil {
		provider.MiddlewareOpts = &jwtmiddleware.Options{}
	}
	if provider.MiddlewareOpts.ValidationKeyGetter == nil {
		provider.MiddlewareOpts.ValidationKeyGetter = provider.validationKeyGetter
	}
	if provider.MiddlewareOpts.SigningMethod == nil {
		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		provider.MiddlewareOpts.SigningMethod = jwt.SigningMethodHS256
	}
	if provider.MiddlewareOpts.ErrorHandler == nil {
		provider.MiddlewareOpts.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err string) {}
	}
	return provider.MiddlewareOpts
}

// Middleware returns middleware object with custom validation procedure
func (provider *Provider) Middleware() *jwtmiddleware.JWTMiddleware {
	return jwtmiddleware.New(*provider.MiddlewareOptions())
}

// ExtractTokenData into the token struct
func (provider *Provider) ExtractTokenData(token *Token) (*TokenData, error) {
	if token == nil || token.Claims == nil {
		return nil, errJWTInvalidToken
	}
	claims := token.Claims.(MapClaims)
	uid := gocast.Uint64(claims[claimUserID])
	acc := gocast.Uint64(claims[claimAccountID])
	sid := gocast.Uint64(claims[claimSocialAccountID])
	exp := gocast.Int64(claims[claimExpiredAt])

	data := &TokenData{
		UserID:          uid,
		AccountID:       acc,
		SocealAccountID: sid,
		ExpireAt:        exp,
	}

	if data.ExpireAt < time.Now().Unix() {
		return nil, errJWTTokenIsExpired
	}
	if data.UserID <= 0 {
		return nil, jwt.ErrInvalidKey
	}
	return data, nil
}

func (provider *Provider) validationKeyGetter(token *Token) (any, error) {
	if token.Claims == nil {
		return nil, jwt.ErrInvalidKey
	}
	claims := token.Claims.(MapClaims)
	uid := claims[claimUserID]
	// acc, _ := claims[claimAccountID]
	exp := claims[claimExpiredAt]

	if gocast.Int64(exp) < time.Now().Unix() {
		return nil, errJWTTokenIsExpired
	}
	if gocast.Int64(uid) <= 0 {
		return nil, jwt.ErrInvalidKey
	}

	return []byte(provider.Secret), nil
}
