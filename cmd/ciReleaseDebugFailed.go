package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var ciReleaseDebugFailedCmd = &cobra.Command{
	Use:   "debug-failed",
	Short: "Debug failed deployment resources",
	Long: `Debug failed deployment by checking:
- OOMKilled containers
- Failed pods with their events and logs
- Not-ready statefulsets with their events
- Not-ready deployments with their events

This command is typically called when a deployment fails to help diagnose the issue.`,
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")

		clientset, err := common.GetKubeClient()
		if err != nil {
			log.Fatalf("failed to get kube client: %v", err)
		}

		fmt.Println()
		hasErrors := false

		// Check for OOMKilled containers
		if checkOOMKilledPods(clientset, namespace, releaseName) {
			hasErrors = true
		}

		// Check for failed pods
		if checkFailedPods(clientset, namespace, releaseName) {
			hasErrors = true
		}

		// Check for not-ready statefulsets
		if checkNotReadyStatefulSets(clientset, namespace, releaseName) {
			hasErrors = true
		}

		// Check for not-ready deployments
		if checkNotReadyDeployments(clientset, namespace, releaseName) {
			hasErrors = true
		}

		if hasErrors {
			log.Fatal("Deployment failures found")
		}
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseDebugFailedCmd)

	ciReleaseDebugFailedCmd.Flags().String("release-name", "", "Release name")
	ciReleaseDebugFailedCmd.Flags().String("namespace", "", "Namespace")

	ciReleaseDebugFailedCmd.MarkFlagRequired("release-name")
	ciReleaseDebugFailedCmd.MarkFlagRequired("namespace")
}

// checkOOMKilledPods checks for pods with OOMKilled containers
func checkOOMKilledPods(clientset *kubernetes.Clientset, namespace, releaseName string) bool {
	labelSelectors := []string{
		fmt.Sprintf("release=%s,cronjob!=true", releaseName),
		fmt.Sprintf("app.kubernetes.io/instance=%s,cronjob!=true", releaseName),
	}

	oomKilledFound := false
	oomKilledPods := make(map[string][]string) // pod -> containers

	for _, selector := range labelSelectors {
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			continue
		}

		for _, pod := range pods.Items {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				isOOMKilled := false
				if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.Reason == "OOMKilled" {
					isOOMKilled = true
				}
				if containerStatus.LastTerminationState.Terminated != nil && containerStatus.LastTerminationState.Terminated.Reason == "OOMKilled" {
					isOOMKilled = true
				}

				if isOOMKilled {
					exists := false
					for _, v := range oomKilledPods[pod.Name] {
						if v == containerStatus.Name {
							exists = true
							break
						}
					}
					if !exists {
						oomKilledPods[pod.Name] = append(oomKilledPods[pod.Name], containerStatus.Name)
					}
					oomKilledFound = true
				}
			}
		}
	}

	if oomKilledFound {
		fmt.Println("Error: following pods run into Out Of Memory (OOM) issues during deployment (need more memory!):")
		for podName, containers := range oomKilledPods {
			for _, container := range containers {
				fmt.Printf("  * %s %s OOMKilled\n", podName, container)
			}
		}
		fmt.Println()
		return true
	}

	return false
}

// checkFailedPods checks for failed pods and displays their events and logs
func checkFailedPods(clientset *kubernetes.Clientset, namespace, releaseName string) bool {
	labelSelectors := []string{
		fmt.Sprintf("release=%s,cronjob!=true", releaseName),
		fmt.Sprintf("app.kubernetes.io/instance=%s,cronjob!=true", releaseName),
	}

	failedPods := []string{}
	podMap := make(map[string]v1.Pod)

	for _, selector := range labelSelectors {
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			continue
		}

		for _, pod := range pods.Items {
			isFailed := false

			// Check if any container is not ready
			if pod.Status.ContainerStatuses != nil {
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if !containerStatus.Ready {
						isFailed = true
						break
					}
				}
			} else {
				// No container statuses means pod hasn't started properly
				isFailed = true
			}

			if isFailed {
				if _, exists := podMap[pod.Name]; !exists {
					exists := false
					for _, v := range failedPods {
						if v == pod.Name {
							exists = true
							break
						}
					}
					if !exists {
						failedPods = append(failedPods, pod.Name)
						podMap[pod.Name] = pod
					}
				}
			}
		}
	}

	if len(failedPods) == 0 {
		return false
	}

	fmt.Println("Error: pod failures:")
	for _, podName := range failedPods {
		pod := podMap[podName]
		fmt.Printf("  * %s / %s\n", namespace, podName)

		// Show events
		fmt.Println("    ---- Events ----")
		showPodEvents(clientset, namespace, podName)

		fmt.Println("    ---- Logs ----")
		showPodLogs(clientset, namespace, pod)
		fmt.Println("    ----")

		fmt.Println()
	}

	return true
}

// showPodEvents displays non-normal events for a pod
func showPodEvents(clientset *kubernetes.Clientset, namespace, podName string) {
	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,type!=Normal", podName),
	})
	if err != nil {
		fmt.Println("Error getting events:", err)
		return
	}

	if len(events.Items) == 0 {
		fmt.Println("No abnormal events found")
		return
	}

	for _, event := range events.Items {
		fmt.Printf("    %s %s/%s: %s\n",
			event.Type,
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.Message)
	}
}

// showPodLogs displays logs for failed containers in a pod
func showPodLogs(clientset *kubernetes.Clientset, namespace string, pod v1.Pod) {
	hasLogs := false

	// Find containers that are not ready
	notReadyContainers := []string{}
	if pod.Status.ContainerStatuses != nil {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				notReadyContainers = append(notReadyContainers, containerStatus.Name)
			}
		}
	}

	if len(notReadyContainers) == 0 {
		fmt.Println("no logs found")
		return
	}

	for _, containerName := range notReadyContainers {
		logOptions := &v1.PodLogOptions{
			Container: containerName,
		}

		req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, logOptions)
		podLogs, err := req.Stream(context.TODO())
		if err != nil {
			continue
		}
		defer podLogs.Close()

		buf := new(strings.Builder)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			continue
		}

		logs := buf.String()
		if logs != "" {
			// Print logs with 4-space indentation for readability
			logLines := strings.Split(strings.TrimRight(logs, "\n"), "\n")
			for _, line := range logLines {
				fmt.Printf("    [%s] %s\n", containerName, line)
			}
			hasLogs = true
		}
	}

	if !hasLogs {
		fmt.Println("no logs found")
	}
}

// checkNotReadyStatefulSets checks for statefulsets that are not ready
func checkNotReadyStatefulSets(clientset *kubernetes.Clientset, namespace, releaseName string) bool {
	labelSelectors := []string{
		fmt.Sprintf("release=%s", releaseName),
		fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
	}

	notReadyStatefulSets := []string{}

	for _, selector := range labelSelectors {
		statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			continue
		}

		for _, sts := range statefulsets.Items {
			if sts.Status.ReadyReplicas < *sts.Spec.Replicas {
				exists := false
				for _, v := range notReadyStatefulSets {
					if v == sts.Name {
						exists = true
						break
					}
				}
				if !exists {
					notReadyStatefulSets = append(notReadyStatefulSets, sts.Name)
				}
			}
		}
	}

	if len(notReadyStatefulSets) == 0 {
		return false
	}

	fmt.Println("Error: statefulset resource failures:")

	for _, stsName := range notReadyStatefulSets {
		fmt.Printf("  * %s / %s\n", namespace, stsName)
		events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s,type!=Normal", stsName),
		})
		if err != nil || len(events.Items) == 0 {
			continue
		}

		fmt.Println("    ---- Events ----")
		for _, event := range events.Items {
			fmt.Printf("    * %s %s/%s: %s\n",
				event.Type,
				event.InvolvedObject.Kind,
				event.InvolvedObject.Name,
				event.Message)
		}
		fmt.Println("    ----")
	}
	fmt.Println()

	return len(notReadyStatefulSets) > 0
}

// checkNotReadyDeployments checks for deployments that are not ready
func checkNotReadyDeployments(clientset *kubernetes.Clientset, namespace, releaseName string) bool {
	labelSelectors := []string{
		fmt.Sprintf("release=%s", releaseName),
		fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
	}

	notReadyDeployments := []string{}

	for _, selector := range labelSelectors {
		deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			continue
		}

		for _, deployment := range deployments.Items {
			if deployment.Status.ReadyReplicas < *deployment.Spec.Replicas {
				// iterate over notReadyDeployments and if deployment is not on the list already, add it
				exists := false
				for _, v := range notReadyDeployments {
					if v == deployment.Name {
						exists = true
						break
					}
				}
				if !exists {
					notReadyDeployments = append(notReadyDeployments, deployment.Name)
				}
			}
		}
	}

	if len(notReadyDeployments) == 0 {
		return false
	}

	fmt.Println("Error: deployment resource failures:")

	for _, deploymentName := range notReadyDeployments {
		fmt.Printf("  * %s / %s\n", namespace, deploymentName)

		events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s,type!=Normal", deploymentName),
		})
		if err != nil || len(events.Items) == 0 {
			continue
		}

		fmt.Println("    ---- Events ----")
		for _, event := range events.Items {
			fmt.Printf("    * %s %s/%s: %s\n",
				event.Type,
				event.InvolvedObject.Kind,
				event.InvolvedObject.Name,
				event.Message)
		}
		fmt.Println("    ----")
	}
	fmt.Println()

	return len(notReadyDeployments) > 0
}
