package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra" // k8s errors and handling
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth provider
	"k8s.io/client-go/tools/clientcmd"
)

var ciReleaseWakeupCmd = &cobra.Command{
	Use:   "wakeup",
	Short: "Wake up a downscaled release",
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("cannot read user home dir")
		}
		kubeConfigPath := homeDir + "/.kube/config"

		kubeConfig, err := os.ReadFile(kubeConfigPath)
		if err != nil {
			log.Fatalf("cannot read kubeConfig from path")
		}

		// k8s go client init logic
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

		// Check if release exists
		_, err = helmClient.GetRelease(releaseName)
		if err != nil {
			if err.Error() == "release: not found" {
				log.Fatalf("Release not found")
			} else {
				log.Fatalf("Cannot get release: %s", err)
			}
		}

		selectorLabels := []string{
			"release",
			"app.kubernetes.io/instance",
		}

		// Restore deployments to original state
		// Select all deployments with label "release: releaseName" and "app.kubernetes.io/instance: releaseName"
		// Restore replicas to original state if found
		log.Println("Scaling deployments")
		deployment_client := clientset.AppsV1().Deployments(namespace)
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			deployment_list, err := deployment_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of deployments: %s", err)
			}

			for _, v := range deployment_list.Items {

				// Load the deployment
				r, err := deployment_client.Get(context.TODO(), v.Name, v1.GetOptions{})
				if err != nil {
					log.Fatalf("Error getting the deployment: %s", err)
				}

				replicas := r.Spec.Replicas
				originalReplicas := r.Annotations["auto-downscale/original-replicas"]
				if originalReplicas == "" {
					originalReplicas = "1"
				}

				if *replicas == 0 && originalReplicas != string(*replicas) {
					log.Printf("Scaling %s to %s replica(s)", r.Name, originalReplicas)

					// Set replicas to originalReplicas
					originalReplicasInt64, err := strconv.ParseInt(originalReplicas, 10, 64)
					if err != nil {
						log.Fatalf("Error converting originalReplicas to int: %s", err)
					}
					originalReplicasInt32 := int32(originalReplicasInt64)
					r.Spec.Replicas = &originalReplicasInt32

					// Unset "auto-downscale/original-replicas" annotation
					delete(r.Annotations, "auto-downscale/original-replicas")

					deployment_client.Update(context.TODO(), r, v1.UpdateOptions{
						TypeMeta:     v1.TypeMeta{},
						DryRun:       nil,
						FieldManager: "",
					})
				}
			}
		}

		// Wait for at least one replica to be ready
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			deployment_list, err := deployment_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of deployments: %s", err)
			}

			for _, v := range deployment_list.Items {

				// Wait up to 2 minutes
				timeout := 120
				// Loop until at least one replica is ready
				for {
					r, err := deployment_client.Get(context.TODO(), v.Name, v1.GetOptions{})
					if err != nil {
						log.Fatalf("Error getting the statefulset: %s", err)
					}
					if r.Status.ReadyReplicas == *r.Spec.Replicas || r.Status.ReadyReplicas > 0 {
						break
					}
					// wait 5 seconds
					time.Sleep(5 * time.Second)
					timeout = timeout - 5

					if timeout <= 0 {
						log.Fatalf("Timeout waiting for %s to be ready", r.Name)
					}
				}

			}
		}

		// Restore statefulset to original state
		// Select all statefulset with label "release: releaseName" and "app.kubernetes.io/instance: releaseName"
		// Restore replicas to original state if found
		log.Println("Scaling statefulsets")
		statefulset_client := clientset.AppsV1().StatefulSets(namespace)
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			statefulset_list, err := statefulset_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of statefulsets: %s", err)
			}

			for _, v := range statefulset_list.Items {

				// Load the statefulset
				r, err := statefulset_client.Get(context.TODO(), v.Name, v1.GetOptions{})
				if err != nil {
					log.Fatalf("Error getting the statefulset: %s", err)
				}

				replicas := r.Spec.Replicas
				originalReplicas := r.Annotations["auto-downscale/original-replicas"]
				if originalReplicas == "" {
					originalReplicas = "1"
				}

				if *replicas == 0 && originalReplicas != string(*replicas) {
					log.Printf("Scaling %s to %s replica(s)", r.Name, originalReplicas)

					// Set replicas to originalReplicas
					originalReplicasInt64, err := strconv.ParseInt(originalReplicas, 10, 64)
					if err != nil {
						log.Fatalf("Error converting originalReplicas to int: %s", err)
					}
					originalReplicasInt32 := int32(originalReplicasInt64)
					r.Spec.Replicas = &originalReplicasInt32

					// Unset "auto-downscale/original-replicas" annotation
					delete(r.Annotations, "auto-downscale/original-replicas")

					statefulset_client.Update(context.TODO(), r, v1.UpdateOptions{
						TypeMeta:     v1.TypeMeta{},
						DryRun:       nil,
						FieldManager: "",
					})
				}
			}
		}

		// Wait for at least one replica to be ready
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			statefulset_list, err := statefulset_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of statefulsets: %s", err)
			}

			for _, v := range statefulset_list.Items {

				// Wait up to 2 minutes
				timeout := 120
				// Loop until at least one replica is ready
				for {
					r, err := statefulset_client.Get(context.TODO(), v.Name, v1.GetOptions{})
					if err != nil {
						log.Fatalf("Error getting the statefulset: %s", err)
					}
					if r.Status.ReadyReplicas == *r.Spec.Replicas || r.Status.ReadyReplicas > 0 {
						break
					}
					// wait 5 seconds
					time.Sleep(5 * time.Second)
					timeout = timeout - 5

					if timeout <= 0 {
						log.Fatalf("Timeout waiting for %s to be ready", r.Name)
					}
				}
			}
		}

		// Unpause cronjobs
		// Select all cronjobs with label "release: releaseName" and "app.kubernetes.io/instance: releaseName"
		log.Println("Resuming cronjobs")
		cronjob_client := clientset.BatchV1().CronJobs(namespace)
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			cronjob_list, err := cronjob_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of cronjobs: %s", err)
			}

			for _, v := range cronjob_list.Items {

				// Load the cronjob
				r, err := cronjob_client.Get(context.TODO(), v.Name, v1.GetOptions{})
				if err != nil {
					log.Fatalf("Error getting the cronjob: %s", err)
				}

				if *r.Spec.Suspend {

					log.Printf("Unpausing %s", r.Name)

					// Unpause cronjob
					r.Spec.Suspend = &[]bool{false}[0]

					cronjob_client.Update(context.TODO(), r, v1.UpdateOptions{
						TypeMeta:     v1.TypeMeta{},
						DryRun:       nil,
						FieldManager: "",
					})
				}
			}
		}

		// Restore services
		log.Println("Restoring services")
		services_client := clientset.CoreV1().Services(namespace)
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			services_list, err := services_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of services: %s", err)
			}

			for _, v := range services_list.Items {

				// Load the service
				r, err := services_client.Get(context.TODO(), v.Name, v1.GetOptions{})
				if err != nil {
					log.Fatalf("Error getting the service: %s", err)
				}

				// If auto-downscale/original-selector is not set, skip
				if r.Annotations["auto-downscale/original-selector"] != "" {

					log.Printf("Resetting %s", r.Name)

					// Recreate service with original definition and change type
					newService := corev1.Service{
						ObjectMeta: v1.ObjectMeta{
							Name:        r.Name,
							Namespace:   r.Namespace,
							Annotations: r.Annotations,
							Labels:      r.Labels,
						},
						Spec: r.Spec,
					}

					// Restore type
					if r.Annotations["auto-downscale/original-type"] != "" {
						newService.Spec.Type = corev1.ServiceType(r.Annotations["auto-downscale/original-type"])
					} else {
						// Fallback to original type when annotation is missing
						newService.Spec.Type = corev1.ServiceType(r.Spec.Type)
					}

					// Remove newService.Spec.ExternalName, we don't use it anymore
					newService.Spec.ExternalName = ""

					// Decode original selector and ports as json
					originalSelector := map[string]string{}
					originalSelectorJson := r.Annotations["auto-downscale/original-selector"]

					// Unmarshal originalSelectorJson to originalSelector
					err := json.Unmarshal([]byte(originalSelectorJson), &originalSelector)
					if err != nil {
						log.Fatalf("Error parsing original selector for %s/%s", namespace, r.Name)
					}

					// Set selector
					newService.Spec.Selector = originalSelector

					// Set ports
					originalPorts := []corev1.ServicePort{}
					// Old downscales did not set ports. If original ports are not set, use current ports.
					if r.Annotations["auto-downscale/original-ports"] == "" {
						newService.Spec.Ports = r.Spec.Ports
					} else {
						// Unmarshal originalPortsJson to originalPorts
						originalPortsJson := r.Annotations["auto-downscale/original-ports"]
						err = json.Unmarshal([]byte(originalPortsJson), &originalPorts)
						if err != nil {
							log.Fatalf("Error parsing original ports for %s/%s", namespace, r.Name)
						}
						// Reset ports
						newService.Spec.Ports = []corev1.ServicePort{}
						// Readd ports, but skip nodeport value
						for _, port := range originalPorts {
							newService.Spec.Ports = append(newService.Spec.Ports, corev1.ServicePort{
								Name:       port.Name,
								Protocol:   port.Protocol,
								Port:       port.Port,
								TargetPort: port.TargetPort,
							})
						}
					}

					// Unset annotations and labels
					delete(newService.ObjectMeta.Annotations, "auto-downscale/original-type")
					delete(newService.ObjectMeta.Annotations, "auto-downscale/original-selector")
					delete(newService.ObjectMeta.Annotations, "auto-downscale/original-ports")
					delete(newService.ObjectMeta.Labels, "auto-downscale/redirected")

					newService.ObjectMeta.Annotations["auto-downscale/down"] = "false"

					err = services_client.Delete(context.TODO(), r.Name, v1.DeleteOptions{
						PropagationPolicy: nil,
					})

					if err != nil {
						log.Fatalf("Error deleting service: %s", err)
					}

					// Create new service
					_, err = services_client.Create(context.TODO(), &newService, v1.CreateOptions{
						TypeMeta:     v1.TypeMeta{},
						DryRun:       nil,
						FieldManager: "",
					})

					if err != nil {
						log.Fatalf("Error creating new service: %s", err)
					}

					log.Printf("Reset %s to original service", r.Name)
				}
			}
		}

		// Gather ingress hostnames
		hostnames := []string{}
		ingress_client := clientset.NetworkingV1().Ingresses(namespace)
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			ingress_list, err := ingress_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of ingresses: %s", err)
			}

			for _, v := range ingress_list.Items {

				// Add ingress hostnames to list
				for _, rule := range v.Spec.Rules {
					// if hostname is not in list, add it
					s := false
					for _, hostname := range hostnames {
						if hostname == rule.Host {
							s = true
							continue
						}
					}
					if !s {
						hostnames = append(hostnames, rule.Host)
					}
				}
			}
		}

		// Print hostnames
		for _, hostname := range hostnames {
			log.Printf("https://%s\n", hostname)
		}
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseWakeupCmd)

	ciReleaseWakeupCmd.Flags().String("release-name", "", "Release name")
	ciReleaseWakeupCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")

	ciReleaseWakeupCmd.MarkFlagRequired("release-name")
	ciReleaseWakeupCmd.MarkFlagRequired("namespace")
}
