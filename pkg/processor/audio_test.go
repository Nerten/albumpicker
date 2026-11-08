package processor

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessAudioFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "audio_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// 1. Test MP3 copy
	mp3Path := filepath.Join(srcDir, "song.mp3")
	if err := os.WriteFile(mp3Path, []byte("fake mp3 data"), 0o644); err != nil {
		t.Fatal(err)
	}
	pic, err := ProcessAudioFile(mp3Path, srcDir, destDir)
	if err != nil {
		t.Errorf("ProcessAudioFile(.mp3) error = %v", err)
	}
	if pic != nil {
		t.Errorf("expected no picture from plain fake mp3, got %v", pic)
	}
	if _, err := os.Stat(filepath.Join(destDir, "song.mp3")); os.IsNotExist(err) {
		t.Error("dest song.mp3 was not created")
	}

	// 2. Test M4A copy
	m4aPath := filepath.Join(srcDir, "song.m4a")
	if err := os.WriteFile(m4aPath, []byte("fake m4a data"), 0o644); err != nil {
		t.Fatal(err)
	}
	pic, err = ProcessAudioFile(m4aPath, srcDir, destDir)
	if err != nil {
		t.Errorf("ProcessAudioFile(.m4a) error = %v", err)
	}
	if pic != nil {
		t.Errorf("expected no picture from plain fake m4a, got %v", pic)
	}
	if _, err := os.Stat(filepath.Join(destDir, "song.m4a")); os.IsNotExist(err) {
		t.Error("dest song.m4a was not created")
	}

	// 3. Test AAC / ALAC copy
	aacPath := filepath.Join(srcDir, "song.aac")
	if err := os.WriteFile(aacPath, []byte("fake aac data"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = ProcessAudioFile(aacPath, srcDir, destDir)
	if err != nil {
		t.Errorf("ProcessAudioFile(.aac) error = %v", err)
	}
}

func TestExtractPictureFromMP3(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mp3_pic_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	expectedImg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x12, 0x34}

	// Build ID3v2.3 tag with APIC frame
	var frameBuf bytes.Buffer
	frameBuf.WriteByte(0) // Latin-1 encoding
	frameBuf.WriteString("image/jpeg\x00")
	frameBuf.WriteByte(3) // Front cover
	frameBuf.WriteString("Cover\x00")
	frameBuf.Write(expectedImg)

	frameData := frameBuf.Bytes()
	tagPayload := make([]byte, 0, 10+len(frameData))
	tagPayload = append(tagPayload, "APIC"...)
	tagPayload = binary.BigEndian.AppendUint32(tagPayload, uint32(len(frameData)))
	tagPayload = append(tagPayload, 0, 0)
	tagPayload = append(tagPayload, frameData...)

	tagSize := len(tagPayload)
	fullMP3 := make([]byte, 0, 10+tagSize+len("audio stream data here"))
	fullMP3 = append(fullMP3, "ID3"...)
	fullMP3 = append(fullMP3,
		3, 0, 0,
		byte((tagSize>>21)&0x7F),
		byte((tagSize>>14)&0x7F),
		byte((tagSize>>7)&0x7F),
		byte(tagSize&0x7F),
	)
	fullMP3 = append(fullMP3, tagPayload...)
	fullMP3 = append(fullMP3, []byte("audio stream data here")...)

	mp3File := filepath.Join(tmpDir, "tagged.mp3")
	if err := os.WriteFile(mp3File, fullMP3, 0o644); err != nil {
		t.Fatal(err)
	}

	extracted, err := ExtractPictureFromMP3(mp3File)
	if err != nil {
		t.Fatalf("ExtractPictureFromMP3() error = %v", err)
	}
	if !bytes.Equal(extracted, expectedImg) {
		t.Errorf("got %v, want %v", extracted, expectedImg)
	}

	// Test non-ID3 file
	plainFile := filepath.Join(tmpDir, "plain.mp3")
	if err := os.WriteFile(plainFile, []byte("no id3 tag here"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = ExtractPictureFromMP3(plainFile)
	if err == nil {
		t.Error("expected error for non-ID3 file")
	}
}

func TestExtractPictureFromM4A(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "m4a_pic_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	expectedImg := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	// Build MP4 atom hierarchy: moov -> udta -> meta -> ilst -> covr -> data
	dataBox := makeBox("data", append(make([]byte, 8), expectedImg...))
	covrBox := makeBox("covr", dataBox)
	ilstBox := makeBox("ilst", covrBox)
	metaBox := makeBox("meta", append(make([]byte, 4), ilstBox...)) // 4 bytes version/flags
	udtaBox := makeBox("udta", metaBox)
	moovBox := makeBox("moov", udtaBox)

	m4aFile := filepath.Join(tmpDir, "tagged.m4a")
	if err := os.WriteFile(m4aFile, moovBox, 0o644); err != nil {
		t.Fatal(err)
	}

	extracted, err := ExtractPictureFromM4A(m4aFile)
	if err != nil {
		t.Fatalf("ExtractPictureFromM4A() error = %v", err)
	}
	if !bytes.Equal(extracted, expectedImg) {
		t.Errorf("got %v, want %v", extracted, expectedImg)
	}

	// Test file without cover
	noCoverFile := filepath.Join(tmpDir, "plain.m4a")
	if err := os.WriteFile(noCoverFile, []byte("plain m4a data"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = ExtractPictureFromM4A(noCoverFile)
	if err == nil {
		t.Error("expected error for M4A without cover")
	}
}

func makeBox(boxType string, content []byte) []byte {
	size := 8 + len(content)
	box := make([]byte, 0, size)
	box = binary.BigEndian.AppendUint32(box, uint32(size))
	box = append(box, boxType...)
	box = append(box, content...)
	return box
}

func TestProcessMP3WithID3_EmbedAndStrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mp3_embed_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldImg := []byte("old big cover")
	newImg := []byte("new small resized cover")

	// Construct MP3 with an old APIC and a TIT2 text frame
	var frameBuf bytes.Buffer
	frameBuf.WriteByte(0)
	frameBuf.WriteString("image/jpeg\x00")
	frameBuf.WriteByte(3)
	frameBuf.WriteString("Cover\x00")
	frameBuf.Write(oldImg)
	apicData := frameBuf.Bytes()

	tagPayload := make([]byte, 0, 100)
	// Add TIT2 frame ("Test Title")
	titleData := []byte("\x00Test Title")
	tagPayload = append(tagPayload, "TIT2"...)
	tagPayload = binary.BigEndian.AppendUint32(tagPayload, uint32(len(titleData)))
	tagPayload = append(tagPayload, 0, 0)
	tagPayload = append(tagPayload, titleData...)

	// Add APIC frame
	tagPayload = append(tagPayload, "APIC"...)
	tagPayload = binary.BigEndian.AppendUint32(tagPayload, uint32(len(apicData)))
	tagPayload = append(tagPayload, 0, 0)
	tagPayload = append(tagPayload, apicData...)

	tagSize := len(tagPayload)
	audioPayload := []byte("audio stream data frames 1234567890")
	fullMP3 := make([]byte, 0, 10+tagSize+len(audioPayload))
	fullMP3 = append(fullMP3, 'I', 'D', '3', 3, 0, 0)
	fullMP3 = append(fullMP3, encodeSynchsafe(tagSize)...)
	fullMP3 = append(fullMP3, tagPayload...)
	fullMP3 = append(fullMP3, audioPayload...)

	srcFile := filepath.Join(tmpDir, "source.mp3")
	if err := os.WriteFile(srcFile, fullMP3, 0o644); err != nil {
		t.Fatal(err)
	}

	// 1. Process with new cover (embedCover)
	destWithCover := filepath.Join(tmpDir, "dest_cover.mp3")
	if err := ProcessMP3WithID3(srcFile, destWithCover, newImg); err != nil {
		t.Fatalf("ProcessMP3WithID3() error = %v", err)
	}
	extractedNew, err := ExtractPictureFromMP3(destWithCover)
	if err != nil {
		t.Fatalf("ExtractPictureFromMP3(destWithCover) error = %v", err)
	}
	if !bytes.Equal(extractedNew, newImg) {
		t.Errorf("got %s, want %s", string(extractedNew), string(newImg))
	}

	// 2. Process without cover (strip APIC, retain other tags)
	destStripped := filepath.Join(tmpDir, "dest_stripped.mp3")
	if err := ProcessMP3WithID3(srcFile, destStripped, nil); err != nil {
		t.Fatalf("ProcessMP3WithID3() strip error = %v", err)
	}
	_, err = ExtractPictureFromMP3(destStripped)
	if err == nil {
		t.Error("expected error extracting picture from stripped MP3, but found one")
	}

	// 3. ProcessAudioFile routing test with embedCover
	destAudioDir := filepath.Join(tmpDir, "audio_dest")
	if err := os.MkdirAll(destAudioDir, 0o755); err != nil {
		t.Fatal(err)
	}
	extractedSrcPic, err := ProcessAudioFile(srcFile, tmpDir, destAudioDir, newImg)
	if err != nil {
		t.Fatalf("ProcessAudioFile with embedCover error: %v", err)
	}
	if !bytes.Equal(extractedSrcPic, oldImg) {
		t.Errorf("expected extracted source pic %q, got %q", oldImg, extractedSrcPic)
	}
}

func TestExtractPictureFromAudioFile(t *testing.T) {
	testDataDir := "../../test_data"
	testFlac := filepath.Join(testDataDir, "01 - test.flac")
	pic, err := ExtractPictureFromAudioFile(testFlac)
	if err != nil {
		t.Fatalf("ExtractPictureFromAudioFile(flac) error = %v", err)
	}
	if len(pic) == 0 {
		t.Error("expected picture data from test flac")
	}

	nonExistent, err := ExtractPictureFromAudioFile("unknown.xyz")
	if err != nil {
		t.Errorf("unexpected error for unsupported ext: %v", err)
	}
	if nonExistent != nil {
		t.Errorf("expected nil for unsupported ext, got %v", nonExistent)
	}
}
