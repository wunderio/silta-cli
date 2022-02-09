package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Generate menu documentation markdown",
	Run: func(cmd *cobra.Command, args []string) {

		// Generate menu documentation markdown
		rootCmd.DisableAutoGenTag = true
		err := doc.GenMarkdownTree(rootCmd, "docs")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("CLI menu tree documentation generated in docs/ folder")
	},
}

func init() {
	rootCmd.AddCommand(docCmd)
}
