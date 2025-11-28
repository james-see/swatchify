package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/james-see/swatchify/internal/cluster"
	"github.com/james-see/swatchify/internal/imageio"
	"github.com/james-see/swatchify/internal/palette"
	"github.com/james-see/swatchify/internal/utils"
)

var (
	// Flags
	numColors    int
	jsonOutput   bool
	pngOutput    string
	excludeWhite bool
	excludeBlack bool
	minContrast  float64
	showPalette  bool
	quality      int
	paletteWidth int
	paletteHeight int
)

// JSONOutput represents the JSON output format
type JSONOutput struct {
	Image  string              `json:"image"`
	Colors []utils.ColorResult `json:"colors"`
}

var rootCmd = &cobra.Command{
	Use:   "swatchify <image>",
	Short: "Extract dominant colors from images",
	Long: `Swatchify extracts the dominant colors from any image using k-means clustering.

Supported formats: JPG, PNG, WebP, GIF, BMP, TIFF

Examples:
  swatchify photo.jpg
  swatchify logo.png -n 8
  swatchify image.png --json
  swatchify brand.png --png palette.png --show`,
	Args: cobra.ExactArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().IntVarP(&numColors, "colors", "n", 5, "Number of dominant colors to extract")
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.Flags().StringVar(&pngOutput, "png", "", "Generate palette PNG at specified path")
	rootCmd.Flags().BoolVar(&excludeWhite, "exclude-white", false, "Exclude colors close to white")
	rootCmd.Flags().BoolVar(&excludeBlack, "exclude-black", false, "Exclude colors close to black")
	rootCmd.Flags().Float64Var(&minContrast, "min-contrast", 0, "Minimum color distance between palette colors")
	rootCmd.Flags().BoolVar(&showPalette, "show", false, "Open generated palette image after creation")
	rootCmd.Flags().IntVar(&quality, "quality", 50, "Downscale quality 0-100 (higher = more accurate, slower)")
	rootCmd.Flags().IntVar(&paletteWidth, "width", 1000, "Palette image width in pixels")
	rootCmd.Flags().IntVar(&paletteHeight, "height", 200, "Palette image height in pixels")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	imagePath := args[0]

	// Validate inputs
	if numColors < 1 {
		return fmt.Errorf("number of colors must be at least 1")
	}
	if numColors > 50 {
		return fmt.Errorf("number of colors cannot exceed 50")
	}
	if quality < 0 || quality > 100 {
		return fmt.Errorf("quality must be between 0 and 100")
	}

	// Load image
	loadOpts := imageio.LoadOptions{
		MaxDimension: imageio.QualityToMaxDimension(quality),
	}
	img, err := imageio.LoadImage(imagePath, loadOpts)
	if err != nil {
		return err
	}

	// Extract pixels
	pixels := imageio.GetPixels(img)
	if len(pixels) == 0 {
		return fmt.Errorf("no pixels found in image")
	}

	// Request more colors than needed if filtering is enabled
	requestColors := numColors
	if excludeWhite || excludeBlack || minContrast > 0 {
		requestColors = numColors * 2
		if requestColors > 50 {
			requestColors = 50
		}
	}

	// Run k-means clustering
	colors := cluster.KMeans(pixels, requestColors)

	// Apply filters
	const colorThreshold = 30.0 // Distance threshold for near-white/black detection
	colors = utils.FilterColors(colors, excludeWhite, excludeBlack, colorThreshold)
	colors = utils.FilterByMinContrast(colors, minContrast)

	// Limit to requested number
	if len(colors) > numColors {
		colors = colors[:numColors]
	}

	if len(colors) == 0 {
		return fmt.Errorf("no colors remaining after filtering")
	}

	// Generate palette PNG if requested
	if pngOutput != "" {
		opts := palette.Options{
			Width:  paletteWidth,
			Height: paletteHeight,
		}
		if err := palette.GeneratePNG(colors, pngOutput, opts); err != nil {
			return fmt.Errorf("failed to generate palette: %w", err)
		}

		if showPalette {
			if err := palette.OpenImage(pngOutput); err != nil {
				// Don't fail on open error, just warn
				fmt.Fprintf(os.Stderr, "Warning: could not open palette: %v\n", err)
			}
		}
	}

	// Output results
	if jsonOutput {
		return outputJSON(imagePath, colors)
	}
	return outputText(colors)
}

func outputText(colors []utils.ColorResult) error {
	for _, c := range colors {
		fmt.Println(c.Hex)
	}
	return nil
}

func outputJSON(imagePath string, colors []utils.ColorResult) error {
	output := JSONOutput{
		Image:  imagePath,
		Colors: colors,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

