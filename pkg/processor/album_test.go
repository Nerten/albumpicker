package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nerten/albumpicker/pkg/config"
)

func TestSelectRandomAlbums(t *testing.T) {
	tests := []struct {
		name     string
		albums   []string
		n        int
		wantLen  int
		wantSame bool
	}{
		{
			name:     "select less than available",
			albums:   []string{"album1", "album2", "album3", "album4"},
			n:        2,
			wantLen:  2,
			wantSame: false,
		},
		{
			name:     "select more than available",
			albums:   []string{"album1", "album2"},
			n:        3,
			wantLen:  2,
			wantSame: true,
		},
		{
			name:     "select exact amount",
			albums:   []string{"album1", "album2", "album3"},
			n:        3,
			wantLen:  3,
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectRandomAlbums(tt.albums, tt.n)
			if len(got) != tt.wantLen {
				t.Errorf("selectRandomAlbums() got len = %v, want %v", len(got), tt.wantLen)
			}
			if tt.wantSame {
				// check if all elements are present
				for _, album := range tt.albums {
					found := false
					for _, selected := range got {
						if album == selected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("selectRandomAlbums() missing album %v", album)
					}
				}
			}
		})
	}
}

func TestProcessAlbum(t *testing.T) {
	// create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "albumpicker_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	albumDir := filepath.Join(srcDir, "testalbum")

	// create test directory structure
	for _, dir := range []string{srcDir, destDir, albumDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// create test FLAC file
	testFlac := filepath.Join(albumDir, "test.flac")
	if err := os.WriteFile(testFlac, []byte("test flac data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// create test cover file
	testCover := filepath.Join(albumDir, "cover.jpg")
	if err := os.WriteFile(testCover, []byte("test jpg data"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Source:          srcDir,
		Destination:     destDir,
		CoverFilenames:  []string{"cover.jpg"},
		OutputCoverName: "cover.jpg",
		CoverHeight:     240,
	}

	err = ProcessAlbum(albumDir, cfg)
	if err != nil {
		t.Errorf("processAlbum() error = %v", err)
	}

	// check if destination files were created
	destAlbum := filepath.Join(destDir, "testalbum")
	destFlac := filepath.Join(destAlbum, "test.flac")
	destCover := filepath.Join(destAlbum, "cover.jpg")

	for _, file := range []string{destFlac, destCover} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}
}

func TestFindAllAlbums(t *testing.T) {
	// create temporary test directory structure
	tmpDir, err := os.MkdirTemp("", "albumpicker_find_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// create test directory structure
	testDirs := map[string][]string{
		"album1":                          {"track1.flac", "track2.flac"},
		"album2":                          {"track1.flac", "cover.jpg"},
		"empty":                           {},
		"noflac":                          {"track1.mp3", "track2.mp3"},
		filepath.Join("nested", "album3"): {"track1.flac"},
	}

	for dir, files := range testDirs {
		dirPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			filePath := filepath.Join(dirPath, file)
			if err := os.WriteFile(filePath, []byte("test data"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	// test finding albums
	albums, err := FindAllAlbums(tmpDir)
	if err != nil {
		t.Errorf("findAllAlbums() error = %v", err)
	}

	// expected albums (directories containing .flac files)
	expected := []string{
		filepath.Join(tmpDir, "album1"),
		filepath.Join(tmpDir, "album2"),
		filepath.Join(tmpDir, "nested", "album3"),
	}

	if len(albums) != len(expected) {
		t.Errorf("findAllAlbums() found %d albums, want %d", len(albums), len(expected))
	}

	// check if all expected albums are found
	foundMap := make(map[string]bool)
	for _, album := range albums {
		foundMap[album] = true
	}

	for _, exp := range expected {
		if !foundMap[exp] {
			t.Errorf("findAllAlbums() missing expected album: %s", exp)
		}
	}
}

func TestProcessAlbums(t *testing.T) {
	// create temporary test directory structure
	tmpDir, err := os.MkdirTemp("", "albumpicker_process_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")

	// create test albums
	albums := []string{"album1", "album2"}
	for _, album := range albums {
		albumDir := filepath.Join(srcDir, album)
		if err := os.MkdirAll(albumDir, 0o755); err != nil {
			t.Fatal(err)
		}

		// create test files
		testFiles := map[string][]byte{
			"track1.flac": []byte("test flac data 1"),
			"track2.flac": []byte("test flac data 2"),
			"cover.jpg":   []byte("test jpg data"),
		}

		for name, data := range testFiles {
			if err := os.WriteFile(filepath.Join(albumDir, name), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	// create test config
	cfg := &config.Config{
		Source:          srcDir,
		Destination:     destDir,
		CoverFilenames:  []string{"cover.jpg"},
		OutputCoverName: "cover.jpg",
		CoverHeight:     240,
	}

	// test processing albums
	albumPaths := []string{
		filepath.Join(srcDir, "album1"),
		filepath.Join(srcDir, "album2"),
	}

	err = ProcessAlbums(albumPaths, cfg)
	if err != nil {
		t.Errorf("processAlbums() error = %v", err)
	}

	// verify destination structure
	for _, album := range albums {
		destAlbum := filepath.Join(destDir, album)
		expectedFiles := []string{
			filepath.Join(destAlbum, "track1.flac"),
			filepath.Join(destAlbum, "track2.flac"),
			filepath.Join(destAlbum, "cover.jpg"),
		}

		for _, file := range expectedFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Expected file %s was not created", file)
			}
		}
	}
}
