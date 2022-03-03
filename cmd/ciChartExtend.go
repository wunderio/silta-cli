package cmd

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var editChartCmd = &cobra.Command{
	Use:   "extend",
	Short: "Extend charts",
	Long:  "Adds subcharts to main chart file",
	Run: func(cmd *cobra.Command, args []string) {

		deploymentFlag, _ := cmd.Flags().GetString("subchart-list-file")
		chartName, _ := cmd.Flags().GetString("chart-name")
		const innerChartFile = "/Chart.yaml"

		if len(deploymentFlag) < 1 && len(chartName) < 1 {
			log.Print("Both options must be passed")
			os.Exit(1)
		}

		// Check if form is: helm_repo/chart
		chartUrl, err := url.Parse(chartName)
		if err != nil {
			log.Fatalf("invalid chart name format: %s", err)
		}

		// Does the chart exist locally
		chartExistsLocally := true
		_, errDir := os.Stat(chartName)
		if os.IsNotExist(errDir) {
			chartExistsLocally = false
		}

		p := strings.SplitN(chartUrl.Path, "/", 2)
		if len(p) > 1 && chartExistsLocally == false && p[0] != "." {
			common.DownloadUntarChart(chartName)
			var l = common.ReadCharts(deploymentFlag)
			var d = common.ReadChartDefinition(p[1] + innerChartFile)
			common.AppendExtraCharts(&l, &d)
			common.WriteChartDefinition(d, p[1]+innerChartFile)
		} else {
			var l = common.ReadCharts(deploymentFlag)
			var d = common.ReadChartDefinition(chartName + innerChartFile)
			common.AppendExtraCharts(&l, &d)
			common.WriteChartDefinition(d, chartName+innerChartFile)
		}

	},
}

func init() {
	ciChartCmd.AddCommand(editChartCmd)
	// Local flags
	editChartCmd.Flags().String("subchart-list-file", "", "Location of custom chart YAML file")
	editChartCmd.Flags().String("chart-name", "", "Chart to edit")
}
