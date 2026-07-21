package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

// cliCredentialCmd is a client-go exec credential plugin. kubectl invokes it to
// obtain the bearer token for silta-proxied clusters. It is hidden from normal
// help output.
var cliCredentialCmd = &cobra.Command{
	Use:    "cli-credential",
	Short:  "Output a Kubernetes ExecCredential for the Silta Hub proxy",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		creds, err := common.LoadCredentials()
		if err != nil {
			fmt.Fprintf(os.Stderr, "silta: %s\n", err)
			os.Exit(1)
		}

		status := map[string]interface{}{
			"token": creds.Token,
		}
		if creds.ExpiresAt != "" {
			status["expirationTimestamp"] = creds.ExpiresAt
		}

		out := map[string]interface{}{
			"apiVersion": "client.authentication.k8s.io/v1",
			"kind":       "ExecCredential",
			"status":     status,
		}

		if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
			fmt.Fprintf(os.Stderr, "silta: failed to encode credential: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	hubCmd.AddCommand(cliCredentialCmd)
}
