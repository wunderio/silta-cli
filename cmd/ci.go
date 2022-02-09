package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ciCmd represents the ci command
var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Silta CI Commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	rootCmd.AddCommand(ciCmd)
}
