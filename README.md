# Swatchify

A fast, cross-platform CLI tool that extracts dominant colors from images using k-means clustering.

## Installation

### Go Install (requires Go 1.21+)

```bash
go install github.com/james-see/swatchify@latest
```

### Homebrew (macOS/Linux)

```bash
brew install james-see/tap/swatchify
```

### Download Binary

Download the latest release for your platform from the [Releases page](https://github.com/james-see/swatchify/releases).

### Build from Source

```bash
git clone https://github.com/james-see/swatchify
cd swatchify
go build -o swatchify .
```

## Usage

```bash
swatchify <image> [flags]
```

### Examples

```bash
# Extract 5 dominant colors (default)
swatchify photo.jpg

# Extract 8 colors
swatchify logo.png -n 8

# Output as JSON
swatchify image.png --json

# Generate a palette PNG with hex labels
swatchify brand.png --png palette.png

# Generate palette and open it
swatchify brand.png --png palette.png --show

# Exclude white and black colors
swatchify mood.png --exclude-white --exclude-black

# Higher quality (slower, more accurate)
swatchify photo.jpg --quality 100

# Pipe JSON to file
swatchify image.jpg --json > palette.json
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--colors` | `-n` | 5 | Number of dominant colors to extract |
| `--json` | | false | Output in JSON format |
| `--png` | | | Generate palette PNG at specified path |
| `--exclude-white` | | false | Exclude colors close to white |
| `--exclude-black` | | false | Exclude colors close to black |
| `--min-contrast` | | 0 | Minimum color distance between palette colors |
| `--show` | | false | Open generated palette image after creation |
| `--quality` | | 50 | Downscale quality 0-100 (higher = more accurate, slower) |
| `--width` | | 1000 | Palette image width in pixels |
| `--height` | | 200 | Palette image height in pixels |

### Supported Formats

- JPEG/JPG
- PNG
- WebP
- GIF (first frame)
- BMP
- TIFF

## Output Formats

### Text (default)

```
#112233
#AABBCC
#FFEEDD
#998877
#341212
```

### JSON

```json
{
  "image": "input.jpg",
  "colors": [
    {"hex": "#112233", "percentage": 34.5},
    {"hex": "#AABBCC", "percentage": 21.0},
    {"hex": "#FFEEDD", "percentage": 18.2},
    {"hex": "#998877", "percentage": 15.1},
    {"hex": "#341212", "percentage": 11.2}
  ]
}
```

### Palette PNG

Generates a horizontal color strip with blocks sized proportionally to color prevalence. Each block displays its hex code with automatic contrast text (white on dark, black on light).

## Performance

- Target execution time: < 300ms for typical images
- Memory footprint: < 100MB
- Images are automatically downscaled for processing speed

## Run as API Server

Start swatchify as an HTTP server for REST API access:

```bash
swatchify serve --port 8080
```

### Endpoints

**POST /extract** — Extract colors from uploaded image

```bash
# Basic usage
curl -X POST -F "image=@photo.jpg" http://localhost:8080/extract

# With options
curl -X POST \
  -F "image=@photo.jpg" \
  -F "colors=8" \
  -F "quality=100" \
  -F "exclude_white=true" \
  http://localhost:8080/extract
```

Response:
```json
{
  "success": true,
  "colors": [
    {"hex": "#112233", "percentage": 34.5},
    {"hex": "#AABBCC", "percentage": 21.0}
  ]
}
```

**GET /health** — Health check

```bash
curl http://localhost:8080/health
# {"status": "ok"}
```

### Form Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `image` | file | required | Image file to analyze |
| `colors` | int | 5 | Number of colors to extract |
| `quality` | int | 50 | Quality 0-100 |
| `exclude_white` | bool | false | Exclude near-white colors |
| `exclude_black` | bool | false | Exclude near-black colors |
| `min_contrast` | float | 0 | Minimum color distance |

## Use as a Library

Swatchify can be imported and used in your Go code:

```go
package main

import (
    "fmt"
    "log"

    "github.com/james-see/swatchify/pkg/swatchify"
)

func main() {
    // Extract with default options (5 colors)
    colors, err := swatchify.ExtractFromFile("photo.jpg", nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, c := range colors {
        fmt.Printf("%s (%.1f%%)\n", c.Hex, c.Percentage)
    }

    // Custom options
    opts := &swatchify.Options{
        NumColors:    8,
        Quality:      100,
        ExcludeWhite: true,
        ExcludeBlack: true,
        MinContrast:  30,
    }
    colors, _ = swatchify.ExtractFromFile("photo.jpg", opts)

    // Generate palette image
    swatchify.GeneratePalette(colors, "palette.png", nil)
}
```

### Library API

```go
// Extract from file path
colors, err := swatchify.ExtractFromFile(path string, opts *Options) ([]Color, error)

// Extract from image.Image
colors, err := swatchify.ExtractFromImage(img image.Image, opts *Options) ([]Color, error)

// Generate palette PNG
err := swatchify.GeneratePalette(colors []Color, outputPath string, opts *PaletteOptions) error
```

## How It Works

1. Load and decode the input image
2. Downscale large images based on quality setting
3. Extract pixel data as RGB vectors
4. Run k-means++ clustering to find dominant colors
5. Sort clusters by population percentage
6. Apply any filters (white/black exclusion, min contrast)
7. Output results in requested format

## License

MIT

## Goose / MCP Extension

swatchify ships as an MCP server, so AI agents like [goose](https://goose-docs.ai/) can extract dominant colors from images as tools.

### Build the MCP server

```bash
go install github.com/james-see/swatchify/cmd/mcp@latest
```

The binary is installed as `mcp` in your `$(go env GOPATH)/bin/`. Rename it to `swatchify-mcp`:

```bash
mv $(go env GOPATH)/bin/mcp $(go env GOPATH)/bin/swatchify-mcp
```

### Add to goose

Add this to `~/.config/goose/config.yaml`:

```yaml
extensions:
  swatchify:
    name: Swatchify
    cmd: swatchify-mcp
    args: []
    enabled: true
    type: stdio
    timeout: 60
```

### Available tools

| Tool | Description |
|------|-------------|
| `extract_colors` | Extract dominant colors from an image (hex, percentage, RGB) |
| `generate_palette` | Generate a color palette PNG from an image |
