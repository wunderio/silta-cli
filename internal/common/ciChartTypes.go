package common

type chartList struct {
	Charts []dependency `yaml:"charts"`
}

type chartDefinition struct {
	APIVersion   string       `yaml:"apiVersion"`
	Dependencies []dependency `yaml:"dependencies"`
	Name         string       `yaml:"name"`
	Version      string       `yaml:"version"`
}

type dependency struct {
	Condition  string `yaml:"condition,omitempty"`
	Name       string `yaml:"name"`
	Repository string `yaml:"repository"`
	Version    string `yaml:"version"`
}
