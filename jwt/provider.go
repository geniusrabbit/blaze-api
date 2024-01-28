package jwt

import (
	"errors"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/demdxx/gocast/v2"
	"github.com/form3tech-oss/jwt-go"
	// "github.com/golang-jwt/jwt/v4"
)

var (
	errJWTTokenIsExpired = errors.New(`JWT token is expired`)
	errJWTInvalidToken   = errors.New(`JWT invalid token`)
)

const (
	claimUserID    = "uid"
	claimAccountID = "acc"
	claimExpiredAt = "exp"
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
	UserID    uint64
	AccountID uint64
	ExpireAt  int64
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

// CreateToken new token for user ID
func (provider *Provider) CreateToken(userID, accountID uint64) (string, error) {
	var err error
	lifetime := provider.TokenLifetime
	if lifetime == 0 {
		lifetime = time.Hour
	}
	//Creating Access Token
	atClaims := jwt.MapClaims{
		claimUserID:    userID,
		claimAccountID: accountID,
		claimExpiredAt: time.Now().Add(lifetime).Unix(),
	}
	opt := provider.MiddlewareOptions()
	at := jwt.NewWithClaims(opt.SigningMethod, atClaims)
	token, err := at.SignedString([]byte(provider.Secret))
	if err != nil {
		return "", err
	}
	return token, nil
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
	uid := claims[claimUserID]
	acc := claims[claimAccountID]
	exp := claims[claimExpiredAt]

	data := &TokenData{
		UserID:    gocast.Number[uint64](uid),
		AccountID: gocast.Number[uint64](acc),
		ExpireAt:  gocast.Number[int64](exp),
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

	if gocast.Number[int64](exp) < time.Now().Unix() {
		return nil, errJWTTokenIsExpired
	}
	if gocast.Number[int64](uid) <= 0 {
		return nil, jwt.ErrInvalidKey
	}

	return []byte(provider.Secret), nil
}
