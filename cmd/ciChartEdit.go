package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var editChartCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit charts",
	Run: func(cmd *cobra.Command, args []string) {

		deploymentFlag, _ := cmd.Flags().GetString("subchart-list-file")
		chartFlag, _ := cmd.Flags().GetString("chart-file")

		if len(deploymentFlag) < 1 && len(chartFlag) < 1 {
			log.Print("Both options must be passed")
			os.Exit(1)
		}
		var l = common.ReadCharts(deploymentFlag)

		var d = common.ReadChartDefinition(chartFlag)
		common.AppendExtraCharts(&l, &d)
		common.WriteChartDefinition(d, chartFlag)
	},
}

func init() {
	ciChartCmd.AddCommand(editChartCmd)
	// Local flags
	editChartCmd.Flags().String("subchart-list-file", "", "Location of custom chart YAML file")
	editChartCmd.Flags().String("chart-file", "", "Charts.yaml file to edit")
}
