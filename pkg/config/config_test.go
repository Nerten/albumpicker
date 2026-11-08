package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "albumpicker_load_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	configFilePath := filepath.Join(tmpDir, "config.yaml")
	yamlData := []byte(`
source: "` + srcDir + `"
destination: "` + destDir + `"
albums_count: 5
cover_filenames:
  - test.jpg
output_cover_filename: output.jpg
cover_height: 480
embed_cover: true
`)
	if err := os.WriteFile(configFilePath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configFilePath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	tests := []struct {
		name   string
		got    any
		want   any
		errMsg string
	}{
		{"Source", cfg.Source, srcDir, "wrong source directory"},
		{"Destination", cfg.Destination, destDir, "wrong destination directory"},
		{"AlbumsCount", cfg.AlbumsCount, 5, "wrong albums count"},
		{"OutputCoverName", cfg.OutputCoverName, "output.jpg", "wrong output cover name"},
		{"CoverHeight", cfg.CoverHeight, 480, "wrong cover height"},
		{"EmbedCover", cfg.EmbedCover, true, "wrong embed cover setting"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.errMsg, tt.got, tt.want)
			}
		})
	}

	if len(cfg.CoverFilenames) != 1 || cfg.CoverFilenames[0] != "test.jpg" {
		t.Errorf("wrong cover filenames: got %v, want %v", cfg.CoverFilenames, []string{"test.jpg"})
	}

	// Also test Validate
	if err := cfg.Validate(); err != nil {
		t.Errorf("cfg.Validate() unexpected error = %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.AlbumsCount != 10 {
		t.Errorf("expected default albums_count=10, got %d", cfg.AlbumsCount)
	}
	if cfg.CoverHeight != 240 {
		t.Errorf("expected default cover_height=240, got %d", cfg.CoverHeight)
	}
	if cfg.OutputCoverName != "cover.jpg" {
		t.Errorf("expected default output_cover_filename=cover.jpg, got %s", cfg.OutputCoverName)
	}
	expectedCovers := []string{"album.jpg", "album.png", "cover.jpg", "cover.png"}
	if len(cfg.CoverFilenames) != len(expectedCovers) {
		t.Fatalf("expected %d cover filenames, got %d", len(expectedCovers), len(cfg.CoverFilenames))
	}
}

func TestLoadConfigDefaultCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "albumpicker_home_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("HOME", tmpDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() default error = %v", err)
	}

	if cfg.AlbumsCount != 10 {
		t.Errorf("expected default count 10, got %d", cfg.AlbumsCount)
	}

	expectedPath := filepath.Join(tmpDir, ".config", "albumpicker", "config.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected default config file to be created at %s", expectedPath)
	}
}

func TestConfigValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "albumpicker_val_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Source:      srcDir,
				Destination: filepath.Join(tmpDir, "dest"),
			},
			wantErr: false,
		},
		{
			name: "missing source",
			cfg: Config{
				Source:      "",
				Destination: filepath.Join(tmpDir, "dest"),
			},
			wantErr: true,
		},
		{
			name: "missing destination",
			cfg: Config{
				Source:      srcDir,
				Destination: "",
			},
			wantErr: true,
		},
		{
			name: "non-existent source",
			cfg: Config{
				Source:      "/non/existent/path/here",
				Destination: filepath.Join(tmpDir, "dest"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadAndValidateConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "albumpicker_load_val_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	configFilePath := filepath.Join(tmpDir, "config.yaml")
	yamlData := []byte("source: " + srcDir + "\ndestination: " + destDir + "\n")
	if err := os.WriteFile(configFilePath, yamlData, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadAndValidateConfig(configFilePath)
	if err != nil {
		t.Fatalf("LoadAndValidateConfig() unexpected error: %v", err)
	}
	if cfg.Source != srcDir {
		t.Errorf("expected source %s, got %s", srcDir, cfg.Source)
	}

	// invalid file
	_, err = LoadAndValidateConfig(filepath.Join(tmpDir, "missing.yaml"))
	if err == nil {
		t.Error("expected error for missing custom config file")
	}
}

