package processor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

//const testDataDir = "../../test_data"

func TestProcessFLACWithLibrary(t *testing.T) {
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "flac_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test paths
	srcFile := filepath.Join(testDataDir, "01 - test.flac")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	err = ProcessFLACWithLibrary(srcFile, testDataDir, destDir)
	if err != nil {
		t.Errorf("ProcessFLACWithLibrary() error = %v", err)
	}

	// Verify output file exists
	destFile := filepath.Join(destDir, "01 - test.flac")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// Verify file size (should be smaller or equal due to removed PICTURE blocks)
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	destInfo, err := os.Stat(destFile)
	if err != nil {
		t.Fatal(err)
	}

	if destInfo.Size() > srcInfo.Size() {
		t.Error("Processed file is larger than source file")
	}
}

func TestSimpleCopyFLACFile(t *testing.T) {
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "flac_copy_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test paths
	srcFile := filepath.Join(testDataDir, "01 - test.flac")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	err = SimpleCopyFLACFile(srcFile, testDataDir, destDir)
	if err != nil {
		t.Errorf("SimpleCopyFLACFile() error = %v", err)
	}

	// Verify output file exists
	destFile := filepath.Join(destDir, "01 - test.flac")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// Verify file contents are identical
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
