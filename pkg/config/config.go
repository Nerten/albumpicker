package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is a set of parameters for albumpicker
type Config struct {
	Source          string   `yaml:"source"`
	Destination     string   `yaml:"destination"`
	AlbumsCount     int      `yaml:"albums_count"`
	CoverFilenames  []string `yaml:"cover_filenames"`
	OutputCoverName string   `yaml:"output_cover_filename"`
	CoverHeight     int      `yaml:"cover_height"`
	EmbedCover      bool     `yaml:"embed_cover"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Source:          "",
		Destination:     "",
		AlbumsCount:     10,
		CoverFilenames:  []string{"album.jpg", "album.png", "cover.jpg", "cover.png"},
		OutputCoverName: "cover.jpg",
		CoverHeight:     240,
		EmbedCover:      false,
	}
}

// GetDefaultConfigPath returns the default config file path: ~/.config/albumpicker/config.yaml
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}
	return filepath.Join(home, ".config", "albumpicker", "config.yaml"), nil
}

// LoadConfig loads the configuration from file (or default path)
func LoadConfig(customPath ...string) (*Config, error) {
	cfg := DefaultConfig()

	var configPath string
	isCustom := len(customPath) > 0 && customPath[0] != ""
	if isCustom {
		configPath = customPath[0]
	} else {
		var err error
		configPath, err = GetDefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("error parsing config file %s: %w", configPath, err)
		}
		fmt.Printf("Config file: %s\n", configPath)
	} else if !isCustom && errors.Is(err, os.ErrNotExist) {
		// save default config if not exists
		if err := SaveConfig(configPath, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not create default config: %s\n", err)
		}
	} else if isCustom {
		return nil, fmt.Errorf("error reading config file %s: %w", configPath, err)
	}

	return cfg, nil
}

// LoadAndValidateConfig loads and validates the configuration
func LoadAndValidateConfig(customPath ...string) (*Config, error) {
	cfg, err := LoadConfig(customPath...)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Source == "" {
		return fmt.Errorf("source directory not specified")
	}
	if c.Destination == "" {
		return fmt.Errorf("destination directory not specified")
	}

	// check if source directory exists
	if _, err := os.Stat(c.Source); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", c.Source)
	}

	// create destination directory if it doesn't exist
	if _, err := os.Stat(c.Destination); os.IsNotExist(err) {
		if err := os.MkdirAll(c.Destination, 0o755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	return nil
}
