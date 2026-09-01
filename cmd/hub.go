package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

// hubCmd groups commands that interact with a Silta Hub.
var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "Interact with a Silta hub",
	Long:  "Authenticate against a Silta hub, manage kubeconfig access and open the hub in a browser.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

// hubDashboardCmd opens the configured hub dashboard in the default browser.
var hubDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the Silta dashboard in a browser",
	Run: func(cmd *cobra.Command, args []string) {
		target := resolveDashboardBrowserURL()
		if target == "" {
			fmt.Println("Error: no Silta hub URL configured. Pass --hub-url or run 'silta config set hub.url <url>'.")
			return
		}

		fmt.Printf("Opening %s\n", target)
		if err := common.OpenBrowser(target); err != nil {
			fmt.Printf("Failed to open browser: %s\n", err)
			fmt.Printf("Open this URL manually: %s\n", target)
		}
	},
}

// resolveDashboardBrowserURL prefers the dashboard-advertised frontend URL,
// falling back to the configured hub URL when it cannot be reached.
func resolveDashboardBrowserURL() string {
	hubURL := resolveHubURL()
	if hubURL == "" {
		if creds, err := common.LoadCredentials(); err == nil {
			hubURL = creds.HubURL
		}
	}
	if hubURL == "" {
		return ""
	}

	client := common.NewHubClient(hubURL, "")
	var info struct {
		FrontendURL string `json:"frontend_url"`
	}
	if _, err := client.GetJSON("/api/cli/info", &info); err == nil && info.FrontendURL != "" {
		return info.FrontendURL
	}
	return hubURL
}

func init() {
	hubCmd.AddCommand(hubDashboardCmd)
	rootCmd.AddCommand(hubCmd)
}
