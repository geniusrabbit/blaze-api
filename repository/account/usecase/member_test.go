package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

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

func TestAccountMemberSuite(t *testing.T) {
	suite.Run(t, &testMemberSuite{})
}
