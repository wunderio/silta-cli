package common

import (
	"context"
	"fmt"
	"log"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	helmAction "helm.sh/helm/v3/pkg/action"
	helmCli "helm.sh/helm/v3/pkg/cli"
	helmRelease "helm.sh/helm/v3/pkg/release"
)

// UninstallRelease removes a Helm release and related resources
func UninstallHelmRelease(kubernetesClient *kubernetes.Clientset, helmClient *helmAction.Configuration, namespace string, releaseName string, deletePVCs bool) error {

	// Do not bail when release removal fails, remove related resources anyway.
	log.Printf("Removing release: %s", releaseName)
	uninstall := helmAction.NewUninstall(helmClient)
	uninstall.KeepHistory = false // Remove release secrets as well
	uninstall.DisableHooks = false
	uninstall.Timeout = 300 // seconds, adjust as needed
	uninstall.Wait = true   // Wait for resources to be deleted

	resp, err := uninstall.Run(releaseName)
	if err != nil {
		log.Printf("Failed to remove helm release: %s", err)
	} else {
		if resp != nil && resp.Info != "" {
			log.Printf("Helm uninstall info: %s", resp.Info)
		}
	}

	// Delete related jobs
	selectorLabels := []string{
		"release",
		"app.kubernetes.io/instance",
	}

	for _, l := range selectorLabels {
		selector := l + "=" + releaseName
		list, _ := kubernetesClient.BatchV1().Jobs(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range list.Items {
			log.Printf("Removing job: %s", v.Name)
			propagationPolicy := v1.DeletePropagationBackground
			kubernetesClient.BatchV1().Jobs(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{PropagationPolicy: &propagationPolicy})
		}
	}

	if deletePVCs {

		// Find and remove related PVC's by release name label
		PVC_client := kubernetesClient.CoreV1().PersistentVolumeClaims(namespace)

		selectorLabels = []string{
			"app",
			"release",
			"app.kubernetes.io/instance",
		}

		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			if l == "app" {
				selector = l + "=" + releaseName + "-es"
			}
			list, _ := PVC_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})

			for _, v := range list.Items {
				log.Printf("Removing PVC: %s", v.Name)
				PVC_client.Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}
	}

	return nil
}

func FailedReleaseCleanup(releaseName string, namespace string) {

	// Helm client init logic
	settings := helmCli.New()
	settings.SetNamespace(namespace) // Ensure Helm uses the correct namespace

	actionConfig := new(helmAction.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	getRelease := helmAction.NewGet(actionConfig)
	release, err := getRelease.Run(releaseName)
	if err != nil {
		return // Release not found or there was an error
	}

	// Check if there's only one revision and it's failed
	if release.Version == 1 && release.Info.Status == helmRelease.StatusFailed {
		fmt.Println("Removing failed first release.")

		uninstall := helmAction.NewUninstall(actionConfig)
		_, err := uninstall.Run(releaseName)
		if err != nil {
			log.Fatalf("failed to uninstall release: %v", err)
		}
	}

	// Workaround for previous Helm release stuck in pending-upgrade state
	if release.Info.Status == helmRelease.StatusPendingUpgrade {

		clientset, err := GetKubeClient()
		if err != nil {
			log.Fatalf("failed to get kube client: %v", err)
		}

		secretName := fmt.Sprintf("sh.helm.release.v1.%s.v%d", releaseName, release.Version)
		fmt.Printf("Deleting secret %s\n", secretName)
		err = clientset.CoreV1().Secrets(namespace).Delete(context.TODO(), secretName, v1.DeleteOptions{})
		if err != nil {
			log.Fatalf("Error deleting secret %s: %s", secretName, err)
		}
	}
}
