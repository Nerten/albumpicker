package processor

import (
	"fmt"
	"github.com/go-flac/go-flac"
	"io"
	"os"
	"path/filepath"
)

// ProcessFLACFile processes a single FLAC file
func ProcessFLACFile(flacFile, srcAlbumPath, destAlbumPath string) error {
	// try to use the FLAC library to process the file
	err := processFLACWithLibrary(flacFile, srcAlbumPath, destAlbumPath)
	if err != nil {
		// if processing with the library fails, fall back to simple copy
		fmt.Fprintf(os.Stderr, "Warning: Failed to process FLAC with library: %v\n", err)
		fmt.Fprintf(os.Stderr, "Falling back to simple copy (PICTURE blocks will not be removed)\n")
		return simpleCopyFLACFile(flacFile, srcAlbumPath, destAlbumPath)
	}
	return nil
}

// processFLACWithLibrary processes a single FLAC file by removing PICTURE blocks and copying it to the destination
func processFLACWithLibrary(flacFile, srcAlbumPath, destAlbumPath string) error {
	// get the relative path from album directory
	relFilePath, err := filepath.Rel(srcAlbumPath, flacFile)
	if err != nil {
		return fmt.Errorf("error getting relative file path: %s", err)
	}

	// create the destination file path
	destFilePath := filepath.Join(destAlbumPath, relFilePath)

	fmt.Printf("  Processing FLAC file: %s\n", relFilePath)

	// parse FLAC file
	file, err := flac.ParseFile(flacFile)
	if err != nil {
		return fmt.Errorf("error parsing FLAC file: %s", err)
	}

	// remove all PICTURE metadata blocks
	var newMetadata []*flac.MetaDataBlock
	for _, block := range file.Meta {
		if block.Type != flac.Picture && block.Type != flac.Padding {
			newMetadata = append(newMetadata, block)
		}
	}
	file.Meta = newMetadata

	// write modified FLAC to destination
	if err := file.Save(destFilePath); err != nil {
		return fmt.Errorf("error saving FLAC file: %s", err)
	}

	return nil
}

// simpleCopyFLACFile is a fallback method that copies the FLAC file without processing it
// This can be used if the go-flac library fails or is not available
func simpleCopyFLACFile(flacFile, srcAlbumPath, destAlbumPath string) error {
	// get the relative path from album directory
	relFilePath, err := filepath.Rel(srcAlbumPath, flacFile)
	if err != nil {
		return fmt.Errorf("error getting relative file path: %s", err)
	}

	// create the destination file path
	destFilePath := filepath.Join(destAlbumPath, relFilePath)

	fmt.Printf("  Copying FLAC file (without processing): %s\n", relFilePath)

	// open source file
	srcFile, err := os.Open(flacFile)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", err)
	}
	defer srcFile.Close()

	// create destination file
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer destFile.Close()

	// copy file contents
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	return nil
}
