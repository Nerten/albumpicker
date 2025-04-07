package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestRunPickCommand(t *testing.T) {
	// get project root directory
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// test data paths
	testDataDir := filepath.Join(projectRoot, "test_data")
	testFlac := filepath.Join(testDataDir, "01 - test.flac")
	testCover := filepath.Join(testDataDir, "cover.jpg")

	// verify test files exist
	if _, err := os.Stat(testFlac); os.IsNotExist(err) {
		t.Fatalf("Test FLAC file not found at %s", testFlac)
	}
	if _, err := os.Stat(testCover); os.IsNotExist(err) {
		t.Fatalf("Test cover file not found at %s", testCover)
	}

	// create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "albumpicker-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	albumDir := filepath.Join(sourceDir, "test-album")

	// create test directory structure
	dirs := []string{sourceDir, destDir, albumDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// copy test files to album directory
	if err := copyFile(testFlac, filepath.Join(albumDir, "01 - test.flac")); err != nil {
		t.Fatalf("Failed to copy test FLAC file: %v", err)
	}
	if err := copyFile(testCover, filepath.Join(albumDir, "cover.jpg")); err != nil {
		t.Fatalf("Failed to copy test cover file: %v", err)
	}

	// create a test FLAC file
	flacPath := filepath.Join(albumDir, "test.flac")
	if err := os.WriteFile(flacPath, []byte("test flac data"), 0o644); err != nil {
		t.Fatalf("Failed to create test FLAC file: %v", err)
	}

	// create a test file in destination to test wipe
	destTestFile := filepath.Join(destDir, "existing-file.txt")
	if err := os.WriteFile(destTestFile, []byte("existing file"), 0o644); err != nil {
		t.Fatalf("Failed to create test file in destination: %v", err)
	}

	tests := []struct {
		name    string
		setup   func(*cobra.Command)
		wantErr bool
		check   func(*testing.T)
	}{
		{
			name: "basic pick command",
			setup: func(cmd *cobra.Command) {
				viper.Set("source", sourceDir)
				viper.Set("destination", destDir)
				viper.Set("albums_count", 1)
				viper.Set("cover_filenames", []string{"cover.jpg"})
				viper.Set("output_cover_filename", "cover.jpg")
				viper.Set("cover_height", 240)
			},
			wantErr: false,
			check: func(t *testing.T) {
				// check if album was copied
				destAlbum := filepath.Join(destDir, "test-album")
				if _, err := os.Stat(destAlbum); os.IsNotExist(err) {
					t.Errorf("Album directory was not created in destination")
				}

				// check if FLAC file was copied
				destFlac := filepath.Join(destAlbum, "test.flac")
				if _, err := os.Stat(destFlac); os.IsNotExist(err) {
					t.Errorf("FLAC file was not copied to destination")
				}

				// check if existing file still exists (not wiped)
				if _, err := os.Stat(destTestFile); os.IsNotExist(err) {
					t.Errorf("Existing file was unexpectedly removed")
				}
			},
		},
		{
			name: "pick with wipe flag",
			setup: func(cmd *cobra.Command) {
				viper.Set("source", sourceDir)
				viper.Set("destination", destDir)
				viper.Set("albums_count", 1)
				cmd.Flags().Set("wipe", "true")
			},
			wantErr: false,
			check: func(t *testing.T) {
				// check if album was copied
				destAlbum := filepath.Join(destDir, "test-album")
				if _, err := os.Stat(destAlbum); os.IsNotExist(err) {
					t.Errorf("Album directory was not created in destination")
				}

				// check if existing file was wiped
				if _, err := os.Stat(destTestFile); !os.IsNotExist(err) {
					t.Errorf("Existing file was not removed despite wipe flag")
				}
			},
		},
		{
			name: "invalid source directory",
			setup: func(cmd *cobra.Command) {
				viper.Set("source", "/nonexistent/path")
				viper.Set("destination", destDir)
			},
			wantErr: true,
		},
		{
			name: "invalid destination directory",
			setup: func(cmd *cobra.Command) {
				viper.Set("source", sourceDir)
				viper.Set("destination", "/nonexistent/path")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset viper and create a new command for testing
			viper.Reset()
			cmd := &cobra.Command{}
			cmd.Flags().Bool("wipe", false, "wipe flag for testing")

			// setup test configuration
			if tt.setup != nil {
				tt.setup(cmd)
			}

			// recreate test file for each test if it was deleted
			if _, err := os.Stat(destTestFile); os.IsNotExist(err) {
				if err := os.WriteFile(destTestFile, []byte("existing file"), 0o644); err != nil {
					t.Fatalf("Failed to recreate test file: %v", err)
				}
			}

			// run the command
			err := runPickCommand(cmd, nil)

			// check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runPickCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// run additional checks
			if !tt.wantErr && tt.check != nil {
				tt.check(t)
			}
		})
	}
}
