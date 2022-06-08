package common

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/k0kubun/pp/v3"
	"gopkg.in/yaml.v2"
)

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

func DownloadUntarChart(chartName *ChartNameVersion) {

	command := ""
	if len(chartName.Version) < 1 {
		command = "helm pull " + chartName.Name + " --untar"
	} else {
		command = "helm pull " + chartName.Name + " --version " + chartName.Version + " --untar"
	}
	helmCmd, err := exec.Command("bash", "-c", command).CombinedOutput()

	if err != nil {
		log.SetFlags(0) //remove timestamp
		log.Fatal(string(helmCmd[:]))
	}

}

// func AppendToChartSchemaFile1(schemaFile string, chartName string) {
// 	file, err := ioutil.ReadFile(schemaFile)
// 	log.Println(schemaFile)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	var data ChartSchema
// 	json.Unmarshal(file, &data)
// 	log.Println(&data)
// }

func AppendToChartSchemaFile1(schemaFile string, chartName string) {
	file, err := ioutil.ReadFile(schemaFile)
	log.Println(schemaFile)
	if err != nil {
		log.Println(err)
	}
	var data interface{}
	json.Unmarshal(file, &data)
	arr := data.(map[string]interface{})

	pp.Println(arr)

	data_out, _ := json.Marshal(arr)
	ioutil.WriteFile(schemaFile+"1", data_out, 0644)
}

func AppendToChartSchemaFile(schemaFile string, chartName string) {
	file, err := ioutil.ReadFile(schemaFile)
	log.Println(schemaFile)
	if err != nil {
		log.Println(err)
	}

	str := string(file)
	str = strings.Replace(str, "\"projectName\"", "\"projectName\" : { \"type\": \"string\" }, \""+chartName+`"`, 1)

	ioutil.WriteFile(schemaFile+"1", []byte(str), 0644)

}
