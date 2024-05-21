package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/wunderio/silta-cli/internal/common"
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

	// Load configuration
	configStore := common.ConfigStore()

	// Set proxy if it's set in the configuration
	proxy := configStore.Get("proxy")
	if proxy != nil {
		proxy := proxy.(string)

		for _, key := range []string{"HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy"} {
			// Set proxy variable unless it's already set in the environment
			if os.Getenv(key) == "" {
				os.Setenv(key, proxy)
			}
		}
	}
}

func bufferedExec(command string, debug bool) {
	if debug {
		fmt.Sprintf("Command (not executed): %s\n", command)
	} else {
		out, err := exec.Command("bash", "-c", command).CombinedOutput()
		fmt.Printf("%s\n", out)
		if err != nil {
			log.Fatal("Error: ", err)
		}
	}
}

func pipedExec(command string, stdOutPrefix string, stdErrPrefix string, debug bool) {
	if debug {
		fmt.Printf("Command (not executed): %s\n", command)
	} else {

		// Flush exec output buffers since this might take a while
		cmd := exec.Command("bash", "-c", command)

		// create a pipe for the output of the script
		cmdOutReader, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal("Error (stdout pipe): ", err)
			return
		}
		cmdErrReader, err := cmd.StderrPipe()
		if err != nil {
			log.Fatal("Error (stderr pipe): ", err)
			return
		}
		outScanner := bufio.NewScanner(cmdOutReader)
		errScanner := bufio.NewScanner(cmdErrReader)
		go func() {
			for errScanner.Scan() {
				fmt.Printf("%s%s\n", stdErrPrefix, errScanner.Text())
			}
		}()
		go func() {
			for outScanner.Scan() {
				fmt.Printf("%s%s\n", stdOutPrefix, outScanner.Text())
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
