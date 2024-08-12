package acr

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/quilla-hq/quilla/constants"
	"github.com/quilla-hq/quilla/extension/credentialshelper"
	"github.com/quilla-hq/quilla/types"
)

func init() {
	credentialshelper.RegisterCredentialsHelper("acr", New())
}

type CredentialsHelper struct {
	enabled bool
}

func New() *CredentialsHelper {
	return &CredentialsHelper{
		enabled: true,
	}
}

func (h *CredentialsHelper) IsEnabled() bool {
	return h.enabled
}

func (h *CredentialsHelper) GetCredentials(image *types.TrackedImage) (*types.Credentials, error) {
	if !h.enabled {
		return nil, errors.New("not initialised")
	}

	if !strings.Contains(image.Image.Registry(), "azurecr.io") && !strings.Contains(image.Image.Registry(), "pkg.dev") {
		return nil, credentialshelper.ErrUnsupportedRegistry
	}

	if credentials, err := readCredentialsFromEnv(); err == nil {
		return credentials, nil
	}

	return nil, errors.New("unable to read credentials from environment")
}

func readCredentialsFromEnv() (*types.Credentials, error) {
	clientId, ok := os.LookupEnv(constants.EnvAzureClientID)
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s environment variable not set", constants.EnvAzureClientID))
	}

	clientSecret, ok := os.LookupEnv(constants.EnvAzureClientSecret)
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s environment variable not set", constants.EnvAzureClientSecret))
	}

	return &types.Credentials{
		Username: clientId,
		Password: clientSecret,
	}, nil
}
