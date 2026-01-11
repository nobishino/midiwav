package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the TOML configuration file structure
type Config struct {
	Targets []Target `toml:"target"`
}

// Target represents a single target directory configuration
type Target struct {
	Dir               string `toml:"dir"`
	DiscordWebhookURL string `toml:"discord_webhook_url"`
	Recursive         bool   `toml:"recursive"`
}

// LoadConfig loads and parses the TOML configuration file
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config file path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(config.Targets) == 0 {
		return nil, fmt.Errorf("no targets defined in config file")
	}

	// Validate targets
	for i, target := range config.Targets {
		if target.Dir == "" {
			return nil, fmt.Errorf("target %d: dir is required", i)
		}
	}

	return &config, nil
}
