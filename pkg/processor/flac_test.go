package processor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-flac/go-flac"
)

func TestProcessFLACWithLibrary(t *testing.T) {
	// create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "flac_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// setup test paths
	testDataDir := "../../test_data"
	srcFile := filepath.Join(testDataDir, "01 - test.flac")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	picData, err := processFLACWithLibrary(srcFile, testDataDir, destDir, nil)
	if err != nil {
		t.Errorf("ProcessFLACWithLibrary() error = %v", err)
	}

	// verify extracted picture
	if len(picData) == 0 {
		t.Error("Expected extracted picture from test FLAC, got empty")
	}

	// verify output file exists
	destFile := filepath.Join(destDir, "01 - test.flac")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// verify destination FLAC has NO picture blocks
	parsedDest, err := flac.ParseFile(destFile)
	if err != nil {
		t.Fatalf("Failed to parse destination FLAC: %v", err)
	}
	for _, block := range parsedDest.Meta {
		if block.Type == flac.Picture {
			t.Error("Destination FLAC file still contains PICTURE block")
		}
	}

	// verify file size (should be smaller due to removed PICTURE blocks)
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	destInfo, err := os.Stat(destFile)
	if err != nil {
		t.Fatal(err)
	}

	if destInfo.Size() >= srcInfo.Size() {
		t.Errorf("Processed file is not smaller than source file (dest=%d, src=%d)", destInfo.Size(), srcInfo.Size())
	}
}

func TestSimpleCopyFLACFile(t *testing.T) {
	// create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "flac_copy_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// setup test paths
	testDataDir := "../../test_data"
	srcFile := filepath.Join(testDataDir, "01 - test.flac")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err = simpleCopyFLACFile(srcFile, testDataDir, destDir)
	if err != nil {
		t.Errorf("SimpleCopyFLACFile() error = %v", err)
	}

	// verify output file exists
	destFile := filepath.Join(destDir, "01 - test.flac")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// verify file contents are identical
	srcData, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	destData, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(srcData, destData) {
		t.Error("Destination file content does not match source")
	}
}

func TestProcessFLACFile(t *testing.T) {
	// create temporary test directories
	tmpDir, err := os.MkdirTemp("", "albumpicker_flac_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// setup test paths
	testDataDir := "../../test_data"
	srcDir := filepath.Clean(testDataDir)
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// use real FLAC file from test_data
	testFlac := filepath.Join(srcDir, "01 - test.flac")
	if _, err := os.Stat(testFlac); os.IsNotExist(err) {
		t.Fatalf("Test FLAC file not found: %s", testFlac)
	}

	picData, err := ProcessFLACFile(testFlac, srcDir, destDir)
	if err != nil {
		t.Errorf("ProcessFLACFile() error = %v", err)
	}
	if len(picData) == 0 {
		t.Error("Expected picture data from ProcessFLACFile")
	}

	// check if destination file exists
	destFile := filepath.Join(destDir, "01 - test.flac")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination FLAC file was not created")
	}
}

func TestProcessFLACFile_EmbedCover(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "albumpicker_flac_embed_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testDataDir := "../../test_data"
	srcFile := filepath.Join(testDataDir, "01 - test.flac")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	dummyCover := []byte("fake jpeg cover art content")
	_, err = ProcessFLACFile(srcFile, testDataDir, destDir, dummyCover)
	if err != nil {
		t.Fatalf("ProcessFLACFile with embedCover error = %v", err)
	}

	destFile := filepath.Join(destDir, "01 - test.flac")
	parsedDest, err := flac.ParseFile(destFile)
	if err != nil {
		t.Fatalf("Failed to parse destination FLAC: %v", err)
	}

	var foundCover bool
	for _, block := range parsedDest.Meta {
		if block.Type != flac.Picture {
			continue
		}
		picType, mimeType, imgData, err := ExtractPictureFromMetaDataBlock(block.Data)
		if err != nil {
			t.Fatalf("Failed to extract embedded picture: %v", err)
		}
		if picType != 3 {
			t.Errorf("Expected picture type 3 (Front Cover), got %d", picType)
		}
		if mimeType != "image/jpeg" {
			t.Errorf("Expected mime type image/jpeg, got %s", mimeType)
		}
		if !bytes.Equal(imgData, dummyCover) {
			t.Errorf("Embedded data mismatch, expected %q, got %q", dummyCover, imgData)
		}
		foundCover = true
	}
	if !foundCover {
		t.Error("Expected embedded PICTURE block in destination FLAC, none found")
	}
}

func TestExtractPictureFromMetaDataBlockErrors(t *testing.T) {
	// too short
	_, _, _, err := ExtractPictureFromMetaDataBlock([]byte{1, 2, 3})
	if err == nil {
		t.Error("expected error for data too short")
	}

	// invalid mime length
	shortMime := make([]byte, 40)
	// set mime length to 1000
	shortMime[4] = 0
	shortMime[5] = 0
	shortMime[6] = 3
	shortMime[7] = 232 // 1000
	_, _, _, err = ExtractPictureFromMetaDataBlock(shortMime)
	if err == nil {
		t.Error("expected error for invalid mime length")
	}
}
