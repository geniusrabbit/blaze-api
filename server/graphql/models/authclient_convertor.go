package models

import (
	"github.com/demdxx/xtypes"
	"github.com/geniusrabbit/blaze-api/model"
)

// FromAuthClientModel to local graphql model
func FromAuthClientModel(acc *model.AuthClient) *AuthClient {
	if acc == nil {
		return nil
	}
	return &AuthClient{
		ID:                 acc.ID,
		AccountID:          acc.AccountID,
		UserID:             acc.UserID,
		Title:              acc.Title,
		Secret:             acc.Secret,
		RedirectURIs:       acc.RedirectURIs,
		GrantTypes:         acc.GrantTypes,
		ResponseTypes:      acc.ResponseTypes,
		Scope:              acc.Scope,
		Audience:           acc.Audience,
		SubjectType:        acc.SubjectType,
		AllowedCORSOrigins: acc.AllowedCORSOrigins,
		Public:             acc.Public,
		CreatedAt:          acc.CreatedAt,
		UpdatedAt:          acc.UpdatedAt,
		DeletedAt:          DeletedAt(acc.DeletedAt),
	}
}

// FromAuthClientModelList converts model list to local model list
func FromAuthClientModelList(list []*model.AuthClient) []*AuthClient {
	return xtypes.SliceApply(list, FromAuthClientModel)
}
