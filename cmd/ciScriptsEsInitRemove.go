package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var ciScriptsEsinitRemoveCmd = &cobra.Command{
	Use:   "elasticsearch-initcontainer-remove",
	Short: "es-init-remove",
	Long:  `Elasticsearch init container removal from all statefulsets in the cluster.`,
	Run: func(cmd *cobra.Command, args []string) {

		namespace, _ := cmd.Flags().GetString("namespace")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		fmt.Printf("Dry run: %t\n", dryRun)

		if namespace == "" {
			fmt.Printf("Namespace: all namespaces\n")
		} else {
			fmt.Printf("Namespace: %s\n", namespace)
		}

		// Try reading KUBECONFIG from environment variable first
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("cannot read user home dir")
			}
			kubeConfigPath = homeDir + "/.kube/config"
		}

		//k8s go client init logic
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			log.Fatalf("cannot read kubeConfig from path: %s", kubeConfigPath)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatalf("cannot initialize k8s client: %s", err)
		}

		// Sanity check - query if daemonset with es-init is installed. It has a specific name, silta-cluster-ds in silta-cluster namespace
		_, err = clientset.AppsV1().DaemonSets("silta-cluster").Get(context.TODO(), "silta-cluster-ds", v1.GetOptions{})
		if err != nil {
			log.Fatal("Elasticsearch init value daemonset is not created, ", err)
		}

		// Select all statefulsets in the namespace
		listOptions := v1.ListOptions{
			LabelSelector: "chart=elasticsearch",
		}
		statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), listOptions)
		if err != nil {
			log.Fatalf("cannot get statefulsets: %s", err)
		}

		fmt.Printf("Elasticsearch statefulsets: %d\n", len(statefulsets.Items))

		matchCounter := 0
		patchCounter := 0

		// Loop through all statefulsets and remove es-init container
		for _, statefulset := range statefulsets.Items {

			// If statefulset has configure-sysctl initContainer, remove it
			for i, container := range statefulset.Spec.Template.Spec.InitContainers {
				if container.Name == "configure-sysctl" {
					fmt.Printf("Removing %s/%s/configure-sysctl ... ", statefulset.Namespace, statefulset.Name)

					if dryRun {
						fmt.Printf("skipping, dry-run is enabled\n")
					} else {

						// Patch statefulset, apply removal of initContainer
						patch := []byte(fmt.Sprintf(`[{"op": "remove", "path": "/spec/template/spec/initContainers/%d"}]`, i))
						_, err = clientset.AppsV1().StatefulSets(statefulset.Namespace).Patch(context.TODO(), statefulset.Name, types.JSONPatchType, patch, v1.PatchOptions{})
						if err != nil {
							log.Printf("cannot patch statefulset, %s", err)
						}

						patchCounter++

						fmt.Printf("removed\n")
					}

					matchCounter++
				}
			}
		}
		fmt.Printf("Total statefulsets matched: %d\n", matchCounter)
		fmt.Printf("Total statefulsets patched: %d\n", patchCounter)
	},
}

func init() {
	ciScriptsEsinitRemoveCmd.Flags().String("namespace", "", "Namespace (optional)")
	ciScriptsEsinitRemoveCmd.Flags().Bool("dry-run", true, "Dry run")

	ciScriptsCmd.AddCommand(ciScriptsEsinitRemoveCmd)
}
