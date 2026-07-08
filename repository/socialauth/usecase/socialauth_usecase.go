package usecase

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	socialAccountModels "github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// UsecaseOption configures the socialauth Usecase.
type UsecaseOption[T user.Model] func(*Usecase[T])

// WithUserProvisioner sets the optional user provisioner.
// If not set, the built-in default flow is used (see [SocialUserProvisioner]).
func WithUserProvisioner[T user.Model](p socialauth.SocialUserProvisioner[T]) UsecaseOption[T] {
	return func(u *Usecase[T]) {
		u.provisioner = p
	}
}

// Usecase provides social auth business logic.
// It depends only on [user.Model] — no Email or Password trait constraints.
type Usecase[T user.Model] struct {
	userRepo       user.Repository[T]
	socAccountRepo socialauth.Repository
	provisioner    socialauth.SocialUserProvisioner[T] // nil = default flow
}

// New creates a socialauth Usecase.
//
//   - socAccountRepo — social account repository
//   - userRepo       — user repository used by the default provisioning flow
//   - opts           — optional [UsecaseOption] values (e.g. [WithUserProvisioner])
func New[T user.Model](
	socAccountRepo socialauth.Repository,
	userRepo user.Repository[T],
	opts ...UsecaseOption[T],
) *Usecase[T] {
	uc := &Usecase[T]{
		userRepo:       userRepo,
		socAccountRepo: socAccountRepo,
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// Get social account by id.
func (u *Usecase[T]) Get(ctx context.Context, id uint64) (*socialAccountModels.AccountSocial, error) {
	if !acl.HaveAccessView(ctx, &socialAccountModels.AccountSocial{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "get social account")
	}
	return u.socAccountRepo.Get(ctx, id)
}

// List social accounts by filter.
func (u *Usecase[T]) List(ctx context.Context, filter *socialauth.Filter) ([]*socialAccountModels.AccountSocial, error) {
	if !acl.HaveAccessList(ctx, &socialAccountModels.AccountSocial{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list social accounts")
	}
	return u.socAccountRepo.List(ctx, filter)
}

// Register links the social account to an owner user.
//
// If the owner user (ownerObj) has no ID yet, the provisioner (if configured)
// or the built-in default flow is used to find or create the owner:
//
//  1. Custom provisioner → delegate entirely to SocialUserProvisioner.EnsureUser.
//  2. Default flow:
//     a. sessionUser is not anonymous → use it as owner.
//     b. sessionUser is anonymous → create a minimal user via userRepo.Create.
func (u *Usecase[T]) Register(ctx context.Context, ownerObj user.Model, accountObj *socialAccountModels.AccountSocial) (uint64, error) {
	owner, ok := ownerObj.(T)
	if !ok {
		return 0, errors.New("socialauth: invalid user type for Register")
	}
	if !acl.HavePermissions(ctx, "account.register") {
		return 0, errors.Wrap(acl.ErrNoPermissions, "register/link social account")
	}

	err := database.ContextTransactionExec(ctx, func(txctx context.Context, _ *gorm.DB) error {
		if owner.GetID() == 0 {
			resolved, err := u.ensureOwner(txctx, owner, accountObj)
			if err != nil {
				return err
			}
			owner = resolved
		}
		accountObj.UserID = owner.GetID()
		aid, err := u.socAccountRepo.Create(txctx, accountObj)
		if err != nil {
			return err
		}
		accountObj.ID = aid
		return nil
	})
	return accountObj.ID, err
}

// ensureOwner resolves or creates the owner user.
func (u *Usecase[T]) ensureOwner(ctx context.Context, current T, accountObj *socialAccountModels.AccountSocial) (T, error) {
	// Custom provisioner takes full control.
	if u.provisioner != nil {
		sessionUser := u.sessionUser(ctx)
		userData := &elogin.UserData{
			ID:    accountObj.SocialID,
			Email: accountObj.Email,
		}
		return u.provisioner.EnsureUser(ctx, sessionUser, accountObj.Provider, userData)
	}

	// Default flow: use non-anonymous session user or create a minimal one.
	sessionUser := u.sessionUser(ctx)
	if !sessionUser.IsAnonymous() {
		return sessionUser, nil
	}

	uid, err := u.userRepo.Create(ctx, current)
	if err != nil {
		return current, errors.Wrap(err, "socialauth: create owner user")
	}
	if setter, ok := any(current).(interface{ SetID(uint64) }); ok {
		setter.SetID(uid)
	}
	return current, nil
}

// sessionUser extracts the typed session user or returns zero value.
func (u *Usecase[T]) sessionUser(ctx context.Context) T {
	var zero T
	if su, ok := any(session.UserModel(ctx)).(T); ok {
		return su
	}
	return zero
}

// Update social account by id.
func (u *Usecase[T]) Update(ctx context.Context, id uint64, account *socialAccountModels.AccountSocial) error {
	if !acl.HaveAccessUpdate(ctx, account) {
		return errors.Wrap(acl.ErrNoPermissions, "update social account")
	}
	return u.socAccountRepo.Update(ctx, id, account)
}

// Token returns social account token by name and account social ID.
func (u *Usecase[T]) Token(ctx context.Context, name string, accountSocialID uint64) (*elogin.Token, error) {
	return u.socAccountRepo.Token(ctx, name, accountSocialID)
}

// SetToken stores a social account token by name and account social ID.
func (u *Usecase[T]) SetToken(ctx context.Context, name string, accountSocialID uint64, token *elogin.Token) error {
	return u.socAccountRepo.SetToken(ctx, name, accountSocialID, token)
}
