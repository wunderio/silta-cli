package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

// hubKubeconfigCmd refreshes the local kubeconfig with the user's current cluster
// access from the Silta hub. It merges silta-<cluster> contexts into the user's kubeconfig, each authenticated via the 'silta hub cli-credential' exec plugin.
var hubKubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Update kubeconfig with Silta cluster access",
	Long:  "Fetch the clusters and namespaces you can access and merge silta-<cluster> contexts into your kubeconfig.",
	Run: func(cmd *cobra.Command, args []string) {
		_, creds, err := common.NewHubClientFromCredentials()
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		if err := syncKubeconfig(creds); err != nil {
			fmt.Printf("Failed to update kubeconfig: %s\n", err)
		}
	},
}

func init() {
	hubCmd.AddCommand(hubKubeconfigCmd)
}
