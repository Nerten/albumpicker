package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config is a set of parameters for albumpicker
type Config struct {
	Source          string
	Destination     string
	AlbumsCount     int
	CoverFilenames  []string
	OutputCoverName string
	CoverHeight     int
}

// LoadConfig loads and validates the configuration from viper
func LoadConfig() (*Config, error) {
	config := &Config{
		Source:          viper.GetString("source"),
		Destination:     viper.GetString("destination"),
		AlbumsCount:     viper.GetInt("albums_count"),
		CoverFilenames:  viper.GetStringSlice("cover_filenames"),
		OutputCoverName: viper.GetString("output_cover_filename"),
		CoverHeight:     viper.GetInt("cover_height"),
	}

	// validate config
	if config.Source == "" {
		return nil, fmt.Errorf("source directory not specified")
	}
	if config.Destination == "" {
		return nil, fmt.Errorf("destination directory not specified")
	}

	// check if source directory exists
	if _, err := os.Stat(config.Source); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist: %s", config.Source)
	}

	// create destination directory if it doesn't exist
	if _, err := os.Stat(config.Destination); os.IsNotExist(err) {
		if err := os.MkdirAll(config.Destination, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create destination directory: %s", err)
		}
	}

	return config, nil
}
