package cmd_test

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

var cliBinaryName = "./silta"

func CliExecTest(t *testing.T, command string, environment []string, testString string, equals bool) {

	// TMP: Do not build
	// cliBinaryName = "go run main.go"
	// environment = os.Environ()

	// fmt.Printf("%s", command)

	cmd := exec.Command("bash", "-c", cliBinaryName+" "+command)

	// PATH is missing from exec env, we'll merge in existing os.Environ() as a base
	mergedEnv := os.Environ()
	for index, _ := range environment {
		mergedEnv = append(mergedEnv, environment[index])
	}

	cmd.Env = mergedEnv

	var out, err bytes.Buffer
	cmd.Stderr = &err
	cmd.Stdout = &out
	cmd.Run()

	if equals == true {
		if out.String() == testString || err.String() == testString {
		} else {
			t.Logf("Error: %s", err.String())
			t.Errorf("Expected :\n '%s' \n Received: \n '%s'\n'%s'", testString, out.String(), err.String())
			d := diffOutput(t, testString, out.String())
			t.Errorf("Diff:\n %s", d)
		}

	} else {
		if strings.Contains(out.String(), testString) || strings.Contains(err.String(), testString) {
		} else {
			t.Logf("Error: %s", err.String())
			t.Errorf("Expected :\n '%s' \n Received: \n '%s'\n'%s'", testString, out.String(), err.String())
			d := diffOutput(t, testString, out.String())
			t.Errorf("Diff:\n %s", d)
		}
	}
}

func diffOutput(t *testing.T, expected string, received string) string {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		B:        difflib.SplitLines(received),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Received",
		ToDate:   "",
		Context:  1,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	return text
}
