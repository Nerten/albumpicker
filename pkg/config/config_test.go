package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig(t *testing.T) {
	// create temporary directory
	tmpDir, err := os.MkdirTemp("", "albumpicker_load_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// create source and destination directories
	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// set up viper values for testing
	viper.Set("source", srcDir)
	viper.Set("destination", destDir)
	viper.Set("albums_count", 5)
	viper.Set("cover_filenames", []string{"test.jpg"})
	viper.Set("output_cover_filename", "output.jpg")
	viper.Set("cover_height", 480)
	viper.Set("concurrency", 2)

	// test LoadConfig
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	// verify config values
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

	// test cover filenames separately due to slice comparison
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
