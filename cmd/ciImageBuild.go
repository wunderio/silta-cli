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
				if imageRepoHost == "gcr.io" || strings.HasSuffix(imageRepoHost, ".gcr.io") || strings.HasSuffix(imageRepoHost, ".pkg.dev") {
					_, useGCloud := os.LookupEnv("SILTA_USE_GCLOUD")
					if useGCloud {
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
						// If gcloud is not used, use docker
						gcpToken := common.GetGCPOAuth2Token()
						repositoryJWT := common.GetJWT(gcpToken, imageRepoHost, common.Image, imageRepoProject, imageIdentifier)
						tags := common.GCPListImageTagSiblings(repositoryJWT, namespace+"-"+imageIdentifier, imageRepoHost, imageRepoProject, imageTag)

						// if there tags are found, use the first one
						if len(tags) > 0 {
							fmt.Println("Image already exists, existing image will be used.")

							if len(extraImageTag) > 0 {
								if !common.HasString(tags, extraImageTag) {
									fmt.Printf("Image %s:%s already exists, but extra tag %s:%s does not exist yet, it will be added.\n", imageUrl, imageTag, imageUrl, extraImageTag)
									// Pull existing image
									err := exec.Command("bash", "-c", fmt.Sprintf("docker pull '%s:%s'", imageUrl, imageTag)).Run()
									if err != nil {
										log.Fatal("Error (docker pull): ", err)
									}
									// Add branch name tag to existing image
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
				} else if strings.HasSuffix(imageRepoHost, ".amazonaws.com") {

					command := fmt.Sprintf("aws ecr describe-images --repository-name='%s' --image-ids='imageTag=%s' 2>&1 > /dev/null", imageUrl, imageTag)
					err := exec.Command("bash", "-c", command).Run()
					if err == nil {
						fmt.Printf("Image %s:%s already exists, existing image will be used.", imageUrl, imageTag)

						// TODO: Add extra tag (branch name) if it does not exist yet
						if len(extraImageTag) > 0 {
							command := fmt.Sprintf("aws ecr describe-images --repository-name='%s' --image-ids='imageTag=%s' 2>&1 > /dev/null", imageUrl, branchName)
							err := exec.Command("bash", "-c", command).Run()
							if err != nil {
								// TODO: docker pull before tagging and pushing

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

					// imageUrl := fmt.Sprintf("%s/%s/%s-%s", imageRepoHost, imageRepoProject, namespace, imageIdentifier)

					auth, err := common.GetDockerAuth(imageRepoHost)
					if err != nil {
						log.Fatal("Error (get docker auth): ", err)
					}

					fmt.Printf("auth: %s\n", auth)
					// Get JWT token using docker hub credentials
					repositoryJWT := common.GetACRJWT(auth["identitytoken"], imageRepoHost, common.Image, imageRepoProject, namespace+"-"+imageIdentifier)

					fmt.Printf("repositoryJWT: %s\n", repositoryJWT)

					tags := common.ACRListImageTagSiblings(repositoryJWT, namespace+"-"+imageIdentifier, imageRepoHost, imageRepoProject, imageTag)

					fmt.Println("tags:", tags)
					os.Exit(0)

					// ------------------------- docker cli

					// // command := fmt.Sprintf("docker manifest inspect '%s:%s' > /dev/null 2>&1", imageUrl, imageTag)
					// command := fmt.Sprintf("docker manifest inspect '%s:%s' | jq -r '.manifests[].platform.digest'", imageUrl, imageTag)

					// checkSumImageDigest, err := exec.Command("bash", "-c", command).Output()

					// fmt.Println("Command:", command)
					// fmt.Printf("Checksum image digest: %s\n", checkSumImageDigest)

					// if err == nil {
					// 	fmt.Printf("Image %s:%s already exists, existing image will be used.", imageUrl, imageTag)

					// 	fmt.Printf("Checksum image digest: %s\n", checkSumImageDigest)

					// 	// Add extra tag (branch name) if it does not exist yet
					// 	if len(extraImageTag) > 0 {

					// 		// Get digest for image with checksum tag
					// 		// Get digest for image with branch name tag
					// 		// If digests are the same, don't push
					// 		// If digests are different, push

					// 		command := fmt.Sprintf("docker manifest inspect '%s/%s/%s-%s:%s' | jq -r '.manifests[0].digest'", imageRepoHost, imageRepoProject, namespace, imageIdentifier, imageTag)
					// 		digest, err := exec.Command("bash", "-c", command).Output()
					// 		if err != nil {

					// 			log.Fatal("Error (docker manifest inspect): ", err)
					// 		}

					// 		fmt.Println("Command:", command)
					// 		fmt.Printf("Digest: %s\n", digest)

					// 		// Get existing tags for that particular image with checksum tag
					// 		// TODO: do this like gcloud, parse json output
					// 		// command := fmt.Sprintf("docker manifest inspect '%s:%s' | jq -r '.manifests[].platform.digest'", imageUrl, imageTag)
					// 		// fmt.Println("Command:", command)
					// 		// existingTags, err := exec.Command("bash", "-c", command).Output()
					// 		// if err != nil {
					// 		// 	log.Fatal("Error (docker manifest inspect): ", err)
					// 		// }
					// 		// fmt.Printf("Existing tags: %s\n", existingTags)

					// 		// -------------------

					// 		// TODO: get existing tags for that particular image with checksum tag and check if branch name tag already exists

					// 		command = fmt.Sprintf("docker manifest inspect '%s/%s/%s-%s:%s' > /dev/null 2>&1", imageRepoHost, imageRepoProject, namespace, imageIdentifier, extraImageTag)
					// 		err = exec.Command("bash", "-c", command).Run()
					// 		if err != nil {
					// 			// Get digest of image
					// 			command := fmt.Sprintf("docker manifest inspect '%s/%s/%s-%s:%s' | jq -r '.manifests[0].digest'", imageRepoHost, imageRepoProject, namespace, imageIdentifier, imageTag)
					// 			digest, err := exec.Command("bash", "-c", command).Output()
					// 			if err != nil {
					// 				log.Fatal("Error (docker manifest inspect): ", err)
					// 			}
					// 			// TODO: docker pull before tagging and pushing

					// 			// Tag image
					// 			err = exec.Command("bash", "-c", fmt.Sprintf("docker tag '%s/%s/%s-%s@%s' '%s/%s/%s-%s:%s'", imageRepoHost, imageRepoProject, namespace, imageIdentifier, digest, imageRepoHost, imageRepoProject, namespace, imageIdentifier, extraImageTag)).Run()
					// 			if err != nil {
					// 				log.Fatal("Error (docker tag): ", err)
					// 			}
					// 			// Push image
					// 			err = exec.Command("bash", "-c", fmt.Sprintf("docker push '%s/%s/%s-%s:%s'", imageRepoHost, imageRepoProject, namespace, imageIdentifier, extraImageTag)).Run()
					// 			if err != nil {
					// 				log.Fatal("Error (docker push): ", err)
					// 			}
					// 		}
					// 	}

					// 	return
					// }

				} else {
					// Generic docker registry, e.g. docker.io
					// Reuse docker cli credentials
					authenticator := remote.WithAuthFromKeychain(authn.DefaultKeychain)
					// Get image tag siblings
					imageTagSiblings := common.ListImageTagSiblings(authenticator, imageUrl, imageTag)

					// Check if image tag already exists
					if common.HasString(imageTagSiblings, imageTag) {
						fmt.Printf("Image %s:%s already exists, existing image will be used.\n", imageUrl, imageTag)

						// Add extra tag (branch name) if it does not exist yet
						if len(extraImageTag) > 0 {
							if !common.HasString(imageTagSiblings, extraImageTag) {
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
