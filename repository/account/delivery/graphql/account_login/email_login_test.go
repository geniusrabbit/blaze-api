package accountlogin

import (
	"context"
	"errors"
	"testing"

	"github.com/geniusrabbit/blaze-api/repository/user"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type fakeEmailRepo struct{ users map[string]*testutil.User }

func (f fakeEmailRepo) GetByEmail(_ context.Context, email string) (*testutil.User, error) {
	return f.users[email], nil
}

type fakePasswordRepo struct {
	user.PasswordRepository[*testutil.User]
	password string
	calls    int
}

func (f *fakePasswordRepo) GetByPassword(_ context.Context, userID uint64, pass string) (*testutil.User, error) {
	f.calls++
	if pass != f.password {
		return nil, userrepo.ErrInvalidPassword
	}
	return testutil.Stub(userID), nil
}

func TestEmailLoginErrorsDoNotRevealAccounts(t *testing.T) {
	t.Parallel()

	passwords := &fakePasswordRepo{password: "right"}
	login := NewEmailPasswordLogin[*testutil.User](
		fakeEmailRepo{users: map[string]*testutil.User{"known@example.com": testutil.Stub(5)}}, passwords)
	ctx := context.Background()

	if u, err := login.Login(ctx, "known@example.com", "right"); err != nil || u.GetID() != 5 {
		t.Fatalf("valid login: user = %v, err = %v", u, err)
	}
	// Before the fix an unknown email failed with "record not found" and a
	// wrong password with "invalid password"
	_, unknownErr := login.Login(ctx, "nobody@example.com", "right")
	_, wrongErr := login.Login(ctx, "known@example.com", "wrong")
	if !errors.Is(unknownErr, ErrInvalidCredentials) || !errors.Is(wrongErr, ErrInvalidCredentials) {
		t.Fatalf("unknown = %v, wrong = %v; want ErrInvalidCredentials for both", unknownErr, wrongErr)
	}
	if passwords.calls != 2 {
		t.Fatalf("password repo calls = %d; an unknown email must not reach it", passwords.calls)
	}
}
