package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantTargets int
		wantErr     bool
	}{
		{
			name: "single target",
			content: `[[target]]
dir = "/path/to/dir1"
discord_webhook_url = "https://discord.com/api/webhooks/XXX/YYY"
`,
			wantTargets: 1,
			wantErr:     false,
		},
		{
			name: "multiple targets",
			content: `[[target]]
dir = "/path/to/dir1"
discord_webhook_url = "https://discord.com/api/webhooks/XXX/YYY"

[[target]]
dir = "/path/to/dir2"
recursive = true
`,
			wantTargets: 2,
			wantErr:     false,
		},
		{
			name: "target with all fields",
			content: `[[target]]
dir = "/path/to/dir1"
discord_webhook_url = "https://discord.com/api/webhooks/XXX/YYY"
recursive = true
`,
			wantTargets: 1,
			wantErr:     false,
		},
		{
			name:        "empty config",
			content:     "",
			wantTargets: 0,
			wantErr:     true,
		},
		{
			name: "no targets",
			content: `
# Just a comment
`,
			wantTargets: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")
			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Load config
			config, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(config.Targets) != tt.wantTargets {
					t.Errorf("LoadConfig() got %d targets, want %d", len(config.Targets), tt.wantTargets)
				}
			}
		})
	}
}

func TestLoadConfig_ValidateFields(t *testing.T) {
	content := `[[target]]
dir = "/path/to/dir1"
discord_webhook_url = "https://discord.com/api/webhooks/XXX/YYY"
recursive = true

[[target]]
dir = "/path/to/dir2"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if len(config.Targets) != 2 {
		t.Fatalf("Expected 2 targets, got %d", len(config.Targets))
	}

	// Validate first target
	if config.Targets[0].Dir != "/path/to/dir1" {
		t.Errorf("Target[0].Dir = %q, want %q", config.Targets[0].Dir, "/path/to/dir1")
	}
	if config.Targets[0].DiscordWebhookURL != "https://discord.com/api/webhooks/XXX/YYY" {
		t.Errorf("Target[0].DiscordWebhookURL = %q, want %q", config.Targets[0].DiscordWebhookURL, "https://discord.com/api/webhooks/XXX/YYY")
	}
	if !config.Targets[0].Recursive {
		t.Errorf("Target[0].Recursive = %v, want true", config.Targets[0].Recursive)
	}

	// Validate second target
	if config.Targets[1].Dir != "/path/to/dir2" {
		t.Errorf("Target[1].Dir = %q, want %q", config.Targets[1].Dir, "/path/to/dir2")
	}
	if config.Targets[1].DiscordWebhookURL != "" {
		t.Errorf("Target[1].DiscordWebhookURL = %q, want empty string", config.Targets[1].DiscordWebhookURL)
	}
	if config.Targets[1].Recursive {
		t.Errorf("Target[1].Recursive = %v, want false", config.Targets[1].Recursive)
	}
}
