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
