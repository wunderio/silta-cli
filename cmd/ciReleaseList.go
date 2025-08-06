package cmd

import (
	"fmt"
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth provider

	"text/tabwriter"
)

var ciReleaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List releases",
	Run: func(cmd *cobra.Command, args []string) {

		namespace, _ := cmd.Flags().GetString("namespace")

		// Try reading KUBECONFIG from environment variable first
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			// If not set, use the default kube config path
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("cannot read user home dir")
				os.Exit(1)
			}
			kubeConfigPath = homeDir + "/.kube/config"
		}

		kubeConfig, err := os.ReadFile(kubeConfigPath)
		if err != nil {
			fmt.Println("cannot read kubeConfig from path")
			os.Exit(1)
		}

		helmOptions := helmclient.Options{
			Namespace:        namespace,
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            false,
			Linting:          false, // Change this to false if you don't want linting.
		}

		//Helm client init logic
		opt := &helmclient.KubeConfClientOptions{
			Options:     &helmOptions,
			KubeContext: "",
			KubeConfig:  kubeConfig,
		}

		helmClient, err := helmclient.NewClientFromKubeConf(opt)
		if err != nil {
			fmt.Println("Cannot create client from kubeConfig")
			os.Exit(1)
		}

		// List Helm releases
		releases, err := helmClient.ListReleasesByStateMask(action.ListAll)
		if err != nil {
			fmt.Printf("Cannot list releases: %s\n", err)
			os.Exit(1)
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprintln(writer, "NAME\tNAMESPACE\tREVISION\tUPDATED\tSTATUS\tCHART\tAPP VERSION")

		for _, release := range releases {
			fmt.Fprintf(writer, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n", release.Name, release.Namespace, release.Version, release.Info.LastDeployed.String(), release.Info.Status.String(), release.Chart.Name(), release.Chart.Metadata.Version)
		}

		writer.Flush()

	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseListCmd)

	ciReleaseListCmd.Flags().StringP("namespace", "n", "", "Project name (namespace, i.e. \"drupal-project\")")
}
