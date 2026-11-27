// Package jwt issues and validates the HS256 session tokens of the API.
package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
)

// MinSecretLength is the minimal length of the signing secret (256 bits for HS256).
const MinSecretLength = 32

var (
	// ErrSecretTooShort is returned when the signing secret is shorter than MinSecretLength.
	ErrSecretTooShort = fmt.Errorf("jwt: secret must be at least %d bytes", MinSecretLength)
	// ErrInvalidToken is returned for malformed, forged, expired or foreign tokens.
	ErrInvalidToken = errors.New("jwt: invalid token")
	// ErrTokenRevoked is returned for revoked tokens.
	ErrTokenRevoked = errors.New("jwt: token is revoked")
)

const (
	claimUserID          = "uid"
	claimAccountID       = "acc"
	claimSocialAccountID = "sid"
)

// signingMethod is the only accepted algorithm; pinning it prevents algorithm
// confusion (e.g. "none" or RS/HS key mixing).
var signingMethod = jwt.SigningMethodHS256

type (
	// Token of JWT session
	Token = jwt.Token
	// MapClaims describes the Claims type that uses map[string]any for JSON decoding
	MapClaims = jwt.MapClaims
)

// TokenData contains extracted token information
type TokenData struct {
	ID              string // jti
	UserID          uint64
	AccountID       uint64
	SocialAccountID uint64
	IssuedAt        int64
	ExpireAt        int64
}

// Provider manages JWT token creation and validation
type Provider struct {
	TokenLifetime time.Duration // Valid time period for tokens (1h by default)
	Secret        string        // HMAC secret, at least MinSecretLength bytes

	// Issuer and Audience are written into tokens and required when set.
	Issuer   string
	Audience string

	// Extractor reads the token from the request. By default only the
	// `Authorization: Bearer` header is used: tokens in URLs leak to logs.
	Extractor func(r *http.Request) (string, error)

	// Revocation enables Revoke/RevokeUserTokens. Without it tokens are valid
	// until they expire.
	Revocation RevocationStore
}

// NewDefaultProvider creates a new JWT provider with default settings.
// The isDebug argument is ignored: tokens are never logged.
func NewDefaultProvider(secret string, tokenLifetime time.Duration, _ bool) *Provider {
	return &Provider{TokenLifetime: tokenLifetime, Secret: secret}
}

// Validate checks the provider configuration.
func (provider *Provider) Validate() error {
	if len(provider.Secret) < MinSecretLength {
		return ErrSecretTooShort
	}
	return nil
}

func (provider *Provider) lifetime() time.Duration {
	return gocast.IfThen(provider.TokenLifetime > time.Minute, provider.TokenLifetime, time.Hour)
}

// CreateToken generates a new signed JWT token for the given user
func (provider *Provider) CreateToken(userID, accountID, socialAccountID uint64) (string, time.Time, error) {
	now := time.Now()
	expireAt := now.Add(provider.lifetime())
	if err := provider.Validate(); err != nil {
		return "", expireAt, err
	}
	if userID == 0 {
		return "", expireAt, errors.New("jwt: user ID is required")
	}

	var jti [16]byte
	if _, err := rand.Read(jti[:]); err != nil {
		return "", expireAt, err
	}
	claims := jwt.MapClaims{
		"jti":       hex.EncodeToString(jti[:]),
		"iat":       now.Unix(),
		"nbf":       now.Unix(),
		"exp":       expireAt.Unix(),
		claimUserID: userID,
	}
	if provider.Issuer != "" {
		claims["iss"] = provider.Issuer
	}
	if provider.Audience != "" {
		claims["aud"] = provider.Audience
	}
	if accountID > 0 {
		claims[claimAccountID] = accountID
	}
	if socialAccountID > 0 {
		claims[claimSocialAccountID] = socialAccountID
	}

	token, err := jwt.NewWithClaims(signingMethod, claims).SignedString([]byte(provider.Secret))
	if err != nil {
		return "", expireAt, err
	}
	return token, expireAt, nil
}

// ExtractToken reads the raw token from the request.
func (provider *Provider) ExtractToken(r *http.Request) (string, error) {
	if provider.Extractor != nil {
		return provider.Extractor(r)
	}
	return tokenextractor.AuthorizationHeaderExtractor(r)
}

// ParseToken verifies the signature and claims of the token and checks that
// it isn't revoked. Errors of the revocation store are wrapped in
// auth.ErrAuthorizationUnavailable.
func (provider *Provider) ParseToken(ctx context.Context, raw string) (*TokenData, error) {
	data, err := provider.parse(raw)
	if err != nil {
		return nil, err
	}
	if provider.Revocation == nil {
		return data, nil
	}
	if data.ID != "" {
		revoked, err := provider.Revocation.IsTokenRevoked(ctx, data.ID)
		if err != nil {
			return nil, fmt.Errorf("%w: jwt revocation check: %w", auth.ErrAuthorizationUnavailable, err)
		}
		if revoked {
			return nil, ErrTokenRevoked
		}
	}
	before, err := provider.Revocation.UserTokensRevokedBefore(ctx, data.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: jwt revocation check: %w", auth.ErrAuthorizationUnavailable, err)
	}
	if !before.IsZero() && data.IssuedAt < before.Unix() {
		return nil, ErrTokenRevoked
	}
	return data, nil
}

func (provider *Provider) parse(raw string) (*TokenData, error) {
	if err := provider.Validate(); err != nil {
		return nil, err
	}
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{signingMethod.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(30 * time.Second),
	}
	if provider.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(provider.Issuer))
	}
	if provider.Audience != "" {
		opts = append(opts, jwt.WithAudience(provider.Audience))
	}
	token, err := jwt.ParseWithClaims(raw, jwt.MapClaims{}, func(*jwt.Token) (any, error) {
		return []byte(provider.Secret), nil
	}, opts...)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	return provider.ExtractTokenData(token)
}

// ExtractTokenData returns the data of a token parsed with this provider.
func (provider *Provider) ExtractTokenData(token *Token) (*TokenData, error) {
	if token == nil || token.Claims == nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	data := &TokenData{
		ID:              gocast.Str(claims["jti"]),
		UserID:          gocast.Uint64(claims[claimUserID]),
		AccountID:       gocast.Uint64(claims[claimAccountID]),
		SocialAccountID: gocast.Uint64(claims[claimSocialAccountID]),
		IssuedAt:        gocast.Int64(claims["iat"]),
		ExpireAt:        gocast.Int64(claims["exp"]),
	}
	if data.UserID == 0 || data.ExpireAt < time.Now().Unix() {
		return nil, ErrInvalidToken
	}
	return data, nil
}

// Revoke invalidates the token until it expires. It needs Revocation.
func (provider *Provider) Revoke(ctx context.Context, raw string) error {
	if provider.Revocation == nil {
		return ErrRevocationDisabled
	}
	data, err := provider.parse(raw)
	if err != nil {
		return err
	}
	if data.ID == "" {
		// Tokens issued before jti existed can only be revoked per user
		return provider.RevokeUserTokens(ctx, data.UserID)
	}
	return provider.Revocation.RevokeToken(ctx, data.ID, time.Unix(data.ExpireAt, 0))
}

// RevokeUserTokens invalidates all tokens of the user issued before now,
// e.g. after a password change. It needs Revocation.
func (provider *Provider) RevokeUserTokens(ctx context.Context, userID uint64) error {
	if provider.Revocation == nil {
		return ErrRevocationDisabled
	}
	now := time.Now()
	return provider.Revocation.RevokeUserTokens(ctx, userID, now, now.Add(provider.lifetime()))
}
