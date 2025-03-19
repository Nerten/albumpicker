package command

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nerten/albumpicker/internal/config"
	"github.com/nerten/albumpicker/internal/processor"
)

// Pick command
var PickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Randomly select and copy FLAC albums",
	RunE:  runPickCommand,
}

func init() {
	// Local flags
	PickCmd.Flags().IntP("count", "n", 0, "Number of albums to select (default 10)")
	PickCmd.Flags().Bool("wipe", false, "Attention!!! Destructive action! Wipe destination directory before copying albums")

	// Bind flags to viper
	viper.BindPFlag("albums_count", PickCmd.Flags().Lookup("count"))
}

// runPickCommand executes the pick command
func runPickCommand(cmd *cobra.Command, args []string) error {
	// Load configuration
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Find all albums in source directory
	fmt.Println("Scanning source directory for FLAC albums...")
	albums, err := findAllAlbums(conf.Source)
	if err != nil {
		return fmt.Errorf("error scanning source directory: %s", err)
	}

	if len(albums) == 0 {
		return fmt.Errorf("no FLAC albums found in source directory")
	}

	fmt.Printf("Found %d albums in total\n", len(albums))

	// Select random albums
	fmt.Printf("Selecting %d random albums...\n", conf.AlbumsCount)
	selectedAlbums := selectRandomAlbums(albums, conf.AlbumsCount)

	// Check wipe flag
	if wipe, _ := cmd.Flags().GetBool("wipe"); wipe {
		// Wipe destination directory
		fmt.Printf("Wiping destination directory: %s\n", conf.Destination)
		if err := os.RemoveAll(conf.Destination); err != nil {
			return fmt.Errorf("failed to wipe destination directory: %s", err)
		}
		// Recreate the destination directory
		if err := os.MkdirAll(conf.Destination, 0755); err != nil {
			return fmt.Errorf("failed to recreate destination directory: %s", err)
		}
	}

	// Process albums
	fmt.Printf("Processing selected %d albums...\n", len(selectedAlbums))
	return processAlbums(selectedAlbums, conf)
}

// findAllAlbums recursively finds all directories containing FLAC files
func findAllAlbums(rootDir string) ([]string, error) {
	var albums []string
	var mutex sync.Mutex

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error accessing path %s: %v\n", path, err)
			return nil // Continue walking despite the error
		}

		// Check for FLAC files in current directory
		if info.IsDir() {
			// Read all entries in the current directory
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil
			}

			// Check for FLAC files directly in this directory
			hasFlac := false
			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) == ".flac" {
					hasFlac = true
					break
				}
			}

			if hasFlac {
				// This directory contains FLAC files, treat it as an album
				mutex.Lock()
				albums = append(albums, path)
				mutex.Unlock()

				// Skip processing subdirectories of this album
				return filepath.SkipDir
			}
		}
		return nil
	})

	return albums, err
}

// selectRandomAlbums randomly selects n albums from the list
func selectRandomAlbums(albums []string, n int) []string {

	// If n is greater than the number of albums, return all albums
	if n >= len(albums) {
		return albums
	}

	// Shuffle the albums
	shuffled := make([]string, len(albums))
	copy(shuffled, albums)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Return the first n albums
	return shuffled[:n]
}

// processAlbums processes the selected albums
func processAlbums(albums []string, config *config.Config) error {
	var errs []error

	for _, albumPath := range albums {
		if err := processAlbum(albumPath, config); err != nil {
			errs = append(errs, fmt.Errorf("error processing album %s: %v", albumPath, err))
		}
	}

	if len(errs) > 0 {
		// Print all errors
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return fmt.Errorf("%d albums failed to process", len(errs))
	}

	fmt.Printf("Successfully processed %d albums\n", len(albums))
	return nil
}

// processAlbum processes a single album
func processAlbum(albumPath string, config *config.Config) error {
	// Check if album path is within source directory
	if !isSubPath(config.Source, albumPath) {
		return fmt.Errorf("album path %s is not within source directory %s", albumPath, config.Source)
	}

	// Get the relative path from source directory
	relPath, err := filepath.Rel(config.Source, albumPath)
	if err != nil {
		return fmt.Errorf("error getting relative path: %s", err)
	}

	// Check if destination album already exists
	destAlbumPath := filepath.Join(config.Destination, relPath)
	if _, err := os.Stat(destAlbumPath); err == nil {
		fmt.Printf("Skipping existing album: %s\n", relPath)
		return nil
	}

	// Create the destination album directory
	if err := os.MkdirAll(destAlbumPath, 0755); err != nil {
		return fmt.Errorf("error creating destination directory: %s", err)
	}

	fmt.Printf("Processing album: %s\n", relPath)

	// Find all FLAC files in the album
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

	// Process FLAC files
	for _, flacFile := range flacFiles {
		if err := processFLACFile(flacFile, albumPath, destAlbumPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error processing FLAC file %s: %v\n", flacFile, err)
			// Continue processing other files despite the error
		}
	}

	// Process cover files
	if err := processCoverFile(albumPath, destAlbumPath, config); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error processing cover for album %s: %v\n", albumPath, err)
		// Continue processing other albums despite the error
	}

	return nil
}

// isSubPath checks if the target path is within the base path
func isSubPath(basePath, targetPath string) bool {
	// Clean and normalize paths
	basePath = filepath.Clean(basePath)
	targetPath = filepath.Clean(targetPath)

	// Get absolute paths
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// Check if target path starts with base path
	return absTarget != absBase && strings.HasPrefix(absTarget, absBase+string(os.PathSeparator))
}

// processFLACFile processes a single FLAC file
func processFLACFile(flacFile, srcAlbumPath, destAlbumPath string) error {
	// Try to use the FLAC library to process the file
	err := processor.ProcessFLACWithLibrary(flacFile, srcAlbumPath, destAlbumPath)
	if err != nil {
		// If processing with the library fails, fall back to simple copy
		fmt.Fprintf(os.Stderr, "Warning: Failed to process FLAC with library: %v\n", err)
		fmt.Fprintf(os.Stderr, "Falling back to simple copy (PICTURE blocks will not be removed)\n")
		return processor.SimpleCopyFLACFile(flacFile, srcAlbumPath, destAlbumPath)
	}
	return nil
}

// processCoverFile processes the album cover
func processCoverFile(srcAlbumPath, destAlbumPath string, config *config.Config) error {

	// Check for cover files
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

		// Try to use the imaging library to process the cover
		err := processor.ProcessImageWithLibrary(coverFile, srcAlbumPath, destAlbumPath, config)
		if err != nil {
			// If processing with the library fails, fall back to simple copy
			fmt.Fprintf(os.Stderr, "Warning: Failed to process cover with imaging library: %v\n", err)
			fmt.Fprintf(os.Stderr, "Falling back to simple copy (cover will not be resized)\n")
			return processor.FallbackCopyCover(coverFile, srcAlbumPath, destAlbumPath, config.OutputCoverName)
		}
	} else {
		fmt.Printf("no cover file found")
	}
	return nil
}
