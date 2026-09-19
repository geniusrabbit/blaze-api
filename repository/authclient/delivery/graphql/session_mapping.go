package graphql

import (
	"github.com/demdxx/xtypes"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/models"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// FromAuthSessionModel converts a domain session to the GraphQL model.
func FromAuthSessionModel(sess *models.AuthSession) *gqlmodels.AuthSession {
	if sess == nil {
		return nil
	}
	return &gqlmodels.AuthSession{
		ID:                    sess.ID,
		Active:                sess.Active,
		ClientID:              sess.ClientID,
		Username:              sess.Username,
		Subject:               sess.Subject,
		RequestID:             sess.RequestID,
		RequestedScope:        []string(sess.RequestedScope),
		GrantedScope:          []string(sess.GrantedScope),
		RequestedAudience:     []string(sess.RequestedAudience),
		GrantedAudience:       []string(sess.GrantedAudience),
		AccessToken:           sess.AccessToken,
		AccessTokenExpiresAt:  sess.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: sess.RefreshTokenExpiresAt,
		CreatedAt:             sess.CreatedAt,
		UpdatedAt:             sess.UpdatedAt,
	}
}

// FromAuthSessionModelList converts a domain session list to GraphQL models.
func FromAuthSessionModelList(list []*models.AuthSession) []*gqlmodels.AuthSession {
	return xtypes.SliceApply(list, FromAuthSessionModel)
}

// FromSessionFilterGraphQL maps a GraphQL filter to the repository filter.
func FromSessionFilterGraphQL(fl *gqlmodels.AuthSessionListFilter) *authclient.SessionFilter {
	if fl == nil {
		return nil
	}
	return &authclient.SessionFilter{
		ID:       fl.ID,
		ClientID: fl.ClientID,
		Username: fl.Username,
		Subject:  fl.Subject,
		Active:   fl.Active,
		Query:    ptrString(fl.Query),
	}
}

// FromSessionOrderGraphQL maps a GraphQL order to the repository order.
func FromSessionOrderGraphQL(ord *gqlmodels.AuthSessionListOrder) *authclient.SessionListOrder {
	if ord == nil {
		return nil
	}
	return &authclient.SessionListOrder{
		ID:                   pkgModels.Order(ord.ID.AsOrder()),
		ClientID:             pkgModels.Order(ord.ClientID.AsOrder()),
		CreatedAt:            pkgModels.Order(ord.CreatedAt.AsOrder()),
		AccessTokenExpiresAt: pkgModels.Order(ord.AccessTokenExpiresAt.AsOrder()),
	}
}

func ptrString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
