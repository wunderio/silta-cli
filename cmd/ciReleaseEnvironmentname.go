package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/wunderio/silta-cli/internal/common"

	"github.com/spf13/cobra"
)

var ciReleaseEnvironmentnameCmd = &cobra.Command{
	Use:   "environmentname",
	Short: "Return environment name",
	Long: `Generate enviornment name based on branchname and release-suffix. 
		This is used in some helm charts`,
	Run: func(cmd *cobra.Command, args []string) {

		branchname, _ := cmd.Flags().GetString("branchname")
		releaseSuffix, _ := cmd.Flags().GetString("release-suffix")

		// Environment value fallback
		if useEnv == true {
			if branchname == "" {
				branchname = os.Getenv("CIRCLE_BRANCH")
			}
		}

		if branchname == "" {
			log.Fatal("Repository branchname not provided")
		}

		siltaEnvironmentName := common.SiltaEnvironmentName(branchname, releaseSuffix)

		fmt.Print(siltaEnvironmentName)
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseEnvironmentnameCmd)

	ciReleaseEnvironmentnameCmd.Flags().String("branchname", "", "Repository branchname that will be used for release name")
	ciReleaseEnvironmentnameCmd.Flags().String("release-suffix", "", "Release name suffix")
}
