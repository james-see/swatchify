package swatchify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFromFile(t *testing.T) {
	imagePath := filepath.Join("..", "..", "examples", "image1.jpg")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Example image not found")
	}

	colors, err := ExtractFromFile(imagePath, nil)
	if err != nil {
		t.Fatalf("ExtractFromFile failed: %v", err)
	}

	if len(colors) == 0 {
		t.Fatal("No colors extracted")
	}

	if len(colors) > 5 {
		t.Errorf("Expected max 5 colors with defaults, got %d", len(colors))
	}

	for _, c := range colors {
		if len(c.Hex) != 7 || c.Hex[0] != '#' {
			t.Errorf("Invalid hex format: %s", c.Hex)
		}
	}
}

func TestExtractFromFile_CustomOptions(t *testing.T) {
	imagePath := filepath.Join("..", "..", "examples", "image1.jpg")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Example image not found")
	}

	opts := &Options{
		NumColors:    8,
		Quality:      100,
		ExcludeWhite: true,
		ExcludeBlack: true,
	}

	colors, err := ExtractFromFile(imagePath, opts)
	if err != nil {
		t.Fatalf("ExtractFromFile failed: %v", err)
	}

	if len(colors) > 8 {
		t.Errorf("Expected max 8 colors, got %d", len(colors))
	}
}

func TestGeneratePalette(t *testing.T) {
	colors := []Color{
		{Hex: "#FF0000", Percentage: 50, R: 255, G: 0, B: 0},
		{Hex: "#00FF00", Percentage: 30, R: 0, G: 255, B: 0},
		{Hex: "#0000FF", Percentage: 20, R: 0, G: 0, B: 255},
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "palette.png")

	err := GeneratePalette(colors, outPath, nil)
	if err != nil {
		t.Fatalf("GeneratePalette failed: %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("Palette file was not created")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.NumColors != 5 {
		t.Errorf("Default NumColors = %d, want 5", opts.NumColors)
	}
	if opts.Quality != 50 {
		t.Errorf("Default Quality = %d, want 50", opts.Quality)
	}
}

