package cmd

import (
	"fmt"
	"log"
	dbg "runtime/debug"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Silta CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", common.Version)
		info, err := dbg.ReadBuildInfo()
		if err == false {
			log.Println("Cant get module info")
		}
		for _, dep := range info.Deps {
			log.Println("%+v", dep)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
