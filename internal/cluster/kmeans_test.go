package cluster

import (
	"testing"
)

func TestKMeans_SingleColor(t *testing.T) {
	// All pixels are the same color
	pixels := make([][3]float64, 100)
	for i := range pixels {
		pixels[i] = [3]float64{255, 0, 0} // Red
	}

	results := KMeans(pixels, 1)
	if len(results) != 1 {
		t.Fatalf("KMeans got %d results, want 1", len(results))
	}

	if results[0].Hex != "#FF0000" {
		t.Errorf("KMeans got hex %s, want #FF0000", results[0].Hex)
	}

	if results[0].Percentage != 100 {
		t.Errorf("KMeans got percentage %f, want 100", results[0].Percentage)
	}
}

func TestKMeans_TwoColors(t *testing.T) {
	pixels := make([][3]float64, 200)
	// Half red, half blue
	for i := 0; i < 100; i++ {
		pixels[i] = [3]float64{255, 0, 0}
	}
	for i := 100; i < 200; i++ {
		pixels[i] = [3]float64{0, 0, 255}
	}

	results := KMeans(pixels, 2)
	if len(results) != 2 {
		t.Fatalf("KMeans got %d results, want 2", len(results))
	}

	// Check that we got both colors (order may vary)
	hasRed := false
	hasBlue := false
	for _, r := range results {
		if r.Hex == "#FF0000" {
			hasRed = true
		}
		if r.Hex == "#0000FF" {
			hasBlue = true
		}
	}

	if !hasRed {
		t.Error("KMeans did not find red")
	}
	if !hasBlue {
		t.Error("KMeans did not find blue")
	}
}

func TestKMeans_DistinctColors(t *testing.T) {
	pixels := make([][3]float64, 500)
	// 5 distinct colors, 100 pixels each
	colors := [][3]float64{
		{255, 0, 0},   // Red
		{0, 255, 0},   // Green
		{0, 0, 255},   // Blue
		{255, 255, 0}, // Yellow
		{255, 0, 255}, // Magenta
	}

	for i := 0; i < 500; i++ {
		pixels[i] = colors[i/100]
	}

	results := KMeans(pixels, 5)
	if len(results) != 5 {
		t.Fatalf("KMeans got %d results, want 5", len(results))
	}

	// Each color should have roughly 20% share
	for _, r := range results {
		if r.Percentage < 15 || r.Percentage > 25 {
			t.Errorf("KMeans got percentage %f for %s, expected ~20", r.Percentage, r.Hex)
		}
	}
}

func TestKMeans_EmptyPixels(t *testing.T) {
	results := KMeans(nil, 5)
	if results != nil {
		t.Errorf("KMeans with nil pixels should return nil, got %+v", results)
	}

	results = KMeans([][3]float64{}, 5)
	if results != nil {
		t.Errorf("KMeans with empty pixels should return nil, got %+v", results)
	}
}

func TestKMeans_SortedByPopulation(t *testing.T) {
	pixels := make([][3]float64, 100)
	// 60 red, 30 green, 10 blue
	for i := 0; i < 60; i++ {
		pixels[i] = [3]float64{255, 0, 0}
	}
	for i := 60; i < 90; i++ {
		pixels[i] = [3]float64{0, 255, 0}
	}
	for i := 90; i < 100; i++ {
		pixels[i] = [3]float64{0, 0, 255}
	}

	results := KMeans(pixels, 3)
	if len(results) < 2 {
		t.Fatalf("KMeans got %d results, want at least 2", len(results))
	}

	// Results should be sorted by percentage descending
	for i := 1; i < len(results); i++ {
		if results[i].Percentage > results[i-1].Percentage {
			t.Errorf("Results not sorted: %f > %f", results[i].Percentage, results[i-1].Percentage)
		}
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		a, b [3]float64
		want float64
	}{
		{[3]float64{0, 0, 0}, [3]float64{0, 0, 0}, 0},
		{[3]float64{255, 0, 0}, [3]float64{255, 0, 0}, 0},
		{[3]float64{0, 0, 0}, [3]float64{255, 0, 0}, 255},
	}

	for _, tt := range tests {
		got := distance(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("distance(%v, %v) = %f, want %f", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCalculateMean(t *testing.T) {
	points := [][3]float64{
		{0, 0, 0},
		{100, 100, 100},
		{200, 200, 200},
	}

	mean := calculateMean(points)
	if mean[0] != 100 || mean[1] != 100 || mean[2] != 100 {
		t.Errorf("calculateMean got %v, want [100 100 100]", mean)
	}
}

func BenchmarkKMeans(b *testing.B) {
	// Create 10000 random-ish pixels
	pixels := make([][3]float64, 10000)
	for i := range pixels {
		pixels[i] = [3]float64{
			float64(i % 256),
			float64((i * 7) % 256),
			float64((i * 13) % 256),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		KMeans(pixels, 5)
	}
}

