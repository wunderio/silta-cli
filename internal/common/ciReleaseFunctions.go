package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
	uninstall.Timeout = 300 * time.Second // seconds, adjust as needed
	uninstall.Wait = true                 // Wait for resources to be deleted

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

func DeleteOrphanedReleaseResources(kubernetesClient *kubernetes.Clientset, helmClient *helmAction.Configuration, namespace string, releaseName string, deletePVCs bool, dryRun bool) error {
	// Select related resources by label selectors
	selectorLabels := []string{
		"release",
		"app.kubernetes.io/instance",
		"app" + "=" + releaseName + "-es",
	}

	for _, l := range selectorLabels {
		selector := l + "=" + releaseName
		// Delete deployments
		dpList, _ := kubernetesClient.AppsV1().Deployments(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range dpList.Items {
			if dryRun {
				fmt.Printf("Dry run: deployment/%s\n", v.Name)
			} else {
				fmt.Printf("Removing deployment/%s\n", v.Name)
				kubernetesClient.AppsV1().Deployments(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete statefulsets
		stsList, _ := kubernetesClient.AppsV1().StatefulSets(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range stsList.Items {
			if dryRun {
				fmt.Printf("Dry run: statefulset/%s\n", v.Name)
			} else {
				log.Printf("Removing statefulset/%s\n", v.Name)
				kubernetesClient.AppsV1().StatefulSets(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete cronjobs
		cjList, _ := kubernetesClient.BatchV1().CronJobs(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range cjList.Items {
			if dryRun {
				fmt.Printf("Dry run: cronjob/%s\n", v.Name)
			} else {
				fmt.Printf("Removing cronjob/%s\n", v.Name)
				kubernetesClient.BatchV1().CronJobs(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete jobs
		jobList, _ := kubernetesClient.BatchV1().Jobs(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range jobList.Items {
			if dryRun {
				fmt.Printf("Dry run: job/%s\n", v.Name)
			} else {
				fmt.Printf("Removing job/%s\n", v.Name)
				kubernetesClient.BatchV1().Jobs(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete horizontal pod autoscalers
		hpaList, _ := kubernetesClient.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range hpaList.Items {
			if dryRun {
				fmt.Printf("Dry run: horizontalpodautoscaler/%s\n", v.Name)
			} else {
				fmt.Printf("Removing horizontalpodautoscaler/%s\n", v.Name)
				kubernetesClient.AutoscalingV1().HorizontalPodAutoscalers(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete ingresses
		ingressList, _ := kubernetesClient.NetworkingV1().Ingresses(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range ingressList.Items {
			if dryRun {
				fmt.Printf("Dry run: ingress/%s\n", v.Name)
			} else {
				fmt.Printf("Removing ingress/%s\n", v.Name)
				kubernetesClient.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete pods
		podList, _ := kubernetesClient.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range podList.Items {
			if dryRun {
				fmt.Printf("Dry run: pod/%s\n", v.Name)
			} else {
				fmt.Printf("Removing pod/%s\n", v.Name)
				kubernetesClient.CoreV1().Pods(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete services
		svcList, _ := kubernetesClient.CoreV1().Services(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range svcList.Items {
			if dryRun {
				fmt.Printf("Dry run: service/%s\n", v.Name)
			} else {
				fmt.Printf("Removing service/%s\n", v.Name)
				kubernetesClient.CoreV1().Services(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete backendconfigs
		bcList, _ := kubernetesClient.NetworkingV1().Ingresses(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range bcList.Items {
			if dryRun {
				fmt.Printf("Dry run: backendconfig/%s\n", v.Name)
			} else {
				fmt.Printf("Removing backendconfig/%s\n", v.Name)
				kubernetesClient.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete configmaps
		cmList, _ := kubernetesClient.CoreV1().ConfigMaps(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range cmList.Items {
			if dryRun {
				fmt.Printf("Dry run: configmap/%s\n", v.Name)
			} else {
				fmt.Printf("Removing configmap/%s\n", v.Name)
				kubernetesClient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete secrets
		secretList, _ := kubernetesClient.CoreV1().Secrets(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range secretList.Items {
			if dryRun {
				fmt.Printf("Dry run: secret/%s\n", v.Name)
			} else {
				fmt.Printf("Removing secret/%s\n", v.Name)
				kubernetesClient.CoreV1().Secrets(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete cerficates
		certList, _ := kubernetesClient.CoreV1().Secrets(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range certList.Items {
			if dryRun {
				fmt.Printf("Dry run: certificate/%s\n", v.Name)
			} else {
				fmt.Printf("Removing certificate/%s\n", v.Name)
				kubernetesClient.CoreV1().Secrets(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete persistent volume claims
		if deletePVCs {
			pvcList, _ := kubernetesClient.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			for _, v := range pvcList.Items {
				if dryRun {
					fmt.Printf("Dry run: persistentvolumeclaim/%s\n", v.Name)
				} else {
					fmt.Printf("Removing persistentvolumeclaim/%s\n", v.Name)
					kubernetesClient.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
				}
			}
		}

		// Delete persistent volumes
		if deletePVCs {
			pvList, _ := kubernetesClient.CoreV1().PersistentVolumes().List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			for _, v := range pvList.Items {
				if dryRun {
					fmt.Printf("Dry run: persistentvolume/%s\n", v.Name)
				} else {
					fmt.Printf("Removing persistentvolume/%s\n", v.Name)
					kubernetesClient.CoreV1().PersistentVolumes().Delete(context.TODO(), v.Name, v1.DeleteOptions{})
					time.Sleep(1 * time.Second)
				}
			}
		}

		// Delete network policies
		npList, _ := kubernetesClient.NetworkingV1().NetworkPolicies(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range npList.Items {
			if dryRun {
				fmt.Printf("Dry run: networkpolicy/%s\n", v.Name)
				continue
			} else {
				fmt.Printf("Removing networkpolicy/%s\n", v.Name)
				kubernetesClient.NetworkingV1().NetworkPolicies(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete rolebindings
		rbList, _ := kubernetesClient.RbacV1().RoleBindings(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range rbList.Items {
			if dryRun {
				fmt.Printf("Dry run: rolebinding/%s\n", v.Name)
			} else {
				fmt.Printf("Removing rolebinding/%s\n", v.Name)
				kubernetesClient.RbacV1().RoleBindings(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete roles
		roleList, _ := kubernetesClient.RbacV1().Roles(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range roleList.Items {
			if dryRun {
				fmt.Printf("Dry run: role/%s\n", v.Name)
			} else {
				fmt.Printf("Removing role/%s\n", v.Name)
				kubernetesClient.RbacV1().Roles(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}

		// Delete serviceaccounts
		saList, _ := kubernetesClient.CoreV1().ServiceAccounts(namespace).List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		for _, v := range saList.Items {
			if dryRun {
				fmt.Printf("Dry run: serviceaccount/%s\n", v.Name)
			} else {
				fmt.Printf("Removing serviceaccount/%s\n", v.Name)
				kubernetesClient.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), v.Name, v1.DeleteOptions{})
			}
		}
	}

	return nil
}
