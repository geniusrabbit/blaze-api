package resolvers

import (
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository/option"
	baseResolver "github.com/geniusrabbit/blaze-api/server/graphql/resolvers"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*baseResolver.Resolver
}

func NewResolver(provider *jwt.Provider, options option.Usecase) *Resolver {
	return &Resolver{
		Resolver: baseResolver.NewResolver(provider, options),
	}
}
