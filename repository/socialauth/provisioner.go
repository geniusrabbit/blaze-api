package socialauth

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// SocialUserProvisioner is an optional hook that determines or creates the owner
// user for a social login callback.
//
// When injected into the socialauth usecase via [WithUserProvisioner], the
// provisioner is called during [Usecase.Register] instead of the built-in
// default flow.  If no provisioner is set the default flow runs automatically.
//
// Default flow (provisioner == nil):
//  1. Look up AccountSocial by (provider, social_id).
//  2. Found → load and return the existing user.
//  3. Not found + sessionUser is not anonymous → use sessionUser as owner.
//  4. Not found + sessionUser is anonymous → create a minimal user via
//     user.Repository.Create and set its ID.
//  5. Always save AccountSocial.UserID = owner.GetID().
type SocialUserProvisioner[TUser user.Model] interface {
	// EnsureUser returns the owner user for the given social login.
	// sessionUser is the currently logged-in user from the request context
	// (may be anonymous).  provider is the OAuth2 provider name.
	// data contains the profile information returned by the OAuth2 provider.
	EnsureUser(ctx context.Context, sessionUser TUser, provider string, data *elogin.UserData) (TUser, error)
}
