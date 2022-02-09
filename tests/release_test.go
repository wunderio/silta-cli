package cmd_test

import (
	"os"
	"testing"
)

func TestReleaseNameCmd(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	// Basic name test
	command := "ci release name --branchname Foo --debug"
	environment := []string{}
	testString := `foo`
	CliExecTest(t, command, environment, testString, true)

	// Basic name + suffix test
	command = "ci release name --branchname Foo --release-suffix bar --debug"
	environment = []string{}
	testString = `foo-bar`
	CliExecTest(t, command, environment, testString, true)

	// Alphanumeric name test
	command = "ci release name --branchname Te_3/s^T --release-suffix bar --debug"
	environment = []string{}
	testString = `te-3-s-t-bar`
	CliExecTest(t, command, environment, testString, true)

	// 39 char long release name test
	command = "ci release name --branchname 111111111122222222223333333333444444444 --debug"
	environment = []string{}
	testString = `111111111122222222223333333333444444444`
	CliExecTest(t, command, environment, testString, true)

	// 40 char long release name test
	command = "ci release name --branchname 1111111111222222222233333333334444444444 --debug"
	environment = []string{}
	testString = `1111111111222222222233333333334444444444`
	CliExecTest(t, command, environment, testString, true)

	// 41 char long release name test
	command = "ci release name --branchname 11111111112222222222333333333344444444445 --release-suffix 123456789012345 --debug"
	environment = []string{}
	testString = `1111111111222222222233-d0c4-1234567-e27a`
	CliExecTest(t, command, environment, testString, true)

	// 41 char release name + 15 char suffix test
	command = "ci release name --branchname 11111111112222222222333333333344444444445 --release-suffix 123456789012345 --debug"
	environment = []string{}
	testString = `1111111111222222222233-d0c4-1234567-e27a`
	CliExecTest(t, command, environment, testString, true)

	// Change dir back to previous
	os.Chdir(wd)
}

func TestReleaseEnvironmentnameCmd(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	// Basic name test
	command := "ci release environmentname --branchname Foo --debug"
	environment := []string{}
	testString := `foo`
	CliExecTest(t, command, environment, testString, true)

	// Basic name + suffix test
	command = "ci release environmentname --branchname Foo --release-suffix bar --debug"
	environment = []string{}
	testString = `foo-bar`
	CliExecTest(t, command, environment, testString, true)

	// Alphanumeric name test
	command = "ci release environmentname --branchname Te_3/s^T --release-suffix bar --debug"
	environment = []string{}
	testString = `te_3/s^t-bar`
	CliExecTest(t, command, environment, testString, true)

	// 39 char long release name test
	command = "ci release environmentname --branchname 111111111122222222223333333333444444444 --debug"
	environment = []string{}
	testString = `111111111122222222223333333333444444444`
	CliExecTest(t, command, environment, testString, true)

	// 40 char long release name test
	command = "ci release environmentname --branchname 1111111111222222222233333333334444444444 --debug"
	environment = []string{}
	testString = `1111111111222222222233333333334444444444`
	CliExecTest(t, command, environment, testString, true)

	// 41 char long release name test
	command = "ci release environmentname --branchname 11111111112222222222333333333344444444445 --release-suffix 123456789012345 --debug"
	environment = []string{}
	testString = `1111111111222222222233-d0c4-1234567-e27a`
	CliExecTest(t, command, environment, testString, true)

	// 41 char release name + 15 char suffix test
	command = "ci release environmentname --branchname 11111111112222222222333333333344444444445 --release-suffix 123456789012345 --debug"
	environment = []string{}
	testString = `1111111111222222222233-d0c4-1234567-e27a`
	CliExecTest(t, command, environment, testString, true)

	// Change dir back to previous
	os.Chdir(wd)
}
func TestReleaseDeployCmd(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	// Basic name test
	command := "ci release deploy"
	environment := []string{}
	testString := `Error: required flag(s)`
	CliExecTest(t, command, environment, testString, false)

	// Test args
	command = `ci release deploy \
		--namespace default \
		--release-name 'test' \
		--chart-name drupal \
		--php-image-url php-image \
		--nginx-image-url nginx-image \
		--shell-image-url shell-image \
		--debug`

	environment = []string{}
	testString = `helm upgrade --install 'test' 'drupal' \
				--repo 'https://storage.googleapis.com/charts.wdr.io' \
				 \
				--cleanup-on-fail \
				--set environmentName='' \
				--set silta-release.branchName='' \
				--set php.image='php-image' \
				--set nginx.image='nginx-image' \
				--set shell.image='shell-image' \
				--set shell.gitAuth.repositoryUrl='' \
				--set shell.gitAuth.keyserver.username='' \
				--set shell.gitAuth.keyserver.password='' \
				--set clusterDomain='' \
				 \
				 \
				 \
				 \
				 \
				--set referenceData.skipMount=true \
				--namespace='default' \
				--values '' \
				--timeout '15m' &> helm-output.log & pid=$!`
	CliExecTest(t, command, environment, testString, false)

	// Test all args
	command = `ci release deploy \
		--release-name 1 \
		--release-suffix 2 \
		--chart-name drupal \
		--chart-repository 3 \
		--chart-version 4 \
		--silta-environment-name 5 \
		--branchname 6 \
		--php-image-url 7 \
		--nginx-image-url 8 \
		--shell-image-url 9 \
		--repository-url 10 \
		--gitauth-username 11 \
		--gitauth-password 12 \
		--cluster-domain 13 \
		--vpn-ip 14 \
		--vpc-native 15 \
		--cluster-type 16 \
		--db-root-pass 17 \
		--db-user-pass 18 \
		--namespace 19 \
		--silta-config 20 \
		--deployment-timeout 21 \
		--debug`
	environment = []string{}
	testString = `helm upgrade --install '1' 'drupal' \
				--repo '3' \
				--version '4' \
				--cleanup-on-fail \
				--set environmentName='5' \
				--set silta-release.branchName='6' \
				--set php.image='7' \
				--set nginx.image='8' \
				--set shell.image='9' \
				--set shell.gitAuth.repositoryUrl='10' \
				--set shell.gitAuth.keyserver.username='11' \
				--set shell.gitAuth.keyserver.password='12' \
				--set clusterDomain='13' \
				--set nginx.noauthips.vpn='14/32' \
				--set cluster.vpcNative='15' \
				--set cluster.type='16' \
				--set mariadb.rootUser.password='17' \
				--set mariadb.db.password='18' \
				--set referenceData.skipMount=true \
				--namespace='19' \
				--values '20' \
				--timeout '21' &> helm-output.log & pid=$!`
	CliExecTest(t, command, environment, testString, false)

	// Change dir back to previous
	os.Chdir(wd)
}
