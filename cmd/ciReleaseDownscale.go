package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth provider
	"k8s.io/client-go/rest"

	helmAction "helm.sh/helm/v3/pkg/action"
	helmCli "helm.sh/helm/v3/pkg/cli"
)

const upscalerProxyName = "silta-cluster-placeholder-upscaler-proxy"

var ciReleaseDownscaleCmd = &cobra.Command{
	Use:   "downscale",
	Short: "Downscale a release",
	Long: `Downscale a release the same way silta-downscaler does it: redirect the release service
to the placeholder upscaler proxy, mark ingress as down, suspend MariaDB resources and cronjobs,
and scale deployments and statefulsets to 0. Use "silta ci release wakeup" to restore it.`,
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")
		placeholderServiceName, _ := cmd.Flags().GetString("placeholder-service-name")
		placeholderServiceNamespace, _ := cmd.Flags().GetString("placeholder-service-namespace")
		placeholderProxyImage, _ := cmd.Flags().GetString("placeholder-proxy-image")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		kubeConfig, err := common.GetKubeConfig()
		if err != nil {
			log.Fatalf("failed to get kube config: %v", err)
		}
		clientset, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			log.Fatalf("failed to get kube client: %v", err)
		}

		// Helm client init logic
		settings := helmCli.New()

		actionConfig := new(helmAction.Configuration)
		if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
			log.Fatalf("%+v", err)
		}
		// Try loading the latest version of the release
		get := helmAction.NewGet(actionConfig)
		_, err = get.Run(releaseName)
		if err != nil {
			log.Fatalf("Release not found: %s", err)
		}

		if dryRun {
			log.Println("Dry run, no changes will be made")
		}

		selectorLabels := []string{
			"release",
			"app.kubernetes.io/instance",
		}

		// Collect ingresses that have auto-downscale enabled
		ingress_client := clientset.NetworkingV1().Ingresses(namespace)
		ingressNames := []string{}
		serviceNames := []string{}
		seen := map[types.UID]bool{}
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			ingress_list, err := ingress_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of ingresses: %s", err)
			}
			for _, v := range ingress_list.Items {
				if seen[v.UID] || v.Annotations["auto-downscale/services"] == "" {
					continue
				}
				seen[v.UID] = true
				ingressNames = append(ingressNames, v.Name)
				if !containsString(serviceNames, v.Annotations["auto-downscale/services"]) {
					serviceNames = append(serviceNames, v.Annotations["auto-downscale/services"])
				}
			}
		}
		if len(ingressNames) == 0 {
			log.Fatalf("Release %s/%s has no ingress with auto-downscale enabled", namespace, releaseName)
		}

		// Resolve placeholder settings from the silta-downscaler cronjob, unless set via flags
		if placeholderServiceName == "" || placeholderServiceNamespace == "" || placeholderProxyImage == "" {
			name, ns, image := discoverPlaceholderSettings(clientset)
			if placeholderServiceName == "" {
				placeholderServiceName = name
			}
			if placeholderServiceNamespace == "" {
				placeholderServiceNamespace = ns
			}
			if placeholderProxyImage == "" {
				placeholderProxyImage = image
			}
		}

		// Redirect services to upscaler proxy
		log.Println("Redirecting services")
		services_client := clientset.CoreV1().Services(namespace)
		for _, serviceName := range serviceNames {
			r, err := services_client.Get(context.TODO(), serviceName, v1.GetOptions{})
			if err != nil {
				log.Fatalf("Error getting the service: %s", err)
			}
			if r.Annotations["auto-downscale/down"] == "true" {
				log.Printf("Service %s is already redirected", r.Name)
				continue
			}

			if !dryRun {
				ensureUpscalerProxy(clientset, namespace, placeholderServiceName, placeholderServiceNamespace, placeholderProxyImage)
			}

			// Do not save NodePort values, as they are automatically set by k8s and we do not depend on them
			originalPorts := []corev1.ServicePort{}
			for _, port := range r.Spec.Ports {
				port.NodePort = 0
				originalPorts = append(originalPorts, port)
			}
			originalPortsJson, err := json.Marshal(originalPorts)
			if err != nil {
				log.Fatalf("Error encoding ports for %s/%s: %s", namespace, r.Name, err)
			}
			originalSelectorJson, err := json.Marshal(r.Spec.Selector)
			if err != nil {
				log.Fatalf("Error encoding selector for %s/%s: %s", namespace, r.Name, err)
			}

			mergePatch, _ := json.Marshal(map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]string{
						"auto-downscale/down":              "true",
						"auto-downscale/original-type":     string(r.Spec.Type),
						"auto-downscale/original-selector": string(originalSelectorJson),
						"auto-downscale/original-ports":    string(originalPortsJson),
					},
					"labels": map[string]string{
						"auto-downscale/redirected": "true",
					},
				},
				"spec": map[string]interface{}{
					"ports": []map[string]interface{}{
						{
							"name":       "http",
							"port":       80,
							"targetPort": 8080,
							"protocol":   "TCP",
						},
					},
				},
			})
			// Replace service selector instead of merging it
			selectorPatch, _ := json.Marshal([]map[string]interface{}{
				{
					"op":    "replace",
					"path":  "/spec/selector",
					"value": map[string]string{"app": upscalerProxyName},
				},
			})

			log.Printf("Redirecting service %s to placeholder service", r.Name)
			if !dryRun {
				_, err = services_client.Patch(context.TODO(), r.Name, types.MergePatchType, mergePatch, v1.PatchOptions{})
				if err != nil {
					log.Fatalf("Error patching service %s: %s", r.Name, err)
				}
				_, err = services_client.Patch(context.TODO(), r.Name, types.JSONPatchType, selectorPatch, v1.PatchOptions{})
				if err != nil {
					log.Fatalf("Error patching service %s selector: %s", r.Name, err)
				}
			}
		}

		// Mark ingresses as down
		ingressPatch := []byte(`{"metadata":{"annotations":{"auto-downscale/down":"true"}}}`)
		for _, ingressName := range ingressNames {
			log.Printf("Marking ingress %s as down", ingressName)
			if !dryRun {
				_, err := ingress_client.Patch(context.TODO(), ingressName, types.MergePatchType, ingressPatch, v1.PatchOptions{})
				if err != nil {
					log.Fatalf("Error patching ingress %s: %s", ingressName, err)
				}
			}
		}

		// Suspend MariaDBs (if any) before scaling their statefulsets to 0
		setMariaDBsSuspended(clientset, kubeConfig, namespace, releaseName, selectorLabels, true, dryRun)

		// Scale deployments to 0
		log.Println("Scaling deployments")
		deployment_client := clientset.AppsV1().Deployments(namespace)
		seen = map[types.UID]bool{}
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			deployment_list, err := deployment_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of deployments: %s", err)
			}

			for _, v := range deployment_list.Items {
				if seen[v.UID] {
					continue
				}
				seen[v.UID] = true
				if v.Spec.Replicas == nil || *v.Spec.Replicas == 0 {
					continue
				}
				log.Printf("Downscaling deployment %s from %d to 0", v.Name, *v.Spec.Replicas)
				if !dryRun {
					_, err := deployment_client.Patch(context.TODO(), v.Name, types.MergePatchType, downscalePatch(*v.Spec.Replicas), v1.PatchOptions{})
					if err != nil {
						log.Fatalf("Error downscaling deployment %s: %s", v.Name, err)
					}
				}
			}
		}

		// Scale statefulsets to 0
		log.Println("Scaling statefulsets")
		statefulset_client := clientset.AppsV1().StatefulSets(namespace)
		seen = map[types.UID]bool{}
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			statefulset_list, err := statefulset_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of statefulsets: %s", err)
			}

			for _, v := range statefulset_list.Items {
				if seen[v.UID] {
					continue
				}
				seen[v.UID] = true
				if v.Spec.Replicas == nil || *v.Spec.Replicas == 0 {
					continue
				}
				log.Printf("Downscaling statefulset %s from %d to 0", v.Name, *v.Spec.Replicas)
				if !dryRun {
					_, err := statefulset_client.Patch(context.TODO(), v.Name, types.MergePatchType, downscalePatch(*v.Spec.Replicas), v1.PatchOptions{})
					if err != nil {
						log.Fatalf("Error downscaling statefulset %s: %s", v.Name, err)
					}
				}
			}
		}

		// Suspend cronjobs
		log.Println("Suspending cronjobs")
		cronjob_client := clientset.BatchV1().CronJobs(namespace)
		suspendPatch := []byte(`{"spec":{"suspend":true}}`)
		seen = map[types.UID]bool{}
		for _, l := range selectorLabels {
			selector := l + "=" + releaseName
			cronjob_list, err := cronjob_client.List(context.TODO(), v1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				log.Fatalf("Error getting the list of cronjobs: %s", err)
			}

			for _, v := range cronjob_list.Items {
				if seen[v.UID] {
					continue
				}
				seen[v.UID] = true
				if v.Spec.Suspend != nil && *v.Spec.Suspend {
					continue
				}
				log.Printf("Suspending cronjob %s", v.Name)
				if !dryRun {
					_, err := cronjob_client.Patch(context.TODO(), v.Name, types.MergePatchType, suspendPatch, v1.PatchOptions{})
					if err != nil {
						log.Fatalf("Error suspending cronjob %s: %s", v.Name, err)
					}
				}
			}
		}

		log.Printf("Release %s/%s downscaled", namespace, releaseName)
	},
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// downscalePatch returns a merge patch that stores the current replica count and scales to 0
func downscalePatch(replicas int32) []byte {
	patch, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				"auto-downscale/original-replicas": strconv.Itoa(int(replicas)),
			},
		},
		"spec": map[string]interface{}{
			"replicas": 0,
		},
	})
	return patch
}

// discoverPlaceholderSettings reads placeholder settings from the silta-downscaler cronjob,
// falling back to silta-cluster defaults when it can't be found
func discoverPlaceholderSettings(clientset *kubernetes.Clientset) (string, string, string) {
	serviceName := "silta-cluster-placeholder-upscaler"
	serviceNamespace := "silta-cluster"
	proxyImage := "wunderio/silta-downscaler:v1-proxy"

	cronjob_list, err := clientset.BatchV1().CronJobs("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Printf("Warning: could not look up downscaler cronjob (%s), using default placeholder settings", err)
		return serviceName, serviceNamespace, proxyImage
	}
	for _, cj := range cronjob_list.Items {
		if !strings.HasSuffix(cj.Name, "-downscale-cron") {
			continue
		}
		for _, c := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
			if c.Name != "downscaler-cron" {
				continue
			}
			serviceNamespace = cj.Namespace
			for _, e := range c.Env {
				switch e.Name {
				case "PLACEHOLDER_SERVICE_NAME":
					if e.Value != "" {
						serviceName = e.Value
					}
				case "PLACEHOLDER_PROXY_IMAGE":
					if e.Value != "" {
						proxyImage = e.Value
					}
				}
			}
			return serviceName, serviceNamespace, proxyImage
		}
	}
	log.Println("Warning: downscaler cronjob not found, using default placeholder settings")
	return serviceName, serviceNamespace, proxyImage
}

// ensureUpscalerProxy creates the upscaler proxy deployment if it does not exist and waits for it to be ready
func ensureUpscalerProxy(clientset *kubernetes.Clientset, namespace string, placeholderServiceName string, placeholderServiceNamespace string, placeholderProxyImage string) {
	deployment_client := clientset.AppsV1().Deployments(namespace)

	_, err := deployment_client.Get(context.TODO(), upscalerProxyName, v1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Fatalf("Error getting upscaler proxy deployment: %s", err)
		}

		log.Printf("Spinning up upscaler proxy deployment in %s namespace", namespace)
		replicas := int32(1)
		enableServiceLinks := false
		labels := map[string]string{"app": upscalerProxyName}
		deployment := &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name: upscalerProxyName,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &v1.LabelSelector{MatchLabels: labels},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{Labels: labels},
					Spec: corev1.PodSpec{
						EnableServiceLinks: &enableServiceLinks,
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: placeholderProxyImage,
								Env: []corev1.EnvVar{
									{Name: "PLACEHOLDER_SERVICE_NAME", Value: placeholderServiceName},
									{Name: "PLACEHOLDER_SERVICE_NAMESPACE", Value: placeholderServiceNamespace},
								},
								Ports: []corev1.ContainerPort{{ContainerPort: 8080}},
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("1m"),
										corev1.ResourceMemory: resource.MustParse("10Mi"),
									},
								},
							},
						},
					},
				},
			},
		}
		_, err = deployment_client.Create(context.TODO(), deployment, v1.CreateOptions{})
		if err != nil {
			log.Fatalf("Error creating upscaler proxy deployment: %s", err)
		}
	}

	// Wait for proxy to be ready so requests don't get dropped with 50x
	// Wait up to 2 minutes
	timeout := 120
	for {
		r, err := deployment_client.Get(context.TODO(), upscalerProxyName, v1.GetOptions{})
		if err != nil {
			log.Fatalf("Error getting upscaler proxy deployment: %s", err)
		}
		if r.Status.ReadyReplicas > 0 {
			break
		}
		// wait 5 seconds
		time.Sleep(5 * time.Second)
		timeout = timeout - 5

		if timeout <= 0 {
			log.Fatalf("Timeout waiting for %s to be ready", upscalerProxyName)
		}
	}
}

// setMariaDBsSuspended suspends or resumes k8s.mariadb.com MariaDB resources belonging to the release, if the CRD is installed
func setMariaDBsSuspended(clientset *kubernetes.Clientset, kubeConfig *rest.Config, namespace string, releaseName string, selectorLabels []string, suspend bool, dryRun bool) {
	const mariadbGroup = "k8s.mariadb.com"

	groups, err := clientset.Discovery().ServerGroups()
	if err != nil {
		log.Printf("Warning: could not discover API groups, skipping MariaDB suspend/resume: %s", err)
		return
	}
	version := ""
	for _, g := range groups.Groups {
		if g.Name == mariadbGroup {
			version = g.PreferredVersion.Version
			if version == "" && len(g.Versions) > 0 {
				version = g.Versions[0].Version
			}
		}
	}
	if version == "" {
		// MariaDB CRD not installed, nothing to manage
		return
	}

	dynamicClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %s", err)
	}
	mariadb_client := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    mariadbGroup,
		Version:  version,
		Resource: "mariadbs",
	}).Namespace(namespace)

	action := "Resuming"
	if suspend {
		action = "Suspending"
	}
	log.Printf("%s MariaDBs", action)
	suspendPatch := []byte(`{"spec":{"suspend":` + strconv.FormatBool(suspend) + `}}`)
	seen := map[types.UID]bool{}
	for _, l := range selectorLabels {
		selector := l + "=" + releaseName
		mariadb_list, err := mariadb_client.List(context.TODO(), v1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return
			}
			log.Fatalf("Error getting the list of MariaDBs: %s", err)
		}

		for _, v := range mariadb_list.Items {
			if seen[v.GetUID()] {
				continue
			}
			seen[v.GetUID()] = true
			suspended, _, _ := unstructured.NestedBool(v.Object, "spec", "suspend")
			if suspended == suspend {
				continue
			}
			log.Printf("%s MariaDB %s", action, v.GetName())
			if !dryRun {
				_, err := mariadb_client.Patch(context.TODO(), v.GetName(), types.MergePatchType, suspendPatch, v1.PatchOptions{})
				if err != nil {
					log.Fatalf("Error patching MariaDB %s: %s", v.GetName(), err)
				}
			}
		}
	}
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseDownscaleCmd)

	ciReleaseDownscaleCmd.Flags().String("release-name", "", "Release name")
	ciReleaseDownscaleCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseDownscaleCmd.Flags().String("placeholder-service-name", "", "Placeholder upscaler service name (default: read from silta-downscaler cronjob)")
	ciReleaseDownscaleCmd.Flags().String("placeholder-service-namespace", "", "Placeholder upscaler service namespace (default: read from silta-downscaler cronjob)")
	ciReleaseDownscaleCmd.Flags().String("placeholder-proxy-image", "", "Upscaler proxy image (default: read from silta-downscaler cronjob)")
	ciReleaseDownscaleCmd.Flags().Bool("dry-run", false, "Print changes without applying them")

	ciReleaseDownscaleCmd.MarkFlagRequired("release-name")
	ciReleaseDownscaleCmd.MarkFlagRequired("namespace")
}
