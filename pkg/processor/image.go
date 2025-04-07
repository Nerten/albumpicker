package processor

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/nerten/albumpicker/pkg/config"
)

// ProcessCoverFile processes the album cover
func ProcessCoverFile(srcAlbumPath, destAlbumPath string, config *config.Config) error {

	// check for cover files
	var coverFile string
	for _, coverName := range config.CoverFilenames {
		candidate := filepath.Join(srcAlbumPath, coverName)
		if _, err := os.Stat(candidate); err == nil {
			coverFile = candidate
			break
		}
	}

	if coverFile != "" {
		fmt.Printf("  Found cover file: %s\n", filepath.Base(coverFile))

		// try to use the imaging library to process the cover
		err := processImageWithLibrary(coverFile, destAlbumPath, config)
		if err != nil {
			// if processing with the library fails, fall back to simple copy
			fmt.Fprintf(os.Stderr, "Warning: Failed to process cover with imaging library: %v\n", err)
			fmt.Fprintf(os.Stderr, "Falling back to simple copy (cover will not be resized)\n")
			return fallbackCopyCover(coverFile, destAlbumPath, config.OutputCoverName)
		}
	} else {
		fmt.Printf("no cover file found")
	}
	return nil
}

// processCoverFile processes the album cover by finding, resizing, and converting it to JPG
func processImageWithLibrary(coverFile, destAlbumPath string, config *config.Config) error {
	// create destination cover file path
	destCoverPath := filepath.Join(destAlbumPath, config.OutputCoverName)

	fmt.Printf("  Processing cover to %dpx height\n", config.CoverHeight)

	// open the source image
	srcImage, err := imaging.Open(coverFile)
	if err != nil {
		return fmt.Errorf("error opening image: %s", err)
	}

	// resize image to the desired height while preserving aspect ratio
	newWidth := int(float64(config.CoverHeight) * float64(srcImage.Bounds().Dx()) / float64(srcImage.Bounds().Dy()))
	resized := imaging.Resize(srcImage, newWidth, config.CoverHeight, imaging.Lanczos)

	// create the destination file
	destFile, err := os.Create(destCoverPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer destFile.Close()

	// save as JPEG with quality 85
	opts := jpeg.Options{Quality: 85}
	if err := jpeg.Encode(destFile, resized, &opts); err != nil {
		return fmt.Errorf("error encoding JPEG: %s", err)
	}

	return nil
}

// fallbackCopyCover is a fallback method that copies the cover file without processing it
// This can be used if the imaging library fails or is not available
func fallbackCopyCover(coverFile, destAlbumPath, outputCoverName string) error {
	// create destination cover file path
	destCoverPath := filepath.Join(destAlbumPath, outputCoverName)

	fmt.Printf("  Copying cover (without processing): %s\n", filepath.Base(coverFile))

	// open source file
	srcFile, err := os.Open(coverFile)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", err)
	}
	defer srcFile.Close()

	// create destination file
	destFile, err := os.Create(destCoverPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer destFile.Close()

	// copy file contents
	_, err = os.ReadFile(coverFile)
	if err != nil {
		return fmt.Errorf("error reading source file: %s", err)
	}

	// open the source image
	img, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("error decoding image: %s", err)
	}

	// reset file pointer
	_, err = srcFile.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error seeking file: %s", err)
	}

	// if the source is already a JPEG, just copy it
	if strings.ToLower(filepath.Ext(coverFile)) == ".jpg" ||
		strings.ToLower(filepath.Ext(coverFile)) == ".jpeg" {
		_, err = srcFile.Seek(0, io.SeekStart)
		if err != nil {
			return fmt.Errorf("error seeking file: %s", err)
		}

		_, err = destFile.ReadFrom(srcFile)
		if err != nil {
			return fmt.Errorf("error copying file: %s", err)
		}
	} else {
		// encode as JPEG with quality 85
		opts := jpeg.Options{Quality: 85}
		if err := jpeg.Encode(destFile, img, &opts); err != nil {
			return fmt.Errorf("error encoding JPEG: %s", err)
		}
	}

	return nil
}
