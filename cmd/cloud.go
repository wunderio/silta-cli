package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cloudCmd represents the cloud command
var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Kubernetes cloud related commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	rootCmd.AddCommand(cloudCmd)
}
