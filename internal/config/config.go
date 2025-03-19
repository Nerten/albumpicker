package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Source          string
	Destination     string
	AlbumsCount     int
	CoverFilenames  []string
	OutputCoverName string
	CoverHeight     int
}

func InitConfig(cfgFile string) error {

	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %s\n", err)
			os.Exit(1)
		}
		// Search config in home directory
		configDir := filepath.Join(home, ".config", "albumpicker")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return err
			}
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		viper.SetDefault("source", "")
		viper.SetDefault("destination", "")
		viper.SetDefault("albums_count", 10)
		viper.SetDefault("cover_filenames", []string{"album.jpg", "album.png", "cover.jpg", "cover.png"})
		viper.SetDefault("output_cover_filename", "cover.jpg")
		viper.SetDefault("cover_height", 240)

		// save default config if not exists
		if err := viper.SafeWriteConfig(); err != nil {
			var configFileAlreadyExistsError viper.ConfigFileAlreadyExistsError
			if !errors.As(err, &configFileAlreadyExistsError) {
				fmt.Fprintf(os.Stderr, "Warning: Could not create default config: %s\n", err)
			}
		}
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}

	return nil
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

	// Validate config
	if config.Source == "" {
		return nil, fmt.Errorf("source directory not specified")
	}
	if config.Destination == "" {
		return nil, fmt.Errorf("destination directory not specified")
	}

	// Check if source directory exists
	if _, err := os.Stat(config.Source); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist: %s", config.Source)
	}

	// Create destination directory if it doesn't exist
	if _, err := os.Stat(config.Destination); os.IsNotExist(err) {
		if err := os.MkdirAll(config.Destination, 0755); err != nil {
			return nil, fmt.Errorf("failed to create destination directory: %s", err)
		}
	}

	return config, nil
}
