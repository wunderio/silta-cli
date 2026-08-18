package cmd_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// hubCredentialEnv writes a silta credentials file into a fresh temp directory
// and returns the environment pointing SILTA_CONFIG_DIR at it. Pass an empty
// expiresAt to omit the expires_at field.
func hubCredentialEnv(t *testing.T, expiresAt string) []string {
	t.Helper()
	dir := t.TempDir()
	creds := "hub_url: https://hub.example.com\ntoken: test-token-123\nusername: test-user\n"
	if expiresAt != "" {
		creds += "expires_at: \"" + expiresAt + "\"\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials"), []byte(creds), 0600); err != nil {
		t.Fatalf("failed to write credentials: %v", err)
	}
	return []string{"SILTA_CONFIG_DIR=" + dir}
}

// runCli runs a silta command and returns the exit code, stdout and stderr.
// Non-zero exit codes are returned, not fatal.
func runCli(t *testing.T, command string, environment []string) (int, string, string) {
	t.Helper()
	cmd := exec.Command("bash", "-c", cliBinaryName+" "+command)
	mergedEnv := os.Environ()
	for _, e := range environment {
		mergedEnv = append(mergedEnv, e)
	}
	cmd.Env = mergedEnv
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("failed to run '%s': %v", command, err)
		}
		code = exitErr.ExitCode()
	}
	return code, out.String(), errOut.String()
}

// TestHubCliCredentialValid checks that a healthy session emits a valid
// ExecCredential for kubectl.
func TestHubCliCredentialValid(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir("..")

	future := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	code, out, errOut := runCli(t, "hub cli-credential", hubCredentialEnv(t, future))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{
		`"apiVersion":"client.authentication.k8s.io/v1"`,
		`"kind":"ExecCredential"`,
		`"token":"test-token-123"`,
		`"expirationTimestamp":"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected stdout to contain %s, got: %s", want, out)
		}
	}
}

// TestHubCliCredentialExpired checks that an expired session fails with a
// relogin hint instead of emitting a stale token.
func TestHubCliCredentialExpired(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir("..")

	past := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	env := hubCredentialEnv(t, past)
	code, out, errOut := runCli(t, "hub cli-credential", env)
	if code == 0 {
		t.Fatalf("expected non-zero exit for expired credentials, got 0 (stdout: %q)", out)
	}
	for _, want := range []string{"session has expired", "silta hub login"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("expected stderr to contain %q, got: %q", want, errOut)
		}
	}
	if strings.Contains(out, "ExecCredential") {
		t.Errorf("expected no credential output for expired credentials, got: %q", out)
	}

	// kubectl re-fetched credentials within the warning window (it re-runs this
	// plugin on every discovery retry); the hint must not print again.
	code2, _, errOut2 := runCli(t, "hub cli-credential", env)
	if code2 == 0 {
		t.Fatalf("expected non-zero exit on repeat invocation, got 0")
	}
	if strings.Contains(errOut2, "session has expired") {
		t.Errorf("expected hint to be suppressed on repeat invocation, got: %q", errOut2)
	}
}

// TestHubLoginPersistsHubURL checks that --hub-url is saved to the
// configuration so later 'silta hub login' runs do not need the flag. The
// login itself fails (nothing listens on the test URL), which is fine.
func TestHubLoginPersistsHubURL(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir("..")

	dir := t.TempDir()
	_, _, errOut := runCli(t, "hub login --hub-url http://127.0.0.1:9", []string{"SILTA_CONFIG_DIR=" + dir})

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("expected hub URL to be persisted to config.yaml: %v (stderr: %q)", err, errOut)
	}
	if !strings.Contains(string(data), "http://127.0.0.1:9") {
		t.Errorf("expected config.yaml to contain the hub URL, got: %s", data)
	}
}

// TestHubCliCredentialNotLoggedIn checks the hint when no credentials exist.
func TestHubCliCredentialNotLoggedIn(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir("..")

	code, _, errOut := runCli(t, "hub cli-credential", []string{"SILTA_CONFIG_DIR=" + t.TempDir()})
	if code == 0 {
		t.Fatalf("expected non-zero exit when not logged in")
	}
	want := "not logged in: run 'silta hub login' first"
	if !strings.Contains(errOut, want) {
		t.Errorf("expected stderr to contain %q, got: %q", want, errOut)
	}
}
