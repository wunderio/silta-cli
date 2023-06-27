package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// TODO: REMOVE ME
func ListImageTags(jwt string, imageName string, imageRepoHost string, imageRepository string) []string {

	requestURL := "https://" + imageRepoHost + "/v2/" + imageRepository + "/" + imageName + "/tags/list"

	if imageRepoHost == "docker.io" {
		requestURL = "https://registry-1.docker.io/v2/" + imageRepository + "/" + imageName + "/tags/list"
	}

	// TODO: REMOVE ME
	fmt.Println("ListImageTags Request URL: ", requestURL)

	//req, err := http.NewRequest("GET", "https://gcr.io/v2/<your-project>/alpine/tags/list", nil)
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

	// TODO: REMOVE ME
	fmt.Println("ListImageTags Response: ", string(response_json))

	// Parsing out token from response
	var j interface{}
	err = json.Unmarshal(response_json, &j)
	m := j.(map[string]interface{})

	if err != nil {
		log.Fatal("Error: ", err)
	}

	tagsArr := m["tags"].([]interface{})
	tags := make([]string, len(tagsArr))
	for i, v := range tagsArr {
		tags[i] = v.(string)
	}
	return tags
}

func ListImageTagSiblings(jwt string, imageName string, imageRepoHost string, imageRepository string, imageTag string) []string {

	if imageRepoHost == "docker.io" {
		imageRepoHost = "registry-1.docker.io"
	}

	// requestURL := "https://" + imageRepoHost + "/v2/" + imageRepository + "/" + imageName + "/tags/list"

	requestURL := "https://" + imageRepoHost + "/v2/" + imageRepository + "/" + imageName + "/manifests/" + imageTag

	// TODO: REMOVE ME
	fmt.Println("ListImageTags Request URL: ", requestURL)

	//req, err := http.NewRequest("GET", "https://gcr.io/v2/<your-project>/alpine/tags/list", nil)
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

	fmt.Printf("ListImageTagSiblings Response: %s\n", string(response_json))

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

	// fmt.Printf("ListImageTagSiblings Response: %s\n", string(response_json))

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
