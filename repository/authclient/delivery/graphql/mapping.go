package graphql

import (
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/repository/authclient/models"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// FromAuthClientModel to local graphql model
func FromAuthClientModel(acc *models.AuthClient) *gqlmodels.AuthClient {
	if acc == nil {
		return nil
	}
	return &gqlmodels.AuthClient{
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
		DeletedAt:          gqlmodels.DeletedAt(acc.DeletedAt),
	}
}

// FromAuthClientModelList converts model list to local model list
func FromAuthClientModelList(list []*models.AuthClient) []*gqlmodels.AuthClient {
	return xtypes.SliceApply(list, FromAuthClientModel)
}

// CreateFillModel fills adasset model from create input.
func CreateFillModel(inp *gqlmodels.AuthClientCreateInput, obj *models.AuthClient) *models.AuthClient {
	obj.ID = ""
	obj.UserID = gocast.PtrAsValue(inp.UserID, 0)
	obj.AccountID = gocast.PtrAsValue(inp.AccountID, 0)
	obj.Title = gocast.PtrAsValue(inp.Title, "")
	obj.Secret = gocast.PtrAsValue(inp.Secret, "")
	obj.RedirectURIs = inp.RedirectURIs
	obj.GrantTypes = inp.GrantTypes
	obj.ResponseTypes = inp.ResponseTypes
	obj.Scope = gocast.PtrAsValue(inp.Scope, "")
	obj.Audience = inp.Audience
	obj.SubjectType = inp.SubjectType
	obj.AllowedCORSOrigins = inp.AllowedCORSOrigins
	obj.Public = gocast.PtrAsValue(inp.Public, false)
	obj.ExpiresAt = gocast.PtrAsValue(inp.ExpiresAt, time.Time{})
	return obj
}

// UpdateFillModel fills adasset model from update input.
func UpdateFillModel(inp *gqlmodels.AuthClientUpdateInput, obj *models.AuthClient) {
	obj.UserID = gocast.PtrAsValue(inp.UserID, obj.UserID)
	obj.AccountID = gocast.PtrAsValue(inp.AccountID, obj.AccountID)
	obj.Title = gocast.PtrAsValue(inp.Title, obj.Title)
	obj.Secret = gocast.PtrAsValue(inp.Secret, obj.Secret)
	if inp.RedirectURIs != nil {
		obj.RedirectURIs = inp.RedirectURIs
	}
	if inp.GrantTypes != nil {
		obj.GrantTypes = inp.GrantTypes
	}
	if inp.ResponseTypes != nil {
		obj.ResponseTypes = inp.ResponseTypes
	}
	obj.Scope = gocast.PtrAsValue(inp.Scope, obj.Scope)
	if inp.Audience != nil {
		obj.Audience = inp.Audience
	}
	obj.SubjectType = gocast.PtrAsValue(inp.SubjectType, obj.SubjectType)
	if inp.AllowedCORSOrigins != nil {
		obj.AllowedCORSOrigins = inp.AllowedCORSOrigins
	}
	obj.Public = gocast.PtrAsValue(inp.Public, obj.Public)
	obj.ExpiresAt = gocast.PtrAsValue(inp.ExpiresAt, obj.ExpiresAt)
}
