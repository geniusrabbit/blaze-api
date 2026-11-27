package graphql

import (
	"testing"

	"github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
)

func TestSessionMappingHidesTokens(t *testing.T) {
	t.Parallel()

	gql := FromSocialAccountSessionModel(&models.AccountSocialSession{
		Name: "default", AccessToken: "provider-access", RefreshToken: "provider-refresh",
	})
	// Before the fix provider tokens were returned to anyone who could view the account
	if gql.AccessToken != "" || gql.RefreshToken != "" || gql.Name != "default" {
		t.Fatalf("session = %+v", gql)
	}
}
