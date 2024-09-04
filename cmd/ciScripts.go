package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ciScriptsCmd = &cobra.Command{
	Use:   "scripts",
	Short: "Convenience scripts for silta",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	rootCmd.AddCommand(ciScriptsCmd)
}
