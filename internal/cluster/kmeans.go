package cluster

import (
	"math"
	"math/rand"
	"sort"

	"github.com/james-see/swatchify/internal/utils"
)

const (
	maxIterations  = 20
	maxSampleSize  = 50000
	convergenceEps = 1.0
)

// Cluster represents a color cluster with its centroid and members
type Cluster struct {
	Centroid [3]float64
	Members  [][3]float64
	Count    int
}

// KMeans performs k-means clustering on pixel data
func KMeans(pixels [][3]float64, k int) []utils.ColorResult {
	if len(pixels) == 0 {
		return nil
	}

	// Sample pixels if too many
	sampled := samplePixels(pixels, maxSampleSize)

	// Initialize centroids using k-means++
	centroids := initializeCentroidsKMeansPP(sampled, k)

	// Run clustering
	var clusters []Cluster
	for iter := 0; iter < maxIterations; iter++ {
		// Assign pixels to nearest centroid
		clusters = assignToClusters(sampled, centroids)

		// Recalculate centroids
		newCentroids := make([][3]float64, k)
		converged := true

		for i, cluster := range clusters {
			if cluster.Count > 0 {
				newCentroids[i] = calculateMean(cluster.Members)
				if distance(centroids[i], newCentroids[i]) > convergenceEps {
					converged = false
				}
			} else {
				// Keep old centroid if cluster is empty
				newCentroids[i] = centroids[i]
			}
		}

		centroids = newCentroids
		if converged {
			break
		}
	}

	// Final assignment
	clusters = assignToClusters(sampled, centroids)

	// Convert to results and sort by population
	return clustersToResults(clusters, len(sampled))
}

// samplePixels randomly samples pixels if there are too many
func samplePixels(pixels [][3]float64, maxSize int) [][3]float64 {
	if len(pixels) <= maxSize {
		return pixels
	}

	sampled := make([][3]float64, maxSize)
	perm := rand.Perm(len(pixels))
	for i := 0; i < maxSize; i++ {
		sampled[i] = pixels[perm[i]]
	}
	return sampled
}

// initializeCentroidsKMeansPP uses k-means++ initialization for better starting points
func initializeCentroidsKMeansPP(pixels [][3]float64, k int) [][3]float64 {
	if len(pixels) == 0 || k == 0 {
		return nil
	}

	centroids := make([][3]float64, 0, k)

	// Pick first centroid randomly
	centroids = append(centroids, pixels[rand.Intn(len(pixels))])

	// Pick remaining centroids with probability proportional to distance squared
	for len(centroids) < k {
		distances := make([]float64, len(pixels))
		totalDist := 0.0

		for i, pixel := range pixels {
			minDist := math.MaxFloat64
			for _, centroid := range centroids {
				d := distance(pixel, centroid)
				if d < minDist {
					minDist = d
				}
			}
			distances[i] = minDist * minDist
			totalDist += distances[i]
		}

		// Pick next centroid
		if totalDist == 0 {
			// All points are already centroids, pick randomly
			centroids = append(centroids, pixels[rand.Intn(len(pixels))])
			continue
		}

		threshold := rand.Float64() * totalDist
		cumulative := 0.0
		for i, d := range distances {
			cumulative += d
			if cumulative >= threshold {
				centroids = append(centroids, pixels[i])
				break
			}
		}
	}

	return centroids
}

// assignToClusters assigns each pixel to its nearest centroid
func assignToClusters(pixels [][3]float64, centroids [][3]float64) []Cluster {
	clusters := make([]Cluster, len(centroids))
	for i := range clusters {
		clusters[i] = Cluster{
			Centroid: centroids[i],
			Members:  make([][3]float64, 0),
		}
	}

	for _, pixel := range pixels {
		minDist := math.MaxFloat64
		minIdx := 0
		for i, centroid := range centroids {
			d := distance(pixel, centroid)
			if d < minDist {
				minDist = d
				minIdx = i
			}
		}
		clusters[minIdx].Members = append(clusters[minIdx].Members, pixel)
		clusters[minIdx].Count++
	}

	return clusters
}

// calculateMean calculates the mean of a set of points
func calculateMean(points [][3]float64) [3]float64 {
	if len(points) == 0 {
		return [3]float64{}
	}

	var sum [3]float64
	for _, p := range points {
		sum[0] += p[0]
		sum[1] += p[1]
		sum[2] += p[2]
	}

	n := float64(len(points))
	return [3]float64{sum[0] / n, sum[1] / n, sum[2] / n}
}

// distance calculates Euclidean distance between two points
func distance(a, b [3]float64) float64 {
	dr := a[0] - b[0]
	dg := a[1] - b[1]
	db := a[2] - b[2]
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// clustersToResults converts clusters to color results sorted by population
func clustersToResults(clusters []Cluster, totalPixels int) []utils.ColorResult {
	results := make([]utils.ColorResult, 0, len(clusters))

	for _, cluster := range clusters {
		if cluster.Count == 0 {
			continue
		}

		r := uint8(math.Round(cluster.Centroid[0]))
		g := uint8(math.Round(cluster.Centroid[1]))
		b := uint8(math.Round(cluster.Centroid[2]))

		percentage := float64(cluster.Count) / float64(totalPixels) * 100

		results = append(results, utils.ColorResult{
			Hex:        utils.RGBToHex(r, g, b),
			Percentage: math.Round(percentage*10) / 10,
			RGB:        utils.RGB{R: r, G: g, B: b},
		})
	}

	// Sort by percentage descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Percentage > results[j].Percentage
	})

	return results
}

