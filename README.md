# Album Picker

A command-line tool for randomly selecting and copying FLAC albums while optimizing cover art.

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