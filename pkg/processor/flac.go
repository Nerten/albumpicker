package processor

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/go-flac/go-flac"
)

// ExtractPictureFromMetaDataBlock extracts picture type, MIME type and raw image bytes from a FLAC Picture metadata block
func ExtractPictureFromMetaDataBlock(data []byte) (pictureType uint32, mimeType string, imgData []byte, err error) {
	if len(data) < 32 {
		return 0, "", nil, errors.New("picture block data too short")
	}
	offset := 0
	pictureType = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	mimeLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if mimeLen < 0 || len(data) < offset+mimeLen+4 {
		return 0, "", nil, errors.New("invalid mime type length")
	}
	mimeType = string(data[offset : offset+mimeLen])
	offset += mimeLen

	descLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if descLen < 0 || len(data) < offset+descLen+16+4 {
		return 0, "", nil, errors.New("invalid description length")
	}
	offset += descLen + 16

	dataLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if dataLen < 0 || len(data) < offset+dataLen {
		return 0, "", nil, errors.New("invalid picture data length")
	}
	imgData = data[offset : offset+dataLen]
	return pictureType, mimeType, imgData, nil
}

// CreateFLACPictureBlock creates a FLAC Picture metadata block
func CreateFLACPictureBlock(imgData []byte, mimeType string, width, height uint32) *flac.MetaDataBlock {
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	mimeLen := len(mimeType)
	dataLen := len(imgData)
	if mimeLen < 0 || uint64(mimeLen) > math.MaxUint32 || dataLen < 0 || uint64(dataLen) > math.MaxUint32 {
		return nil
	}
	blockLen := 4 + 4 + mimeLen + 4 + 16 + 4 + dataLen
	buf := make([]byte, blockLen)
	offset := 0

	// Picture type (3 = Front Cover)
	binary.BigEndian.PutUint32(buf[offset:offset+4], 3)
	offset += 4

	// MIME type length and string
	binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(mimeLen))
	offset += 4
	copy(buf[offset:offset+mimeLen], mimeType)
	offset += mimeLen

	// Description length (0)
	binary.BigEndian.PutUint32(buf[offset:offset+4], 0)
	offset += 4

	// Width & Height
	binary.BigEndian.PutUint32(buf[offset:offset+4], width)
	offset += 4
	binary.BigEndian.PutUint32(buf[offset:offset+4], height)
	offset += 4

	// Color depth (24 bits)
	binary.BigEndian.PutUint32(buf[offset:offset+4], 24)
	offset += 4

	// Indexed colors (0)
	binary.BigEndian.PutUint32(buf[offset:offset+4], 0)
	offset += 4

	// Picture data length and bytes
	binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(dataLen))
	offset += 4
	copy(buf[offset:], imgData)

	return &flac.MetaDataBlock{
		Type: flac.Picture,
		Data: buf,
	}
}

// ProcessFLACFile processes a single FLAC file, removes PICTURE blocks, optionally embeds a cover, and returns extracted picture data if present
func ProcessFLACFile(flacFile, srcAlbumPath, destAlbumPath string, embedCover ...[]byte) ([]byte, error) {
	var toEmbed []byte
	if len(embedCover) > 0 {
		toEmbed = embedCover[0]
	}

	// try to use the FLAC library to process the file
	picData, err := processFLACWithLibrary(flacFile, srcAlbumPath, destAlbumPath, toEmbed)
	if err != nil {
		// if processing with the library fails, fall back to simple copy
		fmt.Fprintf(os.Stderr, "Warning: Failed to process FLAC with library: %v\n", err)
		fmt.Fprintf(os.Stderr, "Falling back to simple copy (PICTURE blocks will not be modified)\n")
		return nil, simpleCopyFLACFile(flacFile, srcAlbumPath, destAlbumPath)
	}
	return picData, nil
}

// processFLACWithLibrary processes a single FLAC file by removing old PICTURE blocks, optionally embedding a new cover, and copying it to the destination
func processFLACWithLibrary(flacFile, srcAlbumPath, destAlbumPath string, embedCover []byte) ([]byte, error) {
	// get the relative path from album directory
	relFilePath, err := filepath.Rel(srcAlbumPath, flacFile)
	if err != nil {
		return nil, fmt.Errorf("error getting relative file path: %w", err)
	}

	// create the destination file path
	destFilePath := filepath.Join(destAlbumPath, relFilePath)

	fmt.Printf("  Processing FLAC file: %s\n", relFilePath)

	// ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destFilePath), 0o755); err != nil {
		return nil, fmt.Errorf("error creating destination directory: %w", err)
	}

	// parse FLAC file
	file, err := flac.ParseFile(flacFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing FLAC file: %w", err)
	}

	var extractedPic []byte
	// remove all PICTURE metadata blocks
	var newMetadata []*flac.MetaDataBlock
	for _, block := range file.Meta {
		if block.Type == flac.Picture {
			picType, _, imgData, err := ExtractPictureFromMetaDataBlock(block.Data)
			if err == nil && len(imgData) > 0 {
				// Prefer Front Cover (type 3) or take first available
				if len(extractedPic) == 0 || picType == 3 {
					extractedPic = imgData
				}
			}
			continue
		}
		if block.Type == flac.Padding {
			continue
		}
		newMetadata = append(newMetadata, block)
	}

	// if embedCover is provided, append new PICTURE metadata block
	if len(embedCover) > 0 {
		if block := CreateFLACPictureBlock(embedCover, "image/jpeg", 0, 0); block != nil {
			newMetadata = append(newMetadata, block)
		}
	}
	file.Meta = newMetadata

	// write modified FLAC to destination
	if err := file.Save(destFilePath); err != nil {
		return nil, fmt.Errorf("error saving FLAC file: %w", err)
	}

	return extractedPic, nil
}

// simpleCopyFLACFile is a fallback method that copies the FLAC file without processing it
// This can be used if the go-flac library fails or is not available
func simpleCopyFLACFile(flacFile, srcAlbumPath, destAlbumPath string) error {
	return simpleCopyAudioFile(flacFile, srcAlbumPath, destAlbumPath)
}
