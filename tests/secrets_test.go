package cmd_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSecretsEncryptDecryptCmd(t *testing.T) {

	// Go to main directory
	wd, _ := os.Getwd()
	os.Chdir("..")

	secretMessage := "test:yes"

	// Create secret
	command := fmt.Sprintf("echo -n '%s' > tests/test-secret", secretMessage)
	exec.Command("bash", "-c", command).Run()

	// Set different filename for encrypted file
	command = "secrets encrypt --file tests/test-secret --secret-key test --output-file tests/test-secret-2"
	environment := []string{}
	testString := "Encrypting tests/test-secret\nSaving encrypted file to tests/test-secret-2\nSuccess\n"
	CliExecTest(t, command, environment, testString, true)

	// Verify file
	out, _ := exec.Command("bash", "-c", "cat tests/test-secret-2").CombinedOutput()
	if !strings.HasPrefix(string(out), "Salted") {
		t.Error("File not encrypted")
	}

	// Encrypt and replace file
	command = "secrets encrypt --file tests/test-secret --secret-key test"
	environment = []string{}
	testString = "Encrypting tests/test-secret\nSuccess\n"
	CliExecTest(t, command, environment, testString, true)

	// Verify file
	out, _ = exec.Command("bash", "-c", "cat tests/test-secret").CombinedOutput()
	if !strings.HasPrefix(string(out), "Salted") {
		t.Error("File not encrypted")
	}

	// Test double-encription prevention
	command = "secrets encrypt --file tests/test-secret --secret-key test"
	environment = []string{}
	testString = "File seems to be been encrypted already, skipping\n"
	CliExecTest(t, command, environment, testString, false)

	// Test decryption using wrong key
	command = "secrets decrypt --file tests/test-secret --secret-key wrongkey"
	environment = []string{}
	testString = "Decryption error: invalid padding"
	CliExecTest(t, command, environment, testString, false)

	// Test decryption using correct key
	command = "secrets decrypt --file tests/test-secret --secret-key test"
	environment = []string{}
	testString = "Decrypting tests/test-secret\nSuccess\n"
	CliExecTest(t, command, environment, testString, true)

	// Verify file
	out, _ = exec.Command("bash", "-c", "cat tests/test-secret").CombinedOutput()
	if string(out) != secretMessage {
		t.Error("Decrypted file incorrect")
	}

	// Set different filename for decrypted file
	command = "secrets decrypt --file tests/test-secret-2 --secret-key test --output-file tests/test-secret"
	environment = []string{}
	testString = "Decrypting tests/test-secret-2\nSaving decrypted file to tests/test-secret\nSuccess\n"
	CliExecTest(t, command, environment, testString, true)

	// Verify file
	out, _ = exec.Command("bash", "-c", "cat tests/test-secret").CombinedOutput()
	if string(out) != secretMessage {
		t.Error("Decrypted file incorrect")
	}

	// Change dir back to previous
	os.Chdir(wd)
}
