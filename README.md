# Album Picker

A command-line tool for randomly selecting and copying music albums (FLAC, MP3, ALAC, AAC, M4A) while optimizing cover art for portable players like iPod (Rockbox & Apple OS).

## Introduction

I planned to write this utility with the help of artificial intelligence as part of the “vibe coding” trend. But things didn't go as I expected, so I had to write it myself in one evening.

I have a collection of audio CDs that I digitized and keep on the server. All the tags are written using MusicBrainz Picard, the covers are embedded, and they are saved in album.jpg or album.png files. Everything is nicely organized in directories like `/path/to/music/AlbumArtist/Year - AlbumName/TrackNumber - TrackName.flac`.

Recently, I got an iPod Video 80GB with Rockbox on board and realized that it's not a good idea to just copy the library to it:
1. Along with the tracks, there are logs, cue files, and scanned printing that are useless on a portable player.
2. Embedded covers are often huge (5–10 MB), which old portable player CPUs and 32/64 MB RAM struggle to decode, causing noticeable UI lag.
3. Too many albums make it hard to focus on listening — you spend more time choosing than listening.

That's why this utility was created: pick random albums from the collection, extract and optimize the cover to a small `cover.jpg` (e.g. 240x240 px) for Rockbox, and strip heavy artwork from audio files.

For **dual-boot iPods** (Rockbox + original Apple OS), it also supports re-embedding the small resized cover (~20–30 KB) into the audio tags with `--embed-cover`, allowing iTunes to generate its `.ithmb` database for Apple OS while keeping Rockbox fast.

## Features

- **Random Album Selection**: Pick and copy *N* random albums with directory hierarchy preserved.
- **Multi-Format Audio Support**: Works with **FLAC**, **MP3**, **ALAC**, **AAC**, and **M4A** containers.
- **Smart Cover Extraction & Optimization**: Automatically detects external artwork or extracts embedded covers from audio files, resizing to baseline JPEG (`cover.jpg`).
- **Artwork Stripping**: Strips heavy embedded artwork from destination audio files by default to save storage and player memory.
- **Dual-Boot iPod / iTunes Mode**: `--embed-cover` embeds the resized artwork into track metadata (FLAC PICTURE, MP3 ID3v2.3 APIC) for iTunes synchronization.
- **Single Album / Artist Copy**: `albumpicker copy` command to copy a specific album or all albums of an artist.
- **Lightweight**: Pure Go standard library CLI (`flag.FlagSet`), zero heavy CLI frameworks.

## Format Support

- [x] FLAC (`.flac`)
- [x] MP3 (`.mp3` with ID3v2 APIC extraction & re-tagging)
- [x] ALAC / AAC (`.m4a`, `.alac`, `.aac` QuickTime container)
- [ ] APE (`.ape`)
- [ ] WavPack (`.wv`)

## Installation

### Binary Releases

Download the appropriate binary for your platform from [the latest release](https://github.com/nerten/albumpicker/releases/latest).

### Go

```sh
go install github.com/nerten/albumpicker@latest
```

## Usage

### Configuration

The config file is automatically created at `~/.config/albumpicker/config.yaml` with default values:

```yaml
source: ""
destination: ""
albums_count: 10
cover_filenames:
  - album.jpg
  - album.png
  - cover.jpg
  - cover.png
output_cover_filename: cover.jpg
cover_height: 240
embed_cover: false
```

Configuration fields:
- `source`: Path to music library root.
- `destination`: Destination directory (e.g., iPod mount path or playlist folder).
- `albums_count`: Default number of albums to pick (default: `10`).
- `cover_filenames`: List of filenames to look for external cover art.
- `output_cover_filename`: Target cover filename in the destination directory (default: `cover.jpg`).
- `cover_height`: Target cover height in pixels (default: `240`).
- `embed_cover`: Whether to embed resized cover art into audio files (default: `false`).

Example:
```yaml
source: "/path/to/music"
destination: "/path/to/picked"
albums_count: 10
cover_filenames:
  - album.jpg
  - album.png
  - cover.jpg
  - cover.png
output_cover_filename: cover.jpg
cover_height: 240
embed_cover: false
```

### Pick Random Albums

```sh
albumpicker pick
```
Copies 10 random albums from `source` to `destination`, creates `cover.jpg` (240px height), and strips heavy embedded tags.

To embed the resized cover into audio files for iTunes / dual-boot:
```sh
albumpicker pick --embed-cover
```

To wipe the destination directory before picking:
```sh
albumpicker pick --wipe
```

### Copy Specific Albums

**Copy a single album:**
```sh
albumpicker copy /path/to/music/Artist/Album
```
or run directly from inside the album folder:
```sh
albumpicker copy .
```

**Copy all albums of an artist:**
```sh
albumpicker copy /path/to/music/Artist
```

**Copy with cover embedding:**
```sh
albumpicker copy --embed-cover /path/to/music/Artist/Album
```

### Command Line Options

#### Common Flags (available on root, `pick`, and `copy`)
- `-c, --config string`: Path to config file (default: `~/.config/albumpicker/config.yaml`)
- `-s, --source string`: Source directory containing music albums
- `-d, --destination string`: Destination directory for copied albums
- `--height int`: Cover image height in pixels (default: `240`)
- `--cover-name string`: Output cover file name (default: `cover.jpg`)
- `--embed-cover`: Embed resized cover art into audio files (for dual-boot/iTunes)

#### `pick` Command Flags
- `-n, --count int`: Number of albums to select (default: `10`)
- `--wipe`: Wipe destination directory before copying (destructive action!)

## Development

### Building

```sh
make build
```

### Testing & Linting

```sh
make lint
make test
```

### Update dependencies

```sh
make update
```
