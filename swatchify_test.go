package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/james-see/swatchify/internal/cluster"
	"github.com/james-see/swatchify/internal/imageio"
	"github.com/james-see/swatchify/internal/palette"
	"github.com/james-see/swatchify/internal/utils"
)

func TestIntegration_ExampleImage(t *testing.T) {
	imagePath := filepath.Join("examples", "image1.jpg")

	// Skip if example image doesn't exist
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Example image not found, skipping integration test")
	}

	// Load image
	opts := imageio.LoadOptions{MaxDimension: 300}
	img, err := imageio.LoadImage(imagePath, opts)
	if err != nil {
		t.Fatalf("Failed to load image: %v", err)
	}

	// Extract pixels
	pixels := imageio.GetPixels(img)
	if len(pixels) == 0 {
		t.Fatal("No pixels extracted from image")
	}

	// Run k-means
	colors := cluster.KMeans(pixels, 5)
	if len(colors) == 0 {
		t.Fatal("No colors extracted")
	}
	if len(colors) > 5 {
		t.Errorf("Got %d colors, expected max 5", len(colors))
	}

	// Verify color format
	for _, c := range colors {
		if len(c.Hex) != 7 || c.Hex[0] != '#' {
			t.Errorf("Invalid hex format: %s", c.Hex)
		}
		if c.Percentage < 0 || c.Percentage > 100 {
			t.Errorf("Invalid percentage: %f", c.Percentage)
		}
	}

	// Verify sorted by percentage
	for i := 1; i < len(colors); i++ {
		if colors[i].Percentage > colors[i-1].Percentage {
			t.Error("Colors not sorted by percentage descending")
		}
	}

	t.Logf("Extracted %d colors from example image", len(colors))
	for _, c := range colors {
		t.Logf("  %s (%.1f%%)", c.Hex, c.Percentage)
	}
}

func TestIntegration_PaletteGeneration(t *testing.T) {
	imagePath := filepath.Join("examples", "image1.jpg")

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Example image not found, skipping integration test")
	}

	// Load and process
	opts := imageio.LoadOptions{MaxDimension: 300}
	img, err := imageio.LoadImage(imagePath, opts)
	if err != nil {
		t.Fatalf("Failed to load image: %v", err)
	}

	pixels := imageio.GetPixels(img)
	colors := cluster.KMeans(pixels, 5)

	// Generate palette
	tmpDir := t.TempDir()
	palettePath := filepath.Join(tmpDir, "palette.png")

	err = palette.GeneratePNG(colors, palettePath, palette.DefaultOptions())
	if err != nil {
		t.Fatalf("Failed to generate palette: %v", err)
	}

	// Verify palette was created
	info, err := os.Stat(palettePath)
	if err != nil {
		t.Fatalf("Palette file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Palette file is empty")
	}

	t.Logf("Generated palette: %s (%d bytes)", palettePath, info.Size())
}

func TestIntegration_Filters(t *testing.T) {
	imagePath := filepath.Join("examples", "image1.jpg")

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Example image not found, skipping integration test")
	}

	opts := imageio.LoadOptions{MaxDimension: 300}
	img, err := imageio.LoadImage(imagePath, opts)
	if err != nil {
		t.Fatalf("Failed to load image: %v", err)
	}

	pixels := imageio.GetPixels(img)
	colors := cluster.KMeans(pixels, 10)

	// Test white/black exclusion
	filtered := utils.FilterColors(colors, true, true, 30)
	t.Logf("After white/black filter: %d -> %d colors", len(colors), len(filtered))

	// Test min contrast filter
	contrastFiltered := utils.FilterByMinContrast(colors, 50)
	t.Logf("After min contrast filter: %d -> %d colors", len(colors), len(contrastFiltered))
}

func BenchmarkIntegration_FullPipeline(b *testing.B) {
	imagePath := filepath.Join("examples", "image1.jpg")

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		b.Skip("Example image not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := imageio.LoadOptions{MaxDimension: 300}
		img, _ := imageio.LoadImage(imagePath, opts)
		pixels := imageio.GetPixels(img)
		cluster.KMeans(pixels, 5)
	}
}

