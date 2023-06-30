package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func ListImageTagSiblings(authenticator remote.Option, imageUrl string, imageTag string) []string {

	// Docker.io registry API does not return digest for tags so we need to get all tags and then get the digest for each tag.
	// Same for ACR

	// Get all image tags
	ref, err := name.ParseReference(imageUrl)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	img, err := remote.List(ref.Context(), authenticator)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	digests := make(map[string][]string)

	// Get image digest for each tag
	for _, tag := range img {

		requestUrl := fmt.Sprintf("%s:%s", imageUrl, tag)
		ref, err := name.ParseReference(requestUrl)
		if err != nil {
			panic(err)
		}
		img, err := remote.Get(ref, authenticator)
		if err != nil {
			panic(err)
		}

		digest := img.Digest.String()
		digests[digest] = append(digests[digest], tag)
	}

	// Iterate digests and find the one that matches the imageTag
	for _, tags := range digests {
		for _, tag := range tags {
			// Return all sibling tags
			if tag == imageTag {
				return tags
			}
		}
	}

	return nil
}
func ACRListImageTagSiblings(jwt string, imageName string, imageRepoHost string, imageRepository string, imageTag string) []string {

	// Docker.io registry API does not return digest for tags so we need to get all tags and then get the digest for each tag.

	requestURL := "https://" + imageRepoHost + "/v2/" + imageRepository + "/" + imageName + "/tags/list"
	// requestURL := "https://" + imageRepoHost + "/v2/" + imageRepository + "/" + imageName + "/manifests/" + imageTag

	// /acr/v1/{name}/_tags

	fmt.Println("requestURL: ", requestURL)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)

	// Accept header string delimited by comma.

	// Request config.digest section
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	defer resp.Body.Close()

	response_json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	fmt.Println(string(response_json))

	// // Get all image tags
	// ref, err := name.ParseReference(imageUrl)
	// if err != nil {
	// 	log.Fatal("Error: ", err)
	// }
	// img, err := remote.List(ref.Context(), authenticator)
	// if err != nil {
	// 	log.Fatal("Error: ", err)
	// }

	// digests := make(map[string][]string)

	// // Get image digest for each tag
	// for _, tag := range img {

	// 	requestUrl := fmt.Sprintf("%s:%s", imageUrl, tag)
	// 	ref, err := name.ParseReference(requestUrl)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	img, err := remote.Get(ref, authenticator)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	digest := img.Digest.String()
	// 	digests[digest] = append(digests[digest], tag)
	// }

	// // Iterate digests and find the one that matches the imageTag
	// for _, tags := range digests {
	// 	for _, tag := range tags {
	// 		// Return all sibling tags
	// 		if tag == imageTag {
	// 			return tags
	// 		}
	// 	}
	// }

	return nil
}

func GCPListImageTagSiblings(jwt string, imageName string, imageRepoHost string, imageRepository string, imageTag string) []string {

	requestURL := "https://" + imageRepoHost + "/v2/" + imageRepository + "/" + imageName + "/tags/list"

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	defer resp.Body.Close()

	response_json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	fmt.Println(string(response_json))

	// Docker hub does not return manifest section with digest to tag relation, use ListImageTagSiblings instead!

	// Response (Google AR via Registry API)
	// {
	//     "child": [],
	//     "manifest": {
	//       "sha256:062908124bf92cccdf3fd2577dca3b79708b11849809d9cbae24c285dc4b75e3": {
	//         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	//         "tag": [
	//           "98682a1f81cf155fb2df167176b5a767f3b09b20",
	//           "dependabot-composer-drupal-core-recommended-9-5-7"
	//         ],
	//         "timeUploadedMs": "1679889705958",
	//         "timeCreatedMs": "1679889694265",
	//         "imageSizeBytes": "16732378"
	//       },
	//       "sha256:0bbe5d2e2ecb7f77f3d60fd0997cef9b24d02e1352c0e3c05e413da863829a71": {
	//         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	//         "tag": ["dependabot-composer-composer-installers-2-2-0"],
	//         "timeUploadedMs": "1663218694437",
	//         "timeCreatedMs": "1663218680307",
	//         "imageSizeBytes": "16655824"
	//       },
	//       (..)
	//     },
	//     "name": "silta-dev/images/drupal-project-k8s-nginx",
	//     "tags": [
	//         "0f039985775951a2aedff4bfac06ed1a1a6b1b6a",
	//         "10c83fd3065a973816d5472edf32d318f7740294",
	//         "branch--master2",
	//         "branch--mastertest",
	//         "master"
	//     ]
	// }

	// Response (Docker Hub via Registry API).
	// {
	//     "name":"jancis/drupal-project-k8s-nginx",
	//     "tags":[
	//         "branch--master",
	//         "branch--master3",
	//         "e9d31f26463df71363cbce03f070d2fb27c53a2b",
	//         "ef206c033456618ce9105f3d16d9f34e04a2ec4e",
	//         "master",
	//         "master2",
	//         "master3"
	//     ]
	// }

	// Parsing out token from response
	type Manifest struct {
		Tag []string `json:"tag"`
	}
	type TagList struct {
		Child    []string            `json:"child"`
		Manifest map[string]Manifest `json:"manifest"`
	}

	var tagList TagList
	err = json.Unmarshal([]byte(response_json), &tagList)
	if err != nil {
		log.Fatal("Error (json unmarshal): ", err)
	}

	for _, v := range tagList.Manifest {
		for tag := range v.Tag {
			if v.Tag[tag] == imageTag {
				return v.Tag
			}
		}
	}
	return nil
}
