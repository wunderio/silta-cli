package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// Access scopes:
//
//	Catalog - registry:catalog:* - listing images
//	Image - repository:<image_name>:pull - info on image. <image_name> must include repository name e.x. silta-dev/
type RegistryAccessScope uint8

const (
	Catalog RegistryAccessScope = iota + 1
	Image
)

func GetGCPOAuth2Token() string {
	// gcp_sa_path - path to GCP service account key in json format
	gcp_sa_path, exists := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !exists {
		log.Fatalln("GOOGLE_APPLICATION_CREDENTIALS is not set.")
	}
	// check if oauth2l binary exists
	_, err := exec.LookPath("oauth2l")
	if err != nil {
		log.Fatalln("oauth2l binary is not found in $PATH. Install it first.")
	}

	// get oauth2 token
	command := "oauth2l fetch --credentials " + gcp_sa_path + " --scope cloud-platform.read-only --cache=\"\""
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		log.Fatal("Error: ", err)
	}
	return string(out)
}

// Returns JWT (JSON Web Token) for accessing GCP Container, Artifact registries.
// Depericated due to name. Use GetJWT() instead
func GetGCPJWT(authToken string, imageRepoHost string, scope RegistryAccessScope, gcpProject string, imageName string) string {
	return GetJWT(authToken, imageRepoHost, scope, gcpProject, imageName)
}

// Returns JWT (JSON Web Token) for docker registries.
//
// If 'scope' is set to 'Catalog', 'projectName' and 'imageName' is not used and can be empty strings
func GetJWT(authToken string, imageRepoHost string, scope RegistryAccessScope, projectName string, imageName string) string {
	// <LOCATION.>gcr.io - container registry ,  need url.QueryEscape
	// <LOCATION>-docker.pkg.dev - artifact registry , dont need url.QueryEscape

	const gcr_substr string = "gcr.io" // container registry domain
	const ar_substr string = "pkg.dev" // artifact registry domain

	requestURL := "https://" + imageRepoHost + "/v2/token?service=" + imageRepoHost + "&scope="

	if imageRepoHost == "docker.io" {
		requestURL = "https://auth.docker.io/token?service=registry.docker.io&scope="
	}

	if scope == Catalog {
		requestURL += "registry:catalog:*"
	} else if scope == Image {
		if !(len(imageName) > 0) || !(len(projectName) > 0) {
			log.Fatal("Error: Image and project(repository) names must be set")
		}
		requestURL += "repository:" + projectName + "/" + imageName + ":pull"
	}

	// TODO: REMOVE ME
	fmt.Println("GetJWT Request URL: ", requestURL)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	if strings.Contains(imageRepoHost, gcr_substr) {
		req.SetBasicAuth(url.QueryEscape("_token"), url.QueryEscape(authToken))
	} else if strings.Contains(imageRepoHost, ar_substr) {
		req.SetBasicAuth("_token", authToken)
	} else {
		req.Header.Set("Authorization", "Basic "+string(authToken))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	defer resp.Body.Close()
	response_json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	// Parsing out token from response
	var response_data map[string]interface{}
	err = json.Unmarshal(response_json, &response_data)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	rawToken, ok := response_data["token"]
	if !ok {
		log.Fatal("Error: couldnt parse key 'token'")
	}
	token, ok := rawToken.(string)
	if !ok {
		log.Fatal("Error: couldnt parse out raw token value")
	}
	return string(token)
}

func HasString(a []string, b string) bool {
	for _, c := range a {
		if c == b {
			return true
		}
	}
	return false
}
