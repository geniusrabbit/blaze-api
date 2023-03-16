package models

import (
	"time"

	"github.com/geniusrabbit/api-template-base/model"
)

// FromAuthClientModel to local graphql model
func FromAuthClientModel(acc *model.AuthClient) *AuthClient {
	var deletedAt *time.Time
	if acc.DeletedAt.Valid {
		deletedAt = &time.Time{}
		*deletedAt = acc.DeletedAt.Time
	}
	return &AuthClient{
		ID:                 acc.ID,
		AccountID:          int(acc.AccountID),
		UserID:             int(acc.UserID),
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
		DeletedAt:          deletedAt,
	}
}

// FromAuthClientModelList converts model list to local model list
func FromAuthClientModelList(list []*model.AuthClient) []*AuthClient {
	userClients := make([]*AuthClient, 0, len(list))
	for _, u := range list {
		userClients = append(userClients, FromAuthClientModel(u))
	}
	return userClients
}
