package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// cloudLoginCmd represents the cloudLogin command
var cloudLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Kubernetes cluster login",
	Long: `Log in to kubernetes cluster using different methods:
	
	* Kubeconfig for custom cluster access 
	Requires:
	  - "--kube-config" flag or "KUBECTL_CONFIG" environment variable
	  
	* Google cloud GKE access
	Requires:
	  - "--cluster-name" flag or "CLUSTER_NAME" environment variable
	  - "--gcp-key-json" flag or "GCLOUD_KEY_JSON" environment variable
	  - "--gcp-project-name" flag or "GCLOUD_PROJECT_NAME" environment variable
	
	Optional parameters:
	  - "--gcp-compute-region" flag or "GCLOUD_COMPUTE_REGION" environment variable
	  - "--gcp-compute-zone" flag or "GCLOUD_COMPUTE_ZONE" environment variable
	  
	* Amazon Web Services EKS access
	Requires:
	  - "--cluster-name" flag or "CLUSTER_NAME" environment variable
	  - "--aws-secret-access-key" flag or "AWS_SECRET_ACCESS_KEY" environment variable
	  - "--aws-region" flag or "AWS_REGION" environment variable

	* Azure Services AKS access
	Requires:
	  - "--cluster-name" flag or "CLUSTER_NAME" environment variable
	  - "--aks-resource-group" flag or "AKS_RESOURCE_GROUP" environment variable
	  - "--aks-tenant-id" flag or "AKS_TENANT_ID" environment variable
	  - "--aks-sp-app-id" flag or "AKS_SP_APP_ID" environment variable
	  - "--aks-sp-password" flag or "AKS_SP_PASSWORD" environment variable
	`,
	Run: func(cmd *cobra.Command, args []string) {

		clusterName, _ := cmd.Flags().GetString("cluster-name")

		kubeConfig, _ := cmd.Flags().GetString("kubeconfig")
		kubeConfigPath, _ := cmd.Flags().GetString("kubeconfigpath")

		gcpKeyJson, _ := cmd.Flags().GetString("gcp-key-json")
		gcpProjectName, _ := cmd.Flags().GetString("gcp-project-name")
		gcpComputeRegion, _ := cmd.Flags().GetString("gcp-compute-region")
		gcpComputeZone, _ := cmd.Flags().GetString("gcp-compute-zone")

		awsSecretAccessKey, _ := cmd.Flags().GetString("aws-secret-access-key")
		awsRegion, _ := cmd.Flags().GetString("aws-region")

		aksResourceGroup, _ := cmd.Flags().GetString("aks-resource-group")
		aksTenantID, _ := cmd.Flags().GetString("aks-tenant-id")
		aksSPAppID, _ := cmd.Flags().GetString("aks-sp-app-id")
		aksSPPass, _ := cmd.Flags().GetString("aks-sp-password")

		// Environment value fallback
		if useEnv == true {
			if len(clusterName) == 0 {
				clusterName = os.Getenv("CLUSTER_NAME")
			}
			if len(kubeConfig) == 0 {
				kubeConfig = os.Getenv("KUBECTL_CONFIG")
			}
			if len(gcpKeyJson) == 0 {
				gcpKeyJson = os.Getenv("GCLOUD_KEY_JSON")
			}
			if len(gcpProjectName) == 0 {
				gcpProjectName = os.Getenv("GCLOUD_PROJECT_NAME")
			}
			if len(gcpComputeRegion) == 0 {
				gcpComputeRegion = os.Getenv("GCLOUD_COMPUTE_REGION")
			}
			if len(gcpComputeZone) == 0 {
				gcpComputeZone = os.Getenv("GCLOUD_COMPUTE_ZONE")
			}
			if len(awsSecretAccessKey) == 0 {
				awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
			}
			if len(awsRegion) == 0 {
				awsRegion = os.Getenv("AWS_REGION")
			}
			if len(aksTenantID) == 0 {
				aksTenantID = os.Getenv("AKS_TENANT_ID")
			}
			if len(aksSPAppID) == 0 {
				aksSPAppID = os.Getenv("AKS_SP_APP_ID")
			}
			if len(aksSPPass) == 0 {
				aksSPPass = os.Getenv("AKS_SP_PASSWORD")
			}
			if len(aksResourceGroup) == 0 {
				aksResourceGroup = os.Getenv("AKS_RESOURCE_GROUP")
			}
		}

		// Require at least one auth method
		if len(kubeConfig) == 0 && len(gcpKeyJson) == 0 && len(awsSecretAccessKey) == 0 && len(aksTenantID) == 0 {
			fmt.Println(cmd.Usage())
			log.Fatal("Configuration method undefined")
		}

		command := ""

		if len(kubeConfig) > 0 {

			// Inject kubeconfig

			// Create kubeconfig folder
			if _, err := os.Stat(filepath.Dir(kubeConfigPath)); os.IsNotExist(err) {
				_ = os.Mkdir(filepath.Dir(kubeConfigPath), 0750)
			}

			// base64decoding Kubeconfig
			config, err := base64.StdEncoding.DecodeString(kubeConfig)
			if err != nil {
				log.Fatal("Error decoding kubeconfig string:", err)
			}

			// Write custom cubeconfig to kube config file
			err = os.WriteFile(kubeConfigPath, config, 0700)
			if err != nil {
				log.Fatal("Error writing kubeconfig:", err)
			}

		} else if len(gcpKeyJson) > 0 {

			// GCP gcloud login

			if len(gcpProjectName) == 0 {
				log.Fatal("GCP project name required (gcp-project-name)")
			}

			if len(clusterName) == 0 {
				log.Fatal("Cluster name required (cluster-name)")
			}

			// Save key
			homedir, _ := os.UserHomeDir()
			gcpKeyFilePath := fmt.Sprintf("%s/%s", homedir, "gcp-service-key.json")
			f, err := os.Create(gcpKeyFilePath)
			if err != nil {
				log.Fatal("Error creating gcp service key file:", err)
			}
			_, err = io.WriteString(f, gcpKeyJson)
			// err := os.WriteFile(gcpKeyFilePath, gcpKeyJson, 0700)
			if err != nil {
				log.Fatal("Error writing to gcp service key file:", err)
			}

			// Authenticate and set compute zone.
			command = fmt.Sprintf("gcloud auth activate-service-account --key-file='%s' --project '%s'; \n", gcpKeyFilePath, gcpProjectName)

			resourceLocation := ""
			if len(gcpComputeRegion) > 0 {
				resourceLocation = fmt.Sprint("--region", gcpComputeRegion)
			} else if len(gcpComputeZone) > 0 {
				resourceLocation = fmt.Sprint("--zone", gcpComputeZone)
			}

			// Updates a kubeconfig file with appropriate credentials and endpoint information.
			command += fmt.Sprintf("gcloud container clusters get-credentials '%s' --project '%s' %s", clusterName, gcpProjectName, resourceLocation)

		} else if awsSecretAccessKey != "" {

			// AWS login

			if len(clusterName) == 0 {
				log.Fatal("Cluster name required (cluster-name)")
			}

			if len(awsRegion) == 0 {
				log.Fatal("Amazon Web Services resource region (aws-region)")
			}

			command = fmt.Sprintf("aws eks update-kubeconfig --name '%s' --region '%s'", clusterName, awsRegion)

		} else if len(aksTenantID) > 0 {

			// Azure Services login

			if len(aksSPAppID) == 0 {
				log.Fatal("Azure Services servicePrincipal app id requred (aks-sp-app-id)")
			}
			if len(aksSPPass) == 0 {
				log.Fatal("Azure Services servicePrincipal password required (aks-sp-password)")
			}
			if len(aksResourceGroup) == 0 {
				log.Fatal("Azure Services resource group required (aks-resource-group)")
			}
			if len(clusterName) == 0 {
				log.Fatal("Cluster name required (cluster-name)")
			}

			command = fmt.Sprintf("az login --service-principal --username '%s' --tenant '%s' --password '%s';", aksSPAppID, aksTenantID, aksSPPass)
			command += fmt.Sprintf("az aks get-credentials --only-show-errors --resource-group '%s' --name '%s' --admin", aksResourceGroup, clusterName)
		}

		// Execute login commands
		if command != "" {
			if debug == true {
				fmt.Printf("Command (not executed): %s\n", command)
			} else {
				out, err := exec.Command("bash", "-c", command).CombinedOutput()
				if err != nil {
					fmt.Printf("Output: %s\n", out)
					log.Fatal("Error: ", err)
				}
				fmt.Printf("%s\n", out)
			}
		}
	},
}

func init() {
	cloudCmd.AddCommand(cloudLoginCmd)

	// Local flags
	cloudLoginCmd.Flags().String("cluster-name", "", "Kubernetes cluster name")
	cloudLoginCmd.Flags().String("kubeconfig", "", "Kubernetes config content (plaintext, base64 encoded)")
	cloudLoginCmd.Flags().String("kubeconfigpath", "~/.kube/config", "Kubernetes config path")
	cloudLoginCmd.Flags().String("gcp-key-json", "", "Google Cloud service account key (plaintext, json)")
	cloudLoginCmd.Flags().String("gcp-project-name", "", "GCP project name (project id)")
	cloudLoginCmd.Flags().String("gcp-compute-region", "", "GCP compute region")
	cloudLoginCmd.Flags().String("gcp-compute-zone", "", "GCP compute zone")
	cloudLoginCmd.Flags().String("aws-secret-access-key", "", "Amazon Web Services IAM account key (string value)")
	cloudLoginCmd.Flags().String("aws-region", "", "Amazon Web Services resource region")
	cloudLoginCmd.Flags().String("aks-resource-group", "", "Azure Services resource group (this is not the AKS RG)")
	cloudLoginCmd.Flags().String("aks-tenant-id", "", "Azure Services tenant id")
	cloudLoginCmd.Flags().String("aks-sp-app-id", "", "Azure Services servicePrincipal app id")
	cloudLoginCmd.Flags().String("aks-sp-password", "", "Azure Services servicePrincipal password")
}
