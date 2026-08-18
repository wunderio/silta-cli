package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var hubClustersJSON bool

// hubClustersCmd lists the clusters the logged-in user may access via the CLI.
var hubClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "List the clusters accessible from your Silta account",
	Long: `List the clusters the logged-in Silta user may access, with the
kubeconfig context name, assigned namespace and (when reported by cluster
inventory) the Kubernetes version.

Use --json for machine-readable output (exit code 0 on success, 1 on
failure).`,
	Run: func(cmd *cobra.Command, args []string) {
		creds, err := common.LoadCredentials()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: not logged in. Run 'silta hub login' first.")
			os.Exit(1)
		}

		client := common.NewHubClient(creds.HubURL, creds.Token)
		resp, err := common.FetchHubClusters(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to fetch clusters from Silta hub: %s\n", err)
			os.Exit(1)
		}

		if hubClustersJSON {
			encoded, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to encode clusters: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(encoded))
			return
		}

		if len(resp.Clusters) == 0 {
			fmt.Println("No cluster access is currently assigned to your account.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "CONTEXT\tNAMESPACE\tKUBERNETES VERSION")
		for _, cluster := range resp.Clusters {
			ns := cluster.Namespace
			if ns == "" {
				ns = "(none)"
			}
			version := cluster.KubernetesVersion
			if version == "" {
				version = "(unknown)"
			}
			fmt.Fprintf(w, "silta-%s\t%s\t%s\n", cluster.ID, ns, version)
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to print clusters: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	hubClustersCmd.Flags().BoolVar(&hubClustersJSON, "json", false, "Output clusters as JSON")
	hubCmd.AddCommand(hubClustersCmd)
}
