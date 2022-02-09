package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var ciReleaseInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print release information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ciReleaseInfo called")

		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")
		githubToken, _ := cmd.Flags().GetString("github-token")
		circlePrNumber, _ := cmd.Flags().GetString("pr-number")
		circlePullRequest, _ := cmd.Flags().GetString("pull-request")
		circleProjectNumber, _ := cmd.Flags().GetString("circle-project-number")
		circleProjectReponame, _ := cmd.Flags().GetString("project-reponame")

		// Use environment variables as fallback
		if useEnv == true {
			if len(githubToken) == 0 {
				githubToken = os.Getenv("GITHUB_TOKEN")
			}
			if len(circlePrNumber) == 0 {
				circlePrNumber = os.Getenv("CIRCLE_PR_NUMBER")
			}
			if len(circlePullRequest) == 0 {
				circlePullRequest = os.Getenv("CIRCLE_PULL_REQUEST")
			}
			if len(circleProjectNumber) == 0 {
				circleProjectNumber = os.Getenv("CIRCLE_PROJECT_USERNAME")
			}
			if len(circleProjectReponame) == 0 {
				circleProjectReponame = os.Getenv("CIRCLE_PROJECT_REPONAME")
			}
		}

		// Post release info to Github
		command := fmt.Sprintf(`
				NAMESPACE='%s'
				RELEASE_NAME='%s'
				GITHUB_TOKEN='%s'
				CIRCLE_PR_NUMBER='%s'
				CIRCLE_PULL_REQUEST='%s'
				CIRCLE_PROJECT_USERNAME='%s'
				CIRCLE_PROJECT_REPONAME='%s'
				
				if [ -n "$GITHUB_TOKEN" ]
				then
					RELEASE_NOTES=$(helm -n "$NAMESPACE" get notes "$RELEASE_NAME")
					if [ -z "$CIRCLE_PR_NUMBER" ]
					then
						CIRCLE_PR_NUMBER="${CIRCLE_PULL_REQUEST//[^0-9]/}"
					fi
					GITHUB_API_URL="https://api.github.com/repos/$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME/issues/$CIRCLE_PR_NUMBER/comments"
					FORMATTED_NOTES=$(echo "$RELEASE_NOTES" | sed 's/\\/\\\\/g' | sed 's/"/\\\"/g' | sed 's/\$/\\n/g' | sed 's/\//\\\//g')
					curl -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: application/json" -X POST -d '{"body": "<details><summary>Release notes for '$RELEASE_NAME'</summary>'"${FORMATTED_NOTES//$'\n'/'\n'}"'</details>"}' $GITHUB_API_URL
				fi
				`, namespace, releaseName, githubToken, circlePrNumber, circlePullRequest, circleProjectNumber, circleProjectReponame)
		pipedExec(command, debug)

		command = fmt.Sprintf(`
				NAMESPACE='%s'
				RELEASE_NAME='%s'
				# Display only the part following NOTES from the helm status.
				helm -n "$NAMESPACE" get notes "$RELEASE_NAME"
			`, namespace, releaseName)
		pipedExec(command, debug)

	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseInfoCmd)

	ciReleaseInfoCmd.Flags().String("release-name", "", "Release name")
	ciReleaseInfoCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")
	ciReleaseInfoCmd.Flags().String("github-token", "", "Github token for posting release status to PR")
	ciReleaseInfoCmd.Flags().String("pr-number", "", "PR number")
	ciReleaseInfoCmd.Flags().String("pull-request", "", "Pull request url")
	ciReleaseInfoCmd.Flags().String("project-reponame", "", "Project repository url")
	ciReleaseInfoCmd.Flags().String("circle-project-number", "", "CircleCI project number")

	ciReleaseInfoCmd.MarkFlagRequired("release-name")
	ciReleaseInfoCmd.MarkFlagRequired("namespace")
}
