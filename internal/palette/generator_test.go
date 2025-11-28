package palette

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/james-see/swatchify/internal/utils"
)

func TestGeneratePNG_Basic(t *testing.T) {
	colors := []utils.ColorResult{
		{Hex: "#FF0000", Percentage: 50, RGB: utils.RGB{R: 255, G: 0, B: 0}},
		{Hex: "#00FF00", Percentage: 30, RGB: utils.RGB{R: 0, G: 255, B: 0}},
		{Hex: "#0000FF", Percentage: 20, RGB: utils.RGB{R: 0, G: 0, B: 255}},
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "palette.png")

	err := GeneratePNG(colors, outPath, DefaultOptions())
	if err != nil {
		t.Fatalf("GeneratePNG failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("Palette PNG was not created")
	}

	// Verify it's a valid PNG
	f, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("Failed to open palette: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("Failed to decode palette PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 1000 || bounds.Dy() != 200 {
		t.Errorf("Palette size = %dx%d, want 1000x200", bounds.Dx(), bounds.Dy())
	}
}

func TestGeneratePNG_CustomSize(t *testing.T) {
	colors := []utils.ColorResult{
		{Hex: "#FF0000", Percentage: 100, RGB: utils.RGB{R: 255, G: 0, B: 0}},
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "palette.png")

	opts := Options{Width: 500, Height: 100}
	err := GeneratePNG(colors, outPath, opts)
	if err != nil {
		t.Fatalf("GeneratePNG failed: %v", err)
	}

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("Failed to open palette: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("Failed to decode palette PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 500 || bounds.Dy() != 100 {
		t.Errorf("Palette size = %dx%d, want 500x100", bounds.Dx(), bounds.Dy())
	}
}

func TestGeneratePNG_EmptyColors(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "palette.png")

	err := GeneratePNG([]utils.ColorResult{}, outPath, DefaultOptions())
	if err == nil {
		t.Error("GeneratePNG should error with empty colors")
	}
}

func TestGeneratePNG_SingleColor(t *testing.T) {
	colors := []utils.ColorResult{
		{Hex: "#123456", Percentage: 100, RGB: utils.RGB{R: 18, G: 52, B: 86}},
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "palette.png")

	err := GeneratePNG(colors, outPath, DefaultOptions())
	if err != nil {
		t.Fatalf("GeneratePNG failed: %v", err)
	}

	// Verify the entire image is the single color
	f, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("Failed to open palette: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("Failed to decode palette PNG: %v", err)
	}

	// Check a pixel in the middle
	r, g, b, _ := img.At(500, 100).RGBA()
	// RGBA returns 16-bit values, shift to 8-bit
	if uint8(r>>8) != 18 || uint8(g>>8) != 52 || uint8(b>>8) != 86 {
		t.Errorf("Pixel color = (%d, %d, %d), want (18, 52, 86)", r>>8, g>>8, b>>8)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Width != 1000 {
		t.Errorf("DefaultOptions Width = %d, want 1000", opts.Width)
	}
	if opts.Height != 200 {
		t.Errorf("DefaultOptions Height = %d, want 200", opts.Height)
	}
}

func TestGetContrastingColor(t *testing.T) {
	// Dark colors should get white text
	dark := utils.RGB{R: 0, G: 0, B: 0}
	if getContrastingColor(dark) != white {
		t.Error("Dark color should get white contrast")
	}

	// Light colors should get black text
	light := utils.RGB{R: 255, G: 255, B: 255}
	if getContrastingColor(light) != black {
		t.Error("Light color should get black contrast")
	}
}

var white = getContrastingColor(utils.RGB{R: 0, G: 0, B: 0})
var black = getContrastingColor(utils.RGB{R: 255, G: 255, B: 255})

