package cmd_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeHubServer serves GET /api/cli/clusters returning the given JSON payload.
func fakeHubServer(t *testing.T, clustersJSON string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cli/clusters" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, clustersJSON)
	}))
	t.Cleanup(srv.Close)
	return srv
}

const testClustersJSON = `{"username":"test-user","clusters":[
  {"id":"cluster-a","server":"https://hub/api/kube/cluster-a","namespace":"team-a","kubernetesVersion":"v1.30.2"},
  {"id":"cluster-b","server":"https://hub/api/kube/cluster-b","namespace":"","kubernetesVersion":""}
]}`

// TestHubClustersTable verifies the human-readable table output of
// 'silta hub clusters', including empty namespace/version fallbacks.
func TestHubClustersTable(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir("..")

	srv := fakeHubServer(t, testClustersJSON)
	code, out, errOut := runCli(t, "hub clusters", hubCredentialEnv(t, srv.URL, ""))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{
		"silta-cluster-a", "team-a", "v1.30.2",
		"silta-cluster-b", "(none)", "(unknown)",
		"KUBERNETES VERSION",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected stdout to contain %q, got: %s", want, out)
		}
	}
}

// TestHubClustersJSON verifies machine-readable output decodes as the
// documented shape and preserves all fields.
func TestHubClustersJSON(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir("..")

	srv := fakeHubServer(t, testClustersJSON)
	code, out, errOut := runCli(t, "hub clusters --json", hubCredentialEnv(t, srv.URL, ""))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	var resp struct {
		Username string `json:"username"`
		Clusters []struct {
			ID                string `json:"id"`
			Server            string `json:"server"`
			Namespace         string `json:"namespace"`
			KubernetesVersion string `json:"kubernetesVersion"`
		} `json:"clusters"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, out)
	}
	if resp.Username != "test-user" {
		t.Errorf("expected username test-user, got %q", resp.Username)
	}
	if len(resp.Clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(resp.Clusters))
	}
	first := resp.Clusters[0]
	if first.ID != "cluster-a" || first.Namespace != "team-a" || first.KubernetesVersion != "v1.30.2" {
		t.Errorf("unexpected first cluster: %+v", first)
	}
	if resp.Clusters[1].ID != "cluster-b" || resp.Clusters[1].Namespace != "" || resp.Clusters[1].KubernetesVersion != "" {
		t.Errorf("unexpected second cluster: %+v", resp.Clusters[1])
	}
}

// TestHubClustersNotLoggedIn verifies the failure path when no credentials
// are stored.
func TestHubClustersNotLoggedIn(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir("..")

	code, out, errOut := runCli(t, "hub clusters", []string{"SILTA_CONFIG_DIR=" + t.TempDir()})
	if code == 0 {
		t.Fatalf("expected non-zero exit when not logged in, got 0 (stdout: %q)", out)
	}
	if !strings.Contains(errOut, "not logged in") {
		t.Errorf("expected stderr to mention 'not logged in', got: %q", errOut)
	}
}

// TestHubKubeconfigPrunesStale verifies 'silta hub kubeconfig' removes silta-*
// contexts the user can no longer access, while preserving the clusters they
// still have and any unrelated entries.
func TestHubKubeconfigPrunesStale(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir("..")

	kubeDir := t.TempDir()
	kubePath := filepath.Join(kubeDir, "config")
	kubeconfig := `apiVersion: v1
kind: Config
current-context: silta-stale
clusters:
- name: silta-stale
  cluster:
    server: https://stale.example.com
- name: silta-keep
  cluster:
    server: https://keep.example.com
- name: unrelated
  cluster:
    server: https://unrelated.example.com
contexts:
- name: silta-stale
  context:
    cluster: silta-stale
    user: silta-stale
    namespace: old-ns
- name: silta-keep
  context:
    cluster: silta-keep
    user: silta-keep
    namespace: keep-ns
- name: unrelated
  context:
    cluster: unrelated
    user: unrelated
users:
- name: silta-stale
  user:
    token: stale
- name: silta-keep
  user:
    token: keep
- name: unrelated
  user:
    token: other
`
	if err := os.WriteFile(kubePath, []byte(kubeconfig), 0600); err != nil {
		t.Fatalf("failed to write kubeconfig: %v", err)
	}

	// The hub reports access to 'keep' only; 'stale' must be pruned.
	keepJSON := `{"username":"test-user","clusters":[
	  {"id":"keep","server":"https://keep.example.com","namespace":"keep-ns","kubernetesVersion":"v1.31.0"}
	]}`
	srv := fakeHubServer(t, keepJSON)
	env := append(hubCredentialEnv(t, srv.URL, ""), "KUBECONFIG="+kubePath)

	code, out, errOut := runCli(t, "hub kubeconfig", env)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "silta-keep") {
		t.Errorf("expected stdout to list silta-keep, got: %s", out)
	}

	data, err := os.ReadFile(kubePath)
	if err != nil {
		t.Fatalf("failed to read updated kubeconfig: %v", err)
	}
	updated := string(data)
	if strings.Contains(updated, "silta-stale") {
		t.Errorf("expected stale context to be pruned, but it remains:\n%s", updated)
	}
	for _, want := range []string{"silta-keep", "unrelated"} {
		if !strings.Contains(updated, want) {
			t.Errorf("expected %q to be preserved, got:\n%s", want, updated)
		}
	}
	// The current context pointed at the pruned entry and must be cleared.
	if strings.Contains(updated, "current-context: silta-stale") {
		t.Errorf("expected current-context to be cleared after pruning silta-stale, got:\n%s", updated)
	}
}
