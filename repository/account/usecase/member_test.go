package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/demdxx/rbac"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/mocks"
	"github.com/geniusrabbit/blaze-api/repository/account/usecase"
	usermocks "github.com/geniusrabbit/blaze-api/repository/user/mocks"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type testMemberSuite struct {
	suite.Suite

	ctx context.Context

	userRepo      *usermocks.MockRepository[*testutil.User]
	accountRepo   *mocks.MockSessionRepository[*testutil.User, *testAccount]
	memberRepo    *mocks.MockMemberRepository[*testutil.User, *testAccount]
	memberUsecase account.MemberUsecase[*testutil.User, *testAccount]
}

func (s *testMemberSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.userRepo = usermocks.NewMockRepository[*testutil.User](ctrl)
	s.accountRepo = mocks.NewMockSessionRepository[*testutil.User, *testAccount](ctrl)
	s.memberRepo = mocks.NewMockMemberRepository[*testutil.User, *testAccount](ctrl)
	s.memberUsecase = usecase.NewMemberUsecase(s.userRepo, s.accountRepo, s.memberRepo)
}

func (s *testMemberSuite) TestFetchListMembers() {
	s.memberRepo.EXPECT().EmptyObject().Return(&account.Member[*testutil.User, *testAccount]{})
	s.memberRepo.EXPECT().
		FetchListMembers(s.ctx, gomock.AssignableToTypeOf((*account.MemberFilter)(nil))).
		Return([]*account.Member[*testutil.User, *testAccount]{
			account.MemberStub[*testutil.User, *testAccount](1, 1, 1),
			account.MemberStub[*testutil.User, *testAccount](2, 1, 2),
		}, nil)

	members, err := s.memberUsecase.FetchListMembers(s.ctx,
		&account.MemberFilter{AccountID: []uint64{1}, UserID: []uint64{1, 2}},
	)

	s.NoError(err)
	s.Equal(2, len(members))
}

func (s *testMemberSuite) TestLinkMember() {
	s.memberRepo.EXPECT().EmptyObject().Return(&account.Member[*testutil.User, *testAccount]{})
	s.memberRepo.EXPECT().
		LinkMember(s.ctx, gomock.AssignableToTypeOf(&testAccount{}),
			true, gomock.AssignableToTypeOf(&testutil.User{})).
		Return(nil)

	accountObj := testAccountStub(1)
	userObj := testutil.Stub(101)
	err := s.memberUsecase.LinkMember(s.ctx, accountObj, true, userObj)
	s.NoError(err)
}

func (s *testMemberSuite) TestUnlinkMember() {
	s.memberRepo.EXPECT().
		UnlinkMember(s.ctx, gomock.AssignableToTypeOf(&testAccount{}),
			gomock.AssignableToTypeOf(&testutil.User{})).
		Return(nil)

	accountObj := testAccountStub(1)
	userObj := testutil.Stub(101)
	err := s.memberUsecase.UnlinkMember(s.ctx, accountObj, userObj)
	s.NoError(err)
}

func (s *testMemberSuite) TestSetMemberRolesHydratesUserAndAccount() {
	member := account.MemberStub[*testutil.User, *testAccount](2, 10, 101)
	acc := testAccountStub(10)
	usr := testutil.Stub(101)

	s.memberRepo.EXPECT().MemberByID(s.ctx, uint64(2)).Return(member, nil)
	s.accountRepo.EXPECT().Get(s.ctx, uint64(10)).Return(acc, nil)
	s.userRepo.EXPECT().Get(s.ctx, uint64(101)).Return(usr, nil)
	s.memberRepo.EXPECT().SetMemberRoles(s.ctx, acc, usr, "editor").Return(nil)

	got, err := s.memberUsecase.SetMemberRoles(s.ctx, 2, "editor")
	s.NoError(err)
	s.Equal(acc, got.Account)
	s.Equal(usr, got.User)
}

func (s *testMemberSuite) TestSetMemberRolesNotFound() {
	s.memberRepo.EXPECT().MemberByID(s.ctx, uint64(99)).Return(nil, nil)

	got, err := s.memberUsecase.SetMemberRoles(s.ctx, 99, "editor")
	s.Nil(got)
	s.ErrorContains(err, "member not found")
}

func (s *testMemberSuite) TestUnlinkAccountMemberHydratesUserAndAccount() {
	member := account.MemberStub[*testutil.User, *testAccount](2, 10, 101)
	acc := testAccountStub(10)
	usr := testutil.Stub(101)

	s.memberRepo.EXPECT().MemberByID(s.ctx, uint64(2)).Return(member, nil)
	s.accountRepo.EXPECT().Get(s.ctx, uint64(10)).Return(acc, nil)
	s.userRepo.EXPECT().Get(s.ctx, uint64(101)).Return(usr, nil)
	s.memberRepo.EXPECT().UnlinkMember(s.ctx, acc, usr).Return(nil)

	s.NoError(s.memberUsecase.UnlinkAccountMember(s.ctx, 2))
}

func (s *testMemberSuite) TestUnlinkAccountMemberNotFound() {
	s.memberRepo.EXPECT().MemberByID(s.ctx, uint64(99)).Return(nil, nil)

	s.ErrorContains(s.memberUsecase.UnlinkAccountMember(s.ctx, 99), "member not found")
}

func (s *testMemberSuite) TestInviteMemberNotFoundAfterLink() {
	acc := testAccountStub(10)
	usr := testutil.Stub(101)

	s.accountRepo.EXPECT().Get(s.ctx, uint64(10)).Return(acc, nil)
	s.userRepo.EXPECT().Get(s.ctx, uint64(101)).Return(usr, nil)
	s.memberRepo.EXPECT().LinkMember(s.ctx, acc, false, usr).Return(nil)
	s.memberRepo.EXPECT().SetMemberRoles(s.ctx, acc, usr).Return(nil)
	s.memberRepo.EXPECT().Member(s.ctx, uint64(101), uint64(10)).Return(nil, nil)

	got, err := s.memberUsecase.InviteMember(s.ctx, 10, 101)
	s.Nil(got)
	s.ErrorContains(err, "member not found")
}

func TestAccountMemberSuite(t *testing.T) {
	suite.Run(t, &testMemberSuite{})
}

// denyAll is a permission checker that grants nothing.
type denyAll struct{}

func (denyAll) CheckPermissions(context.Context, any, ...string) bool              { return false }
func (denyAll) CheckedPermissions(context.Context, any, ...string) rbac.Permission { return nil }
func (denyAll) ChildRoles() []rbac.Role                                            { return nil }
func (denyAll) ChildPermissions() []rbac.Permission                                { return nil }
func (denyAll) Permissions(...string) []rbac.Permission                            { return nil }
func (denyAll) HasPermission(...string) bool                                       { return false }

func TestUnlinkAccountMemberChecksAccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	userRepo := usermocks.NewMockRepository[*testutil.User](ctrl)
	accountRepo := mocks.NewMockSessionRepository[*testutil.User, *testAccount](ctrl)
	memberRepo := mocks.NewMockMemberRepository[*testutil.User, *testAccount](ctrl)
	uc := usecase.NewMemberUsecase(userRepo, accountRepo, memberRepo)

	// The caller is a member of account 1 without delete rights on members of account 10
	callerAcc := testAccountStub(1)
	callerAcc.SetPermissions(denyAll{})
	ctx := session.WithUserAccount(context.Background(), testutil.Stub(5), callerAcc)

	memberRepo.EXPECT().MemberByID(ctx, uint64(2)).Return(account.MemberStub[*testutil.User, *testAccount](2, 10, 101), nil)
	accountRepo.EXPECT().Get(ctx, uint64(10)).Return(testAccountStub(10), nil)
	userRepo.EXPECT().Get(ctx, uint64(101)).Return(testutil.Stub(101), nil)
	// No UnlinkMember call is expected: gomock fails the test if it happens

	if err := uc.UnlinkAccountMember(ctx, 2); !errors.Is(err, acl.ErrNoPermissions) {
		t.Fatalf("err = %v, want %v", err, acl.ErrNoPermissions)
	}
}
