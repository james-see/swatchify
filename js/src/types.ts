/**
 * RGB color representation
 */
export interface RGB {
  r: number;
  g: number;
  b: number;
}

/**
 * Extracted color with hex code and percentage
 */
export interface ExtractedColor {
  /** Hex color code (e.g., "#FF5733") */
  hex: string;
  /** Percentage of image this color represents */
  percentage: number;
  /** RGB values */
  rgb: RGB;
}

/**
 * Options for color extraction
 */
export interface ExtractOptions {
  /** Number of dominant colors to extract (default: 5) */
  numColors?: number;
  /** Quality 0-100, lower = faster but less accurate (default: 50) */
  quality?: number;
  /** Filter out colors close to white */
  excludeWhite?: boolean;
  /** Filter out colors close to black */
  excludeBlack?: boolean;
  /** Minimum color distance between palette colors */
  minContrast?: number;
  /** Distance threshold for white/black detection (default: 30) */
  colorThreshold?: number;
}

/**
 * Input types accepted by extractColors
 */
export type ImageInput =
  | HTMLImageElement
  | HTMLCanvasElement
  | ImageData
  | File
  | Blob
  | string; // URL or data URL
