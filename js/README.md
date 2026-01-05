# swatchme

Extract dominant colors from images using k-means clustering — browser-only, zero dependencies.

## Install

```bash
npm install swatchme
```

## Usage

```ts
import { extractColors, generatePaletteCanvas } from 'swatchme';

// From file input
const file = document.querySelector('input[type="file"]').files[0];
const colors = await extractColors(file);

console.log(colors);
// [
//   { hex: '#4A90D9', percentage: 32.5, rgb: { r: 74, g: 144, b: 217 } },
//   { hex: '#2C3E50', percentage: 24.1, rgb: { r: 44, g: 62, b: 80 } },
//   ...
// ]

// Generate a palette canvas
const canvas = generatePaletteCanvas(colors, 500, 100);
document.body.appendChild(canvas);
```

## API

### `extractColors(input, options?)`

Extract dominant colors from an image.

**Input types:**
- `HTMLImageElement`
- `HTMLCanvasElement`
- `ImageData`
- `File` or `Blob`
- URL string (including data URLs)

**Options:**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `numColors` | number | 5 | Number of colors to extract |
| `quality` | number | 50 | Quality 0-100 (lower = faster) |
| `excludeWhite` | boolean | false | Filter out near-white colors |
| `excludeBlack` | boolean | false | Filter out near-black colors |
| `minContrast` | number | 0 | Minimum distance between colors |
| `colorThreshold` | number | 30 | Threshold for white/black detection |

**Returns:** `Promise<ExtractedColor[]>`

### `generatePaletteCanvas(colors, width?, height?)`

Create a canvas element showing the color palette.

```ts
const canvas = generatePaletteCanvas(colors, 600, 120);
```

### `generatePaletteDataURL(colors, width?, height?)`

Get the palette as a PNG data URL.

```ts
const dataUrl = generatePaletteDataURL(colors);
img.src = dataUrl;
```

### `getContrastColor(rgb)`

Get black or white text color for optimal contrast on a background.

```ts
import { getContrastColor } from 'swatchme';

const textColor = getContrastColor({ r: 74, g: 144, b: 217 });
// '#FFFFFF'
```

## Types

```ts
interface ExtractedColor {
  hex: string;        // e.g., "#4A90D9"
  percentage: number; // e.g., 32.5
  rgb: { r: number; g: number; b: number };
}

interface ExtractOptions {
  numColors?: number;
  quality?: number;
  excludeWhite?: boolean;
  excludeBlack?: boolean;
  minContrast?: number;
  colorThreshold?: number;
}
```

## Development

```bash
cd js

# Install dependencies
npm install

# Run example dev server
npm run dev

# Build library
npm run build
```

## How it works

1. Image is loaded and optionally downscaled based on quality setting
2. Pixels are extracted via Canvas API
3. K-means++ clustering finds dominant color clusters
4. Colors are filtered (white/black exclusion, min contrast)
5. Results sorted by dominance percentage

## License

MIT
