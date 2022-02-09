package cmd

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var ciReleaseNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Return release name",
	Long:  `Generate safe release name based on branchname and release-suffix.`,
	Run: func(cmd *cobra.Command, args []string) {

		branchname, _ := cmd.Flags().GetString("branchname")
		releaseSuffix, _ := cmd.Flags().GetString("release-suffix")

		// Environment value fallback
		if useEnv == true {
			if branchname == "" {
				branchname = os.Getenv("CIRCLE_BRANCH")
			}
		}

		if branchname == "" {
			log.Fatal("Repository branchname not provided")
		}

		// Make sure release name is lowercase without special characters.
		branchnameLower := strings.ToLower(branchname)
		reg, _ := regexp.Compile("[^[:alnum:]]")
		releaseName := reg.ReplaceAllString(branchnameLower, "-")

		suffix := ""
		// TODO: Yes, this part of logic is a bit broken in orb. Keeping it in sync for now, will fix later.
		if len(releaseSuffix) > 0 && len(releaseSuffix+releaseName) > 39 {
			suffix = releaseSuffix

			if len(suffix) > 12 {
				sha256_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(suffix)))
				suffix = fmt.Sprintf("%s-%s", suffix[0:7], sha256_hash[0:4])
			}

			// Maximum length of a release name + release suffix. -1 is for separating '-' char before suffix
			rn_max_length := 40 - len(suffix) - 1

			// Length of a shortened rn_max_length to allow for an appended hash
			rn_cut_length := rn_max_length - 5

			// # If name is too long, truncate it and append a hash
			if len(releaseName) >= rn_max_length {
				sha256_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(branchnameLower)))
				releaseName = fmt.Sprintf("%s-%s", releaseName[0:rn_cut_length], sha256_hash[0:4])
			}
		}

		if len(releaseSuffix) > 0 {
			if len(suffix) > 0 {
				// Using suffix variable for release name
				releaseName = fmt.Sprintf("%s-%s", releaseName, suffix)
			} else {
				// Using parameter for release name
				releaseName = fmt.Sprintf("%s-%s", releaseName, releaseSuffix)
			}
		}

		fmt.Printf("%s", releaseName)
	},
}

func init() {
	ciReleaseCmd.AddCommand(ciReleaseNameCmd)

	ciReleaseNameCmd.Flags().String("branchname", "", "Repository branchname that will be used for release name")
	ciReleaseNameCmd.Flags().String("release-suffix", "", "Release name suffix")
}
