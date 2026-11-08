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

// FindAllAlbums recursively finds all directories containing audio files
func FindAllAlbums(rootDir string) ([]string, error) {
	var albums []string
	var mutex sync.Mutex

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error accessing path %s: %v\n", path, err)
			return nil // continue walking despite the error
		}

		// check for audio files in current directory
		if info.IsDir() {
			// read all entries in the current directory
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil
			}

			// check for audio files directly in this directory
			hasAudio := false
			for _, entry := range entries {
				if isAudioFile(entry) {
					hasAudio = true
					break
				}
			}

			if hasAudio {
				// this directory contains audio files, treat it as an album
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
func ProcessAlbums(albums []string, cfg *config.Config) error {
	var errs []error

	for _, albumPath := range albums {
		if err := ProcessAlbum(albumPath, cfg); err != nil {
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
func ProcessAlbum(albumPath string, cfg *config.Config) error {
	// check if album path is within source directory
	if !isSubPath(cfg.Source, albumPath) {
		return fmt.Errorf("album path %s is not within source directory %s", albumPath, cfg.Source)
	}

	// get the relative path from source directory
	relPath, err := filepath.Rel(cfg.Source, albumPath)
	if err != nil {
		return fmt.Errorf("error getting relative path: %s", err)
	}

	// check if destination album already exists
	destAlbumPath := filepath.Join(cfg.Destination, relPath)
	if _, err := os.Stat(destAlbumPath); err == nil {
		fmt.Printf("Skipping existing album: %s\n", relPath)
		return nil
	}

	// create the destination album directory
	if err := os.MkdirAll(destAlbumPath, 0o755); err != nil {
		return fmt.Errorf("error creating destination directory: %s", err)
	}

	fmt.Printf("Processing album: %s\n", relPath)

	// find all audio files in the album
	entries, err := os.ReadDir(albumPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %s", err)
	}

	var audioFiles []string
	for _, entry := range entries {
		if isAudioFile(entry) {
			audioFiles = append(audioFiles, filepath.Join(albumPath, entry.Name()))
		}
	}

	fmt.Printf("Found %d audio files in album\n", len(audioFiles))

	if len(audioFiles) == 0 {
		return fmt.Errorf("no audio files found in album: %s", relPath)
	}

	var embeddedCover []byte
	// 1. Scan for external cover first
	var coverFile string
	for _, coverName := range cfg.CoverFilenames {
		candidate := filepath.Join(albumPath, coverName)
		if _, err := os.Stat(candidate); err == nil {
			coverFile = candidate
			break
		}
	}

	// 2. If no external cover found, try to extract from the first audio file that has one
	if coverFile == "" {
		for _, audioFile := range audioFiles {
			picData, err := ExtractPictureFromAudioFile(audioFile)
			if err == nil && len(picData) > 0 {
				embeddedCover = picData
				break
			}
		}
	}

	// 3. Process cover file (creates destAlbumPath/cover.jpg for Rockbox and returns resized bytes)
	optimizedCover, err := ProcessCoverFile(albumPath, destAlbumPath, cfg, embeddedCover)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error processing cover for album %s: %v\n", albumPath, err)
		// continue processing audio files despite cover error
	}

	// 4. If embed-cover is enabled and we have an optimized cover, embed it into audio files
	var coverToEmbed []byte
	if cfg.EmbedCover && len(optimizedCover) > 0 {
		coverToEmbed = optimizedCover
	}

	// 5. Process audio files
	for _, audioFile := range audioFiles {
		if _, err := ProcessAudioFile(audioFile, albumPath, destAlbumPath, coverToEmbed); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error processing audio file %s: %v\n", audioFile, err)
			// continue processing other files despite the error
		}
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

// isAudioFile checks if the directory entry is a supported audio file (.flac, .mp3, .m4a, .alac, .aac)
func isAudioFile(entry os.DirEntry) bool {
	if entry.IsDir() || strings.HasPrefix(entry.Name(), "._") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(entry.Name()))
	switch ext {
	case ".flac", ".mp3", ".m4a", ".alac", ".aac":
		return true
	default:
		return false
	}
}
