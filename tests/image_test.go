package cmd_test

import (
	"os"
	"testing"
)

func TestImageLoginCmd(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	// Test env
	command := "ci image login --debug"
	environment := []string{
		"IMAGE_REPO_HOST=foo.bar",
		"GCLOUD_KEY_JSON=baz",
	}
	testString := `echo "baz" | docker login --username "_json_key" --password-stdin https://foo.bar`
	CliExecTest(t, command, environment, testString, false)

	// Test all env
	command = "ci image login --debug"
	environment = []string{
		"IMAGE_REPO_HOST=foo.bar",
		"IMAGE_REPO_USER=111",
		"IMAGE_REPO_PASS=222",
		"GCLOUD_KEY_JSON=333",
		"AWS_SECRET_ACCESS_KEY=444",
		"AWS_REGION=555",
		"AKS_TENANT_ID=666",
		"AKS_SP_APP_ID=777",
		"AKS_SP_PASSWORD=888",
	}
	testString = `IMAGE_REPO_HOST: foo.bar
IMAGE_REPO_TLS: true
IMAGE_REPO_USER: 111
IMAGE_REPO_PASS: 222
GCLOUD_KEY_JSON: 333
AWS_SECRET_ACCESS_KEY: 444
AWS_REGION: 555
AKS_TENANT_ID: 666
AKS_SP_APP_ID: 777
AKS_SP_PASSWORD: 888
Command (not executed): echo "222" | docker login --username "111" --password-stdin https://foo.bar`

	CliExecTest(t, command, environment, testString, false)

	// Test undefined ENV
	command = "ci image login --debug"
	environment = []string{}
	testString = `Docker registry credentials are empty`
	CliExecTest(t, command, environment, testString, false)

	// Test args
	command = "ci image login --image-repo-host foo.bar --gcp-key-json baz --debug"
	environment = []string{}
	testString = `echo "baz" | docker login --username "_json_key" --password-stdin https://foo.bar`
	CliExecTest(t, command, environment, testString, false)

	// Test all args
	command = `ci image login \
			--image-repo-host foo.bar \
			--image-repo-user 111 \
			--image-repo-pass 222 \
			--gcp-key-json 333 \
			--aws-secret-access-key 444 \
			--aws-region 555 \
			--aks-tenant-id 666 \
			--aks-sp-app-id 777 \
			--aks-sp-password 888 \
			--debug`

	environment = []string{}
	testString = `IMAGE_REPO_HOST: foo.bar
IMAGE_REPO_TLS: true
IMAGE_REPO_USER: 111
IMAGE_REPO_PASS: 222
GCLOUD_KEY_JSON: 333
AWS_SECRET_ACCESS_KEY: 444
AWS_REGION: 555
AKS_TENANT_ID: 666
AKS_SP_APP_ID: 777
AKS_SP_PASSWORD: 888
Command (not executed): echo "222" | docker login --username "111" --password-stdin https://foo.bar`
	CliExecTest(t, command, environment, testString, false)

	// Test args+env merge
	command = "ci image login --image-repo-host foo.bar --debug"
	environment = []string{
		"IMAGE_REPO_HOST=bar.bar",
		"GCLOUD_KEY_JSON=baz",
	}
	testString = `echo "baz" | docker login --username "_json_key" --password-stdin https://foo.bar`
	CliExecTest(t, command, environment, testString, false)

	// Change dir back to previous
	os.Chdir(wd)
}

func TestImageUrlCmd(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	// Incomplete flags test
	command := "ci image url"
	environment := []string{}
	testString := `Error: required flag(s)`
	CliExecTest(t, command, environment, testString, false)

	// image-tag flag test
	command = "ci image url --image-repo-host 'foo.bar' --image-repo-project 'silta' --namespace 'baz' --image-identifier 'nginx' --dockerfile 'tests/nginx.Dockerfile' --image-tag=qux"
	environment = []string{}
	testString = `foo.bar/silta/baz-nginx:qux`
	CliExecTest(t, command, environment, testString, true)

	// Change dir back to previous
	os.Chdir(wd)
}
func TestImageBuildCmd(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	// Incomplete flags test
	command := "ci image build"
	environment := []string{}
	testString := `Error: required flag(s)`
	CliExecTest(t, command, environment, testString, false)

	// image-tag flag test
	command = "ci image build --image-repo-host 'foo.bar' --image-repo-project 'silta' --namespace 'baz' --image-identifier 'nginx' --dockerfile 'tests/nginx.Dockerfile' --image-tag=qux --debug"
	environment = []string{}
	testString = `docker push 'foo.bar/silta/baz-nginx:qux'`
	CliExecTest(t, command, environment, testString, false)

	// Change dir back to previous
	os.Chdir(wd)
}
