package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ciReleaseDiffCmd represents the ciReleaseDiff command
var ciReleaseDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff release resources",
	Long: `Release diff command is used to compare the resources of a release with the current state of the cluster.
	
	* Chart allows prepending extra configuration (to helm --values line) via 
	"SILTA_<chart_name>_CONFIG_VALUES" environment variable. It has to be a 
	base64 encoded string of a silta configuration yaml file.
	`,
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

		// Use environment variables as fallback
		if useEnv {
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

		// Uses PrependChartConfigOverrides from "SILTA_<CHART_NAME>_CONFIG_VALUES"
		// environment variable and prepends it to configuration
		chartOverrideFile := common.CreateChartConfigurationFile(chartName)
		if chartOverrideFile != "" {
			defer os.Remove(chartOverrideFile)
			siltaConfig = common.PrependChartConfigOverrides(chartOverrideFile, siltaConfig)
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

		if !debug {
			// Add helm repositories
			command := fmt.Sprintf("helm repo add '%s' '%s'", "wunderio", chartRepository)
			exec.Command("bash", "-c", command).Run()

			// Make sure repositories are up to date
			command = "helm repo update"
			exec.Command("bash", "-c", command).Run()
		}

		if chartName == "simple" || strings.HasSuffix(chartName, "/simple") {

			if len(nginxImageUrl) == 0 {
				log.Fatal("Nginx image url required (nginx-image-url)")
			}

			_, errDir := os.Stat(common.ExtendedFolder + "/simple")
			if !os.IsNotExist(errDir) {
				chartName = common.ExtendedFolder + "/simple"
			}

			fmt.Printf("Diffing %s helm release %s in %s namespace\n", chartName, releaseName, namespace)

			// helm release
			command := fmt.Sprintf(`
			set -Eeuo pipefail

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

			helm diff upgrade --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
				--set environmentName="${SILTA_ENVIRONMENT_NAME}" \
				--set silta-release.branchName="${BRANCHNAME}" \
				--set nginx.image="${NGINX_IMAGE_URL}" \
				--set clusterDomain="${CLUSTER_DOMAIN}" \
				${EXTRA_NOAUTHIPS} \
				${EXTRA_VPCNATIVE} \
				${EXTRA_CLUSTERTYPE} \
				--namespace="${NAMESPACE}" \
				--values "${SILTA_CONFIG}" \
				${EXTRA_HELM_FLAGS}`,
				releaseName, chartName, chartRepository, chartVersionOverride,
				siltaEnvironmentName, branchname, nginxImageUrl,
				clusterDomain, extraNoAuthIPs, vpcNativeOverride, extraClusterType,
				namespace, siltaConfig, helmFlags)
			pipedExec(command, "", "ERROR: ", debug)

		} else if chartName == "frontend" || strings.HasSuffix(chartName, "/frontend") {

			fmt.Printf("Diffing %s helm release %s in %s namespace\n", chartName, releaseName, namespace)

			_, errDir := os.Stat(common.ExtendedFolder + "/frontend")
			if !os.IsNotExist(errDir) {
				chartName = common.ExtendedFolder + "/frontend"
			}

			// helm release
			command := fmt.Sprintf(`
			set -Eeuo pipefail

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
			
			helm diff upgrade --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
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
				${EXTRA_HELM_FLAGS}`,
				releaseName, chartName, chartRepository, chartVersionOverride,
				siltaEnvironmentName, branchname,
				repositoryUrl, gitAuthUsername, gitAuthPassword,
				clusterDomain, namespace,
				extraNoAuthIPs, vpcNativeOverride, extraClusterType,
				dbRootPassOverride, dbUserPassOverride,
				siltaConfig, helmFlags)
			pipedExec(command, "", "ERROR: ", debug)

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

			_, errDir := os.Stat(common.ExtendedFolder + "/drupal")
			if os.IsNotExist(errDir) == false {
				chartName = common.ExtendedFolder + "/drupal"
			}

			// Disable reference data if the required volume is not present.
			referenceDataOverride := ""
			if !debug {

				homeDir, err := os.UserHomeDir()
				if err != nil {
					log.Fatalf("cannot read user home dir")
				}
				kubeConfigPath := homeDir + "/.kube/config"

				//k8s go client init logic
				config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
				if err != nil {
					log.Fatalf("cannot read kubeConfig from path: %s", err)
				}
				clientset, err := kubernetes.NewForConfig(config)
				if err != nil {
					log.Fatalf("cannot initialize k8s client: %s", err)
				}

				// PVC name can be either "*-reference-data" or "*-reference", so we need to check both
				// Unless we parse and merge configuration yaml files, we can't know the exact name of the PVC
				// Check all pvc's in the namespace and see if any of them match the pattern
				pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), v1.ListOptions{})
				if err != nil {
					log.Fatalf("cannot get persistent volume claims: %s", err)
				}
				referenceDataExists := false
				for _, pvc := range pvcs.Items {
					if strings.HasSuffix(pvc.Name, "-reference-data") || strings.HasSuffix(pvc.Name, "-reference") {
						referenceDataExists = true
						break
					}
				}
				if !referenceDataExists {
					referenceDataOverride = "--set referenceData.skipMount=true"
				}
			}

			fmt.Printf("Diffing %s helm release %s in %s namespace\n", chartName, releaseName, namespace)

			command := fmt.Sprintf(`
			set -Eeuo pipefail

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

			helm diff upgrade --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
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
				${EXTRA_HELM_FLAGS}`,
				releaseName, chartName, chartRepository, chartVersionOverride,
				siltaEnvironmentName, branchname,
				phpImageUrl, nginxImageUrl, shellImageUrl,
				repositoryUrl, gitAuthUsername, gitAuthPassword,
				clusterDomain, extraNoAuthIPs, vpcNativeOverride, extraClusterType,
				dbRootPassOverride, dbUserPassOverride, referenceDataOverride, namespace,
				siltaConfig, helmFlags)
			pipedExec(command, "", "ERROR: ", debug)

		} else {
			fmt.Printf("Chart name %s does not match preselected names (drupal, frontend, simple), helm diff step was skipped\n", chartName)
		}
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseDiffCmd)

	ciReleaseDiffCmd.Flags().String("release-name", "", "Release name")
	ciReleaseDiffCmd.Flags().String("release-suffix", "", "Release name suffix for environment name creation")
	ciReleaseDiffCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseDiffCmd.Flags().String("silta-environment-name", "", "Environment name override based on branchname and release-suffix. Used in some helm charts.")
	ciReleaseDiffCmd.Flags().String("branchname", "", "Repository branchname that will be used for release name and environment name creation")
	ciReleaseDiffCmd.Flags().String("db-root-pass", "", "Database password for root account")
	ciReleaseDiffCmd.Flags().String("db-user-pass", "", "Database password for user account")
	ciReleaseDiffCmd.Flags().String("vpn-ip", "", "VPN IP for basic auth allow list")
	ciReleaseDiffCmd.Flags().String("vpc-native", "", "VPC-native cluster (GKE specific)")
	ciReleaseDiffCmd.Flags().String("cluster-type", "", "Cluster type (i.e. gke, aws, aks, other)")
	ciReleaseDiffCmd.Flags().String("chart-version", "", "Diff a specific chart version")
	ciReleaseDiffCmd.Flags().String("php-image-url", "", "PHP image url")
	ciReleaseDiffCmd.Flags().String("nginx-image-url", "", "PHP image url")
	ciReleaseDiffCmd.Flags().String("shell-image-url", "", "PHP image url")
	ciReleaseDiffCmd.Flags().String("repository-url", "", "Repository url (i.e. git@github.com:wunderio/silta.git)")
	ciReleaseDiffCmd.Flags().String("gitauth-username", "", "Gitauth server username")
	ciReleaseDiffCmd.Flags().String("gitauth-password", "", "Gitauth server password")
	ciReleaseDiffCmd.Flags().String("cluster-domain", "", "Base domain for cluster urls (i.e. dev.example.com)")
	ciReleaseDiffCmd.Flags().String("chart-name", "", "Chart name")
	ciReleaseDiffCmd.Flags().String("chart-repository", "https://storage.googleapis.com/charts.wdr.io", "Chart repository")
	ciReleaseDiffCmd.Flags().String("silta-config", "", "Silta release helm chart values")
	ciReleaseDiffCmd.Flags().String("helm-flags", "", "Extra flags for helm release")

	ciReleaseDiffCmd.MarkFlagRequired("release-name")
	ciReleaseDiffCmd.MarkFlagRequired("namespace")
	ciReleaseDiffCmd.MarkFlagRequired("chart-name")
}
