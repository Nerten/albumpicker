package processor

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/nerten/albumpicker/pkg/config"
)

func TestProcessCoverFile(t *testing.T) {
	// create temporary test directories
	tmpDir, err := os.MkdirTemp("", "albumpicker_cover_test")
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

	// verify test cover files exist
	coverFiles := []string{"album.png", "cover.jpg"}
	for _, cover := range coverFiles {
		coverPath := filepath.Join(srcDir, cover)
		if _, err := os.Stat(coverPath); os.IsNotExist(err) {
			t.Fatalf("Test cover file not found: %s", coverPath)
		}
	}

	cfg := &config.Config{
		CoverFilenames:  []string{"album.png", "cover.jpg"},
		OutputCoverName: "cover.jpg",
		CoverHeight:     240,
	}

	err = ProcessCoverFile(srcDir, destDir, cfg)
	if err != nil {
		t.Errorf("processCoverFile() error = %v", err)
	}

	// check if destination file exists
	destFile := filepath.Join(destDir, cfg.OutputCoverName)
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination cover file was not created")
	}

	// verify the output image dimensions
	destImg, err := os.Open(destFile)
	if err != nil {
		t.Fatal(err)
	}
	defer destImg.Close()

	img, _, err := image.Decode(destImg)
	if err != nil {
		t.Fatal(err)
	}

	if img.Bounds().Dy() != cfg.CoverHeight {
		t.Errorf("Output image height = %d, want %d", img.Bounds().Dy(), cfg.CoverHeight)
	}
}

func TestProcessImageWithLibrary(t *testing.T) {
	// create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "image_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// setup test paths
	testDataDir := "../../test_data"
	srcFile := filepath.Join(testDataDir, "album.png")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		OutputCoverName: "cover.jpg",
		CoverHeight:     240,
	}

	err = processImageWithLibrary(srcFile, destDir, cfg)
	if err != nil {
		t.Errorf("ProcessImageWithLibrary() error = %v", err)
	}

	// verify output file
	destFile := filepath.Join(destDir, cfg.OutputCoverName)
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// check image dimensions
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
	// create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "cover_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// setup test paths
	testDataDir := "../../test_data"
	srcFile := filepath.Join(testDataDir, "album.png")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err = fallbackCopyCover(srcFile, destDir, "cover.jpg")
	if err != nil {
		t.Errorf("FallbackCopyCover() error = %v", err)
	}

	// verify output file
	destFile := filepath.Join(destDir, "cover.jpg")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// verify the image can be decoded
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
