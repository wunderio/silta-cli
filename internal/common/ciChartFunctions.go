package common

import (
	"io/ioutil"
	"log"
	"os/exec"

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
	err2 := yaml.Unmarshal(ecfile, &data)

	if err2 != nil {
		log.Fatal(err2)
	}

	return data
}

func AppendExtraCharts(charts *chartList, mainchart *chartDefinition) {

	if len(charts.Charts) < 1 {
		return
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
