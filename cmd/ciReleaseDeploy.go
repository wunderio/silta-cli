package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

// ciReleaseDeployCmd represents the ciReleaseDeploy command
var ciReleaseDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy release",
	Run: func(cmd *cobra.Command, args []string) {

		releaseName, _ := cmd.Flags().GetString("release-name")
		releaseSuffix, _ := cmd.Flags().GetString("release-suffix")
		namespace, _ := cmd.Flags().GetString("namespace")
		siltaEnvironmentName, _ := cmd.Flags().GetString("silta-environment-name")
		branchname, _ := cmd.Flags().GetString("branchname")
		dbRootPass, _ := cmd.Flags().GetString("db-root-pass")
		dbUserPass, _ := cmd.Flags().GetString("db-user-pass")
		vpnIP, _ := cmd.Flags().GetString("vpn-ip")
		vpcNative, _ := cmd.Flags().GetString("vpc-native")
		clusterType, _ := cmd.Flags().GetString("cluster-type")
		chartVersion, _ := cmd.Flags().GetString("chart-version")
		phpImageUrl, _ := cmd.Flags().GetString("php-image-url")
		nginxImageUrl, _ := cmd.Flags().GetString("nginx-image-url")
		shellImageUrl, _ := cmd.Flags().GetString("shell-image-url")
		repositoryUrl, _ := cmd.Flags().GetString("repository-url")
		gitAuthUsername, _ := cmd.Flags().GetString("gitauth-username")
		gitAuthPassword, _ := cmd.Flags().GetString("gitauth-password")
		clusterDomain, _ := cmd.Flags().GetString("cluster-domain")
		chartName, _ := cmd.Flags().GetString("chart-name")
		chartRepository, _ := cmd.Flags().GetString("chart-repository")
		siltaConfig, _ := cmd.Flags().GetString("silta-config")
		helmFlags, _ := cmd.Flags().GetString("helm-flags")
		deploymentTimeout, _ := cmd.Flags().GetString("deployment-timeout")

		// Use environment variables as fallback
		if useEnv == true {
			if len(dbRootPass) == 0 {
				dbRootPass = os.Getenv("DB_ROOT_PASS")
			}
			if len(dbUserPass) == 0 {
				dbUserPass = os.Getenv("DB_USER_PASS")
			}
			if len(vpnIP) == 0 {
				vpnIP = os.Getenv("VPN_IP")
			}
			if len(vpcNative) == 0 {
				vpcNative = os.Getenv("VPC_NATIVE")
			}
			if len(clusterType) == 0 {
				clusterType = os.Getenv("CLUSTER_TYPE")
			}
			if len(siltaEnvironmentName) == 0 {
				siltaEnvironmentName = os.Getenv("SILTA_ENVIRONMENT_NAME")
			}
			if len(siltaEnvironmentName) == 0 {
				siltaEnvironmentName = common.SiltaEnvironmentName(branchname, releaseSuffix)
			}
			if len(repositoryUrl) == 0 {
				repositoryUrl = os.Getenv("CIRCLE_REPOSITORY_URL")
			}
			if len(gitAuthUsername) == 0 {
				gitAuthUsername = os.Getenv("GITAUTH_USERNAME")
			}
			if len(gitAuthPassword) == 0 {
				gitAuthPassword = os.Getenv("GITAUTH_PASSWORD")
			}
			if len(clusterDomain) == 0 {
				clusterDomain = os.Getenv("CLUSTER_DOMAIN")
			}
		}

		if len(deploymentTimeout) == 0 {
			deploymentTimeout = "15m"
		}
		if len(chartRepository) == 0 {
			chartRepository = "https://storage.googleapis.com/charts.wdr.io"
		}

		// Chart value overrides

		// Override Database credentials if specified
		dbRootPassOverride := ""
		if len(dbRootPass) > 0 {
			dbRootPassOverride = fmt.Sprintf("--set mariadb.rootUser.password='%s'", dbRootPass)
		}
		dbUserPassOverride := ""
		if len(dbUserPass) > 0 {
			dbUserPassOverride = fmt.Sprintf("--set mariadb.db.password='%s'", dbUserPass)
		}

		// Skip basic auth for internal VPN if defined in environment
		extraNoAuthIPs := ""
		if len(vpnIP) > 0 {
			extraNoAuthIPs = fmt.Sprintf("--set nginx.noauthips.vpn='%s/32'", vpnIP)
		}

		// Pass VPC-native setting if defined in environment
		vpcNativeOverride := ""
		if len(vpcNative) > 0 {
			vpcNativeOverride = fmt.Sprintf("--set cluster.vpcNative='%s'", vpcNative)
		}

		// Add cluster type if defined in environment
		extraClusterType := ""
		if len(clusterType) > 0 {
			extraClusterType = fmt.Sprintf("--set cluster.type='%s'", clusterType)
		}

		// Allow pinning a specific chart version
		chartVersionOverride := ""
		if len(chartVersion) > 0 {
			chartVersionOverride = fmt.Sprintf("--version '%s'", chartVersion)
		}

		// TODO: Create namespace if it doesn't exist
		// & tag the namespace if it isn't already tagged.
		// TODO: Rewrite

		command := fmt.Sprintf(`
				# Deployed chart version
				NAMESPACE='%s'

				# Create the namespace if it doesn't exist.
				if ! kubectl get namespace "$NAMESPACE" &>/dev/null ; then
					kubectl create namespace "$NAMESPACE"
				fi

				# Tag the namespace if it isn't already tagged.
				if ! kubectl get namespace -l name=$NAMESPACE --no-headers | grep $NAMESPACE &>/dev/null ; then
					kubectl label namespace "$NAMESPACE" "name=$NAMESPACE" --overwrite
				fi

			`, namespace)
		pipedExec(command, debug)

		if debug == false {
			// Add helm repositories
			command := fmt.Sprintf("helm repo add '%s' '%s'", "wunderio", chartRepository)
			exec.Command("bash", "-c", command).Run()

			// Delete existing jobs to prevent getting wrong log output
			command = fmt.Sprintf("kubectl delete job '%s-post-release' --namespace '%s' --ignore-not-found", releaseName, namespace)
			exec.Command("bash", "-c", command).Run()
		}

		if chartName == "simple" || strings.HasSuffix(chartName, "/simple") {

			if len(nginxImageUrl) == 0 {
				log.Fatal("Nginx image url required (nginx-image-url)")
			}

			fmt.Printf("Deploying %s helm release %s in %s namespace\n", chartName, releaseName, namespace)

			// helm release
			command = fmt.Sprintf(`
			set -euo pipefail

			RELEASE_NAME='%s'
			CHART_NAME='%s'
			CHART_REPOSITORY='%s'
			EXTRA_CHART_VERSION='%s'
			SILTA_ENVIRONMENT_NAME='%s'
			BRANCHNAME='%s'
			NGINX_IMAGE_URL='%s'
			CLUSTER_DOMAIN='%s'	
			EXTRA_NOAUTHIPS='%s'
			EXTRA_VPCNATIVE='%s'
			EXTRA_CLUSTERTYPE='%s'
			NAMESPACE='%s'
			SILTA_CONFIG='%s'
			EXTRA_HELM_FLAGS='%s'
			
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
				--wait`,
				releaseName, chartName, chartRepository, chartVersionOverride,
				siltaEnvironmentName, branchname, nginxImageUrl,
				clusterDomain, extraNoAuthIPs, vpcNativeOverride, extraClusterType,
				namespace, siltaConfig, helmFlags)
			pipedExec(command, debug)

		} else if chartName == "frontend" || strings.HasSuffix(chartName, "/frontend") {

			fmt.Printf("Deploying %s helm release %s in %s namespace\n", chartName, releaseName, namespace)

			// helm release
			command = fmt.Sprintf(`
			set -euo pipefail
			
			RELEASE_NAME='%s'
			CHART_NAME='%s'
			CHART_REPOSITORY='%s'
			EXTRA_CHART_VERSION='%s'
			SILTA_ENVIRONMENT_NAME='%s'
			BRANCHNAME='%s'
			GIT_REPOSITORY_URL='%s'
			GITAUTH_USERNAME='%s'
			GITAUTH_PASSWORD='%s'
			CLUSTER_DOMAIN='%s'	
			NAMESPACE='%s'
			EXTRA_NOAUTHIPS='%s'
			EXTRA_VPCNATIVE='%s'
			EXTRA_CLUSTERTYPE='%s'
			EXTRA_DB_ROOT_PASS='%s'
			EXTRA_DB_USER_PASS='%s'
			SILTA_CONFIG='%s'
			EXTRA_HELM_FLAGS='%s'
			DEPLOYMENT_TIMEOUT='%s'

			helm upgrade --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
				--cleanup-on-fail \
				--set environmentName="${SILTA_ENVIRONMENT_NAME}" \
				--set silta-release.branchName="${BRANCHNAME}" \
				--set shell.gitAuth.repositoryUrl="${GIT_REPOSITORY_URL}" \
				--set shell.gitAuth.keyserver.username="${GITAUTH_USERNAME}" \
				--set shell.gitAuth.keyserver.password="${GITAUTH_PASSWORD}" \
				--set clusterDomain="${CLUSTER_DOMAIN}" \
				--namespace="${NAMESPACE}" \
				${EXTRA_NOAUTHIPS} \
				${EXTRA_VPCNATIVE} \
				${EXTRA_CLUSTERTYPE} \
				${EXTRA_DB_ROOT_PASS} \
				${EXTRA_DB_USER_PASS} \
				--values "${SILTA_CONFIG}" \
				${EXTRA_HELM_FLAGS} \
				--timeout "${DEPLOYMENT_TIMEOUT}" &> helm-output.log & pid=$!

				# TODO: Rewrite this part

				echo "Waiting for containers to start"

				TIME_WAITING=0
				LOGS_SHOWN=false
				while true; do
					if [ $LOGS_SHOWN == false ] && kubectl get pod -l job-name="${RELEASE_NAME}-post-release" -n "${NAMESPACE}" --ignore-not-found | grep  -qE "Running|Completed" ; then
					echo ""
					echo "Deployment log:"
					kubectl logs "job/${RELEASE_NAME}-post-release" -n "${NAMESPACE}" -f --timestamps=true || true
					LOGS_SHOWN=true
					fi

					# Helm command is complete.
					if ! ps -p "$pid" > /dev/null; then
					if grep -q BackoffLimitExceeded helm-output.log ; then
						# Don't show BackoffLimitExceeded, it confuses everyone.
						show_failing_pods
						echo "The post-release job failed, see log output above."
					else
						echo "Helm output:"
						cat helm-output.log
					fi
					wait $pid
					break
					fi

					if [ $TIME_WAITING -gt 300 ]; then
					echo "Timeout waiting for resources."
					show_failing_pods
					exit 1
					fi

					echo "."
					sleep 10
					TIME_WAITING=$((TIME_WAITING+10))
				done

				# Wait for resources to be ready
				# Get all deployments and statefulsets in the release and check the status of each one.
				statefulsets=$(kubectl get statefulset -n "$NAMESPACE" -l "release=${RELEASE_NAME}" -o name)
				if [ ! -z "$statefulsets" ]; then
					echo "$statefulsets" | xargs -n 1 kubectl rollout status -n "$NAMESPACE"
				fi
				kubectl get deployment -n "$NAMESPACE" -l "release=${RELEASE_NAME}" -o name | xargs -n 1 kubectl rollout status -n "$NAMESPACE"
				`,
				releaseName, chartName, chartRepository, chartVersionOverride,
				siltaEnvironmentName, branchname,
				repositoryUrl, gitAuthUsername, gitAuthPassword,
				clusterDomain, namespace,
				extraNoAuthIPs, vpcNativeOverride, extraClusterType,
				dbRootPassOverride, dbUserPassOverride,
				siltaConfig, helmFlags,
				deploymentTimeout)
			pipedExec(command, debug)

		} else if chartName == "drupal" || strings.HasSuffix(chartName, "/drupal") {

			if len(phpImageUrl) == 0 {
				log.Fatal("PHP image url required (php-image-url)")
			}
			if len(nginxImageUrl) == 0 {
				log.Fatal("Nginx image url required (nginx-image-url)")
			}
			if len(shellImageUrl) == 0 {
				log.Fatal("Shell image url required (shell-image-url)")
			}

			// Special updates
			// TODO: Rewrite
			command = fmt.Sprintf(`
					# Deployed chart version
					NAMESPACE='%s'
					RELEASE_NAME='%s'
					if helm status -n "$NAMESPACE" "$RELEASE_NAME" > /dev/null  2>&1
					then
						CURRENT_CHART_VERSION=$(helm history -n "$NAMESPACE" "$RELEASE_NAME" --max 1 --output json | jq -r '.[].chart')
						echo "There is an existing chart deployed with version $CURRENT_CHART_VERSION"
					fi

					# Special updates
					function version_lt() { test "$(printf '%%s\n' "$@" | sort -rV | head -n 1)" != "$1"; }

					if [[ -n "$CURRENT_CHART_VERSION" ]] && [[ "$CURRENT_CHART_VERSION" = drupal-* ]]
					then
						if version_lt "$CURRENT_CHART_VERSION" "drupal-0.3.43"
						then
						echo "Recreating statefulset for Mariadb subchart update to 7.x."
						kubectl delete statefulset --cascade=false "$RELEASE_NAME-mariadb" -n "$NAMESPACE"
						fi
					fi
				`, namespace, releaseName)
			pipedExec(command, debug)

			// Clean up failed Helm releases
			// TODO: Rewrite
			command = fmt.Sprintf(`
					NAMESPACE='%s'
					RELEASE_NAME='%s'
					failed_revision=$(helm list -n "$NAMESPACE" --failed --pending --filter="(\s|^)($RELEASE_NAME)(\s|$)" | tail -1 | cut -f3)

					if [[ "$failed_revision" -eq 1 ]]; then
						# Remove any existing post-release hook, since it's technically not part of the release.
						kubectl delete job -n "$NAMESPACE" "$RELEASE_NAME-post-release --ignore-not-found" 2> /dev/null || true

						echo "Removing failed first release."
						helm delete -n "$NAMESPACE" "$RELEASE_NAME"

						echo "Delete persistent volume claims left over from statefulsets."
						kubectl delete pvc -n "$NAMESPACE" -l release="$RELEASE_NAME"
						kubectl delete pvc -n "$NAMESPACE" -l app="$RELEASE_NAME-es"

						echo "Waiting for volumes to be deleted."
						until [[ -z $(kubectl get pv | grep "$NAMESPACE/$RELEASE_NAME-") ]]
						do
						echo "."
						sleep 10
						done
					fi

					# Workaround for previous Helm release stuck in pending state
					pending_release=$(helm list -n "$NAMESPACE" --pending --filter="(\s|^)($RELEASE_NAME)(\s|$)"| tail -1 | cut -f1)

					if [[ "$pending_release" == "$RELEASE_NAME" ]]; then
						secret_to_delete=$(kubectl get secret -l owner=helm,status=pending-upgrade,name="$RELEASE_NAME" -n "$NAMESPACE" | awk '{print $1}' | grep -v NAME)
						kubectl delete secret -n "$NAMESPACE" "$secret_to_delete"
					fi
				`, namespace, releaseName)
			pipedExec(command, debug)

			// Chart value overrides

			// Disable reference data if the required volume is not present.
			referenceDataOverride := ""
			if debug == false {
				command = fmt.Sprintf("kubectl get persistentvolume | grep --extended-regexp '%s/.*-reference-data'", namespace)
				err := exec.Command("bash", "-c", command).Run()
				if err != nil {
					referenceDataOverride = "--set referenceData.skipMount=true"
				}
			}

			fmt.Printf("Deploying %s helm release %s in %s namespace\n", chartName, releaseName, namespace)

			// TODO: rewrite the timeout handling and log printing after helm release
			command = fmt.Sprintf(`
			set -euo pipefail
			
			RELEASE_NAME='%s'
			CHART_NAME='%s'
			CHART_REPOSITORY='%s'
			EXTRA_CHART_VERSION='%s'
			SILTA_ENVIRONMENT_NAME='%s'
			BRANCHNAME='%s'
			PHP_IMAGE_URL='%s'
			NGINX_IMAGE_URL='%s'
			SHELL_IMAGE_URL='%s'
			GIT_REPOSITORY_URL='%s'
			GITAUTH_USERNAME='%s'
			GITAUTH_PASSWORD='%s'
			CLUSTER_DOMAIN='%s'	
			EXTRA_NOAUTHIPS='%s'
			EXTRA_VPCNATIVE='%s'
			EXTRA_CLUSTERTYPE='%s'
			EXTRA_DB_ROOT_PASS='%s'
			EXTRA_DB_USER_PASS='%s'
			EXTRA_REFERENCE_DATA='%s'
			NAMESPACE='%s'
			SILTA_CONFIG='%s'
			EXTRA_HELM_FLAGS='%s'
			DEPLOYMENT_TIMEOUT='%s'

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
				--timeout "${DEPLOYMENT_TIMEOUT}" &> helm-output.log & pid=$!

				# TODO: Rewrite this part
				
				echo "Waiting for containers to start"

				TIME_WAITING=0
				LOGS_SHOWN=false
				while true; do
					if [ $LOGS_SHOWN == false ] && kubectl get pod -l job-name="${RELEASE_NAME}-post-release" -n "${NAMESPACE}" --ignore-not-found | grep  -qE "Running|Completed" ; then
					echo ""
					echo "Deployment log:"
					kubectl logs "job/${RELEASE_NAME}-post-release" -n "${NAMESPACE}" -f --timestamps=true || true
					LOGS_SHOWN=true
					fi

					# Helm command is complete.
					if ! ps -p "$pid" > /dev/null; then
					if grep -q BackoffLimitExceeded helm-output.log ; then
						# Don't show BackoffLimitExceeded, it confuses everyone.
						show_failing_pods
						echo "The post-release job failed, see log output above."
					else
						echo "Helm output:"
						cat helm-output.log
					fi
					wait $pid
					break
					fi

					if [ $TIME_WAITING -gt 300 ]; then
					echo "Timeout waiting for resources."
					show_failing_pods
					exit 1
					fi

					echo "."
					sleep 10
					TIME_WAITING=$((TIME_WAITING+10))
				done

				# Wait for resources to be ready
				# Get all deployments and statefulsets in the release and check the status of each one.
				statefulsets=$(kubectl get statefulset -n "$NAMESPACE" -l "release=${RELEASE_NAME}" -o name)
				if [ ! -z "$statefulsets" ]; then
					echo "$statefulsets" | xargs -n 1 kubectl rollout status -n "$NAMESPACE"
				fi
				kubectl get deployment -n "$NAMESPACE" -l "release=${RELEASE_NAME}" -o name | xargs -n 1 kubectl rollout status -n "$NAMESPACE"
				`,
				releaseName, chartName, chartRepository, chartVersionOverride,
				siltaEnvironmentName, branchname,
				phpImageUrl, nginxImageUrl, shellImageUrl,
				repositoryUrl, gitAuthUsername, gitAuthPassword,
				clusterDomain, extraNoAuthIPs, vpcNativeOverride, extraClusterType,
				dbRootPassOverride, dbUserPassOverride, referenceDataOverride, namespace,
				siltaConfig, helmFlags, deploymentTimeout)
			pipedExec(command, debug)

		} else {
			fmt.Printf("Chart name %s does not match preselected names (drupal, frontend, simple), helm release step was skipped\n", chartName)
		}
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseDeployCmd)

	ciReleaseDeployCmd.Flags().String("release-name", "", "Release name")
	ciReleaseDeployCmd.Flags().String("release-suffix", "", "Release name suffix for environment name creation")
	ciReleaseDeployCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseDeployCmd.Flags().String("silta-environment-name", "", "Environment name override based on branchname and release-suffix. Used in some helm charts.")
	ciReleaseDeployCmd.Flags().String("branchname", "", "Repository branchname that will be used for release name and environment name creation")
	ciReleaseDeployCmd.Flags().String("db-root-pass", "", "Database password for root account")
	ciReleaseDeployCmd.Flags().String("db-user-pass", "", "Database password for user account")
	ciReleaseDeployCmd.Flags().String("vpn-ip", "", "VPN IP for basic auth allow list")
	ciReleaseDeployCmd.Flags().String("vpc-native", "", "VPC-native cluster (GKE specific)")
	ciReleaseDeployCmd.Flags().String("cluster-type", "", "Cluster type (i.e. gke, aws, aks, other)")
	ciReleaseDeployCmd.Flags().String("chart-version", "", "Deploy a specific chart version")
	ciReleaseDeployCmd.Flags().String("php-image-url", "", "PHP image url")
	ciReleaseDeployCmd.Flags().String("nginx-image-url", "", "PHP image url")
	ciReleaseDeployCmd.Flags().String("shell-image-url", "", "PHP image url")
	ciReleaseDeployCmd.Flags().String("repository-url", "", "Repository url (i.e. git@github.com:wunderio/silta.git)")
	ciReleaseDeployCmd.Flags().String("gitauth-username", "", "Gitauth server username")
	ciReleaseDeployCmd.Flags().String("gitauth-password", "", "Gitauth server password")
	ciReleaseDeployCmd.Flags().String("cluster-domain", "", "Base domain for cluster urls (i.e. dev.example.com)")
	ciReleaseDeployCmd.Flags().String("chart-name", "", "Chart name")
	ciReleaseDeployCmd.Flags().String("chart-repository", "", "Chart repository")
	ciReleaseDeployCmd.Flags().String("silta-config", "", "Silta release helm chart values")
	ciReleaseDeployCmd.Flags().String("helm-flags", "", "Extra flags for helm release")
	ciReleaseDeployCmd.Flags().String("deployment-timeout", "", "Helm deployment timeout")

	ciReleaseDeployCmd.MarkFlagRequired("release-name")
	ciReleaseDeployCmd.MarkFlagRequired("namespace")
	ciReleaseDeployCmd.MarkFlagRequired("chart-name")
}
