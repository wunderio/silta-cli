package common

import (
	"errors"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeClient() (*kubernetes.Clientset, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.New("cannot read user home dir")
	}
	kubeConfigPath := homeDir + "/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		// Fall back to in-cluster config
		// use token at /var/run/secrets/kubernetes.io/serviceaccount/token
		// KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined
		config, err = rest.InClusterConfig()
		if err != nil {
			// Still fails, might as well trigger panic() to fail pod
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
