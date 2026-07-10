package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the TOML configuration file structure
type Config struct {
	Targets  []Target `toml:"target"`
	Notation Notation `toml:"notation"`
}

// Target represents a single target directory configuration
type Target struct {
	Dir               string `toml:"dir"`
	DiscordWebhookURL string `toml:"discord_webhook_url"`
	Recursive         bool   `toml:"recursive"`
}

// Notation は楽譜画像の生成に使う外部コマンドの設定。
// コマンドが見つからない場合、楽譜画像の生成はスキップされる（変換や
// Discord投稿は従来どおり行う）。
type Notation struct {
	// VerovioPath はMusicXMLをSVGにレンダリングするVerovioのパス。
	// 省略時は PATH から verovio を探す。
	VerovioPath string `toml:"verovio_path"`
	// SVG2PNGPath はSVGをPNGへ変換するrsvg-convert互換コマンドのパス。
	// 省略時は PATH から rsvg-convert を探す。見つからない場合はSVGのまま扱う。
	SVG2PNGPath string `toml:"svg2png_path"`
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

	if config.Notation.VerovioPath == "" {
		config.Notation.VerovioPath = "verovio"
	}
	if config.Notation.SVG2PNGPath == "" {
		config.Notation.SVG2PNGPath = "rsvg-convert"
	}

	return &config, nil
}
