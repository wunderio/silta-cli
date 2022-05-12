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

		// Use environment variables as fallback
		if useEnv == true {
			if len(siltaEnvironmentName) == 0 {
				siltaEnvironmentName = os.Getenv("SILTA_ENVIRONMENT_NAME")
			}
			if len(siltaEnvironmentName) == 0 {
				siltaEnvironmentName = common.SiltaEnvironmentName(branchname, releaseSuffix)
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

		if chartName == "drupal" || strings.HasSuffix(chartName, "/drupal") {

			fmt.Printf("Deploying %s helm release %s in %s namespace\n", chartName, releaseName, namespace)

			// TODO: rewrite the timeout handling and log printing after helm release
			command := fmt.Sprintf(`
			set -euo pipefail
			
			RELEASE_NAME='%s'
			CHART_NAME='%s'
			CHART_REPOSITORY='%s'
			EXTRA_CHART_VERSION='%s'
			SILTA_ENVIRONMENT_NAME='%s'
			BRANCHNAME='%s'
			NAMESPACE='%s'
			SILTA_CONFIG='%s'
	
			helm upgrade --dry-run --install "${RELEASE_NAME}" "${CHART_NAME}" \
				--repo "${CHART_REPOSITORY}" \
				${EXTRA_CHART_VERSION} \
				--set environmentName="${SILTA_ENVIRONMENT_NAME}" \
				--set silta-release.branchName="${BRANCHNAME}" \
				--set php.image="test:test" \
				--set nginx.image="test:test" \
				--set shell.image="test:test" \
				--namespace="${NAMESPACE}" \
				--values "${SILTA_CONFIG}"
				
				`,
				releaseName, chartName, chartRepository, chartVersionOverride,
				siltaEnvironmentName, branchname,
				namespace, siltaConfig)

			cmd := exec.Command("bash", "-c", command)
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
	ciReleaseValidateCmd.Flags().String("chart-version", "", "Deploy a specific chart version")
	ciReleaseValidateCmd.Flags().String("chart-name", "", "Chart name")
	ciReleaseValidateCmd.Flags().String("chart-repository", "", "Chart repository")
	ciReleaseValidateCmd.Flags().String("silta-config", "", "Silta release helm chart values")

	ciReleaseValidateCmd.MarkFlagRequired("release-name")
	ciReleaseValidateCmd.MarkFlagRequired("namespace")
	ciReleaseValidateCmd.MarkFlagRequired("chart-name")
}
