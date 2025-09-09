package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var ciReleaseCleanfailedCmd = &cobra.Command{
	Use:   "clean-failed",
	Short: "Clean failed releases",
	Run: func(cmd *cobra.Command, args []string) {
		releaseName, _ := cmd.Flags().GetString("release-name")
		namespace, _ := cmd.Flags().GetString("namespace")

		common.FailedReleaseCleanup(releaseName, namespace)
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseCleanfailedCmd)

	ciReleaseCleanfailedCmd.Flags().String("release-name", "", "Release name")
	ciReleaseCleanfailedCmd.Flags().String("namespace", "", "Project name (namespace, i.e. \"drupal-project\")")

	ciReleaseCleanfailedCmd.MarkFlagRequired("release-name")
	ciReleaseCleanfailedCmd.MarkFlagRequired("namespace")
}
