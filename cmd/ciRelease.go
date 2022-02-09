package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ciReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "CI release related commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	ciCmd.AddCommand(ciReleaseCmd)
}
