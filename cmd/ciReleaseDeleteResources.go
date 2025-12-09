package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth provider

	helmAction "helm.sh/helm/v3/pkg/action"
	helmCli "helm.sh/helm/v3/pkg/cli"
)

var ciReleaseDeleteResourcesCmd = &cobra.Command{
	Use:   "delete-resources",
	Short: "Delete orphaned release resources",
	Long: `Deletes release resources based on labels ("release", "app.kubernetes.io/instance" and "app=<release-name>-es" (for Elasticsearch storage))
		This command can be used to clean up resources when helm release configmaps are absent.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")
		deletePVCs, _ := cmd.Flags().GetBool("delete-pvcs")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

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

		fmt.Printf("Finding orphaned resources for release %s in namespace %s\n", releaseName, namespace)

		err = common.DeleteOrphanedReleaseResources(clientset, actionConfig, namespace, releaseName, deletePVCs, dryRun)
		if err != nil {
			log.Fatalf("Error removing a release: %s", err)
		}

	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseDeleteResourcesCmd)

	ciReleaseDeleteResourcesCmd.Flags().String("release-name", "", "Release name")
	ciReleaseDeleteResourcesCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseDeleteResourcesCmd.Flags().Bool("delete-pvcs", true, "Delete PVCs (default: true)")
	ciReleaseDeleteResourcesCmd.Flags().Bool("dry-run", true, "Dry run (default: true)")

	ciReleaseDeleteResourcesCmd.MarkFlagRequired("release-name")
	ciReleaseDeleteResourcesCmd.MarkFlagRequired("namespace")
}
