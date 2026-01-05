import { rgbToHex } from './colors';
import type { ExtractedColor } from './types';

const MAX_ITERATIONS = 20;
const MAX_SAMPLE_SIZE = 50000;
const CONVERGENCE_EPS = 1.0;

type Point = [number, number, number];

interface Cluster {
  centroid: Point;
  members: Point[];
}

/**
 * K-means++ clustering on pixel data
 */
export function kmeans(pixels: Point[], k: number): ExtractedColor[] {
  if (pixels.length === 0 || k <= 0) return [];

  // Sample pixels if too many
  const sampled = samplePixels(pixels, MAX_SAMPLE_SIZE);

  // Initialize centroids using k-means++
  let centroids = initializeCentroidsKMeansPP(sampled, k);

  // Run clustering iterations
  let clusters: Cluster[] = [];
  for (let iter = 0; iter < MAX_ITERATIONS; iter++) {
    clusters = assignToClusters(sampled, centroids);

    const newCentroids: Point[] = [];
    let converged = true;

    for (let i = 0; i < clusters.length; i++) {
      const cluster = clusters[i];
      if (cluster.members.length > 0) {
        const mean = calculateMean(cluster.members);
        newCentroids.push(mean);
        if (distance(centroids[i], mean) > CONVERGENCE_EPS) {
          converged = false;
        }
      } else {
        newCentroids.push(centroids[i]);
      }
    }

    centroids = newCentroids;
    if (converged) break;
  }

  // Final assignment
  clusters = assignToClusters(sampled, centroids);

  return clustersToResults(clusters, sampled.length);
}

function samplePixels(pixels: Point[], maxSize: number): Point[] {
  if (pixels.length <= maxSize) return pixels;

  const sampled: Point[] = [];
  const step = pixels.length / maxSize;
  for (let i = 0; i < maxSize; i++) {
    sampled.push(pixels[Math.floor(i * step)]);
  }
  return sampled;
}

function initializeCentroidsKMeansPP(pixels: Point[], k: number): Point[] {
  if (pixels.length === 0 || k === 0) return [];

  const centroids: Point[] = [];

  // Pick first centroid randomly
  centroids.push(pixels[Math.floor(Math.random() * pixels.length)]);

  // Pick remaining centroids with probability proportional to distance squared
  while (centroids.length < k) {
    const distances: number[] = [];
    let totalDist = 0;

    for (const pixel of pixels) {
      let minDist = Infinity;
      for (const centroid of centroids) {
        const d = distance(pixel, centroid);
        if (d < minDist) minDist = d;
      }
      const distSq = minDist * minDist;
      distances.push(distSq);
      totalDist += distSq;
    }

    if (totalDist === 0) {
      centroids.push(pixels[Math.floor(Math.random() * pixels.length)]);
      continue;
    }

    const threshold = Math.random() * totalDist;
    let cumulative = 0;
    for (let i = 0; i < distances.length; i++) {
      cumulative += distances[i];
      if (cumulative >= threshold) {
        centroids.push(pixels[i]);
        break;
      }
    }
  }

  return centroids;
}

function assignToClusters(pixels: Point[], centroids: Point[]): Cluster[] {
  const clusters: Cluster[] = centroids.map((c) => ({
    centroid: c,
    members: [],
  }));

  for (const pixel of pixels) {
    let minDist = Infinity;
    let minIdx = 0;
    for (let i = 0; i < centroids.length; i++) {
      const d = distance(pixel, centroids[i]);
      if (d < minDist) {
        minDist = d;
        minIdx = i;
      }
    }
    clusters[minIdx].members.push(pixel);
  }

  return clusters;
}

function calculateMean(points: Point[]): Point {
  if (points.length === 0) return [0, 0, 0];

  let r = 0, g = 0, b = 0;
  for (const p of points) {
    r += p[0];
    g += p[1];
    b += p[2];
  }
  const n = points.length;
  return [r / n, g / n, b / n];
}

function distance(a: Point, b: Point): number {
  const dr = a[0] - b[0];
  const dg = a[1] - b[1];
  const db = a[2] - b[2];
  return Math.sqrt(dr * dr + dg * dg + db * db);
}

function clustersToResults(clusters: Cluster[], totalPixels: number): ExtractedColor[] {
  const results: ExtractedColor[] = [];

  for (const cluster of clusters) {
    if (cluster.members.length === 0) continue;

    const [r, g, b] = cluster.centroid.map((v) => Math.round(v));
    const percentage = Math.round((cluster.members.length / totalPixels) * 1000) / 10;

    results.push({
      hex: rgbToHex(r, g, b),
      percentage,
      rgb: { r, g, b },
    });
  }

  // Sort by percentage descending
  results.sort((a, b) => b.percentage - a.percentage);

  return results;
}
