package utils

import (
	"fmt"
	"image/color"
	"math"
)

// RGB represents a color in RGB space
type RGB struct {
	R, G, B uint8
}

// ColorResult represents a dominant color with its percentage
type ColorResult struct {
	Hex        string  `json:"hex"`
	Percentage float64 `json:"percentage"`
	RGB        RGB     `json:"-"`
}

// RGBToHex converts RGB values to a HEX string
func RGBToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// HexToRGB converts a HEX string to RGB values
func HexToRGB(hex string) (RGB, error) {
	var r, g, b uint8
	if len(hex) == 7 && hex[0] == '#' {
		_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return RGB{}, err
		}
	} else if len(hex) == 6 {
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return RGB{}, err
		}
	}
	return RGB{R: r, G: g, B: b}, nil
}

// ColorToRGB converts a color.Color to RGB
func ColorToRGB(c color.Color) RGB {
	r, g, b, _ := c.RGBA()
	return RGB{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
	}
}

// RGBToColor converts RGB to color.Color
func (rgb RGB) ToColor() color.Color {
	return color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

// EuclideanDistance calculates the Euclidean distance between two RGB colors
func EuclideanDistance(c1, c2 RGB) float64 {
	dr := float64(c1.R) - float64(c2.R)
	dg := float64(c1.G) - float64(c2.G)
	db := float64(c1.B) - float64(c2.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// IsNearWhite checks if a color is close to white
func IsNearWhite(c RGB, threshold float64) bool {
	white := RGB{R: 255, G: 255, B: 255}
	return EuclideanDistance(c, white) < threshold
}

// IsNearBlack checks if a color is close to black
func IsNearBlack(c RGB, threshold float64) bool {
	black := RGB{R: 0, G: 0, B: 0}
	return EuclideanDistance(c, black) < threshold
}

// FilterColors filters colors based on exclusion options
func FilterColors(colors []ColorResult, excludeWhite, excludeBlack bool, threshold float64) []ColorResult {
	if !excludeWhite && !excludeBlack {
		return colors
	}

	filtered := make([]ColorResult, 0, len(colors))
	for _, c := range colors {
		if excludeWhite && IsNearWhite(c.RGB, threshold) {
			continue
		}
		if excludeBlack && IsNearBlack(c.RGB, threshold) {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

// FilterByMinContrast removes colors that are too similar to each other
func FilterByMinContrast(colors []ColorResult, minDistance float64) []ColorResult {
	if minDistance <= 0 || len(colors) == 0 {
		return colors
	}

	filtered := []ColorResult{colors[0]}
	for i := 1; i < len(colors); i++ {
		tooClose := false
		for _, kept := range filtered {
			if EuclideanDistance(colors[i].RGB, kept.RGB) < minDistance {
				tooClose = true
				break
			}
		}
		if !tooClose {
			filtered = append(filtered, colors[i])
		}
	}
	return filtered
}

// Luminance calculates relative luminance of a color
func Luminance(c RGB) float64 {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	// sRGB to linear
	if r <= 0.03928 {
		r = r / 12.92
	} else {
		r = math.Pow((r+0.055)/1.055, 2.4)
	}
	if g <= 0.03928 {
		g = g / 12.92
	} else {
		g = math.Pow((g+0.055)/1.055, 2.4)
	}
	if b <= 0.03928 {
		b = b / 12.92
	} else {
		b = math.Pow((b+0.055)/1.055, 2.4)
	}

	return 0.2126*r + 0.7152*g + 0.0722*b
}

