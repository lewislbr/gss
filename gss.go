package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	err := setUpYAML()
	if err != nil {
		log.Fatalf("GSS error: something went wrong with the YAML file: %v ❌\n", err)
	}

	setUpCLI()

	// Check if the directory to serve exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatalf("GSS error: directory %q not found ❌\n", dir)
	}

	err = startServer()
	if err != nil {
		log.Fatalf("GSS error: the server crashed: %v ❌\n", err)
	}
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
		return err
	}

	config := configYAML{}

	// Serialize the YAML content
	err = yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return err
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
	s := setUpServer(port)

	log.Printf("GSS info: serving directory %q on port %v ✅\n", dir, port)

	return s.ListenAndServe()
}

// Configure a basic HTTP server.
func setUpServer(port string) *http.Server {
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
		serveCompressedFile := func(encoding string, extension string) {
			serve := func(encoding string, mimeType string, extension string) {
				w.Header().Add("Content-Encoding", encoding)
				w.Header().Add("Content-Type", mimeType)

				http.ServeFile(w, r, reqFile+extension)
			}

			switch filepath.Ext(reqFile) {
			case ".html":
				serve(encoding, "text/html", extension)
			case ".css":
				serve(encoding, "text/css", extension)
			case ".js":
				serve(encoding, "application/javascript", extension)
			case ".svg":
				serve(encoding, "image/svg+xml", extension)
			default:
				http.ServeFile(w, r, reqFile)
			}
		}

		acceptedEncodings := r.Header.Get("Accept-Encoding")
		files, err := filepath.Glob(dir + "/*")
		if err != nil {
			log.Println(err)
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
