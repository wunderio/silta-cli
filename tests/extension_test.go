package cmd_test

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/wunderio/silta-cli/internal/common"
)

func TestChartExtensionCmd(t *testing.T) {

	var originalCli = cliBinaryName
	cliBinaryName = "../../../silta"

	// Go to main directory
	wd, _ := os.Getwd()

	os.Chdir("assets/extension_test")
	command := "chart extend --chart-name wunderio/frontend --subchart-list-file deployment_options.yml"
	environment := []string{}
	testString := ""
	CliExecTest1(t, command, environment, testString, false)

	log.Print(os.Getwd())
	helmCmd, _ := exec.Command("bash", "-c", "ls").CombinedOutput()
	log.Println(string(helmCmd))
	helmCmd1, _ := exec.Command("bash", "-c", "ls extended-helm-chart").CombinedOutput()
	log.Println(string(helmCmd1))
	schema, err := ioutil.ReadFile("./" + common.ExtendedFolder + "/frontend/values.schema.json")
	chart, err1 := ioutil.ReadFile("./" + common.ExtendedFolder + "/frontend/Chart.yaml")

	cliBinaryName = originalCli

	if err != nil {
		log.Println(err)
		t.Fail()
	}
	if err1 != nil {
		log.Println(err1)
		t.Fail()
	}

	schemaStr := string(schema)
	chartStr := string(chart)
	if strings.Contains(schemaStr, "redis") == false && err == nil {
		log.Println("Redis not present in values.schema.json")
		t.Fail()
	}
	if strings.Contains(chartStr, "redis") == false && err1 == nil {
		log.Println("Redis not present in Chart.yaml")
		t.Fail()
	}

	// Cleanup
	os.RemoveAll(common.ExtendedFolder)

	// Change dir back to previous
	os.Chdir(wd)

}
