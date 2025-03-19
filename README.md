# Album Picker

A command-line tool for randomly selecting and copying FLAC albums while optimizing cover art.

## Introduction

I planned to write this utility with the help of artificial intelligence, as part of the “vibe coding” trend. But things didn't go as I expected, so I had to write it myself in one evening.

I have a collection of audio CDs that I digitized and keep on the server. All the tags are written using MusicBrainz Picard, the covers are embedded, and they are saved in album.jpg or album.png files. Everything is nicely organized in directories like `/path/to/music/AlbumArtist/Year - AlbumName/TrackNumber - TrackName.flac`.

Recently, I got an iPod Video 80gb with RockBox on board and realized that it's not a good idea to just copy the library to it. First of all, along with the tracks there are logs and cue files after digitization, scanned printing, which are useless on a portable player. Secondly, even the covers are of a large size, which the player cannot load into memory to display. Thirdly, and probably the most important thing is that a large number of albums does not allow me to focus on listening to music, and choosing what to listen to, I just start to procrastinate. That's why the idea of this utility was born: copy 10 random albums from the collection, remove unnecessary embedded covers, which will save some space (they are not supported on RockBox anyway) and create a 240 by 240 pixel cover.jpg next to the tracks. 

I hope you will find this utility useful too.

## Features

- Randomly select and copy FLAC albums from a source directory
- Process and optimize album cover art
- Configurable via YAML file or command-line flags
- Support for multiple cover art formats (jpg, png) and names (see Configuration section)

## Installation

```bash
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
```
I recommend to set the `source` and `destination` in the config file.

For example:
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
```

### Pick Random Albums

```bash
albumpicker pick
```
You got in `/path/to/picked` 10 random albums with optimized cover art with folder structure that looks like your music library inside `/path/to/music`

### Copy Single Album

```bash
albumpicker copy /path/to/music/Artist/Album
```
or, if you already inside `/path/to/music/Artist/Album` just run:

```bash
albumpicker copy .
```

### Command Line Options

#### Global flags
- `-c, --config`: Path to config file (default: `~/.config/albumpicker/config.yaml`)
- `-s, --source`: Source directory containing FLAC albums
- `-d, --destination`: Destination directory for copied albums
- `--height`: Cover image height in pixels (default: 240)
- `--cover-name`: Output cover file name

#### `pick` command flags
- `-n, --count`: Number of albums to select (default: 10)
- `--wipe`: Wipe destination directory before copying (pick command only)

## Development

### Building

```bash
go build -o albumpicker
```

### Testing

```bash
go test ./...
```