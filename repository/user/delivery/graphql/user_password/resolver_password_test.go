package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

func TestChangeUserPasswordDeniedForDirectTokenSession(t *testing.T) {
	t.Parallel()

	resolver := &QueryResolverPassword[*testutil.User, any, any, any, any, any]{}
	ctx := session.WithAuthorizerCode(context.Background(), session.AuthorizerCodeDirectToken)

	_, err := resolver.ChangeUserPassword(ctx, "current", "new-password")
	if !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
		t.Fatalf("err = %v, want %v", err, session.ErrDirectTokenSessionForbidden)
	}
}

// fakePasswordRepo records the password passed to CreateWithPassword.
type fakePasswordRepo struct {
	user.PasswordRepository[*testutil.User]
	created  bool
	password string
}

func (f *fakePasswordRepo) CreateWithPassword(_ context.Context, obj *testutil.User, pwd string) (uint64, error) {
	f.created, f.password = true, pwd
	obj.ID = 5
	return 5, nil
}

type fakePasswordUsecase struct {
	user.PasswordUsecase[*testutil.User]
	repo *fakePasswordRepo
}

func (f fakePasswordUsecase) Repo() user.PasswordRepository[*testutil.User] { return f.repo }

type fakeCore struct {
	user.Usecase[*testutil.User]
}

func (fakeCore) Get(_ context.Context, id uint64) (*testutil.User, error) {
	return testutil.Stub(id), nil
}

func newCreateResolver(repo *fakePasswordRepo) *QueryResolverPassword[*testutil.User, uint64, any, uint64, any, any] {
	return NewQueryResolverPassword(QueryResolverPasswordConfig[*testutil.User, uint64, any, uint64, any, any]{
		Core:          fakeCore{},
		Password:      fakePasswordUsecase{repo: repo},
		UserFromInput: func(any, ...pkgModels.ApproveStatus) *testutil.User { return &testutil.User{} },
		NewPayload:    func(_ string, userID uint64, _ uint64) uint64 { return userID },
		ToGraphQL:     func(u *testutil.User) uint64 { return u.GetID() },
	})
}

func TestCreateUserHasNoKnownPassword(t *testing.T) {
	t.Parallel()

	repo := &fakePasswordRepo{}
	userID, err := newCreateResolver(repo).CreateUser(acl.WithNoPermCheck(context.Background()), nil)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if userID != 5 || !repo.created {
		t.Fatalf("user = %d, created = %v", userID, repo.created)
	}
	// Before the fix every user created via GraphQL had the password "GQL create user"
	if repo.password != "" {
		t.Fatalf("password = %q, want none", repo.password)
	}
}

func TestCreateUserChecksCreateAccess(t *testing.T) {
	t.Parallel()

	repo := &fakePasswordRepo{}
	_, err := newCreateResolver(repo).CreateUser(context.Background(), nil)
	if !errors.Is(err, acl.ErrNoPermissions) {
		t.Fatalf("err = %v, want %v", err, acl.ErrNoPermissions)
	}
	if repo.created {
		t.Fatal("user was created without the create permission")
	}
}
