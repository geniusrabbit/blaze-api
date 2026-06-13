package facebook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	oa2 "github.com/geniusrabbit/blaze-api/pkg/auth/elogin/oauth2"
)

const facebookMeURL = "https://graph.facebook.com/v19.0/me?fields=id,name,email,picture&access_token="

// facebookUserDetails represents the user data structure returned by Facebook API.
type facebookUserDetails struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture struct {
		Data struct {
			URL          string `json:"url"`
			Width        int    `json:"width"`
			Height       int    `json:"height"`
			IsSilhouette bool   `json:"is_silhouette"`
		} `json:"data"`
	} `json:"picture"`
}

// FirstName extracts and returns the first name from the full name.
func (dt *facebookUserDetails) FirstName() string {
	name := strings.Split(dt.Name, " ")
	return name[0]
}

// LastName extracts and returns the last name from the full name.
func (dt *facebookUserDetails) LastName() string {
	name := strings.Split(dt.Name, " ")
	if len(name) == 1 {
		return ""
	}
	return name[1]
}

// FacebookUserData fetches user data from Facebook API using the provided access token.
func FacebookUserData(ctx context.Context, token *oauth2.Token, oauth2conf *oauth2.Config) (*elogin.UserData, error) {
	var fbUserDetails facebookUserDetails

	// Create request to fetch user details from Facebook API.
	fbUserDetailsRequest, _ := http.NewRequest("GET", facebookMeURL+url.QueryEscape(token.AccessToken), nil)
	fbUserDetailsResp, fbUserDetailsRespError := http.DefaultClient.Do(fbUserDetailsRequest)

	if fbUserDetailsRespError != nil {
		return nil, errors.Wrap(fbUserDetailsRespError, "Error occurred while getting information from Facebook")
	}
	defer func() {
		_ = fbUserDetailsResp.Body.Close()
	}()

	// Decode the JSON response into facebookUserDetails struct.
	decoderErr := json.NewDecoder(fbUserDetailsResp.Body).Decode(&fbUserDetails)
	if decoderErr != nil {
		return nil, errors.Wrap(decoderErr, "Error occurred while getting information from Facebook")
	}

	return &elogin.UserData{
		ID:         fbUserDetails.ID,
		Email:      fbUserDetails.Email,
		FirstName:  fbUserDetails.FirstName(),
		LastName:   fbUserDetails.LastName(),
		AvatarURL:  fbUserDetails.Picture.Data.URL,
		OAuth2conf: oauth2conf,
		Ext:        map[string]any{"scope": token.Extra("scope")},
	}, nil
}

// NewFacebookConfig creates and returns a new Facebook OAuth2 configuration instance.
func NewFacebookConfig(conf *oauth2.Config) *oa2.Config {
	return &oa2.Config{
		ProviderName: "facebook",
		OAuth2:       conf,
		Extractor:    FacebookUserData,
		StateCode:    "fcodec2024",
	}
}
