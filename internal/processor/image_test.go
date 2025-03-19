package processor

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/nerten/albumpicker/internal/config"
)

const testDataDir = "../../test_data"

func TestProcessImageWithLibrary(t *testing.T) {
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "image_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test paths
	srcFile := filepath.Join(testDataDir, "album.png")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		OutputCoverName: "cover.jpg",
		CoverHeight:     240,
	}

	err = ProcessImageWithLibrary(srcFile, testDataDir, destDir, cfg)
	if err != nil {
		t.Errorf("ProcessImageWithLibrary() error = %v", err)
	}

	// Verify output file
	destFile := filepath.Join(destDir, cfg.OutputCoverName)
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// Check image dimensions
	outFile, err := os.Open(destFile)
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()

	outImg, _, err := image.Decode(outFile)
	if err != nil {
		t.Fatal(err)
	}

	if outImg.Bounds().Dy() != cfg.CoverHeight {
		t.Errorf("Output image height = %d, want %d", outImg.Bounds().Dy(), cfg.CoverHeight)
	}
}

func TestFallbackCopyCover(t *testing.T) {
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "cover_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test paths
	srcFile := filepath.Join(testDataDir, "album.png")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	err = FallbackCopyCover(srcFile, testDataDir, destDir, "cover.jpg")
	if err != nil {
		t.Errorf("FallbackCopyCover() error = %v", err)
	}

	// Verify output file
	destFile := filepath.Join(destDir, "cover.jpg")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// Verify the image can be decoded
	outFile, err := os.Open(destFile)
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()

	_, _, err = image.Decode(outFile)
	if err != nil {
		t.Errorf("Output file is not a valid image: %v", err)
	}
}
