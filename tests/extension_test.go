package cmd_test

import (
	"io/ioutil"
	"log"
	"os"
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
	CliExecTest(t, command, environment, testString, false)

	schema, err := ioutil.ReadFile(common.ExtendedFolder + "/frontend/values.schema.json")
	chart, err1 := ioutil.ReadFile(common.ExtendedFolder + "/frontend/Chart.yaml")

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if err1 != nil {
		log.Println(err1)
		t.FailNow()
	}

	schemaStr := string(schema)
	chartStr := string(chart)
	if strings.Contains(schemaStr, "redis") == false {
		log.Println("Redis not present in values.schema.json")
		t.Fail()
	}
	if strings.Contains(chartStr, "redis") == false {
		log.Println("Redis not present in Chart.yaml")
		t.Fail()
	}

	// Cleanup
	cliBinaryName = originalCli
	os.RemoveAll(common.ExtendedFolder)

	// Change dir back to previous
	os.Chdir(wd)

}
