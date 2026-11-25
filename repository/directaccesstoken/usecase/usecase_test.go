package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

const (
	sessionUserID    = 7
	sessionAccountID = 9
	otherUserID      = 8
)

type grantedPermissions map[string]bool

func (granted grantedPermissions) matches(name string) bool {
	for grant := range granted {
		if grant == name {
			return true
		}
		if prefix, ok := strings.CutSuffix(name, "*"); ok && strings.HasPrefix(grant, prefix) {
			return true
		}
	}
	return false
}

func (granted grantedPermissions) CheckPermissions(_ context.Context, resource any, patterns ...string) bool {
	prefix := ""
	if named, ok := resource.(interface{ RBACResourceName() string }); ok {
		prefix = named.RBACResourceName() + "."
	}
	for _, pattern := range patterns {
		if !granted.matches(prefix + pattern) {
			return false
		}
	}
	return true
}
func (grantedPermissions) CheckedPermissions(context.Context, any, ...string) rbac.Permission {
	return nil
}
func (grantedPermissions) ChildRoles() []rbac.Role                 { return nil }
func (grantedPermissions) ChildPermissions() []rbac.Permission     { return nil }
func (grantedPermissions) Permissions(...string) []rbac.Permission { return nil }
func (grantedPermissions) HasPermission(...string) bool            { return true }

type testAccount struct {
	accountModels.AccountBase
}

func (a *testAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	accountObject := &testAccount{}
	accountObject.SetID(id)
	return accountObject
}

func sessionContext(authorizerCode string, permissions ...string) context.Context {
	accountObject := &testAccount{}
	accountObject.SetID(sessionAccountID)
	granted := grantedPermissions{}
	for _, permission := range permissions {
		granted[permission] = true
	}
	accountObject.SetPermissions(granted)
	ctx := session.WithUserAccount(context.Background(), testutil.Stub(sessionUserID), accountObject)
	return session.WithAuthorizerCode(ctx, authorizerCode)
}

type recordingRepository struct {
	tokens         map[uint64]*models.DirectAccessToken
	generatedUser  *uint64
	listOptions    []directaccesstoken.QOption
	revokeOptions  []directaccesstoken.QOption
	revokeRequests int
}

func (repository *recordingRepository) Get(_ context.Context, id uint64) (*models.DirectAccessToken, error) {
	if token, ok := repository.tokens[id]; ok {
		return token, nil
	}
	return nil, errors.New("not found")
}
func (*recordingRepository) GetByToken(context.Context, string) (*models.DirectAccessToken, error) {
	return nil, nil
}
func (repository *recordingRepository) FetchList(_ context.Context, opts ...directaccesstoken.QOption) ([]*models.DirectAccessToken, error) {
	repository.listOptions = opts
	return nil, nil
}
func (repository *recordingRepository) Count(_ context.Context, opts ...directaccesstoken.QOption) (int64, error) {
	repository.listOptions = opts
	return 0, nil
}
func (repository *recordingRepository) Generate(_ context.Context, userID, accountID uint64, _ string, expiresAt time.Time) (*models.DirectAccessToken, error) {
	repository.generatedUser = &userID
	return &models.DirectAccessToken{ID: 1, UserID: sql.Null[uint64]{V: userID, Valid: userID > 0}, AccountID: accountID, ExpiresAt: expiresAt}, nil
}
func (repository *recordingRepository) Revoke(_ context.Context, opts ...directaccesstoken.QOption) error {
	repository.revokeOptions = opts
	repository.revokeRequests++
	return nil
}

func scopeFilters(opts []directaccesstoken.QOption) []*directaccesstoken.Filter {
	filters := []*directaccesstoken.Filter{}
	for _, opt := range opts {
		if filter, ok := opt.(*directaccesstoken.Filter); ok {
			filters = append(filters, filter)
		}
	}
	return filters
}

func hasScope(opts []directaccesstoken.QOption, accountID, userID uint64) bool {
	for _, filter := range scopeFilters(opts) {
		accountMatches := len(filter.AccountID) == 1 && filter.AccountID[0] == accountID
		userMatches := userID == 0 && len(filter.UserID) == 0 || len(filter.UserID) == 1 && filter.UserID[0] == userID
		if accountMatches && userMatches {
			return true
		}
	}
	return false
}

func expiry() time.Time { return time.Now().Add(time.Hour) }

func TestDirectTokenSessionCannotUseTokenUsecase(t *testing.T) {
	t.Parallel()

	ctx := sessionContext(session.AuthorizerCodeDirectToken, "directaccesstoken.view.all",
		"directaccesstoken.list.all", "directaccesstoken.count.all", "directaccesstoken.create.all",
		"directaccesstoken.delete.all")
	repository := &recordingRepository{tokens: map[uint64]*models.DirectAccessToken{
		1: {ID: 1, AccountID: sessionAccountID, UserID: sql.Null[uint64]{V: sessionUserID, Valid: true}},
	}}
	tokenUsecase := New(repository)

	operations := map[string]func() error{
		"get":      func() error { _, err := tokenUsecase.Get(ctx, 1); return err },
		"list":     func() error { _, err := tokenUsecase.FetchList(ctx); return err },
		"count":    func() error { _, err := tokenUsecase.Count(ctx); return err },
		"generate": func() error { _, err := tokenUsecase.Generate(ctx, 0, sessionAccountID, "", expiry()); return err },
		"revoke":   func() error { return tokenUsecase.Revoke(ctx) },
	}
	for name, operation := range operations {
		if err := operation(); !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
			t.Fatalf("%s: err = %v, want %v", name, err, session.ErrDirectTokenSessionForbidden)
		}
	}
	if repository.generatedUser != nil || repository.revokeRequests != 0 || repository.listOptions != nil {
		t.Fatal("repository was called for a direct token session")
	}
}

func TestGenerateDefaultsToSessionUser(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{}
	_, err := New(repository).Generate(sessionContext("jwt", "directaccesstoken.create.owner"), 0, sessionAccountID, "", expiry())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if repository.generatedUser == nil || *repository.generatedUser != sessionUserID {
		t.Fatalf("generated user = %v, want %d", repository.generatedUser, sessionUserID)
	}
}

func TestGenerateForAnotherUserRequiresCreateAll(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{}
	_, err := New(repository).Generate(sessionContext("jwt", "directaccesstoken.create.owner", "directaccesstoken.create.account"), otherUserID, sessionAccountID, "", expiry())
	if !errors.Is(err, ErrTokenForAnotherUser) {
		t.Fatalf("err = %v, want %v", err, ErrTokenForAnotherUser)
	}
	if repository.generatedUser != nil {
		t.Fatal("token generated for another user")
	}

	_, err = New(repository).Generate(sessionContext("jwt", "directaccesstoken.create.all"), otherUserID, sessionAccountID, "", expiry())
	if err != nil {
		t.Fatalf("generate with create.all: %v", err)
	}
	if repository.generatedUser == nil || *repository.generatedUser != otherUserID {
		t.Fatalf("generated user = %v, want %d", repository.generatedUser, otherUserID)
	}
}

func TestGenerateRequiresSessionUser(t *testing.T) {
	t.Parallel()

	accountObject := &testAccount{}
	accountObject.SetID(sessionAccountID)
	accountObject.SetPermissions(grantedPermissions{"directaccesstoken.create.all": true})
	ctx := session.WithUserAccount(context.Background(), (*testutil.User)(nil), accountObject)

	repository := &recordingRepository{}
	if _, err := New(repository).Generate(ctx, 0, sessionAccountID, "", expiry()); err == nil {
		t.Fatal("generated a token without a user")
	}
	if repository.generatedUser != nil {
		t.Fatal("repository was called")
	}
}

func TestListScopeFollowsPermissionCover(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		permission    string
		scopedAccount uint64
		scopedUser    uint64
	}{
		"owner":   {permission: "directaccesstoken.list.owner", scopedAccount: sessionAccountID, scopedUser: sessionUserID},
		"account": {permission: "directaccesstoken.list.account", scopedAccount: sessionAccountID},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repository := &recordingRepository{}
			callerFilter := &directaccesstoken.Filter{AccountID: []uint64{99}}
			if _, err := New(repository).FetchList(sessionContext("jwt", testCase.permission), callerFilter); err != nil {
				t.Fatalf("list: %v", err)
			}
			if !hasScope(repository.listOptions, testCase.scopedAccount, testCase.scopedUser) {
				t.Fatalf("list is not scoped to account %d user %d", testCase.scopedAccount, testCase.scopedUser)
			}
		})
	}

	repository := &recordingRepository{}
	if _, err := New(repository).FetchList(sessionContext("jwt", "directaccesstoken.list.all")); err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(scopeFilters(repository.listOptions)) != 0 {
		t.Fatal("list.all must not be scoped")
	}

	if _, err := New(&recordingRepository{}).FetchList(sessionContext("jwt")); err == nil {
		t.Fatal("list without permission succeeded")
	}
}

func TestRevokeIsScopedToOwnTokensWithoutDeleteAll(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{}
	if err := New(repository).Revoke(sessionContext("jwt", "directaccesstoken.list.owner"), &directaccesstoken.Filter{ID: []uint64{5}}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !hasScope(repository.revokeOptions, sessionAccountID, sessionUserID) {
		t.Fatal("revoke is not scoped to the session user")
	}

	repository = &recordingRepository{}
	if err := New(repository).Revoke(sessionContext("jwt", "directaccesstoken.delete.account"), &directaccesstoken.Filter{ID: []uint64{5}}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !hasScope(repository.revokeOptions, sessionAccountID, 0) {
		t.Fatal("revoke with delete.account is not scoped to the account")
	}

	repository = &recordingRepository{}
	if err := New(repository).Revoke(sessionContext("jwt", "directaccesstoken.delete.all"), &directaccesstoken.Filter{ID: []uint64{5}}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if len(scopeFilters(repository.revokeOptions)) != 1 {
		t.Fatal("revoke with delete.all must keep only the caller filter")
	}
}

func TestGetOtherUsersTokenRequiresWiderCover(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{tokens: map[uint64]*models.DirectAccessToken{
		1: {ID: 1, AccountID: sessionAccountID, UserID: sql.Null[uint64]{V: sessionUserID, Valid: true}},
		2: {ID: 2, AccountID: sessionAccountID, UserID: sql.Null[uint64]{V: otherUserID, Valid: true}},
		3: {ID: 3, AccountID: 99, UserID: sql.Null[uint64]{V: otherUserID, Valid: true}},
	}}
	tokenUsecase := New(repository)

	ownerContext := sessionContext("jwt", "directaccesstoken.view.owner")
	if _, err := tokenUsecase.Get(ownerContext, 1); err != nil {
		t.Fatalf("own token: %v", err)
	}
	if _, err := tokenUsecase.Get(ownerContext, 2); err == nil {
		t.Fatal("owner cover read another user's token")
	}

	accountContext := sessionContext("jwt", "directaccesstoken.view.account")
	if _, err := tokenUsecase.Get(accountContext, 2); err != nil {
		t.Fatalf("account cover: %v", err)
	}
	if _, err := tokenUsecase.Get(accountContext, 3); err == nil {
		t.Fatal("account cover read another account's token")
	}

	if _, err := tokenUsecase.Get(sessionContext("jwt", "directaccesstoken.view.all"), 3); err != nil {
		t.Fatalf("all cover: %v", err)
	}
}
