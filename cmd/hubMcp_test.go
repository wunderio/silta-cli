package cmd

import "testing"

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"https://hub.example.com/mcp":   "https://hub.example.com/mcp",
		"silta":                         "silta",
		"https://h/mcp?a=1&b=2":         "'https://h/mcp?a=1&b=2'",
		"https://h/$(touch /tmp/pwned)": "'https://h/$(touch /tmp/pwned)'",
		"it's":                          `'it'\''s'`,
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateMCPURL(t *testing.T) {
	valid := []string{"https://hub.example.com/mcp", "http://localhost:4000/mcp"}
	invalid := []string{"", "ftp://hub/mcp", "javascript:alert(1)", "/mcp", "https:///mcp",
		"https://hub/mcp $(id)", "https://hub/mcp\n", "https://hub/\tmcp"}
	for _, u := range valid {
		if err := validateMCPURL(u); err != nil {
			t.Errorf("validateMCPURL(%q) = %v, want nil", u, err)
		}
	}
	for _, u := range invalid {
		if err := validateMCPURL(u); err == nil {
			t.Errorf("validateMCPURL(%q) = nil, want error", u)
		}
	}
}
