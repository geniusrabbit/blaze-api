package rest

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository/account"
	socialAccountModels "github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

type Option func(*Oauth2Wrapper)

// WithErrorRedirectURL sets the error redirect URL
func WithErrorRedirectURL(url string) Option {
	return func(w *Oauth2Wrapper) {
		w.errorRedirectURL = url
	}
}

// WithSuccessRedirectURL sets the success redirect URL
func WithSuccessRedirectURL(url string) Option {
	return func(w *Oauth2Wrapper) {
		w.successRedirectURL = url
	}
}

// WithSocialAuthUsecase sets the social auth usecase
func WithSocialAuthUsecase(usecase socialauth.Usecase) Option {
	return func(w *Oauth2Wrapper) {
		w.socialAuthUsecase = usecase
	}
}

// WithSessionProvider sets the session provider
func WithSessionProvider(provider *jwt.Provider) Option {
	return func(w *Oauth2Wrapper) {
		w.sessProvider = provider
	}
}

// WithAccountResolver resolves default account ID for social login session tokens.
func WithAccountResolver[TAccount account.Model](
	fetchList func(ctx context.Context, filter *account.Filter) ([]TAccount, error),
) Option {
	return func(w *Oauth2Wrapper) {
		w.resolveAccountID = func(ctx context.Context, userID uint64) (uint64, error) {
			acclist, err := fetchList(ctx, &account.Filter{
				UserID: []uint64{userID},
			})
			if err != nil || len(acclist) == 0 {
				return 0, err
			}
			accountID := uint64(0)
			for _, acc := range acclist {
				if acc.GetApprove().IsApproved() {
					return acc.GetID(), nil
				}
				if accountID == 0 {
					accountID = acc.GetID()
				}
			}
			return accountID, nil
		}
	}
}

// WithSocialAccountFactory overrides how the AccountSocial record is built from
// the OAuth2 provider data.  Use this to map provider-specific fields into your
// AccountSocial model without forking the framework.
//
//	socialrest.WithSocialAccountFactory(func(provider string, data *elogin.UserData) *socialaccount.AccountSocial {
//	    return &socialaccount.AccountSocial{
//	        Provider: provider,
//	        SocialID: data.ID,
//	        Username: data.Username,
//	        Avatar:   data.AvatarURL,
//	        Email:    data.Email,
//	    }
//	})
func WithSocialAccountFactory(fn func(provider string, data *elogin.UserData) *socialAccountModels.AccountSocial) Option {
	return func(w *Oauth2Wrapper) {
		w.socialAccountFactory = fn
	}
}

// WithSocialAccountUpdater overrides how an existing AccountSocial is refreshed
// after subsequent logins.  Only non-empty fields in data should be applied.
//
//	socialrest.WithSocialAccountUpdater(func(acc *socialaccount.AccountSocial, data *elogin.UserData) {
//	    if data.AvatarURL != "" { acc.Avatar = data.AvatarURL }
//	})
func WithSocialAccountUpdater(fn func(acc *socialAccountModels.AccountSocial, data *elogin.UserData)) Option {
	return func(w *Oauth2Wrapper) {
		w.socialAccountUpdater = fn
	}
}

// WithUserProvisioner wires a [socialauth.SocialUserProvisioner] into the
// Oauth2Wrapper.  This is the preferred way to customise user creation/linking
// without forking the framework.
//
// The provisioner is stored on the wrapper and passed to the usecase only when
// the usecase implements a compatible interface.  For full control, configure
// the provisioner on the usecase directly via [usecase.WithUserProvisioner].
func WithUserProvisioner[T user.Model](p socialauth.SocialUserProvisioner[T]) Option {
	return func(w *Oauth2Wrapper) {
		// Store as a type-erased wrapper so the non-generic Oauth2Wrapper can
		// hold it.  The provisioner is invoked by the typed usecase internally.
		w.userProvisionerErased = func(ctx context.Context, provider string, data *elogin.UserData) (user.Model, error) {
			var zero T
			if su, ok := any(w.sessionUserFn(ctx)).(T); ok {
				return p.EnsureUser(ctx, su, provider, data)
			}
			return p.EnsureUser(ctx, zero, provider, data)
		}
	}
}
