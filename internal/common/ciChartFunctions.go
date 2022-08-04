package common

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

const ExtendedFolder = "extended-helm-chart"

func ReadCharts(deploymentsFile string) chartList {

	ecfile, err := ioutil.ReadFile(deploymentsFile)
	if err != nil {
		log.Fatal(err)
	}

	var data chartList
	err2 := yaml.Unmarshal(ecfile, &data)

	if err2 != nil {
		log.Fatal(err2)
	}
	return data
}

func ReadChartDefinition(chartFile string) chartDefinition {

	ecfile, err := ioutil.ReadFile(chartFile)
	if err != nil {
		log.Fatal(err)
	}

	var data chartDefinition
	err = yaml.Unmarshal(ecfile, &data)

	if err != nil {
		log.Fatal(err)
	}

	return data
}

func AppendExtraCharts(charts *chartList, mainchart *chartDefinition) {
	if len(charts.Charts) < 1 {
		return
	}

	for _, dependency := range charts.Charts {
		if strings.HasPrefix(dependency.Repository, "https://") {
			command := "helm repo add " + dependency.Name + " " + dependency.Repository
			helmCmd, _ := exec.Command("bash", "-c", command).CombinedOutput()
			log.Print(string(helmCmd[:]))

		}
	}
	mainchart.Dependencies = append(mainchart.Dependencies, charts.Charts...)

	const searchStr = "file://.."
	for i, dependency := range mainchart.Dependencies {
		if strings.HasPrefix(dependency.Repository, "file://..") {
			var finalStr = searchStr + "/" + mainchart.Name + "/charts"
			mainchart.Dependencies[i].Repository = strings.Replace(dependency.Repository, searchStr, finalStr, 1)
		}
	}
}

func WriteChartDefinition(mainchart chartDefinition, ymlfile string) {

	data, err := yaml.Marshal(&mainchart)
	if err != nil {
		log.Fatal(err)
	}

	err2 := ioutil.WriteFile(ymlfile, data, 0644)
	if err2 != nil {
		log.Fatal(err)
	}
}

func DownloadUntarChart(chartName *ChartNameVersion, toExtendedFolder bool) {

	command := ""
	destinationArg := " "

	if toExtendedFolder == true {
		destinationArg = " -d " + ExtendedFolder + " "
	}

	if len(chartName.Version) < 1 {
		command = "helm pull " + chartName.Name + " --untar" + destinationArg
	} else {
		command = "helm pull " + chartName.Name + " --version " + chartName.Version + " --untar" + destinationArg
	}
	helmCmd, err := exec.Command("bash", "-c", command).CombinedOutput()

	if err != nil {
		log.SetFlags(0) //remove timestamp
		log.Fatal(string(helmCmd[:]))
	}

}

func AppendToChartSchemaFile(schemaFile string, chartNames []string) {
	file, err := ioutil.ReadFile(schemaFile)
	log.Println(schemaFile)
	if err != nil {
		log.Println(err)
	}

	var j interface{}
	err = json.Unmarshal(file, &j)
	m := j.(map[string]interface{}) //mapped schema json

	var propertiesArray = m["properties"].(map[string]interface{})
	for _, v := range chartNames {
		propertiesArray[v] = map[string]interface{}{"type": "object"}
	}
	m["properties"] = propertiesArray

	out, _ := json.Marshal(m)
	ioutil.WriteFile(schemaFile, out, 0644)
}

func GetChartNamesFromDependencies(dependencies []dependency) []string {
	names := []string{}

	for _, v := range dependencies {
		names = append(names, v.Name)
	}

	return names
}
