package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

// hubLogoutCmd revokes the stored CLI token and removes local credentials.
var hubLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of the Silta hub",
	Long:  "Revoke the stored CLI token on the Silta hub and remove local credentials.",
	Run: func(cmd *cobra.Command, args []string) {
		client, _, err := common.NewHubClientFromCredentials()
		if err != nil {
			fmt.Println("Not logged in.")
			return
		}

		// Best-effort remote revocation; always clear local credentials.
		if status, err := client.PostJSON("/api/cli/auth/revoke", nil, nil); err != nil && status != http.StatusUnauthorized {
			fmt.Printf("Warning: failed to revoke token on Silta hub: %s\n", err)
		}

		if err := common.DeleteCredentials(); err != nil {
			fmt.Printf("Failed to remove local credentials: %s\n", err)
			return
		}

		fmt.Println("Logged out.")
	},
}

func init() {
	hubCmd.AddCommand(hubLogoutCmd)
}
