package common

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Get image digest from registry
func GetImageTagDigest(authenticator remote.Option, imageUrl string, imageTag string) string {

	requestUrl := fmt.Sprintf("%s:%s", imageUrl, imageTag)
	ref, err := name.ParseReference(requestUrl)
	if err != nil {
		return ""
	}
	// Get image manifest
	img, err := remote.Get(ref, authenticator)
	if err != nil {
		return ""
	}

	// Extract image digest
	digest := img.Digest.String()
	return digest
}
