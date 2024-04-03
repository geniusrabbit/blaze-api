package serverprovider

import (
	"context"
	"database/sql"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/gosql/v2"
	"github.com/guregu/null"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/cache"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
)

type cacher interface {
	Set(ctx context.Context, key string, value any, lifetime time.Duration) error
	Get(ctx context.Context, key string, target any) error
	Del(ctx context.Context, key string) error
}

type userAccessor interface {
	GetByPassword(ctx context.Context, email, password string) (*model.User, error)
}

// DatabaseStorage implements fosite.Storage interface to control Oauth2 and OpenID access
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/storage.go
type DatabaseStorage struct {
	db            *gorm.DB
	userAccessor  userAccessor
	cache         cacher
	cacheLifetime time.Duration
}

// NewDatabaseStorage object accesor
func NewDatabaseStorage(db *gorm.DB, userAccessor userAccessor, cache cacher, cacheLifetime time.Duration) *DatabaseStorage {
	return &DatabaseStorage{
		db:            db,
		userAccessor:  userAccessor,
		cache:         cache,
		cacheLifetime: cacheLifetime,
	}
}

// GetClient object from database
func (s *DatabaseStorage) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	ctxlogger.Get(ctx).Debug("GetClient", zap.String("client_id", id))
	var (
		clientObj model.AuthClient
		err       = s.fromCacheOrSelect(ctx, s.clientCacheKey(id), &clientObj, id)
	)
	if err == sql.ErrNoRows {
		return nil, fosite.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "get client object")
	}
	if clientObj.ExpiresAt.Before(time.Now()) {
		return nil, fosite.ErrInvalidClient.WithHint("get OAuth2 client")
	}
	SetContextTargetClient(ctx, &clientObj)
	client := &fosite.DefaultClient{
		ID:            id,
		Secret:        []byte(clientObj.Secret),
		RedirectURIs:  clientObj.RedirectURIs,
		GrantTypes:    clientObj.GrantTypes,
		ResponseTypes: clientObj.ResponseTypes,
		Scopes:        strings.Split(clientObj.Scope, " "),
		Audience:      clientObj.Audience,
		Public:        clientObj.Public,
	}
	return client, nil
}

// ClientAssertionJWTValid returns an error if the JTI is
// known or the DB check failed and nil if the JTI is not known.
func (s *DatabaseStorage) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	return nil
}

// SetClientAssertionJWT marks a JTI as known for the given
// expiry time. Before inserting the new JTI, it will clean
// up any existing JTIs that have expired as those tokens can
// not be replayed due to the expiry.
func (s *DatabaseStorage) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	return nil
}

// CreateAuthorizeCodeSession stores the authorization request for a given authorization code.
func (s *DatabaseStorage) CreateAuthorizeCodeSession(ctx context.Context, code string, request fosite.Requester) error {
	ctxlogger.Get(ctx).Debug("CreateAuthorizeCodeSession", zap.String("access_token", code))
	return s.newSession(ctx, code, request)
}

// GetAuthorizeCodeSession hydrates the session based on the given code and returns the authorization request.
// If the authorization code has been invalidated with `InvalidateAuthorizeCodeSession`, this
// method should return the ErrInvalidatedAuthorizeCode error.
//
// Make sure to also return the fosite.Requester value when returning the fosite.ErrInvalidatedAuthorizeCode error!
func (s *DatabaseStorage) GetAuthorizeCodeSession(ctx context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
	ctxlogger.Get(ctx).Debug("GetAuthorizeCodeSession", zap.String("access_token", code))
	sess, err := s.getAuthSession(ctx, code, `access_token`)
	if err != nil {
		return nil, err
	}
	return s.toRequest(ctx, sess)
}

// InvalidateAuthorizeCodeSession is called when an authorize code is being used. The state of the authorization
// code should be set to invalid and consecutive requests to GetAuthorizeCodeSession should return the
// ErrInvalidatedAuthorizeCode error.
func (s *DatabaseStorage) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	ctxlogger.Get(ctx).Debug("InvalidateAuthorizeCodeSession", zap.String("access_token", code))
	return s.invalidateSession(ctx, code, `access_token`)
}

// DeleteAuthorizeCodeSession same as InvalidateAuthorizeCodeSession
func (s *DatabaseStorage) DeleteAuthorizeCodeSession(ctx context.Context, code string) error {
	ctxlogger.Get(ctx).Debug("DeleteAuthorizeCodeSession", zap.String("access_token", code))
	return s.invalidateSession(ctx, code, `access_token`)
}

// CreatePKCERequestSession action
func (s *DatabaseStorage) CreatePKCERequestSession(ctx context.Context, code string, request fosite.Requester) error {
	ctxlogger.Get(ctx).Debug("CreatePKCERequestSession", zap.String("access_token", code))
	return s.newSession(ctx, code, request)
}

// GetPKCERequestSession action
func (s *DatabaseStorage) GetPKCERequestSession(ctx context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
	ctxlogger.Get(ctx).Debug("GetPKCERequestSession", zap.String("access_token", code))
	sess, err := s.getAuthSession(ctx, code, `access_token`)
	if err != nil {
		return nil, err
	}
	return s.toRequest(ctx, sess)
}

// DeletePKCERequestSession action
func (s *DatabaseStorage) DeletePKCERequestSession(ctx context.Context, code string) error {
	ctxlogger.Get(ctx).Debug("DeletePKCERequestSession", zap.String("access_token", code))
	return s.invalidateSession(ctx, code, `access_token`)
}

// CreateAccessTokenSession updates session values
func (s *DatabaseStorage) CreateAccessTokenSession(ctx context.Context, signature string, request fosite.Requester) error {
	ctxlogger.Get(ctx).Debug("CreateAccessTokenSession", zap.String("access_token", signature))
	return s.newSession(ctx, signature, request)
}

// GetAccessTokenSession returns request by access token
func (s *DatabaseStorage) GetAccessTokenSession(ctx context.Context, signature string, _ fosite.Session) (fosite.Requester, error) {
	ctxlogger.Get(ctx).Debug("GetAccessTokenSession", zap.String("access_token", signature))
	if signature == `` {
		return nil, fosite.ErrTokenSignatureMismatch
	}
	sess, err := s.getAuthSession(ctx, signature, `access_token`)
	if err != nil {
		return nil, err
	}
	if sess.AccessTokenExpiresAt.Before(time.Now()) {
		return nil, fosite.ErrTokenExpired
	}
	return s.toRequest(ctx, sess)
}

// DeleteAccessTokenSession from DB
func (s *DatabaseStorage) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	ctxlogger.Get(ctx).Debug("DeleteAccessTokenSession", zap.String("access_token", signature))
	return s.dropSession(ctx, signature, `access_token`)
}

// CreateRefreshTokenSession updates session values
func (s *DatabaseStorage) CreateRefreshTokenSession(ctx context.Context, signature string, request fosite.Requester) error {
	ctxlogger.Get(ctx).Debug("CreateRefreshTokenSession", zap.String("refresh_token", signature))
	err := s.updateSessionCB(request.GetID(), func(auth *model.AuthSession) {
		auth.RefreshToken = null.StringFrom(signature)
	})
	return err
}

// GetRefreshTokenSession returns session by refresh token
func (s *DatabaseStorage) GetRefreshTokenSession(ctx context.Context, signature string, _ fosite.Session) (fosite.Requester, error) {
	ctxlogger.Get(ctx).Debug("GetRefreshTokenSession", zap.String("refresh_token", signature))
	if signature == `` {
		return nil, fosite.ErrTokenSignatureMismatch
	}
	sess, err := s.getAuthSession(ctx, signature, `refresh_token`)
	if err != nil {
		return nil, err
	}
	if sess.RefreshTokenExpiresAt.After(time.Now()) {
		return nil, fosite.ErrTokenExpired
	}
	return s.toRequest(ctx, sess)
}

// DeleteRefreshTokenSession from database
func (s *DatabaseStorage) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	ctxlogger.Get(ctx).Debug("DeleteRefreshTokenSession", zap.String("refresh_token", signature))
	return s.dropSession(ctx, signature, `refresh_token`)
}

// CreateImplicitAccessTokenSession invalid method
func (s *DatabaseStorage) CreateImplicitAccessTokenSession(ctx context.Context, code string, req fosite.Requester) error {
	ctxlogger.Get(ctx).Debug("CreateImplicitAccessTokenSession", zap.String("access_token", code))
	panic("CreateImplicitAccessTokenSession is not emplimented")
}

// RevokeRefreshToken revokes a refresh token as specified in:
// https://tools.ietf.org/html/rfc7009#section-2.1
// If the particular
// token is a refresh token and the authorization server supports the
// revocation of access tokens, then the authorization server SHOULD
// also invalidate all access tokens based on the same authorization
// grant (see Implementation Note).
func (s *DatabaseStorage) RevokeRefreshToken(ctx context.Context, requestID string) error {
	ctxlogger.Get(ctx).Debug("RevokeRefreshToken", zap.String("request_id", requestID))
	return s.updateSessionCB(requestID, func(auth *model.AuthSession) {
		auth.AccessTokenExpiresAt = time.Now()
		auth.RefreshTokenExpiresAt = time.Now()
	})
}

// RevokeRefreshTokenMaybeGracePeriod revokes a refresh token as specified in:
// https://tools.ietf.org/html/rfc7009#section-2.1
// If the particular
// token is a refresh token and the authorization server supports the
// revocation of access tokens, then the authorization server SHOULD
// also invalidate all access tokens based on the same authorization
// grant (see Implementation Note).
//
// If the Refresh Token grace period is greater than zero in configuration the token
// will have its expiration time set as UTCNow + GracePeriod.
func (s *DatabaseStorage) RevokeRefreshTokenMaybeGracePeriod(ctx context.Context, requestID string, signature string) error {
	ctxlogger.Get(ctx).Debug("RevokeRefreshTokenMaybeGracePeriod",
		zap.String(`request_id`, requestID),
		zap.String(`signature`, signature))
	return s.RevokeRefreshToken(ctx, requestID)
}

// RevokeAccessToken revokes an access token as specified in:
// https://tools.ietf.org/html/rfc7009#section-2.1
// If the token passed to the request
// is an access token, the server MAY revoke the respective refresh
// token as well.
func (s *DatabaseStorage) RevokeAccessToken(ctx context.Context, requestID string) error {
	ctxlogger.Get(ctx).Debug("RevokeAccessToken", zap.String("request_id", requestID))
	return s.updateSessionCB(requestID, func(auth *model.AuthSession) {
		auth.AccessTokenExpiresAt = time.Now()
	})
}

// Authenticate user by login and secret (:password)
func (s *DatabaseStorage) Authenticate(ctx context.Context, email string, secret string) error {
	ctxlogger.Get(ctx).Debug("Authenticate")
	user, err := s.userAccessor.GetByPassword(ctx, email, secret)
	if err == sql.ErrNoRows {
		return fosite.ErrNotFound
	}
	if user == nil {
		return nil
	}
	SetContextTargetUserID(ctx, user.ID)
	return errors.New("Invalid credentials")
}

///////////////////////////////////////////////////////////////////////////////
/// Session methods
///////////////////////////////////////////////////////////////////////////////

func (s *DatabaseStorage) newSession(ctx context.Context, token string, request fosite.Requester) error {
	var (
		userID    = GetContextTargetUserID(ctx)
		clientObj = GetContextTargetClient(ctx)
		session   = request.GetSession()
		client    = request.GetClient()
	)
	if userID > 0 && clientObj.UserID != userID {
		return fosite.ErrInvalidClient.WithHint("Check session user")
	}
	var (
		sessionObj = &model.AuthSession{
			Active:                true,
			ClientID:              client.GetID(),
			Username:              session.GetUsername(),
			Subject:               session.GetSubject(),
			RequestID:             request.GetID(),
			AccessToken:           token,
			Form:                  request.GetRequestForm().Encode(),
			RequestedScope:        gosql.NullableStringArray(request.GetRequestedScopes()),
			GrantedScope:          gosql.NullableStringArray(request.GetGrantedScopes()),
			RequestedAudience:     gosql.NullableStringArray(request.GetRequestedAudience()),
			GrantedAudience:       gosql.NullableStringArray(request.GetGrantedAudience()),
			AccessTokenExpiresAt:  session.GetExpiresAt(fosite.AccessToken),
			RefreshTokenExpiresAt: session.GetExpiresAt(fosite.RefreshToken),
		}
		err = s.db.Create(sessionObj).Error
	)
	if err != nil {
		return errors.Wrap(err, "create session")
	}
	if err = s.cache.Set(ctx, s.sessCacheKey(token), sessionObj, s.cacheLifetime); err != nil {
		ctxlogger.Get(ctx).Error("put session to cache", zap.Error(err))
	}
	return nil
}

func (s *DatabaseStorage) updateSessionCB(requestID string, cb func(auth *model.AuthSession)) error {
	var (
		sessionObj model.AuthSession
		err        = s.db.Find(&sessionObj, `request_id=?`, requestID).Error
	)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fosite.ErrNotFound
	}
	if err != nil {
		return err
	}
	cb(&sessionObj)
	return s.db.Model(&sessionObj).Where(`request_id=?`, requestID).Updates(&sessionObj).Error
}

func (s *DatabaseStorage) getAuthSession(ctx context.Context, code, fieldName string) (*model.AuthSession, error) {
	var (
		sessionObj model.AuthSession
		err        = s.fromCacheOrSelect(ctx, s.sessCacheKey(code), &sessionObj, fieldName+`=?`, code)
	)
	if err != nil {
		return nil, err
	}
	return &sessionObj, nil
}

// func (s *DatabaseStorage) getSession(ctx context.Context, code, fieldName string) (_ fosite.Session, err error) {
// 	sessionObj, err := s.getAuthSession(code, fieldName)
// 	if err == sql.ErrNoRows {
// 		return nil, fosite.ErrNotFound
// 	}
// 	return s.authToSession(ctx, sessionObj)
// }

func (s *DatabaseStorage) getSessionCode(ctx context.Context, code, fieldName string) (string, error) {
	cacheCode := code
	if fieldName != `access_token` {
		sess, err := s.getAuthSession(ctx, code, fieldName)
		if err != nil {
			return "", err
		}
		cacheCode = sess.AccessToken
	}
	return s.sessCacheKey(cacheCode), nil
}

func (s *DatabaseStorage) dropSession(ctx context.Context, code, fieldName string) error {
	err := s.db.Model((*model.AuthSession)(nil)).Delete(fieldName+`=?`, code).Error
	if err == sql.ErrNoRows {
		return fosite.ErrNotFound
	}
	if err != nil {
		return errors.Wrap(err, `DeleteOpenIDConnectSession`)
	}
	// Drop cache data
	cacheKey, err := s.getSessionCode(ctx, code, fieldName)
	if err != nil {
		return err
	}
	if cacheErr := s.cache.Del(ctx, cacheKey); cacheErr != nil {
		ctxlogger.Get(ctx).Error(`clear session cache`,
			zap.String(`cache_key`, cacheKey), zap.Error(cacheErr))
	}
	return nil
}

func (s *DatabaseStorage) invalidateSession(ctx context.Context, code, fieldName string) error {
	cacheKey, err := s.getSessionCode(ctx, code, fieldName)
	if err != nil {
		return err
	}
	err = s.cache.Del(ctx, cacheKey)
	if err != nil {
		ctxlogger.Get(ctx).Error("incalidate session cache",
			zap.String(`cache_key`, cacheKey), zap.Error(err))
	}
	err = s.db.Model((*model.AuthSession)(nil)).Where(fieldName+`=? AND active=TRUE`, code).Update(`active`, true).Error
	if err == sql.ErrNoRows {
		return fosite.ErrNotFound
	}
	return err
}

func (s *DatabaseStorage) authToSession(ctx context.Context, sessionObj *model.AuthSession) (_ fosite.Session, err error) {
	if !sessionObj.Active {
		err = fosite.ErrInvalidatedAuthorizeCode
	}
	return NewSession(
		ctx,
		sessionObj.Username,
		sessionObj.Subject,
		sessionObj.AccessToken,
		sessionObj.AccessTokenExpiresAt,
		sessionObj.RefreshToken.String,
		sessionObj.RefreshTokenExpiresAt,
	), err
}

func (s *DatabaseStorage) toRequest(ctx context.Context, sessionObj *model.AuthSession) (*fosite.Request, error) {
	session, authErr := s.authToSession(ctx, sessionObj)
	if session == nil {
		return nil, errors.Wrap(authErr, "auth session")
	}

	client, err := s.GetClient(ctx, sessionObj.ClientID)
	if err != nil {
		return nil, errors.Wrap(err, "get client")
	}

	val, err := url.ParseQuery(sessionObj.Form)
	if err != nil {
		return nil, errors.Wrap(err, "parse form data")
	}

	SetContextSession(ctx, sessionObj)
	r := &fosite.Request{
		ID:                sessionObj.RequestID,
		RequestedAt:       sessionObj.CreatedAt,
		Client:            client,
		Session:           session,
		Form:              val,
		RequestedScope:    fosite.Arguments(sessionObj.RequestedScope),
		GrantedScope:      fosite.Arguments(sessionObj.GrantedScope),
		RequestedAudience: fosite.Arguments(sessionObj.RequestedAudience),
		GrantedAudience:   fosite.Arguments(sessionObj.GrantedAudience),
	}
	return r, authErr
}

func (s *DatabaseStorage) sessCacheKey(key string) string {
	return "sess:" + key
}

func (s *DatabaseStorage) clientCacheKey(id string) string {
	return "cli:" + id
}

func (s *DatabaseStorage) fromCacheOrSelect(ctx context.Context, cacheKey string, target any, args ...any) error {
	if err := s.cache.Get(ctx, cacheKey, target); err != nil && err != cache.ErrEntryNotFound {
		ctxlogger.Get(ctx).Error("cache load", zap.String("cache_key", cacheKey), zap.Error(err))
	}
	err := s.db.Find(target, args...).Error
	if err == nil {
		err = s.cache.Set(ctx, cacheKey, target, s.cacheLifetime)
		if err != nil {
			ctxlogger.Get(ctx).Error("set cache", zap.String("cache_key", cacheKey), zap.Error(err))
		}
	}
	return err
}

var _ oauth2.TokenRevocationStorage = (*DatabaseStorage)(nil)
