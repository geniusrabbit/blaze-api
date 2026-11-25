package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
)

var (
	// ErrTokenForAnotherUser is returned when a token is generated for another user without directaccesstoken.create.all.
	ErrTokenForAnotherUser = errors.New("generating an access token for another user requires directaccesstoken.create.all")
	// ErrTokenWithoutUser is returned when a token is generated outside a user session.
	ErrTokenWithoutUser = errors.New("an access token must belong to a user")
)

type tokenScope int

const (
	scopeOwner tokenScope = iota
	scopeAccount
	scopeAll
)

// Usecase handles business logic for direct access tokens.
type Usecase struct {
	repo directaccesstoken.Repository
}

// New creates a new direct access token usecase instance.
func New(repo directaccesstoken.Repository) *Usecase {
	return &Usecase{repo: repo}
}

// Get retrieves a direct access token by ID with permission checks.
// Without the account or all cover only the session user's tokens are visible.
func (u *Usecase) Get(ctx context.Context, id uint64) (*models.DirectAccessToken, error) {
	if err := session.DenyDirectTokenSession(ctx); err != nil {
		return nil, err
	}
	accToken, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, accToken) || !tokenInScope(ctx, accToken, scopeFor(ctx, acl.PermView)) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "get access token")
	}
	return accToken, nil
}

// FetchList retrieves a filtered and paginated list of direct access tokens.
func (u *Usecase) FetchList(ctx context.Context, opts ...directaccesstoken.QOption) ([]*models.DirectAccessToken, error) {
	if err := session.DenyDirectTokenSession(ctx); err != nil {
		return nil, err
	}
	if !acl.HaveAccessList(ctx, &models.DirectAccessToken{AccountID: session.AccountID(ctx)}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list access tokens")
	}
	return u.repo.FetchList(ctx, scopedOptions(ctx, opts, scopeFor(ctx, acl.PermList))...)
}

// Count returns the total count of direct access tokens matching the filter.
func (u *Usecase) Count(ctx context.Context, opts ...directaccesstoken.QOption) (int64, error) {
	if err := session.DenyDirectTokenSession(ctx); err != nil {
		return 0, err
	}
	if !acl.HaveAccessCount(ctx, &models.DirectAccessToken{AccountID: session.AccountID(ctx)}) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "count access tokens")
	}
	return u.repo.Count(ctx, scopedOptions(ctx, opts, scopeFor(ctx, acl.PermCount))...)
}

// Generate creates a new direct access token.
// The token belongs to the session user unless userID names another user,
// which requires directaccesstoken.create.all.
func (u *Usecase) Generate(ctx context.Context, userID, accountID uint64, description string, expiresAt time.Time) (*models.DirectAccessToken, error) {
	if err := session.DenyDirectTokenSession(ctx); err != nil {
		return nil, err
	}
	sessionUserID := session.UserID(ctx)
	if sessionUserID == 0 {
		return nil, ErrTokenWithoutUser
	}
	if userID == 0 {
		userID = sessionUserID
	}
	if userID != sessionUserID && scopeFor(ctx, acl.PermCreate) != scopeAll {
		return nil, ErrTokenForAnotherUser
	}
	if !acl.HaveAccessCreate(ctx, &models.DirectAccessToken{
		UserID:    sql.Null[uint64]{V: userID, Valid: true},
		AccountID: accountID,
	}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "generate access token")
	}
	return u.repo.Generate(ctx, userID, accountID, description, expiresAt)
}

// Revoke revokes direct access tokens matching the filter criteria.
// Without directaccesstoken.delete.account or .all only the session user's tokens are revoked.
func (u *Usecase) Revoke(ctx context.Context, opts ...directaccesstoken.QOption) error {
	if err := session.DenyDirectTokenSession(ctx); err != nil {
		return err
	}
	accountToken := &models.DirectAccessToken{AccountID: session.AccountID(ctx)}
	if !acl.HaveAccessDelete(ctx, accountToken) && !acl.HaveAccessList(ctx, accountToken) {
		return errors.Wrap(acl.ErrNoPermissions, "revoke access tokens")
	}
	return u.repo.Revoke(ctx, scopedOptions(ctx, opts, scopeFor(ctx, acl.PermDelete))...)
}

func scopeFor(ctx context.Context, action string) tokenScope {
	resource := &models.DirectAccessToken{}
	switch {
	case acl.HaveObjectPermissions(ctx, resource, action+".all"):
		return scopeAll
	case acl.HaveObjectPermissions(ctx, resource, action+".account"):
		return scopeAccount
	}
	return scopeOwner
}

func scopedOptions(ctx context.Context, opts []directaccesstoken.QOption, scope tokenScope) []directaccesstoken.QOption {
	switch scope {
	case scopeAll:
		return opts
	case scopeAccount:
		return append(opts, &directaccesstoken.Filter{AccountID: []uint64{session.AccountID(ctx)}})
	}
	return append(opts, &directaccesstoken.Filter{
		AccountID: []uint64{session.AccountID(ctx)},
		UserID:    []uint64{session.UserID(ctx)},
	})
}

func tokenInScope(ctx context.Context, token *models.DirectAccessToken, scope tokenScope) bool {
	switch scope {
	case scopeAll:
		return true
	case scopeAccount:
		return token.AccountID == session.AccountID(ctx)
	}
	return token.AccountID == session.AccountID(ctx) && token.UserID.Valid && token.UserID.V == session.UserID(ctx)
}
