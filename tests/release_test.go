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

	// 40 char long release name test.
	command = "ci release name --branchname 1111111111222222222233333333334444444444 --debug"
	environment = []string{}
	testString = `1111111111222222222233333333334444-808c`
	CliExecTest(t, command, environment, testString, true)

	// 41 char long release name test
	command = "ci release name --branchname 11111111112222222222333333333344444444445 --release-suffix 123456789012345 --debug"
	environment = []string{}
	testString = `1111111111222222222233-d0c4-1234567-e27a`
	CliExecTest(t, command, environment, testString, true)

	// 50 char long release name test
	command = "ci release environmentname --branchname 11111111112222222222333333333344444444445555555555 --debug"
	environment = []string{}
	testString = `1111111111222222222233333333334444-81f3`
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
	testString = `1111111111222222222233333333334444-808c`
	CliExecTest(t, command, environment, testString, true)

	// 50 char long release name test
	command = "ci release environmentname --branchname 11111111112222222222333333333344444444445555555555 --debug"
	environment = []string{}
	testString = `1111111111222222222233333333334444-81f3`
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
	testString = `
			RELEASE_NAME='test'
			CHART_NAME='drupal'
			CHART_REPOSITORY='https://storage.googleapis.com/charts.wdr.io'
			EXTRA_CHART_VERSION=''
			SILTA_ENVIRONMENT_NAME=''
			BRANCHNAME=''
			PHP_IMAGE_URL='php-image'
			NGINX_IMAGE_URL='nginx-image'
			SHELL_IMAGE_URL='shell-image'
			GIT_REPOSITORY_URL=''
			GITAUTH_USERNAME=''
			GITAUTH_PASSWORD=''
			CLUSTER_DOMAIN=''
			EXTRA_NOAUTHIPS=''
			EXTRA_VPCNATIVE=''
			EXTRA_CLUSTERTYPE=''
			EXTRA_DB_ROOT_PASS=''
			EXTRA_DB_USER_PASS=''
			EXTRA_REFERENCE_DATA=''
			NAMESPACE='default'
			SILTA_CONFIG=''
			EXTRA_HELM_FLAGS=''
			DEPLOYMENT_TIMEOUT='15m'
			DEPLOYMENT_TIMEOUT_SECONDS='900'

			# Detect pods in FAILED state
			function show_failing_pods() {
				echo ""
				failed_pods=$(kubectl get pod -l "release=$RELEASE_NAME,cronjob!=true" -n "$NAMESPACE" -o custom-columns="POD:metadata.name,STATE:status.containerStatuses[*].ready" --no-headers | grep -E "<none>|false" | grep -Eo '^[^ ]+')
				if [[ ! -z "$failed_pods" ]] ; then
					echo "Failing pods:"
					while IFS= read -r pod; do
						echo "---- ${NAMESPACE} / ${pod} ----"
						echo "* Events"
						kubectl get events --field-selector involvedObject.name=${pod},type!=Normal --show-kind=true --ignore-not-found=true --namespace ${NAMESPACE}
						echo ""
						echo "* Logs"
						containers=$(kubectl get pods "${pod}" --namespace "${NAMESPACE}" -o json | jq -r 'try .status | .containerStatuses[] | select(.ready == false).name')
						if [[ ! -z "$containers" ]] ; then
							for container in ${containers}; do
								kubectl logs "${pod}" --prefix=true --since="${DEPLOYMENT_TIMEOUT}" --namespace "${NAMESPACE}" -c "${container}"
							done
						else
							echo "no logs found"
						fi

						echo "----"
					done <<< "$failed_pods"

					false
				else
					true
				fi
				rm -f helm-output.log
			}

			trap show_failing_pods ERR

			helm upgrade --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
				--cleanup-on-fail \
				--set environmentName="${SILTA_ENVIRONMENT_NAME}" \
				--set silta-release.branchName="${BRANCHNAME}" \
				--set php.image="${PHP_IMAGE_URL}" \
				--set nginx.image="${NGINX_IMAGE_URL}" \
				--set shell.image="${SHELL_IMAGE_URL}" \
				--set shell.gitAuth.repositoryUrl="${GIT_REPOSITORY_URL}" \
				--set shell.gitAuth.keyserver.username="${GITAUTH_USERNAME}" \
				--set shell.gitAuth.keyserver.password="${GITAUTH_PASSWORD}" \
				--set clusterDomain="${CLUSTER_DOMAIN}" \
				${EXTRA_NOAUTHIPS} \
				${EXTRA_VPCNATIVE} \
				${EXTRA_CLUSTERTYPE} \
				${EXTRA_DB_ROOT_PASS} \
				${EXTRA_DB_USER_PASS} \
				${EXTRA_REFERENCE_DATA} \
				--namespace="${NAMESPACE}" \
				--values "${SILTA_CONFIG}" \
				${EXTRA_HELM_FLAGS} \
				--wait \
				--timeout "${DEPLOYMENT_TIMEOUT}" &> helm-output.log & pid=$!`
	CliExecTest(t, command, environment, testString, false)

	// Test all args (drupal chart)
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
		--helm-flags 21 \
		--deployment-timeout 22m \
		--debug`
	environment = []string{}
	testString = `
			RELEASE_NAME='1'
			CHART_NAME='drupal'
			CHART_REPOSITORY='3'
			EXTRA_CHART_VERSION='--version '4''
			SILTA_ENVIRONMENT_NAME='5'
			BRANCHNAME='6'
			PHP_IMAGE_URL='7'
			NGINX_IMAGE_URL='8'
			SHELL_IMAGE_URL='9'
			GIT_REPOSITORY_URL='10'
			GITAUTH_USERNAME='11'
			GITAUTH_PASSWORD='12'
			CLUSTER_DOMAIN='13'
			EXTRA_NOAUTHIPS='--set nginx.noauthips.vpn='14/32''
			EXTRA_VPCNATIVE='--set cluster.vpcNative='15''
			EXTRA_CLUSTERTYPE='--set cluster.type='16''
			EXTRA_DB_ROOT_PASS='--set mariadb.rootUser.password='17''
			EXTRA_DB_USER_PASS='--set mariadb.db.password='18''
			EXTRA_REFERENCE_DATA=''
			NAMESPACE='19'
			SILTA_CONFIG='20'
			EXTRA_HELM_FLAGS='21'
			DEPLOYMENT_TIMEOUT='22m'
			DEPLOYMENT_TIMEOUT_SECONDS='1320'

			# Detect pods in FAILED state
			function show_failing_pods() {
				echo ""
				failed_pods=$(kubectl get pod -l "release=$RELEASE_NAME,cronjob!=true" -n "$NAMESPACE" -o custom-columns="POD:metadata.name,STATE:status.containerStatuses[*].ready" --no-headers | grep -E "<none>|false" | grep -Eo '^[^ ]+')
				if [[ ! -z "$failed_pods" ]] ; then
					echo "Failing pods:"
					while IFS= read -r pod; do
						echo "---- ${NAMESPACE} / ${pod} ----"
						echo "* Events"
						kubectl get events --field-selector involvedObject.name=${pod},type!=Normal --show-kind=true --ignore-not-found=true --namespace ${NAMESPACE}
						echo ""
						echo "* Logs"
						containers=$(kubectl get pods "${pod}" --namespace "${NAMESPACE}" -o json | jq -r 'try .status | .containerStatuses[] | select(.ready == false).name')
						if [[ ! -z "$containers" ]] ; then
							for container in ${containers}; do
								kubectl logs "${pod}" --prefix=true --since="${DEPLOYMENT_TIMEOUT}" --namespace "${NAMESPACE}" -c "${container}"
							done
						else
							echo "no logs found"
						fi

						echo "----"
					done <<< "$failed_pods"

					false
				else
					true
				fi
				rm -f helm-output.log
			}

			trap show_failing_pods ERR

			helm upgrade --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
				--cleanup-on-fail \
				--set environmentName="${SILTA_ENVIRONMENT_NAME}" \
				--set silta-release.branchName="${BRANCHNAME}" \
				--set php.image="${PHP_IMAGE_URL}" \
				--set nginx.image="${NGINX_IMAGE_URL}" \
				--set shell.image="${SHELL_IMAGE_URL}" \
				--set shell.gitAuth.repositoryUrl="${GIT_REPOSITORY_URL}" \
				--set shell.gitAuth.keyserver.username="${GITAUTH_USERNAME}" \
				--set shell.gitAuth.keyserver.password="${GITAUTH_PASSWORD}" \
				--set clusterDomain="${CLUSTER_DOMAIN}" \
				${EXTRA_NOAUTHIPS} \
				${EXTRA_VPCNATIVE} \
				${EXTRA_CLUSTERTYPE} \
				${EXTRA_DB_ROOT_PASS} \
				${EXTRA_DB_USER_PASS} \
				${EXTRA_REFERENCE_DATA} \
				--namespace="${NAMESPACE}" \
				--values "${SILTA_CONFIG}" \
				${EXTRA_HELM_FLAGS} \
				--wait \
				--timeout "${DEPLOYMENT_TIMEOUT}" &> helm-output.log & pid=$!`
	CliExecTest(t, command, environment, testString, false)

	// Test all args (simple chart)
	command = `ci release deploy \
		--release-name 1 \
		--release-suffix 2 \
		--chart-name simple \
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
		--deployment-timeout 21m \
		--helm-flags 22 \
		--debug`
	environment = []string{}
	testString = `
			RELEASE_NAME='1'
			CHART_NAME='simple'
			CHART_REPOSITORY='3'
			EXTRA_CHART_VERSION='--version '4''
			SILTA_ENVIRONMENT_NAME='5'
			BRANCHNAME='6'
			NGINX_IMAGE_URL='8'
			CLUSTER_DOMAIN='13'
			EXTRA_NOAUTHIPS='--set nginx.noauthips.vpn='14/32''
			EXTRA_VPCNATIVE='--set cluster.vpcNative='15''
			EXTRA_CLUSTERTYPE='--set cluster.type='16''
			NAMESPACE='19'
			SILTA_CONFIG='20'
			EXTRA_HELM_FLAGS='22'
			DEPLOYMENT_TIMEOUT='21m'

			# Detect pods in FAILED state
			function show_failing_pods() {
				echo ""
				failed_pods=$(kubectl get pod -l "release=$RELEASE_NAME,cronjob!=true" -l "app.kubernetes.io/instance=$RELEASE_NAME,cronjob!=true" -n "$NAMESPACE" -o custom-columns="POD:metadata.name,STATE:status.containerStatuses[*].ready" --no-headers | grep -E "<none>|false" | grep -Eo '^[^ ]+')
				if [[ ! -z "$failed_pods" ]] ; then
					echo "Failing pods:"
					while IFS= read -r pod; do
						echo "---- ${NAMESPACE} / ${pod} ----"
						echo "* Events"
						kubectl get events --field-selector involvedObject.name=${pod},type!=Normal --show-kind=true --ignore-not-found=true --namespace ${NAMESPACE}
						echo ""
						echo "* Logs"
						containers=$(kubectl get pods "${pod}" --namespace "${NAMESPACE}" -o json | jq -r 'try .status | .containerStatuses[] | select(.ready == false).name')
						if [[ ! -z "$containers" ]] ; then
							for container in ${containers}; do
								kubectl logs "${pod}" --prefix=true --since="${DEPLOYMENT_TIMEOUT}" --namespace "${NAMESPACE}" -c "${container}"
							done
						else
							echo "no logs found"
						fi

						echo "----"
					done <<< "$failed_pods"

					false
				else
					true
				fi
				# get statefulsets that are not ready
				not_ready_statefulsets=$(kubectl get statefulset -n "$NAMESPACE" -l "release=$RELEASE_NAME" -l "app.kubernetes.io/instance=$RELEASE_NAME" -o custom-columns="NAME:metadata.name,READY:status.readyReplicas,REPLICAS:spec.replicas" --no-headers | grep -E "<none>|false" | grep -Eo '^[^ ]+')
				if [[ ! -z "$not_ready_statefulsets" ]] ; then
					while IFS= read -r statefulset; do
						events=$(kubectl get events --field-selector involvedObject.name=${statefulset},type!=Normal --show-kind=true --ignore-not-found=true --namespace ${NAMESPACE})
						if [[ ! -z "$events" ]] ; then
							echo "---- ${NAMESPACE} / ${statefulset} statefulset events ----"
							echo "$events"
							echo "----"
						fi
					done <<< "$not_ready_statefulsets"

					false
				else
					true
				fi
				# get deployments that are not ready
				not_ready_deployments=$(kubectl get deployment -n "$NAMESPACE" -l "release=$RELEASE_NAME" -l "app.kubernetes.io/instance=$RELEASE_NAME" -o custom-columns="NAME:metadata.name,READY:status.readyReplicas,REPLICAS:spec.replicas" --no-headers | grep -E "<none>|false" | grep -Eo '^[^ ]+')
				if [[ ! -z "$not_ready_deployments" ]] ; then
					while IFS= read -r deployment; do
						events=$(kubectl get events --field-selector involvedObject.name=${deployment},type!=Normal --show-kind=true --ignore-not-found=true --namespace ${NAMESPACE})
						if [[ ! -z "$events" ]] ; then
							echo "---- ${NAMESPACE} / ${deployment} deployment events ----"
							echo "$events"
							echo "----"
						fi
					done <<< "$not_ready_deployments"
					false
				else
					true
				fi
				rm -f helm-output.log
			}

			trap show_failing_pods ERR

			helm upgrade --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
				--cleanup-on-fail \
				--set environmentName="${SILTA_ENVIRONMENT_NAME}" \
				--set silta-release.branchName="${BRANCHNAME}" \
				--set nginx.image="${NGINX_IMAGE_URL}" \
				--set clusterDomain="${CLUSTER_DOMAIN}" \
				${EXTRA_NOAUTHIPS} \
				${EXTRA_VPCNATIVE} \
				${EXTRA_CLUSTERTYPE} \
				--namespace="${NAMESPACE}" \
				--values "${SILTA_CONFIG}" \
				${EXTRA_HELM_FLAGS} \
				--timeout "${DEPLOYMENT_TIMEOUT}" \
				--wait`
	CliExecTest(t, command, environment, testString, false)

	// Test all args (simple chart)
	command = `ci release deploy \
		--release-name 1 \
		--release-suffix 2 \
		--chart-name simple \
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
		--helm-flags 22 \
		--debug`
	environment = []string{}
	testString = `
			RELEASE_NAME='1'
			CHART_NAME='simple'
			CHART_REPOSITORY='3'
			EXTRA_CHART_VERSION='--version '4''
			SILTA_ENVIRONMENT_NAME='5'
			BRANCHNAME='6'
			NGINX_IMAGE_URL='8'
			CLUSTER_DOMAIN='13'
			EXTRA_NOAUTHIPS='--set nginx.noauthips.vpn='14/32''
			EXTRA_VPCNATIVE='--set cluster.vpcNative='15''
			EXTRA_CLUSTERTYPE='--set cluster.type='16''
			NAMESPACE='19'
			SILTA_CONFIG='20'
			EXTRA_HELM_FLAGS='22'
			DEPLOYMENT_TIMEOUT='21'

			# Detect pods in FAILED state
			function show_failing_pods() {
				echo ""
				failed_pods=$(kubectl get pod -l "release=$RELEASE_NAME,cronjob!=true" -l "app.kubernetes.io/instance=$RELEASE_NAME,cronjob!=true" -n "$NAMESPACE" -o custom-columns="POD:metadata.name,STATE:status.containerStatuses[*].ready" --no-headers | grep -E "<none>|false" | grep -Eo '^[^ ]+')
				if [[ ! -z "$failed_pods" ]] ; then
					echo "Failing pods:"
					while IFS= read -r pod; do
						echo "---- ${NAMESPACE} / ${pod} ----"
						echo "* Events"
						kubectl get events --field-selector involvedObject.name=${pod},type!=Normal --show-kind=true --ignore-not-found=true --namespace ${NAMESPACE}
						echo ""
						echo "* Logs"
						containers=$(kubectl get pods "${pod}" --namespace "${NAMESPACE}" -o json | jq -r 'try .status | .containerStatuses[] | select(.ready == false).name')
						if [[ ! -z "$containers" ]] ; then
							for container in ${containers}; do
								kubectl logs "${pod}" --prefix=true --since="${DEPLOYMENT_TIMEOUT}" --namespace "${NAMESPACE}" -c "${container}"
							done
						else
							echo "no logs found"
						fi

						echo "----"
					done <<< "$failed_pods"

					false
				else
					true
				fi
				# get statefulsets that are not ready
				not_ready_statefulsets=$(kubectl get statefulset -n "$NAMESPACE" -l "release=$RELEASE_NAME" -l "app.kubernetes.io/instance=$RELEASE_NAME" -o custom-columns="NAME:metadata.name,READY:status.readyReplicas,REPLICAS:spec.replicas" --no-headers | grep -E "<none>|false" | grep -Eo '^[^ ]+')
				if [[ ! -z "$not_ready_statefulsets" ]] ; then
					while IFS= read -r statefulset; do
						events=$(kubectl get events --field-selector involvedObject.name=${statefulset},type!=Normal --show-kind=true --ignore-not-found=true --namespace ${NAMESPACE})
						if [[ ! -z "$events" ]] ; then
							echo "---- ${NAMESPACE} / ${statefulset} statefulset events ----"
							echo "$events"
							echo "----"
						fi
					done <<< "$not_ready_statefulsets"

					false
				else
					true
				fi
				# get deployments that are not ready
				not_ready_deployments=$(kubectl get deployment -n "$NAMESPACE" -l "release=$RELEASE_NAME" -l "app.kubernetes.io/instance=$RELEASE_NAME" -o custom-columns="NAME:metadata.name,READY:status.readyReplicas,REPLICAS:spec.replicas" --no-headers | grep -E "<none>|false" | grep -Eo '^[^ ]+')
				if [[ ! -z "$not_ready_deployments" ]] ; then
					while IFS= read -r deployment; do
						events=$(kubectl get events --field-selector involvedObject.name=${deployment},type!=Normal --show-kind=true --ignore-not-found=true --namespace ${NAMESPACE})
						if [[ ! -z "$events" ]] ; then
							echo "---- ${NAMESPACE} / ${deployment} deployment events ----"
							echo "$events"
							echo "----"
						fi
					done <<< "$not_ready_deployments"
					false
				else
					true
				fi
				rm -f helm-output.log
			}

			trap show_failing_pods ERR

			helm upgrade --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
				--cleanup-on-fail \
				--set environmentName="${SILTA_ENVIRONMENT_NAME}" \
				--set silta-release.branchName="${BRANCHNAME}" \
				--set nginx.image="${NGINX_IMAGE_URL}" \
				--set clusterDomain="${CLUSTER_DOMAIN}" \
				${EXTRA_NOAUTHIPS} \
				${EXTRA_VPCNATIVE} \
				${EXTRA_CLUSTERTYPE} \
				--namespace="${NAMESPACE}" \
				--values "${SILTA_CONFIG}" \
				${EXTRA_HELM_FLAGS} \
				--timeout "${DEPLOYMENT_TIMEOUT}" \
				--wait`
	CliExecTest(t, command, environment, testString, false)

	// Change dir back to previous
	os.Chdir(wd)
}
