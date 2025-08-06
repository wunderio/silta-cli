package cmd

import (
	"log"
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth provider
	"k8s.io/client-go/tools/clientcmd"
)

var ciReleaseDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a release",
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")
		deletePVCs, _ := cmd.Flags().GetBool("delete-pvcs")

		// Try reading KUBECONFIG from environment variable first
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			// If not set, use the default kube config path
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("cannot read user home dir")
			}
			kubeConfigPath = homeDir + "/.kube/config"

		}

		// Read kubeConfig from file
		if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
			log.Fatalf("kubeConfig file does not exist at path: %s", kubeConfigPath)
		}

		// Read kubeConfig file
		kubeConfig, err := os.ReadFile(kubeConfigPath)
		if err != nil {
			log.Fatalf("cannot read kubeConfig from path: %s", kubeConfigPath)
		}

		//k8s go client init logic
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			log.Fatalf("cannot read kubeConfig from path: %s", err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatalf("cannot initialize k8s client: %s", err)
		}

		//Helm client init logic
		opt := &helmclient.KubeConfClientOptions{
			Options: &helmclient.Options{
				Namespace:        namespace,
				RepositoryCache:  "/tmp/.helmcache",
				RepositoryConfig: "/tmp/.helmrepo",
				Debug:            false,
				Linting:          false, // Change this to false if you don't want linting.
			},
			KubeContext: "",
			KubeConfig:  kubeConfig,
		}

		helmClient, err := helmclient.NewClientFromKubeConf(opt)
		if err != nil {
			log.Fatalf("Cannot create client from kubeConfig")
		}

		err = common.UninstallHelmRelease(clientset, helmClient, releaseName, namespace, deletePVCs)
		if err != nil {
			log.Fatalf("Error removing a release: %s", err)
		}

	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseDeleteCmd)

	ciReleaseDeleteCmd.Flags().String("release-name", "", "Release name")
	ciReleaseDeleteCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseDeleteCmd.Flags().Bool("delete-pvcs", true, "Delete PVCs (default: true)")

	ciReleaseDeleteCmd.MarkFlagRequired("release-name")
	ciReleaseDeleteCmd.MarkFlagRequired("namespace")
}
