package acr

import (
	"testing"

	"github.com/keel-hq/keel/types"
	"github.com/keel-hq/keel/util/image"
)

func TestAzure(t *testing.T) {
	ch := New()

	imgRef, _ := image.Parse("https://mycontainerregistry.azurecr.io/repo/image:latest")
	creds, err := ch.GetCredentials(&types.TrackedImage{
		Image: imgRef,
	})

	if err == nil || creds != nil {
		t.Fatalf("Shouldn't pass")
	}
}
