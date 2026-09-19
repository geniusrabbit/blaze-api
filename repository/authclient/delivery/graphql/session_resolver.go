package graphql

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/repository/authclient"
	authclientrepo "github.com/geniusrabbit/blaze-api/repository/authclient/repository"
	authclientusecase "github.com/geniusrabbit/blaze-api/repository/authclient/usecase"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// SessionQueryResolver implements GraphQL API methods for AuthSession.
type SessionQueryResolver struct {
	sessions authclient.SessionUsecase
}

// NewSessionQueryResolver returns a new session API resolver
func NewSessionQueryResolver(uc authclient.SessionUsecase) *SessionQueryResolver {
	return &SessionQueryResolver{sessions: uc}
}

// NewDefaultSessionQueryResolver returns a session API resolver with the default usecase
func NewDefaultSessionQueryResolver() *SessionQueryResolver {
	return NewSessionQueryResolver(
		authclientusecase.NewSessionUsecase(authclientrepo.NewSessionRepository()),
	)
}

// AuthSession is the resolver for the authSession field.
func (r *SessionQueryResolver) AuthSession(ctx context.Context, id uint64) (*gqlmodels.AuthSession, error) {
	sess, err := r.sessions.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return FromAuthSessionModel(sess), nil
}

// ListAuthSessions is the resolver for the listAuthSessions field.
func (r *SessionQueryResolver) ListAuthSessions(
	ctx context.Context,
	filter *gqlmodels.AuthSessionListFilter,
	order []*gqlmodels.AuthSessionListOrder,
	page *gqlmodels.Page,
) (*AuthSessionConnection, error) {
	return NewAuthSessionConnection(ctx, r.sessions, filter, order, page), nil
}

// DeleteAuthSession is the resolver for the deleteAuthSession field.
func (r *SessionQueryResolver) DeleteAuthSession(ctx context.Context, id uint64) (*gqlmodels.AuthSession, error) {
	sess, err := r.sessions.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := r.sessions.Delete(ctx, id, historylog.Message("GQL delete auth session")); err != nil {
		return nil, err
	}
	return FromAuthSessionModel(sess), nil
}
