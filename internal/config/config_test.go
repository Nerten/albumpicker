package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	// Create temporary directory for test config
	tmpDir, err := os.MkdirTemp("", "albumpicker_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up test environment
	t.Setenv("HOME", tmpDir)

	// Test InitConfig
	err = InitConfig("")
	if err != nil {
		t.Errorf("InitConfig() error = %v", err)
	}

	// Check if config file was created
	configPath := filepath.Join(tmpDir, ".config", "albumpicker", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify default values
	expectedDefaults := map[string]any{
		"source":                "",
		"destination":           "",
		"albums_count":          10,
		"cover_filenames":       []string{"album.jpg", "album.png", "cover.jpg", "cover.png"},
		"output_cover_filename": "cover.jpg",
		"cover_height":          240,
	}

	for key, expected := range expectedDefaults {
		actual := viper.Get(key)
		if actual == nil {
			t.Errorf("Default value for %s was not set", key)
			continue
		}

		switch expected := expected.(type) {
		case []string:
			actualSlice, ok := actual.([]any)
			if !ok {
				t.Errorf("Could not convert %s to []interface{}", key)
				continue
			}
			if len(actualSlice) != len(expected) {
				t.Errorf("Default value for %s has wrong length: got %v, want %v", key, actualSlice, expected)
			}
			for i, v := range expected {
				if actualSlice[i].(string) != v {
					t.Errorf("Default value for %s[%d] is wrong: got %v, want %v", key, i, actualSlice[i], v)
				}
			}
		default:
			if actual != expected {
				t.Errorf("Default value for %s is wrong: got %v, want %v", key, actual, expected)
			}
		}
	}
}

func TestConfigFileContent(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "albumpicker_content_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up test environment
	t.Setenv("HOME", tmpDir)

	// Initialize config
	err = InitConfig("")
	if err != nil {
		t.Fatal(err)
	}

	// Test reading config values
	cfg := &Config{
		Source:          viper.GetString("source"),
		Destination:     viper.GetString("destination"),
		AlbumsCount:     viper.GetInt("albums_count"),
		CoverFilenames:  viper.GetStringSlice("cover_filenames"),
		OutputCoverName: viper.GetString("output_cover_filename"),
		CoverHeight:     viper.GetInt("cover_height"),
	}

	// Verify config struct values
	if cfg.AlbumsCount != 10 {
		t.Errorf("Wrong albums_count: got %v, want %v", cfg.AlbumsCount, 10)
	}

	expectedCoverFiles := []string{"album.jpg", "album.png", "cover.jpg", "cover.png"}
	if len(cfg.CoverFilenames) != len(expectedCoverFiles) {
		t.Errorf("Wrong cover_filenames length: got %v, want %v", len(cfg.CoverFilenames), len(expectedCoverFiles))
	}

	for i, filename := range expectedCoverFiles {
		if cfg.CoverFilenames[i] != filename {
			t.Errorf("Wrong cover filename at index %d: got %v, want %v", i, cfg.CoverFilenames[i], filename)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "albumpicker_load_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source and destination directories
	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Set up viper values for testing
	viper.Set("source", srcDir)
	viper.Set("destination", destDir)
	viper.Set("albums_count", 5)
	viper.Set("cover_filenames", []string{"test.jpg"})
	viper.Set("output_cover_filename", "output.jpg")
	viper.Set("cover_height", 480)
	viper.Set("concurrency", 2)

	// Test LoadConfig
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	// Verify config values
	tests := []struct {
		name   string
		got    interface{}
		want   interface{}
		errMsg string
	}{
		{"Source", cfg.Source, srcDir, "wrong source directory"},
		{"Destination", cfg.Destination, destDir, "wrong destination directory"},
		{"AlbumsCount", cfg.AlbumsCount, 5, "wrong albums count"},
		{"OutputCoverName", cfg.OutputCoverName, "output.jpg", "wrong output cover name"},
		{"CoverHeight", cfg.CoverHeight, 480, "wrong cover height"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.errMsg, tt.got, tt.want)
			}
		})
	}

	// Test cover filenames separately due to slice comparison
	if len(cfg.CoverFilenames) != 1 || cfg.CoverFilenames[0] != "test.jpg" {
		t.Errorf("wrong cover filenames: got %v, want %v", cfg.CoverFilenames, []string{"test.jpg"})
	}
}

func TestLoadConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "missing source",
			setup: func() {
				viper.Set("source", "")
				viper.Set("destination", "testdest")
			},
			wantErr: true,
		},
		{
			name: "missing destination",
			setup: func() {
				viper.Set("source", "testsrc")
				viper.Set("destination", "")
			},
			wantErr: true,
		},
		{
			name: "non-existent source",
			setup: func() {
				viper.Set("source", "/non/existent/path")
				viper.Set("destination", "testdest")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
