package graphql

import (
	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

// FromSocialAccountModel converts a domain AccountSocial model to a GraphQL SocialAccount model.
func FromSocialAccountModel(acc *models.AccountSocial) *gqlmodels.SocialAccount {
	if acc == nil {
		return nil
	}
	return &gqlmodels.SocialAccount{
		ID:        acc.ID,
		UserID:    acc.UserID,
		SocialID:  acc.SocialID,
		Provider:  acc.Provider,
		Username:  acc.Username,
		Email:     acc.Email,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		Avatar:    acc.Avatar,
		Link:      acc.Link,
		Data:      *types.MustNullableJSONFrom(acc.Data.Data),
		Sessions:  FromSocialAccountSessionModelList(acc.Sessions),
		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
		DeletedAt: gqlmodels.DeletedAt(acc.DeletedAt),
	}
}

// FromSocialAccountModelList converts a slice of domain models to GraphQL models.
func FromSocialAccountModelList(list []*models.AccountSocial) []*gqlmodels.SocialAccount {
	return xtypes.SliceApply(list, FromSocialAccountModel)
}

// FromSocialAccountSessionModel converts a domain AccountSocialSession model to a GraphQL model.
func FromSocialAccountSessionModel(sess *models.AccountSocialSession) *gqlmodels.SocialAccountSession {
	if sess == nil {
		return nil
	}
	return &gqlmodels.SocialAccountSession{
		Name:            sess.Name,
		SocialAccountID: sess.AccountSocialID,
		AccessToken:     sess.AccessToken,
		RefreshToken:    sess.RefreshToken,
		Scope:           sess.Scopes,
		ExpiresAt:       gocast.IfThen(sess.ExpiresAt.Valid, &sess.ExpiresAt.Time, nil),
		CreatedAt:       sess.CreatedAt,
		UpdatedAt:       sess.UpdatedAt,
		DeletedAt:       gqlmodels.DeletedAt(sess.DeletedAt),
	}
}

// FromSocialAccountSessionModelList converts a slice of session models to GraphQL models.
func FromSocialAccountSessionModelList(list []*models.AccountSocialSession) []*gqlmodels.SocialAccountSession {
	return xtypes.SliceApply(list, FromSocialAccountSessionModel)
}
