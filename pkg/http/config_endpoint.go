package http

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/quilla-hq/quilla/constants"
)

type AADAuth struct {
	ClientId          string `json:"clientId"`
	Authority         string `json:"authority"`
	ValidateAuthority bool   `json:"validateAuthority"`
	// AuthorityMetadata         *string  `json:"authorityMetadata"`
	// KnownAuthorities          []string `json:"knownAuthorities"`
	RedirectUri string `json:"redirectUri"`
	// PostLogoutRedirectUri     string   `json:"postLogoutRedirectUri"`
	NavigateToLoginRequestUrl bool `json:"navigateToLoginRequestUrl"`
}

type AADCache struct {
	CacheLocation          string `json:"cacheLocation"`
	StoreAuthStateInCookie bool   `json:"storeAuthStateInCookie"`
}

type AADConfig struct {
	Auth   AADAuth  `json:"auth"`
	Cache  AADCache `json:"cache"`
	Scopes []string `json:"scopes"`
}

type AppConfig struct {
	Aad       interface{} `json:"aad"`
	BasicAuth interface{} `json:"basicAuth"`
}

func (s *TriggerServer) configHandler(resp http.ResponseWriter, req *http.Request) {

	config := AppConfig{}

	enabled, err := strconv.ParseBool(os.Getenv(constants.EnvAzureAADEnabled))
	if err == nil && enabled {
		config.Aad = AADConfig{
			Auth: AADAuth{
				ClientId:                  os.Getenv(constants.EnvAzureClientID),
				Authority:                 fmt.Sprintf("https://login.microsoftonline.com/%s/", os.Getenv(constants.EnvAzureTenantID)),
				ValidateAuthority:         true,
				RedirectUri:               os.Getenv(constants.EnvAzureRedirectUri),
				NavigateToLoginRequestUrl: false,
			},
			Cache: AADCache{
				CacheLocation:          "localStorage",
				StoreAuthStateInCookie: true,
			},
			Scopes: []string{},
		}
	} else {
		config.Aad = false
	}

	enabled, err = strconv.ParseBool(os.Getenv(constants.EnvBasicAuthEnabled))
	if err == nil && enabled {
		config.BasicAuth = true
	} else {
		config.BasicAuth = false
	}

	response(config, 200, nil, resp, req)
}
