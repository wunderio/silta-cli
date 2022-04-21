package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ciReleaseCleanfailedCmd = &cobra.Command{
	Use:   "clean-failed",
	Short: "Clean failed releases",
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")

		command := fmt.Sprintf(`
				NAMESPACE='%s'
				RELEASE_NAME='%s'
				
				failed_revision=$(helm list -n "${NAMESPACE}" --failed --pending --filter="(\s|^)(${RELEASE_NAME})(\s|$)" | tail -1 | cut -f3)

				if [[ "${failed_revision}" -eq 1 ]]; then
					# Remove any existing post-release hook, since it's technically not part of the release.
					kubectl delete job -n "${NAMESPACE}" "${RELEASE_NAME}-post-release" 2> /dev/null || true

					echo "Removing failed first release."
					helm delete -n "${NAMESPACE}" "${RELEASE_NAME}"

					echo "Delete persistent volume claims left over from statefulsets."
					kubectl delete pvc -n "${NAMESPACE}" -l release="${RELEASE_NAME}"
					kubectl delete pvc -n "${NAMESPACE}" -l app="${RELEASE_NAME}-es"

					echo -n "Waiting for volumes to be deleted."
					until [[ -z $(kubectl get pv | grep "${NAMESPACE}/${RELEASE_NAME}-") ]]
					do
					echo -n "."
					sleep 5
					done
				fi

				# Workaround for previous Helm release stuck in pending state
				pending_release=$(helm list -n "${NAMESPACE}" --pending --filter="(\s|^)(${RELEASE_NAME})(\s|$)"| tail -1 | cut -f1)

				if [[ "${pending_release}" == "${RELEASE_NAME}" ]]; then
					secret_to_delete=$(kubectl get secret -l owner=helm,status=pending-upgrade,name="${RELEASE_NAME}" -n "${NAMESPACE}" | awk '{print $1}' | grep -v NAME)
					kubectl delete secret -n "${NAMESPACE}" "${secret_to_delete}"
				fi
				`, namespace, releaseName)
		pipedExec(command, debug)
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseCleanfailedCmd)

	ciReleaseCleanfailedCmd.Flags().String("release-name", "", "Release name")
	ciReleaseCleanfailedCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")

	ciReleaseCleanfailedCmd.MarkFlagRequired("release-name")
	ciReleaseCleanfailedCmd.MarkFlagRequired("namespace")
}
