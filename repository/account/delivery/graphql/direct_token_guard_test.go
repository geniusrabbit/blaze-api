package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

type guardAccount struct {
	accountModels.AccountBase
}

func (a *guardAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	accountObject := &guardAccount{}
	accountObject.SetID(id)
	accountObject.ExtendAdminUsers(adminUserIDs...)
	return accountObject
}

func directTokenContext() context.Context {
	accountObject := &guardAccount{}
	accountObject.SetID(9)
	ctx := session.WithUserAccount(context.Background(), testutil.Stub(7), accountObject)
	return session.WithAuthorizerCode(ctx, session.AuthorizerCodeDirectToken)
}

func TestSwitchAccountDeniedForDirectTokenSession(t *testing.T) {
	t.Parallel()

	resolver := &AuthResolver[*testutil.User, *guardAccount]{}
	_, err := resolver.SwitchAccount(directTokenContext(), 10)
	if !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
		t.Fatalf("err = %v, want %v", err, session.ErrDirectTokenSessionForbidden)
	}
}

func TestMemberMutationsDeniedForDirectTokenSession(t *testing.T) {
	t.Parallel()

	resolver := &MemberQueryResolver[*testutil.User, *guardAccount, any]{}
	ctx := directTokenContext()
	mutations := map[string]func() error{
		"invite": func() error {
			_, err := resolver.Invite(ctx, 9, &models.InviteMemberInput{})
			return err
		},
		"update": func() error {
			_, err := resolver.Update(ctx, 1, &models.MemberInput{})
			return err
		},
		"remove": func() error {
			_, err := resolver.Remove(ctx, 1)
			return err
		},
		"approve": func() error {
			_, err := resolver.Approve(ctx, 1, "")
			return err
		},
		"reject": func() error {
			_, err := resolver.Reject(ctx, 1, "")
			return err
		},
	}
	for name, mutation := range mutations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := mutation(); !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
				t.Fatalf("err = %v, want %v", err, session.ErrDirectTokenSessionForbidden)
			}
		})
	}
}
