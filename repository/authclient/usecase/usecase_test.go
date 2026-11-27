package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/demdxx/rbac"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/mocks"
	"github.com/geniusrabbit/blaze-api/repository/authclient/models"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type testSuite struct {
	suite.Suite

	ctx context.Context

	authclientRepo    *mocks.MockRepository
	authclientUsecase authclient.Usecase
}

func (s *testSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.authclientRepo = mocks.NewMockRepository(ctrl)
	s.authclientUsecase = NewAuthclientUsecase(s.authclientRepo)
}

func (s *testSuite) TestGet() {
	s.authclientRepo.EXPECT().Get(s.ctx, "2").
		Return(&models.AuthClient{ID: "2"}, nil)

	role, err := s.authclientUsecase.Get(s.ctx, "2")
	s.NoError(err)
	s.Equal("2", role.ID)
}

func (s *testSuite) TestGetGetError() {
	s.authclientRepo.EXPECT().Get(s.ctx, "2").
		Return(nil, errors.New("test"))

	role, err := s.authclientUsecase.Get(s.ctx, "2")
	s.Error(err)
	s.Nil(role)
}

func (s *testSuite) TestFetchList() {
	s.authclientRepo.EXPECT().
		FetchList(s.ctx, gomock.AssignableToTypeOf(&authclient.Filter{})).
		Return([]*models.AuthClient{{ID: "1"}, {ID: "2"}}, nil)

	roles, err := s.authclientUsecase.FetchList(s.ctx,
		&authclient.Filter{ID: []string{"1", "2"}})
	s.NoError(err)
	s.Equal(2, len(roles))
}

func (s *testSuite) TestCount() {
	s.authclientRepo.EXPECT().
		Count(s.ctx, gomock.AssignableToTypeOf(&authclient.Filter{})).
		Return(int64(2), nil)

	count, err := s.authclientUsecase.Count(s.ctx,
		&authclient.Filter{ID: []string{"1", "2"}})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestCreate() {
	s.authclientRepo.EXPECT().
		Create(s.ctx, gomock.AssignableToTypeOf(&models.AuthClient{}), historylog.Message("create authclient")).
		Return("101", nil)

	id, err := s.authclientUsecase.Create(s.ctx,
		&models.AuthClient{ID: "", Title: "test1"}, historylog.Message("create authclient"))
	s.NoError(err)
	s.Equal(id, "101")
}

func (s *testSuite) TestUpdate() {
	s.authclientRepo.EXPECT().Get(s.ctx, "101").Return(&models.AuthClient{ID: "101", Title: "test"}, nil)
	s.authclientRepo.EXPECT().
		Update(gomock.AssignableToTypeOf(s.ctx),
			"101", gomock.AssignableToTypeOf(&models.AuthClient{}), historylog.Message("update authclient")).
		Return(nil)

	err := s.authclientUsecase.Update(s.ctx, "101",
		&models.AuthClient{Title: "test-test"}, historylog.Message("update authclient"))
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	stype := gomock.AssignableToTypeOf("1")
	s.authclientRepo.EXPECT().
		Get(gomock.AssignableToTypeOf(s.ctx), "1").
		Return(&models.AuthClient{ID: "1"}, nil)
	s.authclientRepo.EXPECT().
		Delete(gomock.AssignableToTypeOf(s.ctx), stype, gomock.AssignableToTypeOf(historylog.Message(""))).
		Return(nil)

	err := s.authclientUsecase.Delete(s.ctx, "1", historylog.Message("delete authclient"))
	s.NoError(err)
}

func (s *testSuite) TestDeleteNotFound() {
	s.authclientRepo.EXPECT().
		Get(s.ctx, "9999").
		Return(nil, sql.ErrNoRows)
	err := s.authclientUsecase.Delete(s.ctx, "9999", historylog.Message("delete authclient"))
	s.EqualError(err, sql.ErrNoRows.Error())
}

func TestRoleSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}

// patternPerms grants exactly the listed permission patterns.
type patternPerms map[string]bool

func (p patternPerms) CheckPermissions(_ context.Context, _ any, patterns ...string) bool {
	return p.HasPermission(patterns...)
}
func (patternPerms) CheckedPermissions(context.Context, any, ...string) rbac.Permission { return nil }
func (patternPerms) ChildRoles() []rbac.Role                                            { return nil }
func (patternPerms) ChildPermissions() []rbac.Permission                                { return nil }
func (patternPerms) Permissions(...string) []rbac.Permission                            { return nil }
func (p patternPerms) HasPermission(patterns ...string) bool {
	for _, pattern := range patterns {
		if p[pattern] {
			return true
		}
	}
	return false
}

type ownerTestAccount struct {
	accountModels.AccountBase
}

func (a *ownerTestAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	acc := &ownerTestAccount{}
	acc.ID = id
	acc.Admins = adminUserIDs
	return acc
}

// ownerCtx is a session of user 5 in account 3 with the given permissions.
func ownerCtx(perms ...string) context.Context {
	acc := &ownerTestAccount{}
	acc.ID = 3
	granted := patternPerms{}
	for _, p := range perms {
		granted[p] = true
	}
	acc.SetPermissions(granted)
	return session.WithUserAccount(context.Background(), testutil.Stub(5), acc)
}

func TestCreateAssignsSessionOwner(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockRepository(gomock.NewController(t))
	uc := NewAuthclientUsecase(repo)
	ctx := ownerCtx("create.*")

	var created *models.AuthClient
	repo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, obj *models.AuthClient, _ ...authclient.QOption) (string, error) {
			created = obj
			return "c1", nil
		})
	if _, err := uc.Create(ctx, &models.AuthClient{Title: "app"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.UserID != 5 || created.AccountID != 3 {
		t.Fatalf("owner = user %d account %d, want user 5 account 3", created.UserID, created.AccountID)
	}
}

func TestCreateRejectsForeignOwner(t *testing.T) {
	t.Parallel()

	// A client of another user would let its creator act as that user
	// through the client_credentials grant.
	for name, obj := range map[string]*models.AuthClient{
		"other user":    {UserID: 7},
		"other account": {AccountID: 9},
	} {
		repo := mocks.NewMockRepository(gomock.NewController(t))
		_, err := NewAuthclientUsecase(repo).Create(ownerCtx("create.*"), obj)
		if !errors.Is(err, acl.ErrNoPermissions) {
			t.Errorf("%s: err = %v, want %v", name, err, acl.ErrNoPermissions)
		}
	}

	// The global permission may set any owner
	repo := mocks.NewMockRepository(gomock.NewController(t))
	ctx := ownerCtx("create.*", "create.all")
	repo.EXPECT().Create(ctx, gomock.Any()).Return("c2", nil)
	if _, err := NewAuthclientUsecase(repo).Create(ctx, &models.AuthClient{UserID: 7, AccountID: 9}); err != nil {
		t.Fatalf("create with create.all: %v", err)
	}
}

func TestUpdateRejectsOwnerChange(t *testing.T) {
	t.Parallel()

	repo := mocks.NewMockRepository(gomock.NewController(t))
	ctx := ownerCtx("update.*")
	repo.EXPECT().Get(ctx, "c1").Return(&models.AuthClient{ID: "c1", UserID: 5, AccountID: 3}, nil)

	err := NewAuthclientUsecase(repo).Update(ctx, "c1", &models.AuthClient{ID: "c1", UserID: 7, AccountID: 3})
	if !errors.Is(err, acl.ErrNoPermissions) {
		t.Fatalf("err = %v, want %v", err, acl.ErrNoPermissions)
	}
}
