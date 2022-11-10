package common

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func ListTags(jwt string, imageName string, imageRepoHost string, imageRepository string) []string {

	requestURL := "https://" + imageRepoHost + "/v2/" + imageRepository + "/" + imageName + "/tags/list"
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
