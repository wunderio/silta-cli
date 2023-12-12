package cmd

import (
	"context"
	"log"
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		//Uninstall Helm release
		uninstallErr := helmClient.UninstallReleaseByName(releaseName)
		if uninstallErr != nil {
			log.Fatalf("Error removing a release:%s", uninstallErr)
		}

		//Delete pre-release jobs
		selectorLabels := []string{
			"release",
			"app.kubernetes.io/instance",
		}

		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			list, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of jobs: %s", err)
			}
			for _, v := range list.Items {
				log.Printf("Deleting job: %s", v.Name)
				clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		//Delete PVCs
		if deletePVCs {

			PVC_client := clientset.CoreV1().PersistentVolumeClaims(namespace)

			selectorLabels := []string{
				"app",
				"release",
				"app.kubernetes.io/instance",
			}

			for _, l := range selectorLabels {
				selector := l + "=" + releaseName
				if l == "app" {
					selector = l + "=" + releaseName + "-es"
				}
				list, err := PVC_client.List(context.TODO(), v1.ListOptions{
					LabelSelector: selector,
				})
				if err != nil {
					log.Fatalf("Error getting the list of PVCs: %s", err)
				}

				for _, v := range list.Items {
					log.Printf("Deleting PVC: %s", v.Name)
					PVC_client.Delete(context.TODO(), v.Name, v1.DeleteOptions{})
				}
			}
		}

	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseDeleteCmd)

	ciReleaseDeleteCmd.Flags().String("release-name", "", "Release name")
	ciReleaseDeleteCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseDeleteCmd.Flags().Bool("delete-pvcs", false, "Delete PVCs (default: false)")

	ciReleaseDeleteCmd.MarkFlagRequired("release-name")
	ciReleaseDeleteCmd.MarkFlagRequired("namespace")
}
