package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const CONFIG_PATH = "./bmux.config.json"

var c *Config

// Load reads the configuration file from disk, parses the JSON content,
// and loads it into the package-level Config variable `c`.
//
// If the config file does not exist, it will attempt to create one with default values.
//
// Returns an error if reading or unmarshalling the file fails.
//
// Example usage:
//
//	err := config.Load()
//	if err != nil {
//	    // handle error
//	}
func Load(override *Config) error {
	_, err := os.Stat(CONFIG_PATH)
	if os.IsNotExist(err) {
		if err := Create(override); err != nil {
			return fmt.Errorf("Load: failed creating config: %w", err)
		}
	}

	file, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		return fmt.Errorf("Load: failed reading json: %w", err)
	}

	err = json.Unmarshal(file, &c)
	if err != nil {
		return fmt.Errorf("Load: failed unmarshalling json: %w", err)
	}
	return nil
}

// Create writes a configuration file with either default values or
// overrides provided by the user.
//
// The file is written in JSON format with indentation for readability.
//
// Returns an error if marshaling or writing to the file fails.
//
// Example usage:
//
//	err := config.Create(&config.Config{Port: "8080"})
func Create(override *Config) error {
	defaultConfig := Config{
		Port:            30000,
		Address:         "0.0.0.0",
		Experimental:    false,
		LogLevel:        "info",
		MaxConnections:  1024,
		ShutdownTimeout: 10,
		EnableMulticore: true,
	}

	if override != nil {
		defaultConfig = *override
	}

	file, err := json.MarshalIndent(&defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("Create: failed marshalling config: %w", err)
	}

	err = os.WriteFile(CONFIG_PATH, file, 0644)
	if err != nil {
		return fmt.Errorf("Create: failed writing config: %w", err)
	}

	return nil
}

// New initializes the package configuration by loading the config file,
// if it hasn't already been loaded.
//
// Returns an error if loading the configuration fails.
//
// Example usage:
//
//	err := config.New()
//	if err != nil {
//	    // handle error
//	}
func New(override *Config) error {
	if c == nil {
		err := Load(override)
		if err != nil {
			return fmt.Errorf("New: failed loading json: %w", err)
		}
	}
	return nil
}
