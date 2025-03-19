package processor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-flac/go-flac"
)

// ProcessFLACWithLibrary processes a single FLAC file by removing PICTURE blocks and copying it to the destination
func ProcessFLACWithLibrary(flacFile, srcAlbumPath, destAlbumPath string) error {
	// Get the relative path from album directory
	relFilePath, err := filepath.Rel(srcAlbumPath, flacFile)
	if err != nil {
		return fmt.Errorf("error getting relative file path: %s", err)
	}

	// Create the destination file path
	destFilePath := filepath.Join(destAlbumPath, relFilePath)

	fmt.Printf("  Processing FLAC file: %s\n", relFilePath)

	// Parse FLAC file
	file, err := flac.ParseFile(flacFile)
	if err != nil {
		return fmt.Errorf("error parsing FLAC file: %s", err)
	}

	// Remove all PICTURE metadata blocks
	var newMetadata []*flac.MetaDataBlock
	for _, block := range file.Meta {
		if block.Type != flac.Picture && block.Type != flac.Padding {
			newMetadata = append(newMetadata, block)
		}
	}
	file.Meta = newMetadata

	// Write modified FLAC to destination
	if err := file.Save(destFilePath); err != nil {
		return fmt.Errorf("error saving FLAC file: %s", err)
	}

	return nil
}

// SimpleCopyFLACFile is a fallback method that copies the FLAC file without processing it
// This can be used if the go-flac library fails or is not available
func SimpleCopyFLACFile(flacFile, srcAlbumPath, destAlbumPath string) error {
	// Get the relative path from album directory
	relFilePath, err := filepath.Rel(srcAlbumPath, flacFile)
	if err != nil {
		return fmt.Errorf("error getting relative file path: %s", err)
	}

	// Create the destination file path
	destFilePath := filepath.Join(destAlbumPath, relFilePath)

	fmt.Printf("  Copying FLAC file (without processing): %s\n", relFilePath)

	// Open source file
	srcFile, err := os.Open(flacFile)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	return nil
}
