package cmd

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Luzifer/go-openssl/v4"
	"github.com/spf13/cobra"
)

var secretsEncryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt secret files",
	Run: func(cmd *cobra.Command, args []string) {
		files, _ := cmd.Flags().GetString("file")
		outputFile, _ := cmd.Flags().GetString("output-file")
		secretKey, _ := cmd.Flags().GetString("secret-key")
		secretKeyEnv, _ := cmd.Flags().GetString("secret-key-env")

		// Use environment variables as fallback
		if useEnv == true {
			if len(secretKey) == 0 {
				secretKey = os.Getenv("SECRET_KEY")
			}
			if len(secretKeyEnv) > 0 {
				secretKey = os.Getenv(secretKeyEnv)
			}
		}

		// Replace comma with whitespace and iterate all whitespace separated values.
		// This also means there can't be commas nor whitespaces in filenames.
		space := regexp.MustCompile(`,\s?|\s+`)
		files = space.ReplaceAllString(files, " ")

		// Allow failing with exit code 0 when no files defined.
		if len(files) == 0 {
			fmt.Println("No input files supplied")
			return
		}

		// Split on whitespace.
		fileList := strings.Split(files, " ")

		// Fail if no secret key is not provided
		if len(secretKey) == 0 {
			fmt.Println("No secret key provided")
			return
		}

		// Encrypt files
		for i := range fileList {
			file := fileList[i]
			fmt.Printf("Encrypting %s\n", file)

			// Read file
			decryptedMsg, _ := os.ReadFile(file)

			// Verify file state
			if strings.HasPrefix(string(decryptedMsg), "Salted") {
				log.Fatal("File seems to be been encrypted already, skipping")
			}

			// Encrypt message
			o := openssl.New()
			// openssl aes-256-cbc -pbkdf2 -in $2.dec -out $2 -pass pass:$ssl_pass
			// openssl aes-256-cbc -pbkdf2 -in $2.dec -out $2 -pass env:SECRET_KEY_ENV
			encryptedMsg64, err := o.EncryptBytes(secretKey, decryptedMsg, openssl.PBKDF2SHA256)
			if err != nil {
				fmt.Printf("An error occurred: %s\n", err)
			}

			// Decode base64 output, we don't use it for encrypted files
			encryptedMsg, _ := base64.StdEncoding.DecodeString(string(encryptedMsg64))

			if len(outputFile) > 0 {
				file = outputFile
				fmt.Printf("Saving encrypted file to %s\n", file)
			}

			// Write back the encrypted file
			f, err := os.Create(file)
			if err != nil {
				log.Fatal("Error creating file: ", err)
			}
			err = f.Truncate(0)
			_, err = f.Seek(0, 0)
			_, err = f.Write(encryptedMsg)
			if err != nil {
				log.Fatal("Error writing to file: ", err)
			}

			fmt.Println("Success")

			f.Close()
		}
	},
}

func init() {
	secretsCmd.AddCommand(secretsEncryptCmd)

	secretsEncryptCmd.Flags().String("file", "", "Decrypted file location. Can have multiple, comma separated paths (i.e. 'silta/secrets.enc,silta/secrets2.enc')")
	secretsEncryptCmd.Flags().String("output-file", "", "Output file location (optional, rewrites original when undefined, don't use with multiple input files)")
	secretsEncryptCmd.Flags().String("secret-key", "", "Secret key (falls back to SECRET_KEY environment variable. Also see: --secret-key-env)")
	secretsEncryptCmd.Flags().String("secret-key-env", "", "Environment variable holding symmetrical decryption key.")

	secretsEncryptCmd.MarkFlagRequired("file")
}
