package processor

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/nerten/albumpicker/internal/config"
)

// processCoverFile processes the album cover by finding, resizing, and converting it to JPG
func ProcessImageWithLibrary(coverFile, srcAlbumPath, destAlbumPath string, config *config.Config) error {
	// Create destination cover file path
	destCoverPath := filepath.Join(destAlbumPath, config.OutputCoverName)

	fmt.Printf("  Processing cover to %dpx height\n", config.CoverHeight)

	// Open the source image
	srcImage, err := imaging.Open(coverFile)
	if err != nil {
		return fmt.Errorf("error opening image: %s", err)
	}

	// Resize image to the desired height while preserving aspect ratio
	newWidth := int(float64(config.CoverHeight) * float64(srcImage.Bounds().Dx()) / float64(srcImage.Bounds().Dy()))
	resized := imaging.Resize(srcImage, newWidth, config.CoverHeight, imaging.Lanczos)

	// Create the destination file
	destFile, err := os.Create(destCoverPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer destFile.Close()

	// Save as JPEG with quality 85
	opts := jpeg.Options{Quality: 85}
	if err := jpeg.Encode(destFile, resized, &opts); err != nil {
		return fmt.Errorf("error encoding JPEG: %s", err)
	}

	return nil
}

// fallbackCopyCover is a fallback method that copies the cover file without processing it
// This can be used if the imaging library fails or is not available
func FallbackCopyCover(coverFile, srcAlbumPath, destAlbumPath, outputCoverName string) error {
	// Create destination cover file path
	destCoverPath := filepath.Join(destAlbumPath, outputCoverName)

	fmt.Printf("  Copying cover (without processing): %s\n", filepath.Base(coverFile))

	// Open source file
	srcFile, err := os.Open(coverFile)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destCoverPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer destFile.Close()

	// Copy file contents
	_, err = os.ReadFile(coverFile)
	if err != nil {
		return fmt.Errorf("error reading source file: %s", err)
	}

	// Open the source image
	img, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("error decoding image: %s", err)
	}

	// Reset file pointer
	srcFile.Seek(0, 0)

	// If the source is already a JPEG, just copy it
	if strings.ToLower(filepath.Ext(coverFile)) == ".jpg" ||
		strings.ToLower(filepath.Ext(coverFile)) == ".jpeg" {
		_, err = srcFile.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("error seeking file: %s", err)
		}

		_, err = destFile.ReadFrom(srcFile)
		if err != nil {
			return fmt.Errorf("error copying file: %s", err)
		}
	} else {
		// Encode as JPEG with quality 85
		opts := jpeg.Options{Quality: 85}
		if err := jpeg.Encode(destFile, img, &opts); err != nil {
			return fmt.Errorf("error encoding JPEG: %s", err)
		}
	}

	return nil
}
