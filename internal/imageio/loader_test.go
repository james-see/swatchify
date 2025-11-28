package imageio

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestQualityToMaxDimension(t *testing.T) {
	tests := []struct {
		quality int
		want    int
	}{
		{0, 100},
		{50, 300},
		{100, 500},
		{-10, 100},  // Clamped to 0
		{150, 500},  // Clamped to 100
	}

	for _, tt := range tests {
		got := QualityToMaxDimension(tt.quality)
		if got != tt.want {
			t.Errorf("QualityToMaxDimension(%d) = %d, want %d", tt.quality, got, tt.want)
		}
	}
}

func TestLoadImage_FileNotFound(t *testing.T) {
	_, err := LoadImage("/nonexistent/path/image.png", DefaultLoadOptions())
	if err == nil {
		t.Error("LoadImage should error on nonexistent file")
	}
}

func TestLoadImage_UnsupportedFormat(t *testing.T) {
	// Create a temp file with unsupported extension
	tmpFile, err := os.CreateTemp("", "test*.xyz")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = LoadImage(tmpFile.Name(), DefaultLoadOptions())
	if err == nil {
		t.Error("LoadImage should error on unsupported format")
	}
}

func TestLoadImage_ValidPNG(t *testing.T) {
	// Create a test PNG
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	// Load it
	loaded, err := LoadImage(imgPath, DefaultLoadOptions())
	if err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	bounds := loaded.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Image size = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}
}

func TestLoadImage_Downscale(t *testing.T) {
	// Create a large test PNG
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "large.png")

	img := image.NewRGBA(image.Rect(0, 0, 1000, 800))
	for y := 0; y < 800; y++ {
		for x := 0; x < 1000; x++ {
			img.Set(x, y, color.RGBA{0, 255, 0, 255})
		}
	}

	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	// Load with small max dimension
	opts := LoadOptions{MaxDimension: 200}
	loaded, err := LoadImage(imgPath, opts)
	if err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	bounds := loaded.Bounds()
	// Should be downscaled to 200x160 (maintaining aspect ratio)
	if bounds.Dx() > 200 || bounds.Dy() > 200 {
		t.Errorf("Image not downscaled properly: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGetPixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	pixels := GetPixels(img)
	if len(pixels) != 100 {
		t.Errorf("GetPixels got %d pixels, want 100", len(pixels))
	}

	// Check first pixel
	if pixels[0][0] != 100 || pixels[0][1] != 150 || pixels[0][2] != 200 {
		t.Errorf("GetPixels first pixel = %v, want [100 150 200]", pixels[0])
	}
}

func TestGetPixels_SkipsTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Half opaque, half transparent
	for y := 0; y < 10; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Opaque
		}
		for x := 5; x < 10; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0}) // Transparent
		}
	}

	pixels := GetPixels(img)
	if len(pixels) != 50 {
		t.Errorf("GetPixels got %d pixels, want 50 (transparent skipped)", len(pixels))
	}
}

func TestDefaultLoadOptions(t *testing.T) {
	opts := DefaultLoadOptions()
	if opts.MaxDimension != 300 {
		t.Errorf("DefaultLoadOptions MaxDimension = %d, want 300", opts.MaxDimension)
	}
}

