package processor

import (
	"bytes"
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

// ProcessCoverFile processes the album cover: checks external files first, then falls back to embeddedCover.
// It saves the cover to destAlbumPath and returns the optimized cover image bytes.
func ProcessCoverFile(srcAlbumPath, destAlbumPath string, cfg *config.Config, embeddedCover ...[]byte) ([]byte, error) {
	// check for cover files
	var coverFile string
	for _, coverName := range cfg.CoverFilenames {
		candidate := filepath.Join(srcAlbumPath, coverName)
		if _, err := os.Stat(candidate); err == nil {
			coverFile = candidate
			break
		}
	}

	if coverFile != "" {
		fmt.Printf("  Found cover file: %s\n", filepath.Base(coverFile))

		// try to use the imaging library to process the cover
		data, err := processImageWithLibrary(coverFile, destAlbumPath, cfg)
		if err != nil {
			// if processing with the library fails, fall back to simple copy
			fmt.Fprintf(os.Stderr, "Warning: Failed to process cover with imaging library: %v\n", err)
			fmt.Fprintf(os.Stderr, "Falling back to simple copy (cover will not be resized)\n")
			return fallbackCopyCover(coverFile, destAlbumPath, cfg.OutputCoverName)
		}
		return data, nil
	}

	// if external cover not found, check embeddedCover
	var embData []byte
	if len(embeddedCover) > 0 && len(embeddedCover[0]) > 0 {
		embData = embeddedCover[0]
	}

	if len(embData) > 0 {
		fmt.Printf("  Extracting and resizing embedded cover (%d bytes)\n", len(embData))
		data, err := ProcessImageBytes(embData, destAlbumPath, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to process embedded cover with imaging library: %v\n", err)
			destCoverPath := filepath.Clean(filepath.Join(destAlbumPath, cfg.OutputCoverName))
			destFile, createErr := os.Create(destCoverPath)
			if createErr != nil {
				return nil, fmt.Errorf("error creating destination file for embedded cover: %w", createErr)
			}
			_, writeErr := destFile.Write(embData)
			_ = destFile.Close()
			if writeErr != nil {
				return nil, fmt.Errorf("error saving raw embedded cover: %w", writeErr)
			}
			return embData, nil
		}
		return data, nil
	}

	fmt.Println("no cover file found")
	return nil, nil
}

// ProcessImageBytes decodes, resizes, and saves an embedded image to destAlbumPath
func ProcessImageBytes(imgData []byte, destAlbumPath string, cfg *config.Config) ([]byte, error) {
	srcImage, err := imaging.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("error decoding embedded image: %w", err)
	}
	return resizeAndSaveCover(srcImage, destAlbumPath, cfg)
}

// processCoverFile processes the album cover by finding, resizing, and converting it to JPG
func processImageWithLibrary(coverFile, destAlbumPath string, cfg *config.Config) ([]byte, error) {
	// open the source image
	srcImage, err := imaging.Open(coverFile)
	if err != nil {
		return nil, fmt.Errorf("error opening image: %w", err)
	}
	return resizeAndSaveCover(srcImage, destAlbumPath, cfg)
}

func resizeAndSaveCover(srcImage image.Image, destAlbumPath string, cfg *config.Config) ([]byte, error) {
	destCoverPath := filepath.Join(destAlbumPath, cfg.OutputCoverName)
	fmt.Printf("  Processing cover to %dpx height\n", cfg.CoverHeight)

	newWidth := int(float64(cfg.CoverHeight) * float64(srcImage.Bounds().Dx()) / float64(srcImage.Bounds().Dy()))
	resized := imaging.Resize(srcImage, newWidth, cfg.CoverHeight, imaging.Lanczos)

	var buf bytes.Buffer
	opts := jpeg.Options{Quality: 85}
	if err := jpeg.Encode(&buf, resized, &opts); err != nil {
		return nil, fmt.Errorf("error encoding JPEG: %w", err)
	}

	destFile, err := os.Create(destCoverPath)
	if err != nil {
		return nil, fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := destFile.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("error writing cover file: %w", err)
	}

	return buf.Bytes(), nil
}

// fallbackCopyCover is a fallback method that copies the cover file without processing it
// This can be used if the imaging library fails or is not available
func fallbackCopyCover(coverFile, destAlbumPath, outputCoverName string) ([]byte, error) {
	destCoverPath := filepath.Join(destAlbumPath, outputCoverName)
	fmt.Printf("  Copying cover (without processing): %s\n", filepath.Base(coverFile))

	srcFile, err := os.Open(coverFile)
	if err != nil {
		return nil, fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destCoverPath)
	if err != nil {
		return nil, fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	var buf bytes.Buffer
	if strings.ToLower(filepath.Ext(coverFile)) == ".jpg" ||
		strings.ToLower(filepath.Ext(coverFile)) == ".jpeg" {
		_, err = srcFile.Seek(0, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("error seeking file: %w", err)
		}

		if _, err := io.Copy(&buf, srcFile); err != nil {
			return nil, fmt.Errorf("error copying file: %w", err)
		}
		if _, err := destFile.Write(buf.Bytes()); err != nil {
			return nil, fmt.Errorf("error writing cover file: %w", err)
		}
	} else {
		opts := jpeg.Options{Quality: 85}
		if err := jpeg.Encode(&buf, img, &opts); err != nil {
			return nil, fmt.Errorf("error encoding JPEG: %w", err)
		}
		if _, err := destFile.Write(buf.Bytes()); err != nil {
			return nil, fmt.Errorf("error writing cover file: %w", err)
		}
	}

	return buf.Bytes(), nil
}
