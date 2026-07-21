package cmd

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var (
	loginHubURL string
	loginDevice bool
)

// hubLoginCmd authenticates the CLI against a Silta hub and stores a token.
var hubLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to a Silta hub",
	Long: `Authenticate the Silta CLI against a Silta hub.

By default a browser window is opened to approve the login. On headless
machines use --device to complete the login using a short code instead.

The Silta hub URL can be provided with --hub-url or stored in the
configuration under 'hub.url'.`,
	Run: func(cmd *cobra.Command, args []string) {
		hubURL := resolveHubURL()
		if hubURL == "" {
			fmt.Println("Error: no Silta hub URL configured. Pass --hub-url or run 'silta config set hub.url <url>'.")
			return
		}

		client := common.NewHubClient(hubURL, "")

		var token *cliTokenResult
		var err error
		if loginDevice {
			token, err = runDeviceLogin(client)
		} else {
			token, err = runBrowserLogin(client, hubURL)
		}
		if err != nil {
			fmt.Printf("Login failed: %s\n", err)
			return
		}

		creds := &common.Credentials{
			HubURL:    strings.TrimRight(hubURL, "/"),
			Token:     token.Token,
			ExpiresAt: token.ExpiresAt,
			Username:  token.Username,
		}
		if err := common.SaveCredentials(creds); err != nil {
			fmt.Printf("Failed to store credentials: %s\n", err)
			return
		}

		fmt.Printf("Logged in as %s.\n", token.Username)

		// Update kubeconfig with the clusters the user can access.
		if err := syncKubeconfig(creds); err != nil {
			fmt.Printf("Warning: logged in but failed to update kubeconfig: %s\n", err)
		}
	},
}

// syncKubeconfig fetches the user's authorized clusters and merges them into the
// local kubeconfig as silta-<cluster> contexts.
func syncKubeconfig(creds *common.Credentials) error {
	client := common.NewHubClient(creds.HubURL, creds.Token)

	var resp struct {
		Clusters []common.SiltaCluster `json:"clusters"`
	}
	if _, err := client.GetJSON("/api/cli/clusters", &resp); err != nil {
		return err
	}

	if len(resp.Clusters) == 0 {
		fmt.Println("No cluster access is currently assigned to your account.")
		return nil
	}

	path, err := common.UpdateKubeconfig(resp.Clusters)
	if err != nil {
		return err
	}

	fmt.Printf("Updated kubeconfig (%s) with %d cluster context(s):\n", path, len(resp.Clusters))
	for _, cluster := range resp.Clusters {
		fmt.Printf("  silta-%s\n", cluster.ID)
	}
	return nil
}

// cliTokenResult mirrors the hub token response.
type cliTokenResult struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	Username  string `json:"username"`
}

// resolveHubURL determines the hub URL from the flag or config.
func resolveHubURL() string {
	if loginHubURL != "" {
		return loginHubURL
	}
	configStore := common.ConfigStore()
	if v := configStore.GetString("hub.url"); v != "" {
		return v
	}
	return ""
}

// randomState returns a URL-safe random string used to guard the loopback flow.
func randomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// runBrowserLogin performs the browser loopback flow.
func runBrowserLogin(client *common.HubClient, hubURL string) (*cliTokenResult, error) {
	// Discover the frontend URL used to build the approval page link.
	var info struct {
		FrontendURL string `json:"frontend_url"`
	}
	if _, err := client.GetJSON("/api/cli/info", &info); err != nil {
		return nil, fmt.Errorf("failed to reach Silta hub: %w", err)
	}
	if info.FrontendURL == "" {
		return nil, fmt.Errorf("Silta hub did not advertise a frontend URL")
	}

	state, err := randomState()
	if err != nil {
		return nil, err
	}

	// Start a loopback server on a random local port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	defer listener.Close()

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", listener.Addr().(*net.TCPAddr).Port)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv := &http.Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch on callback")
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			errCh <- fmt.Errorf("no code returned")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Silta CLI login complete</h2><p>You can close this window and return to your terminal.</p></body></html>")
		codeCh <- code
	})
	srv.Handler = mux

	go srv.Serve(listener)
	defer srv.Shutdown(context.Background())

	approvalURL := fmt.Sprintf("%s/cli-login?redirect_uri=%s&state=%s",
		strings.TrimRight(info.FrontendURL, "/"),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
	)

	fmt.Println("Opening your browser to approve the login...")
	fmt.Printf("If it does not open automatically, visit:\n  %s\n", approvalURL)
	_ = common.OpenBrowser(approvalURL)

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return nil, err
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("timed out waiting for browser approval")
	}

	var token cliTokenResult
	if _, err := client.PostJSON("/api/cli/auth/exchange", map[string]string{"code": code}, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// runDeviceLogin performs the device grant flow.
func runDeviceLogin(client *common.HubClient) (*cliTokenResult, error) {
	var start struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		Interval        int    `json:"interval"`
		ExpiresIn       int    `json:"expires_in"`
	}
	if _, err := client.PostJSON("/api/cli/auth/start", map[string]string{}, &start); err != nil {
		return nil, fmt.Errorf("failed to start device login: %w", err)
	}

	fmt.Printf("To authorize this device, visit:\n  %s\n", start.VerificationURI)
	fmt.Printf("And enter the code: %s\n\n", start.UserCode)

	interval := start.Interval
	if interval <= 0 {
		interval = 5
	}
	deadline := time.Now().Add(time.Duration(start.ExpiresIn) * time.Second)
	if start.ExpiresIn <= 0 {
		deadline = time.Now().Add(10 * time.Minute)
	}

	for time.Now().Before(deadline) {
		time.Sleep(time.Duration(interval) * time.Second)

		var token cliTokenResult
		status, err := client.PostJSON("/api/cli/auth/poll", map[string]string{"device_code": start.DeviceCode}, &token)
		if err != nil && status != http.StatusAccepted && status != 0 {
			// 410 (expired) and other 4xx return an error message.
			if status == http.StatusGone {
				return nil, fmt.Errorf("authorization expired, please try again")
			}
			return nil, err
		}
		if status == http.StatusOK && token.Token != "" {
			return &token, nil
		}
		// status == 202: still pending, keep polling.
	}

	return nil, fmt.Errorf("timed out waiting for approval")
}

func init() {
	hubLoginCmd.Flags().StringVar(&loginHubURL, "hub-url", "", "Silta hub URL (overrides config 'hub.url')")
	hubLoginCmd.Flags().BoolVar(&loginDevice, "device", false, "Use device code flow instead of opening a browser")
	hubCmd.AddCommand(hubLoginCmd)
}
