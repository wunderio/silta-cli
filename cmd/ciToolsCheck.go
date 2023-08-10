package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

var ciToolsCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for required tools",
	Long:  `Displays status of libraries and external binaries required by the tool.`,
	Run: func(cmd *cobra.Command, args []string) {

		bins := make(map[string]string)
		bins["helm"] = "version"
		bins["kubectl"] = "version"

		for bin, cmd := range bins {
			_, err := exec.LookPath(bin)
			if err != nil {
				log.Println(bin + " not found in $PATH")
			} else {
				log.Println(bin + " is installed")
				if cmd != "" {
					command := fmt.Sprintf("%s %s", bin, cmd)
					pipedExec(command, debug)
				}
			}
		}
	},
}

func init() {
	ciToolsCmd.AddCommand(ciToolsCheckCmd)

}
