package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var configFile = common.ConfigStore()

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Silta configuration commands",
	Long: `Silta configuration commands, allows setting and getting configuration values. 
Configuration is persistent and is stored in file "` + configFile.ConfigFileUsed() + `".`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
