package acr

import (
	"testing"

	"github.com/quilla-hq/quilla/types"
	"github.com/quilla-hq/quilla/util/image"
)

func TestAzure(t *testing.T) {
	ch := New()

	imgRef, _ := image.Parse("https://mycontainerregistry.azurecr.io/repo/image:latest")
	creds, err := ch.GetCredentials(&types.TrackedImage{
		Image: imgRef,
	})

	if err == nil || creds != nil {
		t.Log(creds)
		t.Fatalf("Shouldn't pass")
	}
}
