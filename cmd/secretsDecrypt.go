package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// secretsDecryptCmd represents the secretsDecrypt command
var secretsDecryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt encrypted files",
	Run: func(cmd *cobra.Command, args []string) {
		files, _ := cmd.Flags().GetString("files")
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

		// Split on comma.
		fileList := strings.Split(files, ",")

		// Decrypt files
		for i := range fileList {
			file := fileList[i]
			fmt.Printf("Decrypting %s\n", file)

			// TODO: Convert to golang crypto functions
			command := fmt.Sprintf(`
				export tmp=$(mktemp)
				export SECRET_KEY_ENV='%s'
				export FILE='%s'
				openssl enc -d -aes-256-cbc -pbkdf2 -in "$FILE" -out "$tmp" -pass env:SECRET_KEY_ENV
				# Check encryption status
				if [[ $? -eq 0 ]]; then
					echo "Success"
					mv -v "$tmp" "$FILE" > /dev/null
					chmod a+r "$FILE"
				else
					echo "Error decrypting secret"
					exit 1
				fi
				`, secretKey, file,
			)

			out, err := exec.Command("bash", "-c", command).CombinedOutput()
			fmt.Printf("%s", out)
			if err != nil {
				log.Fatal("Error (file checksum): ", err)
			}
		}
	},
}

func init() {
	secretsCmd.AddCommand(secretsDecryptCmd)

	secretsDecryptCmd.Flags().String("files", "", "Encrypted file location. Can have multiple, comma separated paths (i.e. 'silta/secrets.enc,silta/secrets2.enc')")
	secretsDecryptCmd.Flags().String("secret-key", "", "Secret key (falls back to SECRET_KEY environment variable. Also see: --secret-key-env)")
	secretsDecryptCmd.Flags().String("secret-key-env", "", "Environment variable holding symmetrical decryption key.")

	secretsDecryptCmd.MarkFlagRequired("files")
}
