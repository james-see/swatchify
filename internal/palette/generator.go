package palette

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"runtime"

	"github.com/james-see/swatchify/internal/utils"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Options contains options for palette generation
type Options struct {
	Width  int
	Height int
}

// DefaultOptions returns default palette options
func DefaultOptions() Options {
	return Options{
		Width:  1000,
		Height: 200,
	}
}

// GeneratePNG creates a horizontal color strip palette image with hex labels
func GeneratePNG(colors []utils.ColorResult, outputPath string, opts Options) error {
	if len(colors) == 0 {
		return fmt.Errorf("no colors to generate palette from")
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// Calculate total percentage for proportional widths
	totalPct := 0.0
	for _, c := range colors {
		totalPct += c.Percentage
	}

	// If percentages don't sum to useful value, use equal widths
	if totalPct < 1 {
		totalPct = float64(len(colors))
		for i := range colors {
			colors[i].Percentage = 1
		}
	}

	// Track block positions for text drawing
	type block struct {
		x, width int
		hex      string
		rgb      utils.RGB
	}
	blocks := make([]block, 0, len(colors))

	// Draw color blocks
	x := 0
	for i, c := range colors {
		// Calculate block width proportionally
		var blockWidth int
		if i == len(colors)-1 {
			// Last block takes remaining space to avoid rounding gaps
			blockWidth = opts.Width - x
		} else {
			blockWidth = int(float64(opts.Width) * (c.Percentage / totalPct))
		}

		// Convert to color.RGBA
		clr := color.RGBA{R: c.RGB.R, G: c.RGB.G, B: c.RGB.B, A: 255}

		// Fill the block
		for py := 0; py < opts.Height; py++ {
			for px := x; px < x+blockWidth && px < opts.Width; px++ {
				img.Set(px, py, clr)
			}
		}

		blocks = append(blocks, block{x: x, width: blockWidth, hex: c.Hex, rgb: c.RGB})
		x += blockWidth
	}

	// Draw hex labels on each block
	for _, b := range blocks {
		// Choose contrasting text color (white or black)
		textColor := getContrastingColor(b.rgb)

		// Calculate text position (centered in block)
		textWidth := len(b.hex) * 7 // basicfont is ~7px per char
		textX := b.x + (b.width-textWidth)/2
		textY := opts.Height / 2

		// Ensure text doesn't go off the left edge
		if textX < b.x+4 {
			textX = b.x + 4
		}

		drawLabel(img, textX, textY, b.hex, textColor)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Encode as PNG
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// getContrastingColor returns white or black based on background luminance
func getContrastingColor(bg utils.RGB) color.Color {
	// Calculate relative luminance
	luminance := utils.Luminance(bg)
	if luminance > 0.5 {
		return color.Black
	}
	return color.White
}

// drawLabel draws text on the image at the specified position
func drawLabel(img draw.Image, x, y int, label string, col color.Color) {
	point := fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

// OpenImage opens the generated image with the system default viewer
func OpenImage(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Start()
}

