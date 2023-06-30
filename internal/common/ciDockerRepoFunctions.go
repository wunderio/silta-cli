package common

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func GetImageTagDigest(authenticator remote.Option, imageUrl string, imageTag string) string {

	// Get image digest for each tag
	requestUrl := fmt.Sprintf("%s:%s", imageUrl, imageTag)
	ref, err := name.ParseReference(requestUrl)
	if err != nil {
		return ""
	}
	img, err := remote.Get(ref, authenticator)
	if err != nil {
		return ""
	}

	digest := img.Digest.String()
	return digest
}
