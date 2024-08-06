package acr

import (
	"errors"
	"os"
	"strings"

	"github.com/keel-hq/keel/extension/credentialshelper"
	"github.com/keel-hq/keel/types"
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
	clientId, ok := os.LookupEnv("AZURE_CLIENT_ID")
	if !ok {
		return nil, errors.New("AZURE_CLIENT_ID environment variable not set")
	}

	clientSecret, ok := os.LookupEnv("AZURE_CLIENT_SECRET")
	if !ok {
		return nil, errors.New("AZURE_CLIENT_SECRET environment variable not set")
	}

	return &types.Credentials{
		Username: clientId,
		Password: clientSecret,
	}, nil
}
