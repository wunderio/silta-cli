package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ciChartCmd = &cobra.Command{
	Use:   "chart",
	Short: "CI chart commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	rootCmd.AddCommand(ciChartCmd)
}
