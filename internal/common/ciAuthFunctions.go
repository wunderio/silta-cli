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

// Get authentication token from docker config.json
func GetDockerAuth(imageRepoHost string) (map[string]string, error) {

	// Load docker config file
	dockerConfigFile := fmt.Sprintf("%s/.docker/config.json", os.Getenv("HOME"))

	dockerConfigFileContent, err := ioutil.ReadFile(dockerConfigFile)
	if err != nil {
		return nil, fmt.Errorf("error (ioutil.ReadFile): %w", err)
	}

	type DockerConfigFile struct {
		Auths map[string]map[string]string `json:"auths"`
	}

	var dockerConfigFileJson DockerConfigFile

	err = json.Unmarshal(dockerConfigFileContent, &dockerConfigFileJson)
	if err != nil {
		return nil, fmt.Errorf("error (json.Unmarshal on docker config.json): %w", err)
	}

	// Get authentification credentials for the specified image repository host
	for key, value := range dockerConfigFileJson.Auths {
		// Fall back to docker.io if no authentification credentials are found for the specified image repository host
		if strings.HasPrefix(imageRepoHost, key) || (imageRepoHost == "docker.io" && strings.HasPrefix(key, "https://index.docker.io/")) {
			// auth = value.(map[string]interface{})["auth"].(string)
			return value, nil
		}
	}

	return nil, fmt.Errorf("error: No authentification credentials found for image repository host %w", imageRepoHost)
}

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
func GetACRJWT(authToken string, imageRepoHost string, scope RegistryAccessScope, projectName string, imageName string) string {

	requestURL := "https://" + imageRepoHost + "/oauth2/token"

	type Data struct {
		Scope        string `json:"scope"`
		Service      string `json:"service"`
		GrantType    string `json:"grant_type"`
		RefreshToken string `json:"refresh_token"`
	}

	data := Data{
		Service:      imageRepoHost,
		GrantType:    "refresh_token",
		RefreshToken: authToken,
	}

	if scope == Catalog {
		data.Scope = "registry:catalog:*"
	} else if scope == Image {
		if !(len(imageName) > 0) || !(len(projectName) > 0) {
			log.Fatal("Error: Image and project(repository) names must be set")
		}
		data.Scope = "repository:" + projectName + "/" + imageName + ":pull"
	}

	// reqData, _ := json.Marshal(data)
	// Convert bytes to string.
	// s := string(b)

	// JSON body
	// reqData := []byte(`{
	// 	"scope": "repository:silta-images/drupal-project-k8s-nginx:pull",
	// 	"service": "siltaimageregistry.azurecr.io",
	// 	"grant_type": "refresh_token",
	// 	"refresh_token": "` + authToken + `",
	// }`)

	// ------------------

	// apiUrl := "https://api.com"
	// resource := "/user/"
	reqData := url.Values{}
	reqData.Set("scope", "repository:silta-images/drupal-project-k8s-nginx:pull")
	reqData.Set("service", "siltaimageregistry.azurecr.io")
	reqData.Set("grant_type", "refresh_token")
	reqData.Set("refresh_token", authToken)

	// u, _ := url.ParseRequestURI(requestURL)
	// u.Path = resource
	// urlStr := u.String() // "https://api.com/user/"

	// client := &http.Client{}
	// r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	// r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	// r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// resp, _ := client.Do(r)

	// ---------------

	fmt.Println(data)

	// // Create a HTTP post request
	// r, err := http.NewRequest("POST", posturl, bytes.NewBuffer(body))

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(reqData.Encode()))
	// req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(reqData))
	if err != nil {
		log.Fatalln("Request error: ", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// if strings.Contains(imageRepoHost, gcr_substr) {
	// 	req.SetBasicAuth(url.QueryEscape("_token"), url.QueryEscape(authToken))
	// } else if strings.Contains(imageRepoHost, ar_substr) {
	// 	req.SetBasicAuth("_token", authToken)
	// } else {
	// 	req.Header.Set("Authorization", "Basic "+string(authToken))
	// }

	// TODO: REMOVE ME
	fmt.Println(requestURL)

	client := &http.Client{}
	resp, err := client.Do(req)
	// resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	defer resp.Body.Close()
	response_json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	// TODO: REMOVE ME
	fmt.Printf("Response: %s\n", response_json)

	// Parsing out token from response
	var response_data map[string]interface{}
	err = json.Unmarshal(response_json, &response_data)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	rawToken, ok := response_data["access_token"]
	if !ok {
		log.Fatal("Error: couldnt parse key 'token'")
	}
	token, ok := rawToken.(string)
	if !ok {
		log.Fatal("Error: couldnt parse out raw token value")
	}
	return string(token)
}

// Returns JWT (JSON Web Token) for docker registries.
//
// If 'scope' is set to 'Catalog', 'projectName' and 'imageName' is not used and can be empty strings
func GetJWT(authToken string, imageRepoHost string, scope RegistryAccessScope, projectName string, imageName string) string {
	// <LOCATION.>gcr.io - container registry ,  need url.QueryEscape
	// <LOCATION>-docker.pkg.dev - artifact registry , dont need url.QueryEscape

	const gcr_substr string = "gcr.io" // container registry domain
	const ar_substr string = "pkg.dev" // artifact registry domain

	requestURL := "https://" + imageRepoHost + "/v2/token?service=" + imageRepoHost

	if imageRepoHost == "docker.io" {
		requestURL = "https://auth.docker.io/token?service=registry.docker.io"
	}

	if scope == Catalog {
		requestURL += "&scope=registry:catalog:*"
	} else if scope == Image {
		if !(len(imageName) > 0) || !(len(projectName) > 0) {
			log.Fatal("Error: Image and project(repository) names must be set")
		}
		requestURL += "&scope=repository:" + projectName + "/" + imageName + ":pull"
	}

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
