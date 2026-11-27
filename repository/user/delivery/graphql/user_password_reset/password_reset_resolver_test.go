package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/messanger"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
	userusecase "github.com/geniusrabbit/blaze-api/repository/user/usecase"
)

type fakeEmail struct {
	users map[string]*testutil.User
	err   error
}

func (f fakeEmail) GetByEmail(_ context.Context, email string) (*testutil.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.users[email], nil
}

type fakePassword struct {
	user.PasswordUsecase[*testutil.User]
	resetErr error
}

func (f fakePassword) ResetPassword(_ context.Context, userID uint64) (*user.UserPasswordReset, *testutil.User, error) {
	if f.resetErr != nil {
		return nil, nil, f.resetErr
	}
	return &user.UserPasswordReset{UserID: userID, Token: "tok"}, testutil.Stub(userID), nil
}

func resetCtx(sent *int) context.Context {
	return messanger.WithMessanger(context.Background(),
		messanger.MessangerFunc(func(context.Context, string, []string, map[string]any) error {
			*sent++
			return nil
		}))
}

func TestResetUserPasswordDoesNotRevealAccounts(t *testing.T) {
	t.Parallel()

	known := map[string]*testutil.User{"known@example.com": testutil.Stub(5)}
	cases := map[string]struct {
		email    fakeEmail
		password fakePassword
		address  string
		wantSent int
	}{
		"registered":           {email: fakeEmail{users: known}, address: "known@example.com", wantSent: 1},
		"unknown":              {email: fakeEmail{users: known}, address: "nobody@example.com"},
		"lookup not permitted": {email: fakeEmail{err: acl.ErrNoPermissions.WithMessage("view user")}, address: "known@example.com"},
		"reset not permitted":  {email: fakeEmail{users: known}, password: fakePassword{resetErr: acl.ErrNoPermissions}, address: "known@example.com"},
	}
	var wantMessage string
	for name, tc := range cases {
		sent := 0
		r := NewPasswordResetQueryResolver(PasswordResetQueryResolverConfig[*testutil.User]{Email: tc.email, Password: tc.password})
		resp, err := r.ResetUserPassword(resetCtx(&sent), tc.address)
		// Before the fix an existing but inaccessible account answered with an
		// error, an unknown one with success
		if err != nil {
			t.Fatalf("%s: err = %v", name, err)
		}
		if wantMessage == "" {
			wantMessage = *resp.Message
		}
		if *resp.Message != wantMessage || sent != tc.wantSent {
			t.Errorf("%s: message = %q, sent = %d; want %q, %d", name, *resp.Message, sent, wantMessage, tc.wantSent)
		}
	}
}

func TestUpdateResetedUserPasswordHidesInaccessibleAccount(t *testing.T) {
	t.Parallel()

	r := NewPasswordResetQueryResolver(PasswordResetQueryResolverConfig[*testutil.User]{
		Email: fakeEmail{err: acl.ErrNoPermissions},
	})
	if _, err := r.UpdateResetedUserPassword(context.Background(), "tok", "known@example.com", "new-password"); !errors.Is(err, userusecase.ErrInvalidPasswordResetCode) {
		t.Fatalf("err = %v, want ErrInvalidPasswordResetCode", err)
	}
}
