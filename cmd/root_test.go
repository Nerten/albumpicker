package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nerten/albumpicker/pkg/config"
)

func TestDefaultConfigCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "albumpicker_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("HOME", tmpDir)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".config", "albumpicker", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	if cfg.AlbumsCount != 10 {
		t.Errorf("expected albums_count=10, got %d", cfg.AlbumsCount)
	}
	if cfg.CoverHeight != 240 {
		t.Errorf("expected cover_height=240, got %d", cfg.CoverHeight)
	}
	if cfg.OutputCoverName != "cover.jpg" {
		t.Errorf("expected output_cover_filename=cover.jpg, got %s", cfg.OutputCoverName)
	}
	expectedCovers := []string{"album.jpg", "album.png", "cover.jpg", "cover.png"}
	if len(cfg.CoverFilenames) != len(expectedCovers) {
		t.Fatalf("expected %d cover filenames, got %d", len(expectedCovers), len(cfg.CoverFilenames))
	}
}

func TestConfigFileContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "albumpicker_content_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("HOME", tmpDir)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

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

func TestExecute(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{"albumpicker"},
			wantErr: false,
		},
		{
			name:    "version command",
			args:    []string{"albumpicker", "--version"},
			wantErr: false,
		},
		{
			name:    "version short flag",
			args:    []string{"albumpicker", "-v"},
			wantErr: false,
		},
		{
			name:    "version subcommand",
			args:    []string{"albumpicker", "version"},
			wantErr: false,
		},
		{
			name:    "help command",
			args:    []string{"albumpicker", "--help"},
			wantErr: false,
		},
		{
			name:    "help subcommand",
			args:    []string{"albumpicker", "help"},
			wantErr: false,
		},
		{
			name:    "help pick subcommand",
			args:    []string{"albumpicker", "help", "pick"},
			wantErr: false,
		},
		{
			name:    "help copy subcommand",
			args:    []string{"albumpicker", "help", "copy"},
			wantErr: false,
		},
		{
			name:    "invalid flag",
			args:    []string{"albumpicker", "--invalid-flag"},
			wantErr: true,
		},
		{
			name:    "unknown subcommand",
			args:    []string{"albumpicker", "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			exitCode := 0
			oldOsExit := osExit
			osExit = func(code int) { exitCode = code }
			defer func() { osExit = oldOsExit }()

			Execute("test-version")
			w.Close()

			if tt.wantErr && exitCode == 0 {
				t.Errorf("Execute() expected error but got none")
			} else if !tt.wantErr && exitCode != 0 {
				t.Errorf("Execute() unexpected error with exit code %d", exitCode)
			}

			output, _ := io.ReadAll(r)
			if strings.Contains(tt.name, "version") && !contains(string(output), "albumpicker version test-version") {
				t.Errorf("Version output doesn't contain expected version: %s", string(output))
			}
		})
	}
}

func TestCustomConfigFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "albumpicker_custom_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	customConfigPath := filepath.Join(tmpDir, "custom_config.yaml")
	customConfig := []byte(`
source: "/custom/source"
destination: "/custom/destination"
albums_count: 5
cover_height: 300
`)
	if err := os.WriteFile(customConfigPath, customConfig, 0o600); err != nil {
		t.Fatalf("Failed to write custom config file: %v", err)
	}

	cfg, err := config.LoadConfig(customConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg.Source != "/custom/source" {
		t.Errorf("Custom source not loaded, got: %s", cfg.Source)
	}
	if cfg.Destination != "/custom/destination" {
		t.Errorf("Custom destination not loaded, got: %s", cfg.Destination)
	}
	if cfg.AlbumsCount != 5 {
		t.Errorf("Custom albums_count not loaded, got: %d", cfg.AlbumsCount)
	}
	if cfg.CoverHeight != 300 {
		t.Errorf("Custom cover_height not loaded, got: %d", cfg.CoverHeight)
	}
}

func TestSubcommandUsages(t *testing.T) {
	printPickUsage()
	printCopyUsage()

	if err := runPickCommand([]string{"--help"}); err != nil {
		t.Errorf("runPickCommand(--help) error = %v", err)
	}
	if err := runCopyCommand([]string{"--help"}); err != nil {
		t.Errorf("runCopyCommand(--help) error = %v", err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
