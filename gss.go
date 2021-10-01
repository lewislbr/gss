package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

var (
	dir     = "dist"
	headers = map[string]string{}
	port    = "80"
)

type configYAML struct {
	Dir     string            `yaml:"directory,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Port    string            `yaml:"port,omitempty"`
}

func main() {
	setUpLogger()

	err := setUpYAML()
	if err != nil {
		log.Fatal().Msgf("Error retrieving YAML config: %v", err)
	}

	setUpCLI()

	// Check if the directory to serve exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatal().Msgf("Directory %q not found", dir)
	}

	err = startServer()
	if err != nil {
		log.Fatal().Msgf("Error starting server: %v", err)
	}
}

func setUpLogger() {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.With().Str("app", "GSS").Logger()
}

// Enable configuration via YAML file.
func setUpYAML() error {
	configFile := "gss.yaml"

	// Check if there is a config file
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}

	// Read the file
	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading YAML file: %w", err)
	}

	config := configYAML{}

	// Serialize the YAML content
	err = yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return fmt.Errorf("unmarshaling YAML file: %w", err)
	}

	// Assign non-empty values
	if config.Dir != "" {
		dir = config.Dir
	}
	if len(config.Headers) != 0 {
		headers = config.Headers
	}
	if config.Port != "" {
		port = config.Port
	}

	return nil
}

// Enable configuration via CLI flags.
func setUpCLI() {
	d := flag.String("d", dir, "Path to the directory to serve.")
	p := flag.String("p", port, "Port where to run the server.")

	flag.Parse()

	// Assign non-empty values
	if *d != "" {
		dir = *d
	}
	if *p != "" {
		port = *p
	}
}

// Initialize the server.
func startServer() error {
	s := setUpServer()

	log.Info().Msgf("Serving directory %q on port %v", dir, port)

	return s.ListenAndServe()
}

// Configure a basic HTTP server.
func setUpServer() *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      addHeaders(serveSPA(dir)),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// Serve static files from a directory.
func serveSPA(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqFile := filepath.Join(dir, filepath.Clean(r.URL.Path))

		// Send the index if the root path is requested.
		if filepath.Clean(r.URL.Path) == "/" {
			reqFile = reqFile + "/index.html"
		}

		// Send a 404 if a file with extension is not found, and the index if it has no extension,
		// as it will likely be a SPA route.
		if _, err := os.Stat(reqFile); os.IsNotExist(err) {
			if filepath.Ext(reqFile) != "" {
				w.WriteHeader(http.StatusNotFound)

				return
			}

			reqFile = filepath.Join(dir, "index.html")
		}

		// Serve pre-compressed file with appropriate headers and extension.
		serveCompressedFile := func(encoding, extension string) {
			serve := func(mimeType string) {
				w.Header().Add("Content-Encoding", encoding)
				w.Header().Add("Content-Type", mimeType)

				http.ServeFile(w, r, reqFile+extension)
			}

			switch filepath.Ext(reqFile) {
			case ".html":
				serve("text/html")
			case ".css":
				serve("text/css")
			case ".js":
				serve("application/javascript")
			case ".svg":
				serve("image/svg+xml")
			default:
				http.ServeFile(w, r, reqFile)
			}
		}

		acceptedEncodings := r.Header.Get("Accept-Encoding")
		files, err := filepath.Glob(dir + "/*")
		if err != nil {
			log.Error().Msgf("Error getting files to serve: %v", err)
		}

		brotli := "br"
		brotliExt := ".br"

		// If the request accepts brotli, and the directory contains brotli files, serve them.
		if strings.Contains(acceptedEncodings, brotli) {
			for _, f := range files {
				if f == reqFile+brotliExt {
					serveCompressedFile(brotli, brotliExt)

					return
				}
			}
		}

		gzip := "gzip"
		gzipExt := ".gz"

		// If the request accepts gzip, and the directory contains gzip files, serve them.
		if strings.Contains(acceptedEncodings, gzip) {
			for _, f := range files {
				if f == reqFile+gzipExt {
					serveCompressedFile(gzip, gzipExt)

					return
				}
			}
		}

		// If the request does not accept compressed files, or the directory does not contain compressed files,
		// serve the file as is.
		http.ServeFile(w, r, reqFile)
	}
}

// Add custom headers to the response.
func addHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Server", "GSS")

		for k, v := range headers {
			w.Header().Add(k, v)
		}

		h.ServeHTTP(w, r)
	}
}
