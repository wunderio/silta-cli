package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get configuration",
	Long: `This will print configuration information. If no arguments are provided, the entire configuration file will be printed.
If a single argument is provided, the value of the configuration key will be printed. Supports nested keys using dot notation.`,
	Run: func(cmd *cobra.Command, args []string) {
		configStore := common.ConfigStore()

		if len(args) < 1 {
			cfg := configStore.ConfigFileUsed()
			// Read raw file content and print it
			content, err := os.ReadFile(cfg)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%s", content)

		}

		if len(args) == 1 {
			// Print single configuration item
			key := args[0]
			value := configStore.Get(key)
			if value != nil {
				fmt.Printf("%s", value)
			}
		}
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
}
