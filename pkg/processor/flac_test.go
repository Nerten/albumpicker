package processor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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

	err = processFLACWithLibrary(srcFile, testDataDir, destDir)
	if err != nil {
		t.Errorf("ProcessFLACWithLibrary() error = %v", err)
	}

	// verify output file exists
	destFile := filepath.Join(destDir, "01 - test.flac")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// verify file size (should be smaller or equal due to removed PICTURE blocks)
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

	err = ProcessFLACFile(testFlac, srcDir, destDir)
	if err != nil {
		t.Errorf("processFLACFile() error = %v", err)
	}

	// check if destination file exists
	destFile := filepath.Join(destDir, "01 - test.flac")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination FLAC file was not created")
	}

	// verify file size (should be smaller or equal due to removed PICTURE blocks)
	srcInfo, err := os.Stat(testFlac)
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
