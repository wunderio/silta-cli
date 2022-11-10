package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Silta CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", common.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
