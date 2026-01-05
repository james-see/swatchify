import { colorDistance, isNearWhite, isNearBlack, getContrastColor } from './colors';
import { kmeans } from './kmeans';
import type { ExtractedColor, ExtractOptions, ImageInput, RGB } from './types';

export type { ExtractedColor, ExtractOptions, ImageInput, RGB };
export { getContrastColor } from './colors';

const DEFAULT_OPTIONS: Required<ExtractOptions> = {
  numColors: 5,
  quality: 50,
  excludeWhite: false,
  excludeBlack: false,
  minContrast: 0,
  colorThreshold: 30,
};

/**
 * Convert quality (0-100) to max dimension (100-500px)
 */
function qualityToMaxDimension(quality: number): number {
  const q = Math.max(0, Math.min(100, quality));
  return 100 + q * 4;
}

/**
 * Load image from various input types and return as HTMLImageElement
 */
async function loadImage(input: ImageInput): Promise<HTMLImageElement> {
  if (input instanceof HTMLImageElement) {
    if (input.complete) return input;
    return new Promise((resolve, reject) => {
      input.onload = () => resolve(input);
      input.onerror = reject;
    });
  }

  if (input instanceof HTMLCanvasElement) {
    const img = new Image();
    img.src = input.toDataURL();
    return new Promise((resolve, reject) => {
      img.onload = () => resolve(img);
      img.onerror = reject;
    });
  }

  if (input instanceof ImageData) {
    const canvas = document.createElement('canvas');
    canvas.width = input.width;
    canvas.height = input.height;
    const ctx = canvas.getContext('2d')!;
    ctx.putImageData(input, 0, 0);
    const img = new Image();
    img.src = canvas.toDataURL();
    return new Promise((resolve, reject) => {
      img.onload = () => resolve(img);
      img.onerror = reject;
    });
  }

  if (input instanceof Blob) {
    const url = URL.createObjectURL(input);
    const img = new Image();
    img.src = url;
    return new Promise((resolve, reject) => {
      img.onload = () => {
        URL.revokeObjectURL(url);
        resolve(img);
      };
      img.onerror = (e) => {
        URL.revokeObjectURL(url);
        reject(e);
      };
    });
  }

  // String URL
  const img = new Image();
  img.crossOrigin = 'anonymous';
  img.src = input;
  return new Promise((resolve, reject) => {
    img.onload = () => resolve(img);
    img.onerror = reject;
  });
}

/**
 * Get pixels from image, optionally downscaling for performance
 */
function getPixels(img: HTMLImageElement, maxDim: number): [number, number, number][] {
  let { width, height } = img;

  // Downscale if needed
  if (width > maxDim || height > maxDim) {
    if (width > height) {
      height = Math.round((height * maxDim) / width);
      width = maxDim;
    } else {
      width = Math.round((width * maxDim) / height);
      height = maxDim;
    }
  }

  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext('2d')!;
  ctx.drawImage(img, 0, 0, width, height);

  const imageData = ctx.getImageData(0, 0, width, height);
  const data = imageData.data;
  const pixels: [number, number, number][] = [];

  for (let i = 0; i < data.length; i += 4) {
    const a = data[i + 3];
    if (a === 0) continue; // Skip transparent pixels
    pixels.push([data[i], data[i + 1], data[i + 2]]);
  }

  return pixels;
}

/**
 * Filter colors based on options
 */
function filterColors(
  colors: ExtractedColor[],
  opts: Required<ExtractOptions>
): ExtractedColor[] {
  let result = colors;

  // Filter white/black
  if (opts.excludeWhite || opts.excludeBlack) {
    result = result.filter((c) => {
      if (opts.excludeWhite && isNearWhite(c.rgb, opts.colorThreshold)) return false;
      if (opts.excludeBlack && isNearBlack(c.rgb, opts.colorThreshold)) return false;
      return true;
    });
  }

  // Filter by min contrast
  if (opts.minContrast > 0) {
    const filtered: ExtractedColor[] = [];
    for (const color of result) {
      const tooClose = filtered.some(
        (kept) => colorDistance(color.rgb, kept.rgb) < opts.minContrast
      );
      if (!tooClose) filtered.push(color);
    }
    result = filtered;
  }

  return result;
}

/**
 * Extract dominant colors from an image.
 *
 * @param input - Image source (HTMLImageElement, Canvas, ImageData, File, Blob, or URL)
 * @param options - Extraction options
 * @returns Array of extracted colors sorted by dominance
 *
 * @example
 * ```ts
 * // From file input
 * const colors = await extractColors(fileInput.files[0]);
 *
 * // From URL
 * const colors = await extractColors('https://example.com/image.jpg');
 *
 * // With options
 * const colors = await extractColors(img, {
 *   numColors: 8,
 *   excludeWhite: true,
 *   quality: 75
 * });
 * ```
 */
export async function extractColors(
  input: ImageInput,
  options?: ExtractOptions
): Promise<ExtractedColor[]> {
  const opts = { ...DEFAULT_OPTIONS, ...options };

  const img = await loadImage(input);
  const maxDim = qualityToMaxDimension(opts.quality);
  const pixels = getPixels(img, maxDim);

  if (pixels.length === 0) return [];

  // Request more colors if filtering
  let requestColors = opts.numColors;
  if (opts.excludeWhite || opts.excludeBlack || opts.minContrast > 0) {
    requestColors = Math.min(opts.numColors * 2, 50);
  }

  // Run k-means clustering
  let colors = kmeans(pixels, requestColors);

  // Apply filters
  colors = filterColors(colors, opts);

  // Limit to requested count
  return colors.slice(0, opts.numColors);
}

/**
 * Generate a palette canvas element from extracted colors
 */
export function generatePaletteCanvas(
  colors: ExtractedColor[],
  width = 500,
  height = 100
): HTMLCanvasElement {
  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext('2d')!;

  if (colors.length === 0) return canvas;

  // Calculate total percentage
  const totalPct = colors.reduce((sum, c) => sum + c.percentage, 0) || colors.length;

  let x = 0;
  for (let i = 0; i < colors.length; i++) {
    const color = colors[i];
    const blockWidth =
      i === colors.length - 1
        ? width - x
        : Math.round((width * color.percentage) / totalPct);

    // Fill block
    ctx.fillStyle = color.hex;
    ctx.fillRect(x, 0, blockWidth, height);

    // Draw hex label
    ctx.fillStyle = getContrastColor(color.rgb);
    ctx.font = '12px monospace';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    const textX = x + blockWidth / 2;
    if (blockWidth > 50) {
      ctx.fillText(color.hex, textX, height / 2);
    }

    x += blockWidth;
  }

  return canvas;
}

/**
 * Generate palette as data URL
 */
export function generatePaletteDataURL(
  colors: ExtractedColor[],
  width = 500,
  height = 100
): string {
  return generatePaletteCanvas(colors, width, height).toDataURL('image/png');
}
