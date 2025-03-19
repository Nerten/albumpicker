package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestRunCopyCommand(t *testing.T) {
	// Get project root directory
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Test data paths
	testDataDir := filepath.Join(projectRoot, "test_data")
	testFlac := filepath.Join(testDataDir, "01 - test.flac")
	testCover := filepath.Join(testDataDir, "album.png")
	if _, err := os.Stat(testCover); os.IsNotExist(err) {
		testCover = filepath.Join(testDataDir, "cover.jpg")
	}

	// Verify test files exist
	if _, err := os.Stat(testFlac); os.IsNotExist(err) {
		t.Fatalf("Test FLAC file not found at %s", testFlac)
	}
	if _, err := os.Stat(testCover); os.IsNotExist(err) {
		t.Fatalf("Test cover file not found at %s", testCover)
	}

	// Create temporary directories
	tmpDir, err := os.MkdirTemp("", "albumpicker-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test directory structure
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	albumDir := filepath.Join(sourceDir, "artist", "test - album")

	if err := os.MkdirAll(albumDir, 0755); err != nil {
		t.Fatalf("Failed to create album directory: %v", err)
	}

	// Copy test files to album directory
	if err := copyFile(testFlac, filepath.Join(albumDir, "01 - test.flac")); err != nil {
		t.Fatalf("Failed to copy test FLAC file: %v", err)
	}
	if err := copyFile(testCover, filepath.Join(albumDir, filepath.Base(testCover))); err != nil {
		t.Fatalf("Failed to copy test cover file: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		setup   func()
		wantErr bool
	}{
		{
			name: "copy album with absolute path",
			args: []string{albumDir},
			setup: func() {
				viper.Set("source", sourceDir)
				viper.Set("destination", destDir)
				viper.Set("output_cover_filename", "album.jpg")
				viper.Set("cover_filenames", []string{"album.jpg", "album.png", "cover.jpg", "cover.png"})
				viper.Set("cover_height", 240)
			},
			wantErr: false,
		},
		{
			name: "invalid source directory",
			args: []string{"/nonexistent/path"},
			setup: func() {
				viper.Set("source", "/nonexistent")
				viper.Set("destination", destDir)
			},
			wantErr: true,
		},
		{
			name: "missing configuration",
			args: []string{albumDir},
			setup: func() {
				viper.Reset()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and setup test configuration
			viper.Reset()
			if tt.setup != nil {
				tt.setup()
			}

			// Create a new command for testing
			cmd := &cobra.Command{}
			err := runCopyCommand(cmd, tt.args)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runCopyCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check if album was copied
				destAlbum := filepath.Join(destDir, "artist", "test - album")
				if _, err := os.Stat(destAlbum); os.IsNotExist(err) {
					t.Errorf("Album directory was not created in destination")
				}

				// Check if FLAC file was copied
				destFlac := filepath.Join(destAlbum, "01 - test.flac")
				if _, err := os.Stat(destFlac); os.IsNotExist(err) {
					t.Errorf("FLAC file was not copied to destination")
				}

				// Check if cover file was copied
				destCover := filepath.Join(destAlbum, "album.jpg")
				if _, err := os.Stat(destCover); os.IsNotExist(err) {
					t.Errorf("Cover file was not copied to destination")
				}
			}
		})
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
