package common

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

// credentialsFileName is the on-disk name of the CLI credentials file. It is
// kept separate from config.yaml and stored with 0600 permissions.
const credentialsFileName = "credentials"

// Credentials holds the silta hub-issued CLI token and related metadata.
type Credentials struct {
	HubURL    string `yaml:"hub_url"`
	Token     string `yaml:"token"`
	ExpiresAt string `yaml:"expires_at"`
	Username  string `yaml:"username"`
}

// credentialsPath returns the absolute path to the credentials file.
func credentialsPath() string {
	return filepath.Join(ConfigDir(), credentialsFileName)
}

// LoadCredentials reads the stored CLI credentials. It returns an error if no
// credentials have been saved yet.
func LoadCredentials() (*Credentials, error) {
	data, err := os.ReadFile(credentialsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in: run 'silta login' first")
		}
		return nil, err
	}
	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}
	return &creds, nil
}

// SaveCredentials writes the CLI credentials to disk with 0600 permissions.
func SaveCredentials(creds *Credentials) error {
	data, err := yaml.Marshal(creds)
	if err != nil {
		return err
	}
	return os.WriteFile(credentialsPath(), data, 0600)
}

// DeleteCredentials removes the stored CLI credentials, if present.
func DeleteCredentials() error {
	err := os.Remove(credentialsPath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Expired reports whether the credentials have an expiry in the past.
func (c *Credentials) Expired() bool {
	if c.ExpiresAt == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, c.ExpiresAt)
	if err != nil {
		return false
	}
	return time.Now().After(t)
}
