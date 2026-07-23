// Package main provides a swatchify MCP server for goose and other MCP clients.
//
// Exposes color extraction tools: extract_colors and generate_palette.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/james-see/swatchify/pkg/swatchify"
)

type extractColorsArgs struct {
	FilePath     string  `json:"file_path"      jsonschema:"Path to the image file"`
	NumColors    float64 `json:"num_colors,omitempty"     jsonschema:"Number of dominant colors to extract. Default 5"`
	Quality      float64 `json:"quality,omitempty"        jsonschema:"Downscale quality 0-100. Lower is faster. Default 50"`
	ExcludeWhite bool    `json:"exclude_white,omitempty"  jsonschema:"Filter out colors near white"`
	ExcludeBlack bool    `json:"exclude_black,omitempty"  jsonschema:"Filter out colors near black"`
}

type generatePaletteArgs struct {
	InputPath  string  `json:"input_path"           jsonschema:"Path to the source image file"`
	OutputPath string  `json:"output_path"           jsonschema:"Path for the output palette PNG"`
	NumColors  float64 `json:"num_colors,omitempty"  jsonschema:"Number of colors in palette. Default 5"`
	Width      float64 `json:"width,omitempty"       jsonschema:"Palette image width in pixels. Default 1000"`
	Height     float64 `json:"height,omitempty"      jsonschema:"Palette image height in pixels. Default 200"`
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "swatchify",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "extract_colors",
		Description: "Extract dominant colors from an image file using k-means clustering. Returns hex colors with percentage dominance. Supports JPG, PNG, WebP, GIF, BMP, TIFF.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args extractColorsArgs) (*mcp.CallToolResult, any, error) {
		if args.FilePath == "" {
			return errorResult("file_path is required"), nil, nil
		}

		opts := swatchify.DefaultOptions()
		if args.NumColors > 0 {
			opts.NumColors = int(args.NumColors)
		}
		if args.Quality > 0 {
			opts.Quality = int(args.Quality)
		}
		opts.ExcludeWhite = args.ExcludeWhite
		opts.ExcludeBlack = args.ExcludeBlack

		colors, err := swatchify.ExtractFromFile(args.FilePath, opts)
		if err != nil {
			return errorResult(fmt.Sprintf("Error extracting colors: %v", err)), nil, nil
		}

		text := fmt.Sprintf("Extracted %d colors from %s:\n", len(colors), args.FilePath)
		for _, c := range colors {
			text += fmt.Sprintf("  %s (%.1f%%) rgb(%d,%d,%d)\n", c.Hex, c.Percentage, c.R, c.G, c.B)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_palette",
		Description: "Generate a color palette PNG image from an input image. Creates a horizontal strip of the dominant colors as a new PNG file.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args generatePaletteArgs) (*mcp.CallToolResult, any, error) {
		if args.InputPath == "" {
			return errorResult("input_path is required"), nil, nil
		}
		if args.OutputPath == "" {
			return errorResult("output_path is required"), nil, nil
		}

		opts := swatchify.DefaultOptions()
		if args.NumColors > 0 {
			opts.NumColors = int(args.NumColors)
		}

		colors, err := swatchify.ExtractFromFile(args.InputPath, opts)
		if err != nil {
			return errorResult(fmt.Sprintf("Error extracting colors: %v", err)), nil, nil
		}

		palOpts := swatchify.DefaultPaletteOptions()
		if args.Width > 0 {
			palOpts.Width = int(args.Width)
		}
		if args.Height > 0 {
			palOpts.Height = int(args.Height)
		}

		if err := swatchify.GeneratePalette(colors, args.OutputPath, palOpts); err != nil {
			return errorResult(fmt.Sprintf("Error generating palette: %v", err)), nil, nil
		}

		text := fmt.Sprintf("OK: generated palette with %d colors at %s\n", len(colors), args.OutputPath)
		for _, c := range colors {
			text += fmt.Sprintf("  %s (%.1f%%)\n", c.Hex, c.Percentage)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("swatchify-mcp: %v", err)
		os.Exit(1)
	}
}