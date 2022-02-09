package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// secretsCmd represents the secrets command
var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage encrypted secret files",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	rootCmd.AddCommand(secretsCmd)
}
