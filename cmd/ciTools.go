package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ciToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "CI tooling",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	rootCmd.AddCommand(ciToolsCmd)
}
