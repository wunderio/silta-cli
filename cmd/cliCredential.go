package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

// expiredWarnWindow is how long the expiry hint stays suppressed after it is
// printed. kubectl retries discovery several times per invocation, re-running
// this plugin each time, so without this the hint would print once per retry.
const expiredWarnWindow = 5 * time.Minute

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

		if creds.Expired() {
			if shouldPrintExpiredHint() {
				fmt.Fprintf(os.Stderr, "silta: session has expired (expired at %s). Log in again: silta hub login (add --device on headless machines)\n", creds.ExpiresAt)
			}
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

// shouldPrintExpiredHint reports whether the session-expiry hint should be
// shown now, recording that it was shown under a marker file in the
// configuration directory so that repeat invocations within
// expiredWarnWindow (kubectl re-fetches credentials on every discovery retry)
// stay quiet. File errors fail open: print the hint.
func shouldPrintExpiredHint() bool {
	path := filepath.Join(common.ConfigDir(), ".cli-credential-expired")

	if data, err := os.ReadFile(path); err == nil {
		if ts, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data))); err == nil {
			if time.Since(ts) < expiredWarnWindow {
				return false
			}
		}
	}

	if err := os.WriteFile(path, []byte(time.Now().UTC().Format(time.RFC3339)), 0600); err != nil {
		return true
	}
	return true
}

func init() {
	hubCmd.AddCommand(cliCredentialCmd)
}
