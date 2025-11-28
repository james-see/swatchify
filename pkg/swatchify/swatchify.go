// Package swatchify extracts dominant colors from images using k-means clustering.
//
// Example usage:
//
//	colors, err := swatchify.ExtractFromFile("image.jpg", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, c := range colors {
//	    fmt.Printf("%s (%.1f%%)\n", c.Hex, c.Percentage)
//	}
package swatchify

import (
	"image"

	"github.com/james-see/swatchify/internal/cluster"
	"github.com/james-see/swatchify/internal/imageio"
	"github.com/james-see/swatchify/internal/palette"
	"github.com/james-see/swatchify/internal/utils"
)

// Color represents a dominant color extracted from an image.
type Color struct {
	Hex        string  `json:"hex"`
	Percentage float64 `json:"percentage"`
	R, G, B    uint8   `json:"-"`
}

// Options configures color extraction behavior.
type Options struct {
	// NumColors is the number of dominant colors to extract (default: 5)
	NumColors int

	// Quality controls downscale size for speed/accuracy tradeoff (0-100, default: 50)
	Quality int

	// ExcludeWhite filters out colors close to white
	ExcludeWhite bool

	// ExcludeBlack filters out colors close to black
	ExcludeBlack bool

	// MinContrast is the minimum color distance between palette colors
	MinContrast float64

	// ColorThreshold is the distance threshold for white/black detection (default: 30)
	ColorThreshold float64
}

// DefaultOptions returns sensible default extraction options.
func DefaultOptions() *Options {
	return &Options{
		NumColors:      5,
		Quality:        50,
		ColorThreshold: 30,
	}
}

// ExtractFromFile extracts dominant colors from an image file.
// Supported formats: JPEG, PNG, WebP, GIF, BMP, TIFF.
func ExtractFromFile(path string, opts *Options) ([]Color, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	loadOpts := imageio.LoadOptions{
		MaxDimension: imageio.QualityToMaxDimension(opts.Quality),
	}

	img, err := imageio.LoadImage(path, loadOpts)
	if err != nil {
		return nil, err
	}

	return ExtractFromImage(img, opts)
}

// ExtractFromImage extracts dominant colors from an image.Image.
func ExtractFromImage(img image.Image, opts *Options) ([]Color, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	pixels := imageio.GetPixels(img)
	if len(pixels) == 0 {
		return nil, nil
	}

	// Request more colors if filtering
	requestColors := opts.NumColors
	if opts.ExcludeWhite || opts.ExcludeBlack || opts.MinContrast > 0 {
		requestColors = opts.NumColors * 2
		if requestColors > 50 {
			requestColors = 50
		}
	}

	// Run clustering
	results := cluster.KMeans(pixels, requestColors)

	// Apply filters
	threshold := opts.ColorThreshold
	if threshold == 0 {
		threshold = 30
	}
	results = utils.FilterColors(results, opts.ExcludeWhite, opts.ExcludeBlack, threshold)
	results = utils.FilterByMinContrast(results, opts.MinContrast)

	// Limit to requested count
	if len(results) > opts.NumColors {
		results = results[:opts.NumColors]
	}

	// Convert to public type
	colors := make([]Color, len(results))
	for i, r := range results {
		colors[i] = Color{
			Hex:        r.Hex,
			Percentage: r.Percentage,
			R:          r.RGB.R,
			G:          r.RGB.G,
			B:          r.RGB.B,
		}
	}

	return colors, nil
}

// PaletteOptions configures palette image generation.
type PaletteOptions struct {
	Width  int
	Height int
}

// DefaultPaletteOptions returns default palette dimensions.
func DefaultPaletteOptions() *PaletteOptions {
	return &PaletteOptions{
		Width:  1000,
		Height: 200,
	}
}

// GeneratePalette creates a palette PNG image from colors.
func GeneratePalette(colors []Color, outputPath string, opts *PaletteOptions) error {
	if opts == nil {
		opts = DefaultPaletteOptions()
	}

	// Convert to internal type
	internal := make([]utils.ColorResult, len(colors))
	for i, c := range colors {
		internal[i] = utils.ColorResult{
			Hex:        c.Hex,
			Percentage: c.Percentage,
			RGB:        utils.RGB{R: c.R, G: c.G, B: c.B},
		}
	}

	return palette.GeneratePNG(internal, outputPath, palette.Options{
		Width:  opts.Width,
		Height: opts.Height,
	})
}

