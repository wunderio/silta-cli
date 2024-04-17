package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
)

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration",
	Long: `This will set configuration information. The first argument is the key and the second argument is the value.
If the key already exists, the value will be overwritten. Supports nested keys using dot notation.
Usage: silta config set <key> <value>
Example: silta config set mykey
Example: silta config set mykey myvalue
Example: silta config set mykey.subkey myvalue
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		key := args[0]
		value := strings.Join(args[1:], " ")

		configStore := common.ConfigStore()
		configStore.Set(key, value)

		// Create a new configuration file if it doesn't exist
		err := configStore.WriteConfig()
		if err != nil {
			log.Fatalf("Error writing config file, %s", err)
		}

		fmt.Println("Configuration set")
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
