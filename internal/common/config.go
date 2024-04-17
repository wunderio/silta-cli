package common

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

func ConfigStore() viper.Viper {

	// Configuration subpath based on OS
	siltaConfigDir := ".config/silta"
	if os := os.Getenv("OS"); os == "Windows_NT" {
		siltaConfigDir = ".silta"
	}
	if os := os.Getenv("OS"); os == "darwin" {
		siltaConfigDir = "Library/Application Support/silta"
	}

	// Get the user's home directory
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error getting user home directory, %s", err)
	}
	configDir := filepath.Join(usr.HomeDir, siltaConfigDir)

	// Create the configuration directory if it doesn't exist
	_, err = os.Stat(configDir)
	if !os.IsExist(err) {
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			log.Fatalf("Error creating config directory, %s", err)
		}
	}

	// Set the configuration file
	viper.SetConfigFile(filepath.Join(configDir, "config.yaml"))
	viper.AddConfigPath(configDir)

	// Read the configuration file if it exists
	viper.ReadInConfig()

	return *viper.GetViper()
}
