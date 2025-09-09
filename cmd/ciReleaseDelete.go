package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth provider

	helmAction "helm.sh/helm/v3/pkg/action"
	helmCli "helm.sh/helm/v3/pkg/cli"
)

var ciReleaseDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a release",
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")
		deletePVCs, _ := cmd.Flags().GetBool("delete-pvcs")

		clientset, err := common.GetKubeClient()
		if err != nil {
			log.Fatalf("failed to get kube client: %v", err)
		}

		// Helm client init logic
		settings := helmCli.New()
		settings.SetNamespace(namespace) // Ensure Helm uses the correct namespace

		actionConfig := new(helmAction.Configuration)
		if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
			log.Printf("%+v", err)
			os.Exit(1)
		}

		err = common.UninstallHelmRelease(clientset, actionConfig, namespace, releaseName, deletePVCs)
		if err != nil {
			log.Fatalf("Error removing a release: %s", err)
		}

	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseDeleteCmd)

	ciReleaseDeleteCmd.Flags().String("release-name", "", "Release name")
	ciReleaseDeleteCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseDeleteCmd.Flags().Bool("delete-pvcs", true, "Delete PVCs (default: true)")

	ciReleaseDeleteCmd.MarkFlagRequired("release-name")
	ciReleaseDeleteCmd.MarkFlagRequired("namespace")
}
