package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ciImageCmd = &cobra.Command{
	Use:   "image",
	Short: "CI (docker) image commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	ciCmd.AddCommand(ciImageCmd)
}
