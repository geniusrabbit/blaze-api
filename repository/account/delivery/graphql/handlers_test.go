package graphql

import (
	"testing"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type baseOnlyUser struct {
	userModels.UserBase
}

func (u *baseOnlyUser) NewWithID(id uint64) user.Model {
	return &baseOnlyUser{UserBase: userModels.UserBase{ID: id}}
}

func TestUserEmail(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()
		if got := userEmail((*testutil.User)(nil)); got != "" {
			t.Fatalf("userEmail(nil) = %q, want empty", got)
		}
	})

	t.Run("email model", func(t *testing.T) {
		t.Parallel()
		u := testutil.StubWithEmail("admin@example.com", pkgModels.ApprovedApproveStatus)
		if got := userEmail(u); got != "admin@example.com" {
			t.Fatalf("userEmail() = %q, want admin@example.com", got)
		}
	})

	t.Run("non email model", func(t *testing.T) {
		t.Parallel()
		u := &baseOnlyUser{}
		u.SetID(1)
		if got := userEmail(u); got != "" {
			t.Fatalf("userEmail(non-email) = %q, want empty", got)
		}
	})
}
