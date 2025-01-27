package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var ciReleaseCleanfailedCmd = &cobra.Command{
	Use:   "clean-failed",
	Short: "Clean failed releases",
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")

		// ----

		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("cannot read user home dir")
		}
		kubeConfigPath := homeDir + "/.kube/config"

		kubeConfig, err := os.ReadFile(kubeConfigPath)
		if err != nil {
			log.Fatalf("cannot read kubeConfig from path")
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

		// Get release info
		release, err := helmClient.GetRelease(releaseName)
		if err != nil {
			return // Release not found or there was an error
		}

		// Check if there's only one revision and it's failed
		if release.Version == 1 && release.Info.Status == "failed" {

			fmt.Println("Removing failed first release.")

			// Remove release
			common.UninstallHelmRelease(clientset, helmClient, releaseName, namespace, true)
		}

		// Workaround for previous Helm release stuck in pending state
		// This is a workaround for a known issue with Helm where a release can get stuck in a pending-upgrade state
		// and the secret is not deleted. This is a workaround to delete the secret if it exists.
		if release.Info.Status == "pending-upgrade" {
			secretName := fmt.Sprintf("%s.%s.v%d", releaseName, namespace, release.Version)
			if err == nil {
				fmt.Printf("Deleting secret %s\n", secretName)
				err := clientset.CoreV1().Secrets(namespace).Delete(context.TODO(), secretName, v1.DeleteOptions{})
				if err != nil {
					log.Fatalf("Error deleting secret %s: %s", secretName, err)
				}
			}
		}
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseCleanfailedCmd)

	ciReleaseCleanfailedCmd.Flags().String("release-name", "", "Release name")
	ciReleaseCleanfailedCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")

	ciReleaseCleanfailedCmd.MarkFlagRequired("release-name")
	ciReleaseCleanfailedCmd.MarkFlagRequired("namespace")
}
