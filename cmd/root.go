package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	useEnv bool
	debug  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "silta",
	Short: "Silta CLI",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent flags
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Print variables, do not execute external commands, rather print them")
	rootCmd.PersistentFlags().BoolVar(&useEnv, "use-env", true, "Use environment variables for value assignment")
}

func bufferedExec(command string, debug bool) string {
	out := ""

	if debug == true {
		out = fmt.Sprintf("Command (not executed): %s\n", command)
	} else {
		out, err := exec.Command("bash", "-c", command).CombinedOutput()
		if err != nil {
			out = []byte(fmt.Sprintf("Output: %s\n", out))
			out = []byte(fmt.Sprintf("Error: %x\n", err))
		}
	}

	return out
}

func pipedExec(command string, debug bool) {
	if debug == true {
		fmt.Printf("Command (not executed): %s\n", command)
	} else {

		// Flush exec output buffers since this might take a while
		cmd := exec.Command("bash", "-c", command)

		// create a pipe for the output of the script
		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal("Error (pipe): ", err)
			return
		}
		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				fmt.Printf("  %s\n", scanner.Text())
			}
		}()
		err = cmd.Start()
		if err != nil {
			log.Fatal("Error (Start): ", err)
		}
		err = cmd.Wait()
		if err != nil {
			log.Fatal("Error (Wait): ", err)
		}
	}
}
