package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	config "github.com/nerten/albumpicker/pkg/config"
	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	// create temporary directory for test config
	tmpDir, err := os.MkdirTemp("", "albumpicker_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// set up test environment
	t.Setenv("HOME", tmpDir)

	// test InitConfig
	viper.Reset() // reset viper for each test
	initConfig()

	// check if config file was created
	configPath := filepath.Join(tmpDir, ".config", "albumpicker", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// verify default values
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
	// create temporary directory
	tmpDir, err := os.MkdirTemp("", "albumpicker_content_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// set up test environment
	t.Setenv("HOME", tmpDir)

	// initialize config
	viper.Reset() // reset viper for each test
	initConfig()

	// test reading config values
	cfg := &config.Config{
		Source:          viper.GetString("source"),
		Destination:     viper.GetString("destination"),
		AlbumsCount:     viper.GetInt("albums_count"),
		CoverFilenames:  viper.GetStringSlice("cover_filenames"),
		OutputCoverName: viper.GetString("output_cover_filename"),
		CoverHeight:     viper.GetInt("cover_height"),
	}

	// verify config struct values
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
	// save original os.Args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version command",
			args:    []string{"albumpicker", "--version"},
			wantErr: false,
		},
		{
			name:    "help command",
			args:    []string{"albumpicker", "--help"},
			wantErr: false,
		},
		{
			name:    "invalid flag",
			args:    []string{"albumpicker", "--invalid-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up test args
			os.Args = tt.args

			// capture stdout/stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			// execute with test version
			exitCode := 0
			oldOsExit := osExit
			osExit = func(code int) { exitCode = code }
			defer func() { osExit = oldOsExit }()

			Execute("test-version")
			w.Close()

			// check exit code
			if tt.wantErr && exitCode == 0 {
				t.Errorf("Execute() expected error but got none")
			} else if !tt.wantErr && exitCode != 0 {
				t.Errorf("Execute() unexpected error with exit code %d", exitCode)
			}

			// read output
			var output []byte
			output, _ = io.ReadAll(r)
			if tt.name == "version command" && !contains(string(output), "albumpicker version test-version") {
				t.Errorf("Version output doesn't contain expected version: %s", string(output))
			}
		})
	}
}

func TestCustomConfigFile(t *testing.T) {
	// create temporary directory for test config
	tmpDir, err := os.MkdirTemp("", "albumpicker_custom_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// create custom config file
	customConfigPath := filepath.Join(tmpDir, "custom_config.yaml")
	customConfig := []byte(`
source: "/custom/source"
destination: "/custom/destination"
albums_count: 5
cover_height: 300
`)
	if err := os.WriteFile(customConfigPath, customConfig, 0o644); err != nil {
		t.Fatalf("Failed to write custom config file: %v", err)
	}

	// set config file path
	cfgFile = customConfigPath

	// initialize config
	viper.Reset() // reset viper for each test
	initConfig()

	// verify custom values were loaded
	if src := viper.GetString("source"); src != "/custom/source" {
		t.Errorf("Custom source not loaded, got: %s", src)
	}
	if dest := viper.GetString("destination"); dest != "/custom/destination" {
		t.Errorf("Custom destination not loaded, got: %s", dest)
	}
	if count := viper.GetInt("albums_count"); count != 5 {
		t.Errorf("Custom albums_count not loaded, got: %d", count)
	}
	if height := viper.GetInt("cover_height"); height != 300 {
		t.Errorf("Custom cover_height not loaded, got: %d", height)
	}

	// reset for other tests
	cfgFile = ""
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
