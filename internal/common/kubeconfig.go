package common

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// SiltaCluster describes a hub-proxied cluster the CLI can access.
type SiltaCluster struct {
	ID        string `json:"id"`
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

// contextName returns the kubeconfig context/cluster/user name for a cluster id.
func contextName(clusterID string) string {
	return "silta-" + clusterID
}

// UpdateKubeconfig merges silta-<cluster> contexts into the user's kubeconfig,
// each authenticated via the 'silta cli-credential' exec plugin. Existing
// unrelated entries are preserved.
func UpdateKubeconfig(clusters []SiltaCluster) (string, error) {
	pathOptions := clientcmd.NewDefaultPathOptions()
	config, err := pathOptions.GetStartingConfig()
	if err != nil {
		return "", err
	}

	self, err := os.Executable()
	if err != nil || self == "" {
		// Fall back to the command name resolved from PATH.
		self = "silta"
	}

	for _, cluster := range clusters {
		name := contextName(cluster.ID)

		config.Clusters[name] = &clientcmdapi.Cluster{
			Server: cluster.Server,
		}

		config.AuthInfos[name] = &clientcmdapi.AuthInfo{
			Exec: &clientcmdapi.ExecConfig{
				APIVersion:      "client.authentication.k8s.io/v1",
				Command:         self,
				Args:            []string{"hub", "cli-credential"},
				InteractiveMode: clientcmdapi.IfAvailableExecInteractiveMode,
			},
		}

		config.Contexts[name] = &clientcmdapi.Context{
			Cluster:   name,
			AuthInfo:  name,
			Namespace: cluster.Namespace,
		}
	}

	if err := clientcmd.ModifyConfig(pathOptions, *config, true); err != nil {
		return "", fmt.Errorf("failed to update kubeconfig: %w", err)
	}

	return pathOptions.GetDefaultFilename(), nil
}
