package imageio

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp" // WebP support
)

// LoadOptions contains options for loading images
type LoadOptions struct {
	MaxDimension int // Maximum dimension for downscaling
}

// DefaultLoadOptions returns default load options
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{
		MaxDimension: 300,
	}
}

// QualityToMaxDimension converts quality (0-100) to max dimension (100-500px)
func QualityToMaxDimension(quality int) int {
	if quality < 0 {
		quality = 0
	}
	if quality > 100 {
		quality = 100
	}
	// Map 0-100 to 100-500
	return 100 + (quality * 4)
}

// LoadImage loads an image from the given path and optionally downscales it
func LoadImage(path string, opts LoadOptions) (image.Image, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(path))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".tiff": true,
		".tif":  true,
	}
	if !validExts[ext] {
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	// Open and decode image
	img, err := imaging.Open(path, imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	// Downscale if needed
	if opts.MaxDimension > 0 {
		img = downscale(img, opts.MaxDimension)
	}

	return img, nil
}

// downscale reduces image size while maintaining aspect ratio
func downscale(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check if downscaling is needed
	if width <= maxDim && height <= maxDim {
		return img
	}

	// Calculate new dimensions
	var newWidth, newHeight int
	if width > height {
		newWidth = maxDim
		newHeight = int(float64(height) * float64(maxDim) / float64(width))
	} else {
		newHeight = maxDim
		newWidth = int(float64(width) * float64(maxDim) / float64(height))
	}

	// Use fast Lanczos resizing
	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}

// GetPixels extracts all pixels from an image as RGB values
func GetPixels(img image.Image) [][3]float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	pixels := make([][3]float64, 0, width*height)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// Skip fully transparent pixels
			if a == 0 {
				continue
			}
			pixels = append(pixels, [3]float64{
				float64(r >> 8),
				float64(g >> 8),
				float64(b >> 8),
			})
		}
	}

	return pixels
}

