package cmd_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// fakeMcpHubServer serves /api/cli/info (with mcp_url when advertise is true)
// and answers the OAuth protected resource metadata with wellKnownStatus.
func fakeMcpHubServer(t *testing.T, advertise bool, wellKnownStatus int) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/cli/info":
			w.Header().Set("Content-Type", "application/json")
			if advertise {
				fmt.Fprintf(w, `{"frontend_url":"https://dash","mcp_url":"%s/mcp"}`, srv.URL)
			} else {
				fmt.Fprint(w, `{"frontend_url":"https://dash"}`)
			}
		case "/.well-known/oauth-protected-resource":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(wellKnownStatus)
			fmt.Fprint(w, `{}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// mcpTestEnv returns an environment with an empty silta config directory.
func mcpTestEnv(t *testing.T) []string {
	t.Helper()
	return []string{"SILTA_CONFIG_DIR=" + t.TempDir()}
}

func chdirRepoRoot(t *testing.T) {
	t.Helper()
	wd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(wd) })
	os.Chdir("..")
}

func TestHubMcpAllClients(t *testing.T) {
	chdirRepoRoot(t)
	srv := fakeMcpHubServer(t, true, http.StatusOK)

	code, out, errOut := runCli(t, "hub mcp --hub-url "+srv.URL, mcpTestEnv(t))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %q)", code, errOut)
	}
	mcpURL := srv.URL + "/mcp"
	for _, want := range []string{
		"Silta MCP endpoint: " + mcpURL,
		"claude mcp add --transport http --scope user silta " + mcpURL,
		"opencode mcp auth silta",
		"~/.cursor/mcp.json",
		".vscode/mcp.json",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestHubMcpURLOnly(t *testing.T) {
	chdirRepoRoot(t)
	srv := fakeMcpHubServer(t, true, http.StatusOK)

	code, out, errOut := runCli(t, "hub mcp --url --hub-url "+srv.URL, mcpTestEnv(t))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %q)", code, errOut)
	}
	if strings.TrimSpace(out) != srv.URL+"/mcp" {
		t.Errorf("expected only the URL, got %q", out)
	}
}

func TestHubMcpSingleClientCustomName(t *testing.T) {
	chdirRepoRoot(t)
	srv := fakeMcpHubServer(t, true, http.StatusOK)

	code, out, errOut := runCli(t, "hub mcp --client opencode --name silta-dev --hub-url "+srv.URL, mcpTestEnv(t))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %q)", code, errOut)
	}
	if !strings.Contains(out, `"silta-dev": { "type": "remote"`) || !strings.Contains(out, "opencode mcp auth silta-dev") {
		t.Errorf("expected opencode instructions with custom name, got:\n%s", out)
	}
	if strings.Contains(out, "Cursor") || strings.Contains(out, "Claude Code") {
		t.Errorf("expected only the opencode block, got:\n%s", out)
	}
}

func TestHubMcpFallbackForOlderDashboard(t *testing.T) {
	chdirRepoRoot(t)
	srv := fakeMcpHubServer(t, false, http.StatusOK)

	code, out, errOut := runCli(t, "hub mcp --url --hub-url "+srv.URL, mcpTestEnv(t))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %q)", code, errOut)
	}
	if strings.TrimSpace(out) != srv.URL+"/mcp" {
		t.Errorf("expected fallback URL, got %q", out)
	}
}

func TestHubMcpDisabled(t *testing.T) {
	chdirRepoRoot(t)
	srv := fakeMcpHubServer(t, false, http.StatusNotFound)

	code, _, errOut := runCli(t, "hub mcp --hub-url "+srv.URL, mcpTestEnv(t))
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "MCP is not enabled") {
		t.Errorf("expected 'not enabled' message, got %q", errOut)
	}
}

func TestHubMcpInvalidFlags(t *testing.T) {
	chdirRepoRoot(t)
	srv := fakeMcpHubServer(t, true, http.StatusOK)

	for _, args := range []string{
		"--client emacs",
		"--name 'bad name'",
		"--scope global",
	} {
		code, _, errOut := runCli(t, "hub mcp "+args+" --hub-url "+srv.URL, mcpTestEnv(t))
		if code != 1 || !strings.Contains(errOut, "Error:") {
			t.Errorf("%s: expected exit 1 with an error, got %d (%q)", args, code, errOut)
		}
	}
}

func TestHubMcpClaudeNotInstalled(t *testing.T) {
	chdirRepoRoot(t)
	srv := fakeMcpHubServer(t, true, http.StatusOK)

	// An empty PATH hides any installed 'claude' command; the binary itself
	// is started by relative path.
	env := append(mcpTestEnv(t), "PATH="+t.TempDir())
	code, _, errOut := runCli(t, "hub mcp --client claude --hub-url "+srv.URL, env)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "claude mcp add --transport http --scope user silta "+srv.URL+"/mcp") {
		t.Errorf("expected the manual command, got %q", errOut)
	}
}
