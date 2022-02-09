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

	make := exec.Command("make", "build")
	err := make.Run()
	if err != nil {
		fmt.Printf("could not make binary for %s: %v", cliBinaryName, err)
		os.Exit(1)
	}

	// Change dir back to previous
	os.Chdir(wd)
}
