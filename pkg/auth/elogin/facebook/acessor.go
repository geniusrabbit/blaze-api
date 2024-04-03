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

func (dt *facebookUserDetails) FirstName() string {
	name := strings.Split(dt.Name, " ")
	return name[0]
}

func (dt *facebookUserDetails) LastName() string {
	name := strings.Split(dt.Name, " ")
	if len(name) == 1 {
		return ""
	}
	return name[1]
}

// FacebookUserData returns user data from Facebook API using access token
func FacebookUserData(ctx context.Context, token *oauth2.Token, oauth2conf *oauth2.Config) (*elogin.UserData, error) {
	var fbUserDetails facebookUserDetails

	fbUserDetailsRequest, _ := http.NewRequest("GET", facebookMeURL+url.QueryEscape(token.AccessToken), nil)
	fbUserDetailsResp, fbUserDetailsRespError := http.DefaultClient.Do(fbUserDetailsRequest)

	if fbUserDetailsRespError != nil {
		return nil, errors.Wrap(fbUserDetailsRespError, "Error occurred while getting information from Facebook")
	}
	defer fbUserDetailsResp.Body.Close()

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

// NewFacebookConfig creates a new instance of Facebook oauth2 configuration
func NewFacebookConfig(conf *oauth2.Config) *oa2.Config {
	return &oa2.Config{
		ProviderName: "facebook",
		OAuth2:       conf,
		Extractor:    FacebookUserData,
		StateCode:    "fcodec2024",
	}
}
