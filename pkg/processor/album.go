package processor

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nerten/albumpicker/pkg/config"
)

// FindAllAlbums recursively finds all directories containing FLAC files
func FindAllAlbums(rootDir string) ([]string, error) {
	var albums []string
	var mutex sync.Mutex

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error accessing path %s: %v\n", path, err)
			return nil // continue walking despite the error
		}

		// check for FLAC files in current directory
		if info.IsDir() {
			// read all entries in the current directory
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil
			}

			// check for FLAC files directly in this directory
			hasFlac := false
			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) == ".flac" {
					hasFlac = true
					break
				}
			}

			if hasFlac {
				// this directory contains FLAC files, treat it as an album
				mutex.Lock()
				albums = append(albums, path)
				mutex.Unlock()

				// skip processing subdirectories of this album
				return filepath.SkipDir
			}
		}
		return nil
	})

	return albums, err
}

// SelectRandomAlbums randomly selects n albums from the list
func SelectRandomAlbums(albums []string, n int) []string {

	// if n is greater than the number of albums, return all albums
	if n >= len(albums) {
		return albums
	}

	// shuffle the albums
	shuffled := make([]string, len(albums))
	copy(shuffled, albums)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// return the first n albums
	return shuffled[:n]
}

// ProcessAlbums processes the selected albums
func ProcessAlbums(albums []string, config *config.Config) error {
	var errs []error

	for _, albumPath := range albums {
		if err := ProcessAlbum(albumPath, config); err != nil {
			errs = append(errs, fmt.Errorf("error processing album %s: %v", albumPath, err))
		}
	}

	if len(errs) > 0 {
		// print all errors
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return fmt.Errorf("%d albums failed to process", len(errs))
	}

	fmt.Printf("Successfully processed %d albums\n", len(albums))
	return nil
}

// ProcessAlbum processes a single album
func ProcessAlbum(albumPath string, config *config.Config) error {
	// check if album path is within source directory
	if !isSubPath(config.Source, albumPath) {
		return fmt.Errorf("album path %s is not within source directory %s", albumPath, config.Source)
	}

	// get the relative path from source directory
	relPath, err := filepath.Rel(config.Source, albumPath)
	if err != nil {
		return fmt.Errorf("error getting relative path: %s", err)
	}

	// check if destination album already exists
	destAlbumPath := filepath.Join(config.Destination, relPath)
	if _, err := os.Stat(destAlbumPath); err == nil {
		fmt.Printf("Skipping existing album: %s\n", relPath)
		return nil
	}

	// create the destination album directory
	if err := os.MkdirAll(destAlbumPath, 0o755); err != nil {
		return fmt.Errorf("error creating destination directory: %s", err)
	}

	fmt.Printf("Processing album: %s\n", relPath)

	// find all FLAC files in the album
	entries, err := os.ReadDir(albumPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %s", err)
	}

	var flacFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".flac" {
			flacFiles = append(flacFiles, filepath.Join(albumPath, entry.Name()))
		}
	}

	fmt.Printf("Found %d FLAC files in album\n", len(flacFiles))

	if len(flacFiles) == 0 {
		return fmt.Errorf("no FLAC files found in album: %s", relPath)
	}

	// process FLAC files
	for _, flacFile := range flacFiles {
		if err := ProcessFLACFile(flacFile, albumPath, destAlbumPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error processing FLAC file %s: %v\n", flacFile, err)
			// continue processing other files despite the error
		}
	}

	// process cover files
	if err := ProcessCoverFile(albumPath, destAlbumPath, config); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error processing cover for album %s: %v\n", albumPath, err)
		// continue processing other albums despite the error
	}

	return nil
}

// isSubPath checks if the target path is within the base path
func isSubPath(basePath, targetPath string) bool {
	// clean and normalize paths
	basePath = filepath.Clean(basePath)
	targetPath = filepath.Clean(targetPath)

	// get absolute paths
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// check if target path starts with base path
	return absTarget != absBase && strings.HasPrefix(absTarget, absBase+string(os.PathSeparator))
}
