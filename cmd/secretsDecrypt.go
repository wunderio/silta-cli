package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"encoding/base64"

	"github.com/Luzifer/go-openssl/v4"
	"github.com/spf13/cobra"
)

var secretsDecryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt encrypted files",
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

		// Allow failing with exit code 0 when no files defined.
		if len(files) == 0 {
			fmt.Println("No input files supplied")
			return
		}

		// Split on comma.
		fileList := strings.Split(files, ",")

		// Decrypt files
		for i := range fileList {
			file := fileList[i]
			fmt.Printf("Decrypting %s\n", file)

			// Read encrypted file
			encryptedMsg, err := os.ReadFile(file)
			if err != nil {
				log.Fatal("Error: ", err)
			}

			// Verify file state
			if !strings.HasPrefix(string(encryptedMsg), "Salted") {
				log.Fatal("File does not appear to have been encrypted, salt header missing")
			}

			// Encode to base64 because library requires it
			encryptedMsg64 := base64.StdEncoding.EncodeToString(encryptedMsg)

			// Decrypt file content
			// 	openssl enc -d -aes-256-cbc -pbkdf2 -in "$FILE" -out "$tmp" -pass env:SECRET_KEY_ENV
			o := openssl.New()
			decryptedMessage, err := o.DecryptBytes(secretKey, []byte(encryptedMsg64), openssl.PBKDF2SHA256)
			if err != nil {
				log.Fatal("Decryption error: ", err)
			}

			if len(outputFile) > 0 {
				file = outputFile
				fmt.Printf("Saving decrypted file to %s\n", file)
			}

			// Write back the decrypted file
			f, err := os.Create(file)
			if err != nil {
				log.Fatal("Error writing file: ", err)
			}
			err = f.Truncate(0)
			_, err = f.Seek(0, 0)
			_, err = f.Write(decryptedMessage)
			if err != nil {
				log.Fatal("Error writing file: ", err)
			}

			fmt.Println("Success")

			f.Close()
		}
	},
}

func init() {
	secretsCmd.AddCommand(secretsDecryptCmd)

	secretsDecryptCmd.Flags().String("file", "", "Encrypted file location. Can have multiple, comma separated paths (i.e. 'silta/secrets.enc,silta/secrets2.enc')")
	secretsDecryptCmd.Flags().String("output-file", "", "Output file location (optional, rewrites original when undefined, don't use with multiple input files)")
	secretsDecryptCmd.Flags().String("secret-key", "", "Secret key (falls back to SECRET_KEY environment variable. Also see: --secret-key-env)")
	secretsDecryptCmd.Flags().String("secret-key-env", "", "Environment variable holding symmetrical decryption key.")

	secretsDecryptCmd.MarkFlagRequired("files")
}
