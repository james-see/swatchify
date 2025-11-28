package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/james-see/swatchify/pkg/swatchify"
)

var (
	port int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server for color extraction API",
	Long: `Start an HTTP server that provides a REST API for color extraction.

Endpoints:
  POST /extract    Extract colors from uploaded image
  GET  /health     Health check endpoint

Example:
  swatchify serve --port 8080
  curl -X POST -F "image=@photo.jpg" http://localhost:8080/extract
  curl -X POST -F "image=@photo.jpg" -F "colors=8" http://localhost:8080/extract`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
}

func runServe(cmd *cobra.Command, args []string) error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", handleHealth)

	// Extract colors endpoint
	mux.HandleFunc("/extract", handleExtract)

	// Root info
	mux.HandleFunc("/", handleRoot)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🎨 Swatchify server starting on http://localhost%s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /extract  - Extract colors from image")
	fmt.Println("  GET  /health   - Health check")
	fmt.Println()

	server := &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(loggingMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server.ListenAndServe()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s %s %v\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"name":    "swatchify",
		"version": "0.3.0",
		"endpoints": map[string]string{
			"POST /extract": "Extract dominant colors from image",
			"GET /health":   "Health check",
		},
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// ExtractResponse is the JSON response for color extraction
type ExtractResponse struct {
	Success bool              `json:"success"`
	Colors  []swatchify.Color `json:"colors,omitempty"`
	Error   string            `json:"error,omitempty"`
}

func handleExtract(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(ExtractResponse{
			Success: false,
			Error:   "Method not allowed. Use POST.",
		})
		return
	}

	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ExtractResponse{
			Success: false,
			Error:   "Failed to parse form: " + err.Error(),
		})
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ExtractResponse{
			Success: false,
			Error:   "No image file provided. Use form field 'image'.",
		})
		return
	}
	defer file.Close()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "swatchify-*"+getExtension(header.Filename))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ExtractResponse{
			Success: false,
			Error:   "Failed to create temp file",
		})
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy uploaded file to temp
	if _, err := io.Copy(tmpFile, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ExtractResponse{
			Success: false,
			Error:   "Failed to save uploaded file",
		})
		return
	}

	// Parse options from form
	opts := parseOptions(r)

	// Extract colors
	colors, err := swatchify.ExtractFromFile(tmpFile.Name(), opts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ExtractResponse{
			Success: false,
			Error:   "Failed to extract colors: " + err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(ExtractResponse{
		Success: true,
		Colors:  colors,
	})
}

func parseOptions(r *http.Request) *swatchify.Options {
	opts := swatchify.DefaultOptions()

	if n := r.FormValue("colors"); n != "" {
		if num, err := strconv.Atoi(n); err == nil && num > 0 && num <= 50 {
			opts.NumColors = num
		}
	}

	if q := r.FormValue("quality"); q != "" {
		if qual, err := strconv.Atoi(q); err == nil && qual >= 0 && qual <= 100 {
			opts.Quality = qual
		}
	}

	if r.FormValue("exclude_white") == "true" || r.FormValue("exclude_white") == "1" {
		opts.ExcludeWhite = true
	}

	if r.FormValue("exclude_black") == "true" || r.FormValue("exclude_black") == "1" {
		opts.ExcludeBlack = true
	}

	if mc := r.FormValue("min_contrast"); mc != "" {
		if contrast, err := strconv.ParseFloat(mc, 64); err == nil {
			opts.MinContrast = contrast
		}
	}

	return opts
}

func getExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ".tmp"
}

