package cmd

import (
	"os"
	"path/filepath"
	"testing"
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

	// create a dummy config file in tmpDir
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("cover_filenames:\n  - cover.jpg\noutput_cover_filename: cover.jpg\ncover_height: 240\n"), 0o644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(*testing.T)
	}{
		{
			name:    "basic pick command",
			args:    []string{"-c", configFile, "-s", sourceDir, "-d", destDir, "-n", "1"},
			wantErr: false,
			check: func(t *testing.T) {
				destAlbum := filepath.Join(destDir, "test-album")
				if _, err := os.Stat(destAlbum); os.IsNotExist(err) {
					t.Errorf("Album directory was not created in destination")
				}

				destFlac := filepath.Join(destAlbum, "test.flac")
				if _, err := os.Stat(destFlac); os.IsNotExist(err) {
					t.Errorf("FLAC file was not copied to destination")
				}

				if _, err := os.Stat(destTestFile); os.IsNotExist(err) {
					t.Errorf("Existing file was unexpectedly removed")
				}
			},
		},
		{
			name:    "pick with wipe flag",
			args:    []string{"-c", configFile, "-s", sourceDir, "-d", destDir, "-n", "1", "--wipe"},
			wantErr: false,
			check: func(t *testing.T) {
				destAlbum := filepath.Join(destDir, "test-album")
				if _, err := os.Stat(destAlbum); os.IsNotExist(err) {
					t.Errorf("Album directory was not created in destination")
				}

				if _, err := os.Stat(destTestFile); !os.IsNotExist(err) {
					t.Errorf("Existing file was not removed despite wipe flag")
				}
			},
		},
		{
			name:    "pick with embed-cover flag",
			args:    []string{"-c", configFile, "-s", sourceDir, "-d", destDir, "-n", "1", "--embed-cover"},
			wantErr: false,
			check: func(t *testing.T) {
				destAlbum := filepath.Join(destDir, "test-album")
				if _, err := os.Stat(destAlbum); os.IsNotExist(err) {
					t.Errorf("Album directory was not created in destination")
				}
			},
		},
		{
			name:    "invalid source directory",
			args:    []string{"-c", configFile, "-s", "/nonexistent/path", "-d", destDir},
			wantErr: true,
		},
		{
			name:    "missing source and destination",
			args:    []string{"-c", configFile},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// recreate test file for each test if it was deleted
			if _, err := os.Stat(destTestFile); os.IsNotExist(err) {
				if err := os.WriteFile(destTestFile, []byte("existing file"), 0o644); err != nil {
					t.Fatalf("Failed to recreate test file: %v", err)
				}
			}

			err := runPickCommand(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runPickCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t)
			}
		})
	}
}
