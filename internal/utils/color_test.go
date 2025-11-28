package utils

import (
	"image/color"
	"testing"
)

func TestRGBToHex(t *testing.T) {
	tests := []struct {
		r, g, b uint8
		want    string
	}{
		{0, 0, 0, "#000000"},
		{255, 255, 255, "#FFFFFF"},
		{255, 0, 0, "#FF0000"},
		{0, 255, 0, "#00FF00"},
		{0, 0, 255, "#0000FF"},
		{18, 52, 86, "#123456"},
		{171, 205, 239, "#ABCDEF"},
	}

	for _, tt := range tests {
		got := RGBToHex(tt.r, tt.g, tt.b)
		if got != tt.want {
			t.Errorf("RGBToHex(%d, %d, %d) = %s, want %s", tt.r, tt.g, tt.b, got, tt.want)
		}
	}
}

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		hex     string
		want    RGB
		wantErr bool
	}{
		{"#000000", RGB{0, 0, 0}, false},
		{"#FFFFFF", RGB{255, 255, 255}, false},
		{"#FF0000", RGB{255, 0, 0}, false},
		{"#123456", RGB{18, 52, 86}, false},
		{"123456", RGB{18, 52, 86}, false},
		{"#abcdef", RGB{171, 205, 239}, false},
	}

	for _, tt := range tests {
		got, err := HexToRGB(tt.hex)
		if (err != nil) != tt.wantErr {
			t.Errorf("HexToRGB(%s) error = %v, wantErr %v", tt.hex, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("HexToRGB(%s) = %+v, want %+v", tt.hex, got, tt.want)
		}
	}
}

func TestColorToRGB(t *testing.T) {
	tests := []struct {
		c    color.Color
		want RGB
	}{
		{color.RGBA{255, 0, 0, 255}, RGB{255, 0, 0}},
		{color.RGBA{0, 255, 0, 255}, RGB{0, 255, 0}},
		{color.RGBA{0, 0, 255, 255}, RGB{0, 0, 255}},
		{color.White, RGB{255, 255, 255}},
		{color.Black, RGB{0, 0, 0}},
	}

	for _, tt := range tests {
		got := ColorToRGB(tt.c)
		if got != tt.want {
			t.Errorf("ColorToRGB(%+v) = %+v, want %+v", tt.c, got, tt.want)
		}
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		c1, c2 RGB
		want   float64
	}{
		{RGB{0, 0, 0}, RGB{0, 0, 0}, 0},
		{RGB{255, 255, 255}, RGB{255, 255, 255}, 0},
		{RGB{0, 0, 0}, RGB{255, 0, 0}, 255},
		{RGB{0, 0, 0}, RGB{0, 255, 0}, 255},
		{RGB{0, 0, 0}, RGB{0, 0, 255}, 255},
	}

	for _, tt := range tests {
		got := EuclideanDistance(tt.c1, tt.c2)
		if got != tt.want {
			t.Errorf("EuclideanDistance(%+v, %+v) = %f, want %f", tt.c1, tt.c2, got, tt.want)
		}
	}
}

func TestIsNearWhite(t *testing.T) {
	tests := []struct {
		c         RGB
		threshold float64
		want      bool
	}{
		{RGB{255, 255, 255}, 10, true},
		{RGB{250, 250, 250}, 20, true},
		{RGB{200, 200, 200}, 20, false},
		{RGB{0, 0, 0}, 50, false},
	}

	for _, tt := range tests {
		got := IsNearWhite(tt.c, tt.threshold)
		if got != tt.want {
			t.Errorf("IsNearWhite(%+v, %f) = %v, want %v", tt.c, tt.threshold, got, tt.want)
		}
	}
}

func TestIsNearBlack(t *testing.T) {
	tests := []struct {
		c         RGB
		threshold float64
		want      bool
	}{
		{RGB{0, 0, 0}, 10, true},
		{RGB{5, 5, 5}, 20, true},
		{RGB{50, 50, 50}, 20, false},
		{RGB{255, 255, 255}, 50, false},
	}

	for _, tt := range tests {
		got := IsNearBlack(tt.c, tt.threshold)
		if got != tt.want {
			t.Errorf("IsNearBlack(%+v, %f) = %v, want %v", tt.c, tt.threshold, got, tt.want)
		}
	}
}

func TestFilterColors(t *testing.T) {
	colors := []ColorResult{
		{Hex: "#FFFFFF", RGB: RGB{255, 255, 255}},
		{Hex: "#000000", RGB: RGB{0, 0, 0}},
		{Hex: "#FF0000", RGB: RGB{255, 0, 0}},
		{Hex: "#00FF00", RGB: RGB{0, 255, 0}},
	}

	// Exclude white
	filtered := FilterColors(colors, true, false, 30)
	if len(filtered) != 3 {
		t.Errorf("FilterColors (exclude white) got %d colors, want 3", len(filtered))
	}

	// Exclude black
	filtered = FilterColors(colors, false, true, 30)
	if len(filtered) != 3 {
		t.Errorf("FilterColors (exclude black) got %d colors, want 3", len(filtered))
	}

	// Exclude both
	filtered = FilterColors(colors, true, true, 30)
	if len(filtered) != 2 {
		t.Errorf("FilterColors (exclude both) got %d colors, want 2", len(filtered))
	}

	// Exclude neither
	filtered = FilterColors(colors, false, false, 30)
	if len(filtered) != 4 {
		t.Errorf("FilterColors (exclude neither) got %d colors, want 4", len(filtered))
	}
}

func TestFilterByMinContrast(t *testing.T) {
	colors := []ColorResult{
		{Hex: "#FF0000", RGB: RGB{255, 0, 0}},
		{Hex: "#FE0000", RGB: RGB{254, 0, 0}}, // Very similar to first
		{Hex: "#00FF00", RGB: RGB{0, 255, 0}}, // Different
	}

	// With high min contrast, similar colors should be filtered
	filtered := FilterByMinContrast(colors, 50)
	if len(filtered) != 2 {
		t.Errorf("FilterByMinContrast got %d colors, want 2", len(filtered))
	}

	// With zero min contrast, all should remain
	filtered = FilterByMinContrast(colors, 0)
	if len(filtered) != 3 {
		t.Errorf("FilterByMinContrast (0) got %d colors, want 3", len(filtered))
	}
}

func TestLuminance(t *testing.T) {
	tests := []struct {
		c    RGB
		want float64
	}{
		{RGB{0, 0, 0}, 0},
		{RGB{255, 255, 255}, 1},
	}

	for _, tt := range tests {
		got := Luminance(tt.c)
		diff := got - tt.want
		if diff < -0.01 || diff > 0.01 {
			t.Errorf("Luminance(%+v) = %f, want %f", tt.c, got, tt.want)
		}
	}
}

