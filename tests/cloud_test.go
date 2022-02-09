package cmd_test

import (
	"os"
	"testing"
)

func TestCloudLoginCmd(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	// Custom kubeconfig
	command := "cloud login --kubeconfig `echo \"TEST\" | base64` --kubeconfigpath tmpkubeconfig --debug; cat tmpkubeconfig; rm tmpkubeconfig"
	environment := []string{}
	testString := "TEST\n"
	CliExecTest(t, command, environment, testString, true)

	// Custom kubeconfig as env variable
	command = "cloud login --kubeconfigpath tmpkubeconfig --debug; cat tmpkubeconfig; rm tmpkubeconfig"
	environment = []string{
		// echo "TEST2" | base64 => "VEVTVDIK"
		"KUBECTL_CONFIG=VEVTVDIK",
	}
	testString = "TEST2\n"
	CliExecTest(t, command, environment, testString, true)

	// TODO: test gcp, aws and aks
	t.Log("TODO: test gcp, aws and aks")

	// Change dir back to previous
	os.Chdir(wd)
}
