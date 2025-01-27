package common

import (
	"context"
	"log"

	helmclient "github.com/mittwald/go-helm-client"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// UninstallRelease removes a Helm release and related resources
// Note: namespace is inferred from the helmclient.Options struct but set here for kubernetes clientset actions
func UninstallHelmRelease(clientset *kubernetes.Clientset, helmClient helmclient.Client, releaseName string, namespace string, deletePVCs bool) error {

	// Uninstall helm release. Namespace and other context is provided via the
	// helmclient.Options struct when instantiating a client.
	// Do not bail when release removal fails, remove related resources anyway.
	err := helmClient.UninstallReleaseByName(releaseName)
	if err != nil {
		log.Printf("Failed to remove helm release: %s", err)
	}

	// Delete related jobs
	selectorLabels := []string{
		"release",
		"app.kubernetes.io/instance",
	}

	for _, l := range selectorLabels {
		selector := l + "=" + releaseName
		list, _ := clientset.BatchV1().Jobs(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range list.Items {
			log.Printf("Removing job: %s", v.Name)
			propagationPolicy := v1.DeletePropagationBackground
			clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{PropagationPolicy: &propagationPolicy})
		}
	}

	if deletePVCs {

		// Find and remove related PVC's by release name label
		PVC_client := clientset.CoreV1().PersistentVolumeClaims(namespace)

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
