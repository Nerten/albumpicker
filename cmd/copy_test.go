package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCopyCommand(t *testing.T) {
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	testDataDir := filepath.Join(projectRoot, "test_data")
	testFlac := filepath.Join(testDataDir, "01 - test.flac")
	testCover := filepath.Join(testDataDir, "album.png")
	if _, err := os.Stat(testCover); os.IsNotExist(err) {
		testCover = filepath.Join(testDataDir, "cover.jpg")
	}

	if _, err := os.Stat(testFlac); os.IsNotExist(err) {
		t.Fatalf("Test FLAC file not found at %s", testFlac)
	}
	if _, err := os.Stat(testCover); os.IsNotExist(err) {
		t.Fatalf("Test cover file not found at %s", testCover)
	}

	tmpDir, err := os.MkdirTemp("", "albumpicker-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	albumDir := filepath.Join(sourceDir, "artist", "test - album")

	if err := os.MkdirAll(albumDir, 0o755); err != nil {
		t.Fatalf("Failed to create album directory: %v", err)
	}

	if err := copyFile(testFlac, filepath.Join(albumDir, "01 - test.flac")); err != nil {
		t.Fatalf("Failed to copy test FLAC file: %v", err)
	}
	if err := copyFile(testCover, filepath.Join(albumDir, filepath.Base(testCover))); err != nil {
		t.Fatalf("Failed to copy test cover file: %v", err)
	}

	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := []byte(`
cover_filenames:
  - album.jpg
  - album.png
  - cover.jpg
  - cover.png
output_cover_filename: album.jpg
cover_height: 240
`)
	if err := os.WriteFile(configFile, configContent, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "copy album with absolute path",
			args:    []string{"-c", configFile, "-s", sourceDir, "-d", destDir, albumDir},
			wantErr: false,
		},
		{
			name:    "copy album with embed cover flag",
			args:    []string{"-c", configFile, "-s", sourceDir, "-d", destDir, "--embed-cover", albumDir},
			wantErr: false,
		},
		{
			name:    "invalid source directory",
			args:    []string{"-c", configFile, "-s", "/nonexistent", "-d", destDir, albumDir},
			wantErr: true,
		},
		{
			name:    "missing album path argument",
			args:    []string{"-c", configFile, "-s", sourceDir, "-d", destDir},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCopyCommand(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runCopyCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				destAlbum := filepath.Join(destDir, "artist", "test - album")
				if _, err := os.Stat(destAlbum); os.IsNotExist(err) {
					t.Errorf("Album directory was not created in destination")
				}

				destFlac := filepath.Join(destAlbum, "01 - test.flac")
				if _, err := os.Stat(destFlac); os.IsNotExist(err) {
					t.Errorf("FLAC file was not copied to destination")
				}

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
	return os.WriteFile(dst, data, 0o644)
}
