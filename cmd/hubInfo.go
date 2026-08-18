package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var infoVerify bool

// hubInfoCmd reports the current Silta hub login state from the locally
// stored credentials.
var hubInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current Silta Hub login state",
	Run: func(cmd *cobra.Command, args []string) {
		creds, err := common.LoadCredentials()
		if err != nil {
			fmt.Println("Not logged in. Run 'silta hub login' first.")
			return
		}

		fmt.Printf("Logged in as: %s\n", creds.Username)
		fmt.Printf("Silta Hub URL:      %s\n", creds.HubURL)
		if creds.ExpiresAt != "" {
			state := "valid"
			if creds.Expired() {
				state = "EXPIRED - run 'silta hub login' again"
			}
			fmt.Printf("Token expiry: %s (%s)\n", creds.ExpiresAt, state)
		}

		if !infoVerify {
			return
		}

		client := common.NewHubClient(creds.HubURL, creds.Token)
		var resp struct {
			Username string                `json:"username"`
			Clusters []common.SiltaCluster `json:"clusters"`
		}
		if _, err := client.GetJSON("/api/cli/clusters", &resp); err != nil {
			fmt.Printf("\nServer verification failed: %s\n", err)
			return
		}

		fmt.Printf("\nServer confirmed session for: %s\n", resp.Username)
		if len(resp.Clusters) == 0 {
			fmt.Println("No cluster access is currently assigned to your account.")
			return
		}
		fmt.Printf("Cluster access (%d):\n", len(resp.Clusters))
		for _, cluster := range resp.Clusters {
			ns := cluster.Namespace
			if ns == "" {
				ns = "(none)"
			}
			fmt.Printf("  silta-%s [namespace: %s]\n", cluster.ID, ns)
		}
	},
}

func init() {
	hubInfoCmd.Flags().BoolVar(&infoVerify, "verify", false, "Verify the token with the silta hub and list cluster access")
	hubCmd.AddCommand(hubInfoCmd)
}
