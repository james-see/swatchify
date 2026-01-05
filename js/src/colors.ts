import type { RGB } from './types';

/**
 * Convert RGB to hex string
 */
export function rgbToHex(r: number, g: number, b: number): string {
  const toHex = (n: number) => Math.round(n).toString(16).padStart(2, '0').toUpperCase();
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
}

/**
 * Parse hex string to RGB
 */
export function hexToRgb(hex: string): RGB {
  const clean = hex.replace('#', '');
  return {
    r: parseInt(clean.slice(0, 2), 16),
    g: parseInt(clean.slice(2, 4), 16),
    b: parseInt(clean.slice(4, 6), 16),
  };
}

/**
 * Euclidean distance between two RGB colors
 */
export function colorDistance(c1: RGB, c2: RGB): number {
  const dr = c1.r - c2.r;
  const dg = c1.g - c2.g;
  const db = c1.b - c2.b;
  return Math.sqrt(dr * dr + dg * dg + db * db);
}

/**
 * Check if color is near white
 */
export function isNearWhite(c: RGB, threshold: number): boolean {
  return colorDistance(c, { r: 255, g: 255, b: 255 }) < threshold;
}

/**
 * Check if color is near black
 */
export function isNearBlack(c: RGB, threshold: number): boolean {
  return colorDistance(c, { r: 0, g: 0, b: 0 }) < threshold;
}

/**
 * Calculate relative luminance (for contrast calculations)
 */
export function luminance(c: RGB): number {
  const toLinear = (v: number) => {
    const s = v / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * toLinear(c.r) + 0.7152 * toLinear(c.g) + 0.0722 * toLinear(c.b);
}

/**
 * Get contrasting text color (black or white) for a background
 */
export function getContrastColor(bg: RGB): string {
  return luminance(bg) > 0.5 ? '#000000' : '#FFFFFF';
}
