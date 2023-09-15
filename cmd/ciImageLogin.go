package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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

  * General (optional):
    - "--image-repo-user" flag or "IMAGE_REPO_USER" environment variable: (Docker) container image repository user
    - "--image-repo-pass" flag or "IMAGE_REPO_PASS" environment variable: (Docker) container image repository password

  * Google Cloud:
    - "--gcp-key-json" flag or "GCP_KEY_JSON" environment variable: Google Cloud service account key (string value)

  * Amazon Web Services:
    - "--aws-secret-access-key" flag or "AWS_SECRET_ACCESS_KEY" environment variable: Amazon Web Services IAM account key (string value)
		- "--aws-region" flag or "AWS_REGION" environment variable: Region of the container repository

  * Azure Services:
    - "--aks-tenant-id" flag or "AKS_TENANT_ID" environment variable: Azure Services tenant id
    - "--aks-sp-app-id" flag or "AKS_SP_APP_ID" environment variable: Azure Services servicePrincipal app id
    - "--aks-sp-password" flag or "AKS_SP_PASSWORD" environment variable: Azure Services servicePrincipal password
`,
	Run: func(cmd *cobra.Command, args []string) {

		// Read flags into variables
		imageRepoHost, _ := cmd.Flags().GetString("image-repo-host")
		imageRepoTLS, _ := cmd.Flags().GetBool("image-repo-tls")
		imageRepoUser, _ := cmd.Flags().GetString("image-repo-user")
		imageRepoPass, _ := cmd.Flags().GetString("image-repo-pass")
		gcpKeyJson, _ := cmd.Flags().GetString("gcp-key-json")
		awsSecretAccessKey, _ := cmd.Flags().GetString("aws-secret-access-key")
		awsRegion, _ := cmd.Flags().GetString("aws-region")
		aksTenantID, _ := cmd.Flags().GetString("aks-tenant-id")
		aksSPAppID, _ := cmd.Flags().GetString("aks-sp-app-id")
		aksSPPass, _ := cmd.Flags().GetString("aks-sp-password")

		// Use environment variables as fallback
		if useEnv {
			if len(gcpKeyJson) == 0 {
				gcpKeyJson = os.Getenv("GCLOUD_KEY_JSON")
			}
			if len(gcpKeyJson) == 0 {
				gcpKeyJson = os.Getenv("GCP_KEY_JSON")
			}
			if len(awsSecretAccessKey) == 0 {
				awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
			}
			if len(awsRegion) == 0 {
				awsRegion = os.Getenv("AWS_REGION")
			}
			if len(imageRepoHost) == 0 {
				imageRepoHost = os.Getenv("IMAGE_REPO_HOST")
			}
			if len(imageRepoUser) == 0 {
				imageRepoUser = os.Getenv("IMAGE_REPO_USER")
			}
			if len(imageRepoPass) == 0 {
				imageRepoPass = os.Getenv("IMAGE_REPO_PASS")
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
			fmt.Println("IMAGE_REPO_TLS:", imageRepoTLS)
			fmt.Println("IMAGE_REPO_USER:", imageRepoUser)
			fmt.Println("IMAGE_REPO_PASS:", imageRepoPass)
			fmt.Println("GCLOUD_KEY_JSON:", gcpKeyJson)
			fmt.Println("AWS_SECRET_ACCESS_KEY:", awsSecretAccessKey)
			fmt.Println("AWS_REGION:", awsRegion)
			fmt.Println("AKS_TENANT_ID:", aksTenantID)
			fmt.Println("AKS_SP_APP_ID:", aksSPAppID)
			fmt.Println("AKS_SP_PASSWORD:", aksSPPass)
		}

		if len(imageRepoUser) == 0 && len(imageRepoPass) == 0 && len(gcpKeyJson) == 0 && len(awsSecretAccessKey) == 0 && len(aksSPPass) == 0 {
			log.Fatal("Docker registry credentials are empty, have you set a context for this CircleCI job correctly?")
		} else {

			command := ""

			if imageRepoUser != "" {
				// Allow insecure registries (local dev)
				protocol := "https://"
				if !imageRepoTLS {
					protocol = "http://"
				}
				// User && pass login
				command = fmt.Sprintf("echo %q | docker login --username %q --password-stdin %s%s", imageRepoPass, imageRepoUser, protocol, imageRepoHost)

			} else if gcpKeyJson != "" {
				// GCR login
				if !strings.Contains(imageRepoHost, "://") {
					imageRepoHost = "https://" + imageRepoHost
				}
				command = fmt.Sprintf("echo %q | docker login --username %q --password-stdin %s", gcpKeyJson, "_json_key", imageRepoHost)

			} else if awsSecretAccessKey != "" {
				//Get AWS Account ID
				awsAccountId, err := exec.Command("aws sts get-caller-identity --query \"Account\" --output text --no-cli-pager").Output()
				if err != nil {
					log.Fatal("Error:", err)
				}
				// ECR login
				command = fmt.Sprintf("aws ecr get-login-password --region %q | docker login --username AWS --password-stdin %s.dkr.ecr.%q.amazonaws.com", awsRegion, awsAccountId, awsRegion)
				// TODO: use aws cli v2
				// command = fmt.Sprintf("echo %q | docker login --username AWS --password-stdin %s", awsSecretAccessKey, imageRepoHost)

			} else if aksSPPass != "" {
				// ACR Login
				command = fmt.Sprintf("echo %q | docker login --username %q --password-stdin %s", aksSPPass, aksSPAppID, imageRepoHost)
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
	ciImageLoginCmd.Flags().Bool("image-repo-tls", true, "(Docker) container image repository url tls (enabled by default)")
	ciImageLoginCmd.Flags().String("image-repo-user", "", "(Docker) container image repository username")
	ciImageLoginCmd.Flags().String("image-repo-pass", "", "(Docker) container image repository password")
	ciImageLoginCmd.Flags().String("gcp-key-json", "", "Google Cloud service account key (plaintext, json)")
	ciImageLoginCmd.Flags().String("aws-secret-access-key", "", "Amazon Web Services IAM account key (string value)")
	ciImageLoginCmd.Flags().String("aws-region", "", "Elastic Container Registry region (string value)")
	ciImageLoginCmd.Flags().String("aks-tenant-id", "", "Azure Services tenant id")
	ciImageLoginCmd.Flags().String("aks-sp-app-id", "", "Azure Services servicePrincipal app id")
	ciImageLoginCmd.Flags().String("aks-sp-password", "", "Azure Services servicePrincipal password")
	ciImageLoginCmd.Flags().SortFlags = false
}
