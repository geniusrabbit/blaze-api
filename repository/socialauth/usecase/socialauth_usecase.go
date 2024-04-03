package usecase

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var ErrLinkToExistsUser = errors.New("to connect social account to exists user, need to be authorized")

type Usecase struct {
	userRepo       user.Repository
	socAccountRepo socialauth.Repository
}

func New(userRepo user.Repository, socAccountRepo socialauth.Repository) *Usecase {
	return &Usecase{
		userRepo:       userRepo,
		socAccountRepo: socAccountRepo,
	}
}

// Get social account by id
func (u *Usecase) Get(ctx context.Context, id uint64) (*model.AccountSocial, error) {
	if !acl.HaveAccessView(ctx, &model.AccountSocial{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "get social account")
	}
	return u.socAccountRepo.Get(ctx, id)
}

// List social accounts by filter
func (u *Usecase) List(ctx context.Context, filter *socialauth.Filter) ([]*model.AccountSocial, error) {
	if !acl.HaveAccessList(ctx, &model.AccountSocial{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list social accounts")
	}
	return u.socAccountRepo.List(ctx, filter)
}

// Register new social account and link it to the user
func (u *Usecase) Register(ctx context.Context, ownerObj *model.User, accountObj *model.AccountSocial) (uint64, error) {
	if !acl.HavePermissions(ctx, "account.register") {
		return 0, errors.Wrap(acl.ErrNoPermissions, "register/link social account")
	}
	// Execute all operations in transaction
	err := database.ContextTransactionExec(ctx, func(txctx context.Context, tx *gorm.DB) error {
		// If user not exists then create it
		if ownerObj.ID == 0 {
			if existsUser, err := u.userRepo.GetByEmail(txctx, ownerObj.Email); existsUser != nil && err == nil {
				return ErrLinkToExistsUser
			}

			// Create user in the database to link social account
			uid, err := u.userRepo.Create(txctx, ownerObj, "")
			if err != nil {
				return err
			}
			ownerObj.ID = uid
		}

		// Link social account to the user
		accountObj.UserID = ownerObj.ID

		// Create social account
		aid, err := u.socAccountRepo.Create(txctx, accountObj)
		if err != nil {
			return err
		}
		accountObj.ID = aid

		return nil
	})
	return accountObj.ID, err
}

// Update social account by id
func (u *Usecase) Update(ctx context.Context, id uint64, account *model.AccountSocial) error {
	if !acl.HaveAccessUpdate(ctx, account) {
		return errors.Wrap(acl.ErrNoPermissions, "update social account")
	}
	return u.socAccountRepo.Update(ctx, id, account)
}

// Token returns social account token by id
func (u *Usecase) Token(ctx context.Context, name string, accountSocialID uint64) (*elogin.Token, error) {
	return u.socAccountRepo.Token(ctx, name, accountSocialID)
}

// SetToken sets social account token by id
func (u *Usecase) SetToken(ctx context.Context, name string, accountSocialID uint64, token *elogin.Token) error {
	return u.socAccountRepo.SetToken(ctx, name, accountSocialID, token)
}
