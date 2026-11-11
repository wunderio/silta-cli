package common

import "strings"

// EscapeSingleQuoted makes a value safe to embed inside a single-quoted bash
// string literal. A single quote cannot be escaped inside single quotes, so
// each one is rewritten as: close quote, escaped quote, reopen quote ('\”).
// Every other character is literal inside single quotes, so nothing else
// needs handling. The value itself is preserved exactly.
func EscapeSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", `'\''`)
}
