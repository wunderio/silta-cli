package common

import (
	"encoding/json"
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

func GetGoogleOAuth2Token() string {
	// gcp_sa_path - path to GCP service account key in json format
	gcp_sa_path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	log.Println(gcp_sa_path)
	command := "oauth2l fetch --credentials " + gcp_sa_path + " --scope cloud-platform.read-only --cache=\"\""
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		log.Fatal("Error: ", err)
	}
	log.Println(string(out))
	return string(out)
}

// Returns JWT (JSON Web Token) for accessing GCP Container, Artifact registries.
//
// If 'scope' is set to 'Catalog', 'gcpProject' and 'imageName' is not used and can be empty strings
func GetGoogleJWT(oauth2Token string, imageRepoHost string, scope RegistryAccessScope, gcpProject string, imageName string) string {
	// <LOCATION.>gcr.io - container registry ,  need url.QueryEscape
	// <LOCATION>-docker.pkg.dev - artifact registry , dont need url.QueryEscape

	const gcr_substr string = "gcr.io" // container registry domain
	const ar_substr string = "pkg.dev" // artifact registry domain

	requestURL := "https://" + imageRepoHost + "/v2/token?service=" + imageRepoHost + "&scope="
	if scope == Catalog {
		requestURL += "registry:catalog:*"
	} else if scope == Image {
		if !(len(imageName) > 0) || !(len(gcpProject) > 0) {
			log.Fatal("Error: Image and project(repository) names must be set")
		}
		requestURL += "repository:" + gcpProject + "/" + imageName + ":pull"
	}

	//req, err := http.NewRequest("GET", "https://"+imageRepoHost+"/v2/token?service="+imageRepoHost+"&scope=registry:catalog:*", nil)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	if strings.Contains(imageRepoHost, gcr_substr) {
		req.SetBasicAuth(url.QueryEscape("_token"), url.QueryEscape(oauth2Token))
	} else if strings.Contains(imageRepoHost, ar_substr) {
		req.SetBasicAuth("_token", oauth2Token)
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
