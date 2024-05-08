package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var ciImageUrlCmd = &cobra.Command{
	Use:   "url",
	Short: "Calculate container image url based on build content",
	Run: func(cmd *cobra.Command, args []string) {

		// Calculate docker image tag

		imageRepoHost, _ := cmd.Flags().GetString("image-repo-host")
		imageRepoProject, _ := cmd.Flags().GetString("image-repo-project")
		namespace, _ := cmd.Flags().GetString("namespace")
		imageIdentifier, _ := cmd.Flags().GetString("image-identifier")
		imageTag, _ := cmd.Flags().GetString("image-tag")
		imageTagPrefix, _ := cmd.Flags().GetString("image-tag-prefix")
		dockerfile, _ := cmd.Flags().GetString("dockerfile")
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
				--absolute-names \
				--exclude='vendor/composer' \
				--exclude='vendor/autoload.php' \
				--mtime='2000-01-01 00:00Z' \
				--clamp-mtime \
				-cf - '%s' '%s' | sha1sum | cut -c 1-40 | tr -d $'\n'`,
				excludeDockerignore, buildPath, dockerfile)

			fileListing, err := exec.Command("bash", "-c", command).CombinedOutput()
			if err != nil {
				log.Fatal("Error (imageTag): ", err)
			}

			// Unless golang calculates checksum itself, passing plain output uses just too much memory.
			imageTag = string(fileListing)

			// Add prefix if it is specified
			if len(imageTagPrefix) > 0 {
				imageTag = imageTagPrefix + string('-') + imageTag
			}

			// Calculate hash sum
			// sha1_hash := fmt.Sprintf("%x", sha1.Sum([]byte(fileListing)))
			// imageTag = sha1_hash[0:40]
		}

		// Return Image url and tag
		fmt.Printf("%s:%s", imageUrl, imageTag)

	},
}

func init() {
	ciImageCmd.AddCommand(ciImageUrlCmd)

	ciImageUrlCmd.Flags().String("image-repo-host", "", "(Docker) container image repository url")
	ciImageUrlCmd.Flags().String("image-repo-project", "", "(Docker) image repository project (project name, i.e. \"silta\")")
	ciImageUrlCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciImageUrlCmd.Flags().String("image-identifier", "", "Docker image identifier (i.e. \"php\")")
	ciImageUrlCmd.Flags().String("image-tag", "", "Docker image tag (optional)")
	ciImageUrlCmd.Flags().String("image-tag-prefix", "", "Prefix for Docker image tag (optional)")
	ciImageUrlCmd.Flags().String("dockerfile", "", "Dockerfile (relative path)")
	ciImageUrlCmd.Flags().String("build-path", "", "Docker image build path")

	ciImageUrlCmd.MarkFlagRequired("image-repo-host")
	ciImageUrlCmd.MarkFlagRequired("image-repo-project")
	ciImageUrlCmd.MarkFlagRequired("namespace")
	ciImageUrlCmd.MarkFlagRequired("image-identifier")
	ciImageUrlCmd.MarkFlagRequired("dockerfile")
	ciImageUrlCmd.MarkFlagRequired("image-repo-host")
}
