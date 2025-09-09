package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth provider

	"text/tabwriter"

	helmAction "helm.sh/helm/v3/pkg/action"
	helmCli "helm.sh/helm/v3/pkg/cli"
)

var ciReleaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List releases",
	Run: func(cmd *cobra.Command, args []string) {

		namespace, _ := cmd.Flags().GetString("namespace")

		settings := helmCli.New()

		actionConfig := new(helmAction.Configuration)
		if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
			log.Printf("%+v", err)
			os.Exit(1)
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprintln(writer, "NAME\tNAMESPACE\tREVISION\tUPDATED\tSTATUS\tCHART\tAPP VERSION")

		client := helmAction.NewList(actionConfig)
		client.All = true
		client.SetStateMask()
		results, err := client.Run()
		if err != nil {
			log.Printf("%+v", err)
			os.Exit(1)
		}

		for _, rel := range results {
			fmt.Fprintf(writer, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n", rel.Name, rel.Namespace, rel.Version, rel.Info.LastDeployed.String(), rel.Info.Status.String(), rel.Chart.Name(), rel.Chart.Metadata.Version)
		}
		writer.Flush()

	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseListCmd)

	ciReleaseListCmd.Flags().StringP("namespace", "n", "", "Project name (namespace, i.e. \"drupal-project\")")
}
