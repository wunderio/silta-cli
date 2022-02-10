package cmd_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestBuild(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	out, err := exec.Command("bash", "-c", "make build").CombinedOutput()
	fmt.Printf("Output: %s\n", out)
	if err != nil {
		fmt.Printf("Could not make binary for %s: %v", cliBinaryName, err)
		os.Exit(1)
	}

	// Change dir back to previous
	os.Chdir(wd)
}
