package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// imageLoginCmd represents the login command
var ciImageLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Image repository login",
	Long: `Login to (docker) image repository. 
	
Use either flags or environment variables for authentication. 

Available flags and environment variables:

  * General (required):
    - "--image-repo-host" flag or "IMAGE_REPO_HOST" environment variable: (Docker) container image repository url

  * Google Cloud:
    - "--gcp-key-json" flag or "GCP_KEY_JSON" environment variable: Google Cloud service account key (string value)

  * Amazon Web Services:
    - "--aws-secret-access-key" flag or "AWS_SECRET_ACCESS_KEY" environment variable: Amazon Web Services IAM account key (string value)

  * Azure Services:
    - "--aks-tenant-id" flag or "AKS_TENANT_ID" environment variable: Azure Services tenant id
    - "--aks-sp-app-id" flag or "AKS_SP_APP_ID" environment variable: Azure Services servicePrincipal app id
    - "--aks-sp-password" flag or "AKS_SP_PASSWORD" environment variable: Azure Services servicePrincipal password
`,
	Run: func(cmd *cobra.Command, args []string) {

		// Read flags into variables
		imageRepoHost, _ := cmd.Flags().GetString("image-repo-host")
		gcpKeyJson, _ := cmd.Flags().GetString("gcp-key-json")
		awsSecretAccessKey, _ := cmd.Flags().GetString("aws-secret-access-key")
		aksTenantID, _ := cmd.Flags().GetString("aks-tenant-id")
		aksSPAppID, _ := cmd.Flags().GetString("aks-sp-app-id")
		aksSPPass, _ := cmd.Flags().GetString("aks-sp-password")

		// Use environment variables as fallback
		if useEnv == true {
			if len(gcpKeyJson) == 0 {
				gcpKeyJson = os.Getenv("GCLOUD_KEY_JSON")
			}
			if len(gcpKeyJson) == 0 {
				gcpKeyJson = os.Getenv("GCP_KEY_JSON")
			}
			if len(awsSecretAccessKey) == 0 {
				awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
			}
			if len(imageRepoHost) == 0 {
				imageRepoHost = os.Getenv("IMAGE_REPO_HOST")
			}
			if len(imageRepoHost) == 0 {
				imageRepoHost = os.Getenv("DOCKER_REPO_HOST")
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
		}

		if debug == true {
			// Print variables
			fmt.Println("IMAGE_REPO_HOST:", imageRepoHost)
			fmt.Println("GCLOUD_KEY_JSON:", gcpKeyJson)
			fmt.Println("AWS_SECRET_ACCESS_KEY:", awsSecretAccessKey)
			fmt.Println("AKS_TENANT_ID:", aksTenantID)
			fmt.Println("AKS_SP_APP_ID:", aksSPAppID)
			fmt.Println("AKS_SP_PASSWORD:", aksSPPass)
		}

		if len(gcpKeyJson) == 0 && len(awsSecretAccessKey) == 0 && len(aksSPPass) == 0 {
			log.Fatal("Docker registry credentials are empty, have you set a context for this CircleCI job correctly?")
		} else {

			command := ""

			if gcpKeyJson != "" {
				// GCR login
				command = fmt.Sprintf("echo %q | docker login --username _json_key --password-stdin https://%s", gcpKeyJson, imageRepoHost)

			} else if awsSecretAccessKey != "" {
				// ECR login
				command = fmt.Sprintf("aws ecr get-login --no-include-email | bash")

			} else if aksSPPass != "" {
				// AKS & ACR Login
				command = fmt.Sprintf("az login --service-principal --username '%s' --tenant '%s' --password '%s';\n", aksSPAppID, aksTenantID, aksSPPass)
				command += fmt.Sprintf("az acr login --name '%s' --only-show-errors", imageRepoHost)
			}

			if command != "" {
				if debug == true {
					fmt.Printf("Command (not executed): %s\n", command)
				} else {
					out, err := exec.Command("bash", "-c", command).CombinedOutput()
					if err != nil {
						log.Fatal("Error: ", err)
					}
					fmt.Printf("Output: %s\n", out)
				}
			}
		}
	},
}

func init() {
	ciImageCmd.AddCommand(ciImageLoginCmd)

	ciImageLoginCmd.Flags().String("image-repo-host", "", "(Docker) container image repository url")
	ciImageLoginCmd.Flags().String("gcp-key-json", "", "Google Cloud service account key (plaintext, json)")
	ciImageLoginCmd.Flags().String("aws-secret-access-key", "", "Amazon Web Services IAM account key (string value)")
	ciImageLoginCmd.Flags().String("aks-tenant-id", "", "Azure Services tenant id")
	ciImageLoginCmd.Flags().String("aks-sp-app-id", "", "Azure Services servicePrincipal app id")
	ciImageLoginCmd.Flags().String("aks-sp-password", "", "Azure Services servicePrincipal password")
	ciImageLoginCmd.Flags().SortFlags = false
}
