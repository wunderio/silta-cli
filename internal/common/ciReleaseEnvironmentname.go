package common

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func SiltaEnvironmentName(branchname string, releaseSuffix string) string {
	// Lowercase without special characters.
	siltaEnvironmentName := strings.ToLower(branchname)

	suffix := ""
	if len(releaseSuffix+siltaEnvironmentName) > 39 {
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
		if len(siltaEnvironmentName) >= rn_max_length {
			sha256_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(siltaEnvironmentName)))
			siltaEnvironmentName = fmt.Sprintf("%s-%s", siltaEnvironmentName[0:rn_cut_length], sha256_hash[0:4])
		}
	}

	if len(releaseSuffix) > 0 {
		if len(suffix) > 0 {
			// Using suffix variable for release name
			siltaEnvironmentName = fmt.Sprintf("%s-%s", siltaEnvironmentName, suffix)
		} else {
			// Using parameter for release name
			siltaEnvironmentName = fmt.Sprintf("%s-%s", siltaEnvironmentName, releaseSuffix)
		}
	}

	return siltaEnvironmentName
}
