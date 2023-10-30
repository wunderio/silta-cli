package common

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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

// GetChartName reduces chart name to a simple name
func GetChartName(chartName string) string {
	// Charts can be stored stored locally, can't use remote repository name as chart name
	if chartName == "simple" || strings.HasSuffix(chartName, "/simple") {
		return "simple"
	} else if chartName == "frontend" || strings.HasSuffix(chartName, "/frontend") {
		return "frontend"
	} else if chartName == "drupal" || strings.HasSuffix(chartName, "/drupal") {
		return "drupal"
	}
	return ""
}

// Creates a configuration file that can be used for helm release
func CreateChartConfigurationFile(configuration string) string {
	rawConfig, err := base64.StdEncoding.DecodeString(configuration)
	if err != nil {
		log.Fatal("base64 decoding failed for silta configuration overrides file")
	}

	// Write configuration to temporary file
	chartConfigOverrideFile, err := os.CreateTemp("", "silta-config-*")
	if err != nil {
		log.Fatal("failed to create temporary values file")
	}

	if _, err := chartConfigOverrideFile.Write(rawConfig); err != nil {
		log.Fatal("failed to write to temporary values file")
	}

	return chartConfigOverrideFile.Name()
}

// Uses PrependChartConfigOverrides from "SILTA_" + strings.ToUpper(GetChartName(chartName)) + "_CONFIG_VALUES"
// environment variable and prepends it to configuration
func PrependChartConfigOverrides(chartName string, configuration string) string {
	// If chart config override is not empty, decode base64 value and write to temporary file
	chartConfigOverride := os.Getenv("SILTA_" + strings.ToUpper(GetChartName(chartName)) + "_CONFIG_VALUES")

	if len(chartConfigOverride) > 0 {
		chartOverrideFile := CreateChartConfigurationFile(chartConfigOverride)
		defer os.Remove(chartOverrideFile)

		// Prepend override to configuration
		if len(configuration) > 0 {
			configuration = chartOverrideFile + "," + configuration
		} else {
			configuration = chartOverrideFile
		}
	}

	return configuration
}
