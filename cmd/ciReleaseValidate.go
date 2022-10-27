package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

// ciReleaseValidateCmd represents the ciReleaseValidate command
var ciReleaseValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate release",
	Run: func(cmd *cobra.Command, args []string) {

		releaseName, _ := cmd.Flags().GetString("release-name")
		releaseSuffix, _ := cmd.Flags().GetString("release-suffix")
		namespace, _ := cmd.Flags().GetString("namespace")
		siltaEnvironmentName, _ := cmd.Flags().GetString("silta-environment-name")
		branchname, _ := cmd.Flags().GetString("branchname")
		chartVersion, _ := cmd.Flags().GetString("chart-version")
		chartName, _ := cmd.Flags().GetString("chart-name")
		chartRepository, _ := cmd.Flags().GetString("chart-repository")
		siltaConfig, _ := cmd.Flags().GetString("silta-config")
		vpnIP, _ := cmd.Flags().GetString("vpn-ip")
		vpcNative, _ := cmd.Flags().GetString("vpc-native")
		clusterType, _ := cmd.Flags().GetString("cluster-type")

		// Use environment variables as fallback
		if useEnv == true {
			if len(siltaEnvironmentName) == 0 {
				siltaEnvironmentName = os.Getenv("SILTA_ENVIRONMENT_NAME")
			}
			if len(siltaEnvironmentName) == 0 {
				siltaEnvironmentName = common.SiltaEnvironmentName(branchname, releaseSuffix)
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
		}

		if len(chartRepository) == 0 {
			chartRepository = "https://storage.googleapis.com/charts.wdr.io"
		}

		// Chart value overrides

		// Allow pinning a specific chart version
		chartVersionOverride := ""
		if len(chartVersion) > 0 {
			chartVersionOverride = fmt.Sprintf("--version '%s'", chartVersion)
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

		//Remove release stuck in pending stage
		command := fmt.Sprintf(`
			NAMESPACE='%s'
			RELEASE_NAME='%s'

			# Workaround for previous Helm release stuck in pending state
			echo "Looking for and deleting malformed secrets"
			pending_release=$(helm list -n "$NAMESPACE" --pending --filter="(\s|^)($RELEASE_NAME)(\s|$)"| tail -1 | cut -f1)

			if [[ "$pending_release" == "$RELEASE_NAME" ]]; then
				upgrade_secret_to_delete=$(kubectl get secret -l owner=helm,status=pending-upgrade,name="$RELEASE_NAME" -n "$NAMESPACE" | awk '{print $1}' | grep -v NAME)
				install_secret_to_delete=$(kubectl get secret -l owner=helm,status=pending-install,name="$RELEASE_NAME" -n "$NAMESPACE" | awk '{print $1}' | grep -v NAME)
				
				if [[ ! -z "$upgrade_secret_to_delete" ]] ; then
					kubectl delete secret -n "$NAMESPACE" "$upgrade_secret_to_delete"
				fi
				if [[ ! -z "$install_secret_to_delete" ]] ; then
					kubectl delete secret -n "$NAMESPACE" "$install_secret_to_delete"
				fi
			fi
			`, namespace, releaseName)

		pipedExec(command, debug)

		if chartName == "drupal" || strings.HasSuffix(chartName, "/drupal") {

			_, errDir := os.Stat(common.ExtendedFolder + "/drupal")
			if os.IsNotExist(errDir) == false {
				chartName = common.ExtendedFolder + "/drupal"
			}

			fmt.Printf("Deploying %s helm release %s in %s namespace\n", chartName, releaseName, namespace)

			// TODO: rewrite the timeout handling and log printing after helm release
			command := fmt.Sprintf(`
			set -Eeuo pipefail

			RELEASE_NAME='%s'
			CHART_NAME='%s'
			CHART_REPOSITORY='%s'
			EXTRA_CHART_VERSION='%s'
			SILTA_ENVIRONMENT_NAME='%s'
			BRANCHNAME='%s'
			NAMESPACE='%s'
			SILTA_CONFIG='%s'
			EXTRA_NOAUTHIPS='%s'
			EXTRA_VPCNATIVE='%s'
			EXTRA_CLUSTERTYPE='%s'

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
			}

			trap show_failing_pods ERR

			helm upgrade --dry-run --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
				--set environmentName="${SILTA_ENVIRONMENT_NAME}" \
				--set silta-release.branchName="${BRANCHNAME}" \
				--set php.image="test:test" \
				--set nginx.image="test:test" \
				--set shell.image="test:test" \
				--namespace="${NAMESPACE}" \
				${EXTRA_NOAUTHIPS} \
				${EXTRA_VPCNATIVE} \
				${EXTRA_CLUSTERTYPE} \
				--values "${SILTA_CONFIG}"
				
				`,
				releaseName, chartName, chartRepository, chartVersionOverride,
				siltaEnvironmentName, branchname,
				namespace, siltaConfig, extraNoAuthIPs, vpcNativeOverride, extraClusterType)

			cmd := exec.Command("bash", "-c", command)
			// StdOutPipe omitted to avoid exposing secrets
			cmdErrReader, err := cmd.StderrPipe()
			if err != nil {
				log.Fatal("Error (stderr pipe): ", err)
				return
			}
			errScanner := bufio.NewScanner(cmdErrReader)
			go func() {
				for errScanner.Scan() {
					fmt.Printf("ERROR: %s\n", errScanner.Text())
				}
			}()
			err = cmd.Start()
			if err != nil {
				log.Fatal("Error (Start): ", err)
			}
			err = cmd.Wait()
			if err != nil {
				log.Fatal("Error (Wait): ", err)
			}

		} else {
			fmt.Printf("Chart name %s does not match \"drupal\", helm validation step was skipped\n", chartName)
		}
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseValidateCmd)

	ciReleaseValidateCmd.Flags().String("release-name", "", "Release name")
	ciReleaseValidateCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseValidateCmd.Flags().String("silta-environment-name", "", "Environment name override based on branchname and release-suffix. Used in some helm charts.")
	ciReleaseValidateCmd.Flags().String("branchname", "", "Repository branchname that will be used for release name and environment name creation")
	ciReleaseValidateCmd.Flags().String("vpn-ip", "", "VPN IP for basic auth allow list")
	ciReleaseValidateCmd.Flags().String("vpc-native", "", "VPC-native cluster (GKE specific)")
	ciReleaseValidateCmd.Flags().String("cluster-type", "", "Cluster type (i.e. gke, aws, aks, other)")
	ciReleaseValidateCmd.Flags().String("chart-version", "", "Deploy a specific chart version")
	ciReleaseValidateCmd.Flags().String("chart-name", "", "Chart name")
	ciReleaseValidateCmd.Flags().String("chart-repository", "", "Chart repository")
	ciReleaseValidateCmd.Flags().String("silta-config", "", "Silta release helm chart values")

	ciReleaseValidateCmd.MarkFlagRequired("release-name")
	ciReleaseValidateCmd.MarkFlagRequired("namespace")
	ciReleaseValidateCmd.MarkFlagRequired("chart-name")
}
