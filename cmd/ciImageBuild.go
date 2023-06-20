package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

// buildCmd represents the build command
var ciImageBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build and push container image",
	Run: func(cmd *cobra.Command, args []string) {

		imageRepoHost, _ := cmd.Flags().GetString("image-repo-host")
		imageRepoProject, _ := cmd.Flags().GetString("image-repo-project")
		namespace, _ := cmd.Flags().GetString("namespace")
		branchName, _ := cmd.Flags().GetString("branchname")
		imageIdentifier, _ := cmd.Flags().GetString("image-identifier")
		imageTag, _ := cmd.Flags().GetString("image-tag")
		dockerfile, _ := cmd.Flags().GetString("dockerfile")
		reuseExisting, _ := cmd.Flags().GetBool("image-reuse")
		buildPath, _ := cmd.Flags().GetString("build-path")

		// Use environment variables as fallback
		if useEnv == true {
			if len(imageRepoHost) == 0 {
				imageRepoHost = os.Getenv("IMAGE_REPO_HOST")
			}
			if len(imageRepoHost) == 0 {
				imageRepoHost = os.Getenv("DOCKER_REPO_HOST")
			}
			if len(imageRepoProject) == 0 {
				imageRepoProject = os.Getenv("DOCKER_REPO_PROJ")
			}
			if len(namespace) == 0 {
				namespace = os.Getenv("NAMESPACE")
			}
		}

		imageUrl := fmt.Sprintf("%s/%s/%s-%s", imageRepoHost, imageRepoProject, namespace, imageIdentifier)

		// Only use .dockerignore files if they exist
		excludeDockerignore := ""
		if _, err := os.Stat(fmt.Sprintf("%s/.dockerignore", buildPath)); err == nil {
			excludeDockerignore = fmt.Sprintf("--exclude-from='%s'/.dockerignore", buildPath)
		}

		// If no path is specified, build from an empty directory
		if len(buildPath) == 0 {
			// Create kubeconfig folder
			if _, err := os.Stat("/tmp/empty"); os.IsNotExist(err) {
				_ = os.Mkdir("/tmp/empty", 0775)
			}
			buildPath = "/tmp/empty"
		}

		// No tag has been defined
		// Calculate a hash sum of files in the folder except those ignored by docker.
		// Also make sure modification time or order play no role.
		if len(imageTag) == 0 {
			command := fmt.Sprintf(`tar \
				--sort=name %s \
				--exclude='vendor/composer' \
				--exclude='vendor/autoload.php' \
				--mtime='2000-01-01 00:00Z' \
				--clamp-mtime \
				-cf - '%s' '%s' | sha1sum | cut -c 1-40 | tr -d $'\n'`,
				excludeDockerignore, buildPath, dockerfile)

			fileListing, err := exec.Command("bash", "-c", command).CombinedOutput()
			if err != nil {
				log.Fatal("Error (file checksum): ", err)
			}
			// Unless golang calculates checksum itself, passing plain output uses just too much memory.
			imageTag = string(fileListing)

			// Calculate hash sum
			// sha1_hash := fmt.Sprintf("%x", sha1.Sum([]byte(fileListing)))
			// imageTag = sha1_hash[0:40]
		}

		// Add extra image tag for image identification
		extraImageTag := ""
		if len(branchName) > 0 {
			// Make sure release name is lowercase without special characters.
			branchName = strings.ToLower(branchName)
			reg, _ := regexp.Compile("[^[:alnum:]]")
			branchName = reg.ReplaceAllString(branchName, "-")
			extraImageTag = fmt.Sprintf("--tag '%s:%s'", imageUrl, branchName)
		}

		// Reuse existing image if it exists

		if !debug {
			if reuseExisting {
				if imageRepoHost == "gcr.io" || strings.HasSuffix(imageRepoHost, ".gcr.io") || strings.HasSuffix(imageRepoHost, ".pkg.dev") {
					_, useGCloud := os.LookupEnv("SILTA_USE_GCLOUD")
					if useGCloud {
						command := fmt.Sprintf("gcloud container images list-tags '%s' --filter='tags:%s' --format=json | grep -q '\"%s\"';", imageUrl, imageTag, imageTag)
						err := exec.Command("bash", "-c", command).Run()

						if err == nil {
							fmt.Printf("Image %s:%s already exists, existing image will be used.", imageUrl, imageTag)

							// Add extra tag (branch name) if it does not exist yet
							if len(extraImageTag) > 0 {
								command := fmt.Sprintf("gcloud container images list-tags '%s' --filter='tags:%s' --format=json | grep -q '\"%s\"';", imageUrl, branchName, branchName)
								err := exec.Command("bash", "-c", command).Run()
								if err != nil {
									// Adding branch name tag to existing image
									command := fmt.Sprintf("gcloud container images add-tag '%s:%s' '%s:%s'", imageUrl, imageTag, imageUrl, branchName)
									err := exec.Command("bash", "-c", command).Run()
									if err != nil {
										log.Fatal("Error (gcloud tag): ", err)
									}
								}
							}

							return
						}

					} else {

						gcpToken := common.GetGCPOAuth2Token()
						repositoryJWT := common.GetGCPJWT(gcpToken, imageRepoHost, common.Image, imageRepoProject, imageIdentifier)
						tags := common.GCPListTags(repositoryJWT, namespace+"-"+imageIdentifier, imageRepoHost, imageRepoProject)
						if common.HasString(tags, imageTag) {
							fmt.Printf("Image %s:%s already exists, existing image will be used.", imageUrl, imageTag)

							// Add extra tag (branch name) if it does not exist yet
							if len(extraImageTag) > 0 {
								if !common.HasString(tags, branchName) {
									// Adding branch name tag to existing image
									err := exec.Command("bash", "-c", fmt.Sprintf("docker tag '%s:%s' '%s:%s'", imageUrl, imageTag, imageUrl, branchName)).Run()
									if err != nil {
										log.Fatal("Error (docker tag): ", err)
									}
									err = exec.Command("bash", "-c", fmt.Sprintf("docker push '%s:%s'", imageUrl, branchName)).Run()
									if err != nil {
										log.Fatal("Error (docker push): ", err)
									}
								}
							}

							return
						}
					}
				} else if strings.HasSuffix(imageRepoHost, ".amazonaws.com") {

					command := fmt.Sprintf("aws ecr describe-images --repository-name='%s' --image-ids='imageTag=%s' 2>&1 > /dev/null", imageUrl, imageTag)
					err := exec.Command("bash", "-c", command).Run()
					if err == nil {
						fmt.Printf("Image %s:%s already exists, existing image will be used.", imageUrl, imageTag)

						// Add extra tag (branch name) if it does not exist yet
						if len(extraImageTag) > 0 {
							command := fmt.Sprintf("aws ecr describe-images --repository-name='%s' --image-ids='imageTag=%s' 2>&1 > /dev/null", imageUrl, branchName)
							err := exec.Command("bash", "-c", command).Run()
							if err != nil {
								// Adding branch name tag to existing image
								err := exec.Command("bash", "-c", fmt.Sprintf("docker tag '%s:%s' '%s:%s'", imageUrl, imageTag, imageUrl, branchName)).Run()
								if err != nil {
									log.Fatal("Error (docker tag): ", err)
								}
								err = exec.Command("bash", "-c", fmt.Sprintf("docker push '%s:%s'", imageUrl, branchName)).Run()
								if err != nil {
									log.Fatal("Error (docker push): ", err)
								}
							}
						}

						return
					}
				} else if strings.HasSuffix(imageRepoHost, ".azurecr.io") {

					imageUrl := fmt.Sprintf("%s/%s/%s-%s", imageRepoHost, imageRepoProject, namespace, imageIdentifier)

					command := fmt.Sprintf("docker manifest inspect '%s/%s/%s-%s:%s' > /dev/null 2>&1", imageRepoHost, imageRepoProject, namespace, imageIdentifier, imageTag)
					err := exec.Command("bash", "-c", command).Run()
					if err == nil {
						fmt.Printf("Image %s:%s already exists, existing image will be used.", imageUrl, imageTag)

						// Add extra tag (branch name) if it does not exist yet
						if len(extraImageTag) > 0 {
							command := fmt.Sprintf("docker manifest inspect '%s/%s/%s-%s:%s' > /dev/null 2>&1", imageRepoHost, imageRepoProject, namespace, imageIdentifier, branchName)
							err := exec.Command("bash", "-c", command).Run()
							if err != nil {
								// Get digest of image
								command := fmt.Sprintf("docker manifest inspect '%s/%s/%s-%s:%s' | jq -r '.manifests[0].digest'", imageRepoHost, imageRepoProject, namespace, imageIdentifier, imageTag)
								digest, err := exec.Command("bash", "-c", command).Output()
								if err != nil {

									log.Fatal("Error (docker manifest inspect): ", err)
								}
								// Tag image
								err = exec.Command("bash", "-c", fmt.Sprintf("docker tag '%s/%s/%s-%s@%s' '%s/%s/%s-%s:%s'", imageRepoHost, imageRepoProject, namespace, imageIdentifier, digest, imageRepoHost, imageRepoProject, namespace, imageIdentifier, branchName)).Run()
								if err != nil {
									log.Fatal("Error (docker tag): ", err)
								}
								// Push image
								err = exec.Command("bash", "-c", fmt.Sprintf("docker push '%s/%s/%s-%s:%s'", imageRepoHost, imageRepoProject, namespace, imageIdentifier, branchName)).Run()
								if err != nil {
									log.Fatal("Error (docker push): ", err)
								}
							}
						}

						return
					}

				} else {
					// Generic docker registry, e.g. docker.io
					imageUrl := fmt.Sprintf("%s/%s-%s", imageRepoHost, imageRepoProject, imageIdentifier)

					// Use generic docker registry API to check if image exists
					imageRepoUser := os.Getenv("IMAGE_REPO_USER")
					imageRepoPassword := os.Getenv("IMAGE_REPO_PASSWORD")

					err := exec.Command("bash", "-c", fmt.Sprintf("curl -s -f -lSL -I -o /dev/null -w '%%{http_code}' -u '%s:%s' https://%s/v2/%s/tags/list", imageRepoUser, imageRepoPassword, imageRepoHost, imageUrl)).Run()
					if err == nil {
						fmt.Printf("Image %s:%s already exists, existing image will be used.", imageUrl, imageTag)
						// TODO: Add extra tag (branch name) if it does not exist yet
						return
					}

				}
			}
		}

		// Run docker build
		command := fmt.Sprintf("docker build --tag '%s:%s' %s -f '%s' %s", imageUrl, imageTag, extraImageTag, dockerfile, buildPath)
		pipedExec(command, debug)

		// Create AWS/ECR repository (ECR requires a dedicated repository per project)
		if strings.HasSuffix(imageRepoHost, ".amazonaws.com") {
			command = fmt.Sprintf("aws ecr describe-repositories --repository-name '%s'", imageUrl)
			err := exec.Command("bash", "-c", command).Run()
			if err != nil {
				command = fmt.Sprintf("aws ecr create-repository --repository-name '%s'", imageUrl)
				err = exec.Command("bash", "-c", command).Run()
				if err != nil {
					log.Fatal("Error (aws ecr create-repository): ", err)
				}
			}
		}

		// Image push
		command = fmt.Sprintf("docker push '%s:%s'", imageUrl, imageTag)
		pipedExec(command, debug)

		// Push extra tags
		if len(branchName) > 0 {
			command = fmt.Sprintf("docker push '%s:%s'", imageUrl, branchName)
			pipedExec(command, debug)
		}
	},
}

func init() {
	ciImageCmd.AddCommand(ciImageBuildCmd)

	ciImageBuildCmd.Flags().String("image-repo-host", "", "(Docker) container image repository url")
	ciImageBuildCmd.Flags().String("image-repo-project", "", "(Docker) image repository project (project name, i.e. \"silta\")")
	ciImageBuildCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciImageBuildCmd.Flags().String("branchname", "", "Branch name (used as an extra tag for image identification)")
	ciImageBuildCmd.Flags().String("image-identifier", "", "Docker image identifier (i.e. \"php\")")
	ciImageBuildCmd.Flags().String("image-tag", "", "Docker image tag (optional, check '--image-reuse' flag)")
	ciImageBuildCmd.Flags().String("dockerfile", "", "Dockerfile (relative path)")
	ciImageBuildCmd.Flags().String("build-path", "", "Docker image build path")
	ciImageBuildCmd.Flags().Bool("image-reuse", true, "Do not rebuild image if identical image:tag exists in remote")

	ciImageBuildCmd.MarkFlagRequired("image-repo-host")
	ciImageBuildCmd.MarkFlagRequired("image-repo-project")
	ciImageBuildCmd.MarkFlagRequired("namespace")
	ciImageBuildCmd.MarkFlagRequired("image-identifier")
	ciImageBuildCmd.MarkFlagRequired("dockerfile")
}
