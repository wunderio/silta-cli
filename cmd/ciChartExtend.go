package cmd

import (
	"log"
	"net/url"
	"os"
	"os/exec"
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
		chartVersion, err1 := cmd.Flags().GetString("chart-version")
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

		c := common.ChartNameVersion{Name: chartName}
		if err1 == nil {
			c.Version = chartVersion
		} else {
			c.Version = ""
		}

		p := strings.Split(chartUrl.Path, "/")
		if debug == true {
			log.Println("p[0] ", p[0])
			log.Println("p[1] ", p[1])
			log.Println("deploymentFlag ", deploymentFlag)
			log.Println("c.Name ", c.Name)
		}

		var helmErr error
		var helmCmd []byte
		_, errDir = os.Stat(common.ExtendedFolder)
		if os.IsNotExist(errDir) {
			os.Mkdir(common.ExtendedFolder, 0744)
		}

		if len(p) == 2 && chartExistsLocally == false && p[0] != "." {
			//p[0] - name of the repo
			//p[1] - name of the chart itself
			if debug {
				log.Print(common.ExtendedFolder + "/" + p[1] + innerChartFile)
				log.Print("Chart doesnt exist locally")
			}
			common.DownloadUntarChart(&c, true)
			var l = common.ReadCharts(deploymentFlag)
			var d = common.ReadChartDefinition(common.ExtendedFolder + "/" + p[1] + innerChartFile)
			common.AppendExtraCharts(&l, &d)
			common.WriteChartDefinition(d, common.ExtendedFolder+"/"+p[1]+innerChartFile)
			for _, v := range l.Charts {
				common.AppendToChartSchemaFile(common.ExtendedFolder+"/"+p[1]+"/values.schema.json", v.Name)
			}
			helmCmdString := "helm dep update " + common.ExtendedFolder + "/" + p[1]
			helmCmd, helmErr = exec.Command("bash", "-c", helmCmdString).CombinedOutput()
		} else {

			if chartExistsLocally {
				if debug == true {
					log.Println("chartName ", chartName)
					log.Println("common.ExtendedFolder+ / p[len(p)-1] ", common.ExtendedFolder+"/"+p[len(p)-1])
				}
				err := os.Rename(chartName, common.ExtendedFolder+"/"+p[len(p)-1])
				if err != nil {
					log.Println("Cant move chart directory")
				}
			}
			var l = common.ReadCharts(deploymentFlag)
			var d = common.ReadChartDefinition(common.ExtendedFolder + "/" + p[len(p)-1] + innerChartFile)
			common.AppendExtraCharts(&l, &d)
			common.WriteChartDefinition(d, common.ExtendedFolder+"/"+p[len(p)-1]+innerChartFile)
			for _, v := range l.Charts {
				common.AppendToChartSchemaFile(common.ExtendedFolder+"/"+p[len(p)-1]+"/values.schema.json", v.Name)
			}
			helmCmdString := "helm dep update " + common.ExtendedFolder + "/" + p[len(p)-1]
			helmCmd, helmErr = exec.Command("bash", "-c", helmCmdString).CombinedOutput()
		}

		if helmErr != nil {
			log.SetFlags(0) //remove timestamp
			log.Fatal(string(helmCmd[:]))
		}

	},
}

func init() {
	ciChartCmd.AddCommand(editChartCmd)
	// Local flags
	editChartCmd.Flags().String("subchart-list-file", "", "Location of custom chart YAML file")
	editChartCmd.Flags().String("chart-name", "", "Source chart")
	editChartCmd.Flags().String("chart-version", "", "Chart version")

	editChartCmd.MarkFlagRequired("subchart-list-file")
	editChartCmd.MarkFlagRequired("chart-name")
}
