package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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
			extraImageTag = fmt.Sprintf("branch--%s", branchName)
		}

		// Reuse existing image if it exists
		if !debug {
			if reuseExisting {
				_, useGCloud := os.LookupEnv("SILTA_USE_GCLOUD")

				if useGCloud && (imageRepoHost == "gcr.io" || strings.HasSuffix(imageRepoHost, ".gcr.io") || strings.HasSuffix(imageRepoHost, ".pkg.dev")) {

					// Get image tags via gcloud
					command := fmt.Sprintf("gcloud container images list-tags '%s' --filter='tags:%s' --format=json", imageUrl, imageTag)
					output, err := exec.Command("bash", "-c", command).CombinedOutput()
					if err != nil {
						log.Fatal("Error (gcloud list-tags): ", err)
					}
					// Unmarshal or Decode the JSON to the interface.
					type TagList struct {
						Digest string   `json:"digest"`
						Tags   []string `json:"tags"`
					}
					var taglist []TagList
					err = json.Unmarshal([]byte(output), &taglist)
					if err != nil {
						log.Fatal("Error (json unmarshal): ", err)
					}
					var tagExists bool = false
					var extraTagExists bool = false
					for _, tag := range taglist {
						for _, t := range tag.Tags {
							if t == imageTag {
								tagExists = true
							}
							if len(extraImageTag) > 0 && t == extraImageTag {
								extraTagExists = true
							}
						}
					}
					// If tag exists in taglist, return and don't rebuild.
					if tagExists {
						fmt.Printf("Image %s:%s already exists, existing image will be used.\n", imageUrl, imageTag)
						// Add extra tag if it does not exist yet
						if len(extraImageTag) > 0 && !extraTagExists {
							fmt.Printf("Image %s:%s already exists, but extra tag %s:%s does not exist yet, it will be added.\n", imageUrl, imageTag, imageUrl, extraImageTag)
							command := fmt.Sprintf("gcloud container images add-tag '%s:%s' '%s:%s'", imageUrl, imageTag, imageUrl, extraImageTag)
							err = exec.Command("bash", "-c", command).Run()
							if err != nil {
								log.Fatal("Error (gcloud add-tag): ", err)
							}
						}
						return
					}
				} else {
					// Generic docker registry, e.g. docker.io
					// Supports ACR, AR, GCR and ECR

					// Reuse docker cli credentials
					authenticator := remote.WithAuthFromKeychain(authn.DefaultKeychain)

					imageTag_digest := common.GetImageTagDigest(authenticator, imageUrl, imageTag)

					if imageTag_digest != "" {
						fmt.Printf("Image %s:%s already exists, existing image will be used.\n", imageUrl, imageTag)

						// Add extra tag (branch name) if it does not exist yet
						if len(extraImageTag) > 0 {

							extraImageTag_digest := common.GetImageTagDigest(authenticator, imageUrl, extraImageTag)

							imageTag_digest := common.GetImageTagDigest(authenticator, imageUrl, imageTag)
							if extraImageTag_digest == "" || extraImageTag_digest != imageTag_digest {
								// Have to pull images, manifest creation is unreliable due to digest differences
								// https://github.com/docker/hub-feedback/issues/1925
								fmt.Printf("Image tag %s:%s already exists, but extra tag %s:%s does not exist yet, it will be added.\n", imageUrl, imageTag, imageUrl, extraImageTag)
								// Pull image, tag it and push it
								err := exec.Command("bash", "-c", fmt.Sprintf("docker pull '%s:%s'", imageUrl, imageTag)).Run()
								if err != nil {
									log.Fatal("Error (docker pull): ", err)
								}
								err = exec.Command("bash", "-c", fmt.Sprintf("docker tag '%s:%s' '%s:%s'", imageUrl, imageTag, imageUrl, extraImageTag)).Run()
								if err != nil {
									log.Fatal("Error (docker tag): ", err)
								}
								err = exec.Command("bash", "-c", fmt.Sprintf("docker push '%s:%s'", imageUrl, extraImageTag)).Run()
								if err != nil {
									log.Fatal("Error (docker push): ", err)
								}
							}
						}
						return
					}
				}
			}
		}

		// Run docker build
		extraImageTagString := ""
		if len(extraImageTag) > 0 {
			extraImageTagString = fmt.Sprintf("--tag '%s:%s'", imageUrl, extraImageTag)
		}
		command := fmt.Sprintf("docker build --tag '%s:%s' %s -f '%s' %s", imageUrl, imageTag, extraImageTagString, dockerfile, buildPath)
		pipedExec(command, debug)

		// Create AWS/ECR repository (ECR requires a dedicated repository per project)
		if strings.HasSuffix(imageRepoHost, ".amazonaws.com") {

			command = fmt.Sprintf("aws ecr describe-repositories --repository-name '%s/%s-%s'", imageRepoProject, namespace, imageIdentifier)
			err := exec.Command("bash", "-c", command).Run()
			if err != nil {

				command = fmt.Sprintf("aws ecr create-repository --repository-name '%s/%s-%s'", imageRepoProject, namespace, imageIdentifier)
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
			command = fmt.Sprintf("docker push '%s:%s'", imageUrl, extraImageTag)
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
