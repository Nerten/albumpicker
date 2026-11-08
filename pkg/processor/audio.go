package processor

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-flac/go-flac"
)

// ProcessAudioFile routes audio file processing based on extension (.flac, .mp3, .m4a, .alac, .aac)
func ProcessAudioFile(audioFile, srcAlbumPath, destAlbumPath string, embedCover ...[]byte) ([]byte, error) {
	var toEmbed []byte
	if len(embedCover) > 0 {
		toEmbed = embedCover[0]
	}

	ext := strings.ToLower(filepath.Ext(audioFile))
	switch ext {
	case ".flac":
		return ProcessFLACFile(audioFile, srcAlbumPath, destAlbumPath, toEmbed)
	case ".mp3":
		return ProcessMP3File(audioFile, srcAlbumPath, destAlbumPath, toEmbed)
	case ".m4a", ".alac", ".aac":
		return ProcessM4AFile(audioFile, srcAlbumPath, destAlbumPath)
	default:
		return nil, simpleCopyAudioFile(audioFile, srcAlbumPath, destAlbumPath)
	}
}

// ProcessMP3File extracts embedded ID3 APIC cover (if present), modifies/embeds APIC, and copies MP3
func ProcessMP3File(mp3File, srcAlbumPath, destAlbumPath string, embedCover ...[]byte) ([]byte, error) {
	var toEmbed []byte
	if len(embedCover) > 0 {
		toEmbed = embedCover[0]
	}

	picData, _ := ExtractPictureFromMP3(mp3File)

	relFilePath, err := filepath.Rel(srcAlbumPath, mp3File)
	if err != nil {
		return nil, fmt.Errorf("error getting relative file path: %w", err)
	}
	destFilePath := filepath.Join(destAlbumPath, relFilePath)

	fmt.Printf("  Processing MP3 file: %s\n", relFilePath)

	if err := ProcessMP3WithID3(mp3File, destFilePath, toEmbed); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to process MP3 with ID3: %v, falling back to simple copy\n", err)
		return picData, simpleCopyAudioFile(mp3File, srcAlbumPath, destAlbumPath)
	}

	return picData, nil
}

// ProcessMP3WithID3 modifies an MP3 file's ID3v2 tag: strips existing APIC frames,
// optionally embeds new cover art, and writes to destPath preserving audio frames.
func ProcessMP3WithID3(srcPath, destPath string, embedCover []byte) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	header := make([]byte, 10)
	var audioOffset int64
	var keptFrames []byte

	n, err := io.ReadFull(srcFile, header)
	if err == nil && n == 10 && string(header[0:3]) == "ID3" {
		version := header[3]
		tagSize := int(header[6]&0x7F)<<21 | int(header[7]&0x7F)<<14 | int(header[8]&0x7F)<<7 | int(header[9]&0x7F)
		if tagSize > 0 && tagSize <= 50*1024*1024 {
			tagData := make([]byte, tagSize)
			if _, readErr := io.ReadFull(srcFile, tagData); readErr == nil {
				audioOffset = 10 + int64(tagSize)
				offset := 0
				for offset+10 <= len(tagData) {
					frameID := string(tagData[offset : offset+4])
					if frameID[0] == 0 {
						break
					}
					var frameSize int
					if version == 4 {
						frameSize = int(tagData[offset+4]&0x7F)<<21 | int(tagData[offset+5]&0x7F)<<14 | int(tagData[offset+6]&0x7F)<<7 | int(tagData[offset+7]&0x7F)
					} else {
						frameSize = int(binary.BigEndian.Uint32(tagData[offset+4 : offset+8]))
					}
					totalFrameLen := 10 + frameSize
					if frameSize <= 0 || offset+totalFrameLen > len(tagData) {
						break
					}
					if frameID != "APIC" {
						keptFrames = append(keptFrames, tagData[offset:offset+totalFrameLen]...)
					}
					offset += totalFrameLen
				}
			}
		}
	}

	if _, err := srcFile.Seek(audioOffset, io.SeekStart); err != nil {
		return fmt.Errorf("error seeking audio stream: %w", err)
	}

	if len(embedCover) > 0 {
		if apicFrame := buildAPICFrame(embedCover); len(apicFrame) > 0 {
			keptFrames = append(keptFrames, apicFrame...)
		}
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	if len(keptFrames) > 0 {
		newHeader := make([]byte, 10)
		copy(newHeader[0:3], "ID3")
		newHeader[3] = 3 // ID3v2.3
		newHeader[4] = 0 // revision
		newHeader[5] = 0 // flags
		synchsafe := encodeSynchsafe(len(keptFrames))
		copy(newHeader[6:10], synchsafe)

		if _, err := destFile.Write(newHeader); err != nil {
			return fmt.Errorf("error writing ID3 header: %w", err)
		}
		if _, err := destFile.Write(keptFrames); err != nil {
			return fmt.Errorf("error writing ID3 frames: %w", err)
		}
	}

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("error copying audio frames: %w", err)
	}

	return nil
}

func encodeSynchsafe(size int) []byte {
	return []byte{
		byte((size >> 21) & 0x7F),
		byte((size >> 14) & 0x7F),
		byte((size >> 7) & 0x7F),
		byte(size & 0x7F),
	}
}

func buildAPICFrame(imgData []byte) []byte {
	mime := "image/jpeg\x00"
	contentSize := 1 + len(mime) + 1 + 1 + len(imgData)
	if contentSize < 0 || contentSize > math.MaxUint32 {
		return nil
	}
	frame := make([]byte, 10+contentSize)
	copy(frame[0:4], "APIC")
	binary.BigEndian.PutUint32(frame[4:8], uint32(contentSize))
	frame[8] = 0
	frame[9] = 0

	offset := 10
	frame[offset] = 0 // ISO-8859-1
	offset++
	copy(frame[offset:], mime)
	offset += len(mime)
	frame[offset] = 3 // Cover (front)
	offset++
	frame[offset] = 0 // Description: empty string
	offset++
	copy(frame[offset:], imgData)

	return frame
}

// ProcessM4AFile extracts embedded cover from M4A/ALAC/AAC (if present) and copies the file
func ProcessM4AFile(m4aFile, srcAlbumPath, destAlbumPath string) ([]byte, error) {
	picData, _ := ExtractPictureFromM4A(m4aFile)
	err := simpleCopyAudioFile(m4aFile, srcAlbumPath, destAlbumPath)
	if err != nil {
		return nil, err
	}
	return picData, nil
}

// simpleCopyAudioFile copies an audio file from srcAlbumPath to destAlbumPath preserving relative path
func simpleCopyAudioFile(audioFile, srcAlbumPath, destAlbumPath string) error {
	relFilePath, err := filepath.Rel(srcAlbumPath, audioFile)
	if err != nil {
		return fmt.Errorf("error getting relative file path: %w", err)
	}

	destFilePath := filepath.Join(destAlbumPath, relFilePath)
	fmt.Printf("  Copying audio file: %s\n", relFilePath)

	// ensure destination directory exists (for nested artist/album folders)
	if err := os.MkdirAll(filepath.Dir(destFilePath), 0o755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	srcFile, err := os.Open(audioFile)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}

// ExtractPictureFromMP3 extracts embedded APIC picture from an MP3 file
func ExtractPictureFromMP3(filePath string) ([]byte, error) {
	srcFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	header := make([]byte, 10)
	if _, err := io.ReadFull(srcFile, header); err != nil {
		return nil, err
	}
	if string(header[0:3]) != "ID3" {
		return nil, errors.New("no ID3 tag")
	}

	version := header[3]
	tagSize := int(header[6]&0x7F)<<21 | int(header[7]&0x7F)<<14 | int(header[8]&0x7F)<<7 | int(header[9]&0x7F)
	if tagSize <= 0 || tagSize > 50*1024*1024 {
		return nil, errors.New("invalid tag size")
	}

	tagData := make([]byte, tagSize)
	if _, err := io.ReadFull(srcFile, tagData); err != nil {
		return nil, err
	}

	offset := 0
	for offset+10 <= len(tagData) {
		frameID := string(tagData[offset : offset+4])
		if frameID[0] == 0 {
			break
		}

		var frameSize int
		if version == 4 {
			frameSize = int(tagData[offset+4]&0x7F)<<21 | int(tagData[offset+5]&0x7F)<<14 | int(tagData[offset+6]&0x7F)<<7 | int(tagData[offset+7]&0x7F)
		} else {
			frameSize = int(binary.BigEndian.Uint32(tagData[offset+4 : offset+8]))
		}

		offset += 10
		if frameSize <= 0 || offset+frameSize > len(tagData) {
			break
		}

		if frameID == "APIC" {
			frameData := tagData[offset : offset+frameSize]
			if len(frameData) > 4 {
				enc := frameData[0]
				idx := 1
				for idx < len(frameData) && frameData[idx] != 0 {
					idx++
				}
				idx++ // skip null byte

				if idx < len(frameData) {
					idx++ // skip picture type
					if enc == 1 || enc == 2 {
						for idx+1 < len(frameData) && (frameData[idx] != 0 || frameData[idx+1] != 0) {
							idx += 2
						}
						idx += 2
					} else {
						for idx < len(frameData) && frameData[idx] != 0 {
							idx++
						}
						idx++
					}

					if idx < len(frameData) {
						return frameData[idx:], nil
					}
				}
			}
		}

		offset += frameSize
	}

	return nil, errors.New("no APIC frame found")
}

// ExtractPictureFromM4A extracts embedded cover from M4A/ALAC/AAC QuickTime container
func ExtractPictureFromM4A(filePath string) ([]byte, error) {
	srcFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	stat, err := srcFile.Stat()
	if err != nil {
		return nil, err
	}
	return findCoverInBoxes(srcFile, 0, stat.Size())
}

func findCoverInBoxes(r io.ReaderAt, start, length int64) ([]byte, error) {
	offset := start
	end := start + length
	hdr := make([]byte, 8)

	for offset+8 <= end {
		if _, err := r.ReadAt(hdr, offset); err != nil {
			break
		}
		boxSize := int64(binary.BigEndian.Uint32(hdr[0:4]))
		boxType := string(hdr[4:8])

		headerSize := int64(8)
		if boxSize == 1 {
			largeHdr := make([]byte, 8)
			if _, err := r.ReadAt(largeHdr, offset+8); err != nil {
				break
			}
			val := binary.BigEndian.Uint64(largeHdr)
			if val > 1<<62 {
				break
			}
			boxSize = int64(val)
			headerSize = 16
		} else if boxSize == 0 {
			boxSize = end - offset
		}

		if boxSize < headerSize || offset+boxSize > end {
			break
		}

		contentOffset := offset + headerSize
		contentLength := boxSize - headerSize

		switch boxType {
		case "moov", "udta", "ilst", "covr":
			data, err := findCoverInBoxes(r, contentOffset, contentLength)
			if err == nil && len(data) > 0 {
				return data, nil
			}
		case "meta":
			if contentLength > 4 {
				data, err := findCoverInBoxes(r, contentOffset+4, contentLength-4)
				if err == nil && len(data) > 0 {
					return data, nil
				}
			}
		case "data":
			if contentLength > 8 {
				imgData := make([]byte, contentLength-8)
				if _, err := r.ReadAt(imgData, contentOffset+8); err == nil {
					return imgData, nil
				}
			}
		}

		offset += boxSize
	}
	return nil, errors.New("no cover found in M4A")
}

// ExtractPictureFromAudioFile extracts embedded artwork from supported audio files
func ExtractPictureFromAudioFile(filePath string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".flac":
		return ExtractPictureFromFLAC(filePath)
	case ".mp3":
		return ExtractPictureFromMP3(filePath)
	case ".m4a", ".alac", ".aac":
		return ExtractPictureFromM4A(filePath)
	default:
		return nil, nil
	}
}

// ExtractPictureFromFLAC extracts embedded artwork from a FLAC file
func ExtractPictureFromFLAC(filePath string) ([]byte, error) {
	file, err := flac.ParseFile(filePath)
	if err != nil {
		return nil, err
	}
	var extractedPic []byte
	for _, block := range file.Meta {
		if block.Type == flac.Picture {
			picType, _, imgData, err := ExtractPictureFromMetaDataBlock(block.Data)
			if err == nil && len(imgData) > 0 {
				if len(extractedPic) == 0 || picType == 3 {
					extractedPic = imgData
				}
			}
		}
	}
	if len(extractedPic) == 0 {
		return nil, errors.New("no picture found in FLAC")
	}
	return extractedPic, nil
}
