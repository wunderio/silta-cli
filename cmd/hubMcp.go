package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var (
	hubMcpClient  string
	hubMcpName    string
	hubMcpScope   string
	hubMcpURLOnly bool
)

// mcpClients are the AI clients 'silta hub mcp' prints setup for, in order.
var mcpClients = []string{"claude", "opencode", "cursor", "vscode"}

var mcpServerNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// hubMcpCmd helps connect AI clients to the Silta dashboard MCP server.
var hubMcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Connect AI clients (Claude Code, opencode, Cursor, VS Code) to the Silta MCP server",
	Long: `Show how to connect AI clients to the Silta dashboard MCP server, which gives
them read-only access to the releases, pods, logs, metrics and traces you can
see in the dashboard.

The AI client logs in by itself: on first use it opens a browser where you sign
in to the dashboard and approve the client. No 'silta hub login' is needed.

Without --client, setup instructions for every supported client are printed.
With --client claude, the server is registered in Claude Code directly when the
'claude' command is installed.

Use --url to print only the MCP endpoint URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		if hubMcpClient != "" && !containsString(mcpClients, hubMcpClient) {
			fmt.Fprintf(os.Stderr, "Error: --client must be one of: %s\n", strings.Join(mcpClients, ", "))
			os.Exit(1)
		}
		if !mcpServerNamePattern.MatchString(hubMcpName) {
			fmt.Fprintln(os.Stderr, "Error: --name may only contain letters, digits, '-' and '_'")
			os.Exit(1)
		}
		if hubMcpScope != "local" && hubMcpScope != "user" && hubMcpScope != "project" {
			fmt.Fprintln(os.Stderr, "Error: --scope must be one of: local, user, project")
			os.Exit(1)
		}

		hubURL := resolveHubURL()
		if hubURL == "" {
			if creds, err := common.LoadCredentials(); err == nil {
				hubURL = creds.HubURL
			}
		}
		if hubURL == "" {
			fmt.Fprintln(os.Stderr, "Error: no Silta hub URL configured. Pass --hub-url or run 'silta config set hub.url <url>'.")
			os.Exit(1)
		}

		mcpURL, err := discoverMCPURL(hubURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		if hubMcpURLOnly {
			fmt.Println(mcpURL)
			return
		}

		if hubMcpClient == "claude" {
			if debug {
				fmt.Printf("Command (not executed): %s\n", claudeAddCommand(hubMcpName, hubMcpScope, mcpURL))
				return
			}
			if err := addToClaudeCode(hubMcpName, hubMcpScope, mcpURL); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				fmt.Fprintf(os.Stderr, "Run it manually:\n  %s\n", claudeAddCommand(hubMcpName, hubMcpScope, mcpURL))
				os.Exit(1)
			}
			fmt.Println("Added. Start Claude Code, run /mcp, select '" + hubMcpName + "' and choose Authenticate.")
			return
		}

		fmt.Printf("Silta MCP endpoint: %s\n", mcpURL)
		fmt.Println("On first use the client opens a browser to sign in to the dashboard and approve it.")
		clients := mcpClients
		if hubMcpClient != "" {
			clients = []string{hubMcpClient}
		}
		for _, client := range clients {
			fmt.Println()
			fmt.Print(mcpClientInstructions(client, hubMcpName, hubMcpScope, mcpURL))
		}
	},
}

// discoverMCPURL returns the MCP endpoint of a hub. Dashboards advertise it as
// mcp_url in /api/cli/info when MCP is enabled; for dashboards that predate
// that field, the OAuth protected resource metadata reveals whether MCP is on.
func discoverMCPURL(hubURL string) (string, error) {
	client := common.NewHubClient(hubURL, "")
	var info struct {
		MCPURL string `json:"mcp_url"`
	}
	if _, err := client.GetJSON("/api/cli/info", &info); err == nil && info.MCPURL != "" {
		if err := validateMCPURL(info.MCPURL); err != nil {
			return "", err
		}
		return info.MCPURL, nil
	}

	base := strings.TrimRight(hubURL, "/")
	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Get(base + "/.well-known/oauth-protected-resource")
	if err != nil {
		return "", fmt.Errorf("failed to reach Silta hub: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("MCP is not enabled on this Silta hub (%s)", base)
	}
	return base + "/mcp", nil
}

// validateMCPURL checks that a hub-advertised MCP endpoint is a plain HTTP(S)
// URL, since it ends up in printed commands and client configuration.
func validateMCPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
		strings.IndexFunc(raw, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return fmt.Errorf("Silta hub advertised an invalid MCP URL: %q", raw)
	}
	return nil
}

var shellSafePattern = regexp.MustCompile(`^[A-Za-z0-9_./:@%+=,-]+$`)

// shellQuote quotes s for a POSIX shell unless it only has safe characters.
func shellQuote(s string) string {
	if shellSafePattern.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// claudeAddCommand is the Claude Code command that registers the server.
func claudeAddCommand(name, scope, mcpURL string) string {
	return fmt.Sprintf("claude mcp add --transport http --scope %s %s %s", shellQuote(scope), shellQuote(name), shellQuote(mcpURL))
}

// addToClaudeCode registers the MCP server with the claude CLI.
func addToClaudeCode(name, scope, mcpURL string) error {
	path, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("the 'claude' command was not found in PATH")
	}
	c := exec.Command(path, "mcp", "add", "--transport", "http", "--scope", scope, name, mcpURL)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("'claude mcp add' failed: %w", err)
	}
	return nil
}

// mcpClientInstructions returns the setup instructions for one client.
func mcpClientInstructions(client, name, scope, mcpURL string) string {
	switch client {
	case "claude":
		return fmt.Sprintf(`Claude Code:
  %s
  Then run /mcp in Claude Code, select '%s' and choose Authenticate.
  (Or let this CLI do it: silta hub mcp --client claude)
`, claudeAddCommand(name, scope, mcpURL), name)
	case "opencode":
		return fmt.Sprintf(`opencode (opencode.json):
  {
    "mcp": {
      "%s": { "type": "remote", "url": "%s" }
    }
  }
  Then run: opencode mcp auth %s
`, name, mcpURL, name)
	case "cursor":
		return fmt.Sprintf(`Cursor (~/.cursor/mcp.json):
  { "mcpServers": { "%s": { "url": "%s" } } }
`, name, mcpURL)
	case "vscode":
		return fmt.Sprintf(`VS Code (.vscode/mcp.json):
  { "servers": { "%s": { "type": "http", "url": "%s" } } }
`, name, mcpURL)
	}
	return ""
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func init() {
	hubMcpCmd.Flags().StringVar(&hubMcpClient, "client", "", "Only set up this client: claude, opencode, cursor or vscode (claude registers the server directly)")
	hubMcpCmd.Flags().StringVar(&hubMcpName, "name", "silta", "Server name to use in the client configuration")
	hubMcpCmd.Flags().StringVar(&hubMcpScope, "scope", "user", "Claude Code configuration scope: local, user or project")
	hubMcpCmd.Flags().BoolVar(&hubMcpURLOnly, "url", false, "Print only the MCP endpoint URL")
	hubMcpCmd.Flags().StringVar(&loginHubURL, "hub-url", "", "Silta hub URL (defaults to the configured hub.url)")
	hubCmd.AddCommand(hubMcpCmd)
}
