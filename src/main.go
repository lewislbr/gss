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

// startServer initializes the server.
func startServer() error {
	httpServer := customHTTPServer(port, addHeaders(serveSPA(dir)))

	log.Printf("GSS info: serving directory %q on port %v ✅\n", dir, port)

	err := httpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

// setUpYAML enables configuration via YAML file.
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

	// Check if values are empty
	if config.Dir == "" || len(config.Headers) == 0 || config.Port == "" {
		log.Println("GSS warning: some YAML config values are empty ⚠️")
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

// setUpCLI enables configuration via CLI flags.
func setUpCLI() {
	d := flag.String("d", dir, "Path to the directory to serve.")
	p := flag.String("p", port, "Port where to run the server.")

	flag.Parse()

	// Check if flags are set up
	if *d == dir && *p == port {
		return
	}

	// Assign non-empty values
	if *d != "" {
		dir = *d
	}
	if *p != "" {
		port = *p
	}
}

// customHTTPServer configures a basic HTTP server
func customHTTPServer(port string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      h,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// serveSPA serves files from a directory, defaulting to the index if the root
// is requested or a file is not found, leaving it for the SPA to handle. If
// the directory contains pre-compressed brotli or gzip files those are served
// instead for the file types that accept them.
func serveSPA(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqFile := filepath.Join(dir, filepath.Clean(r.URL.Path))

		if filepath.Clean(r.URL.Path) == "/" {
			reqFile = reqFile + "/index.html"
		}
		if _, err := os.Stat(reqFile); os.IsNotExist(err) {
			reqFile = filepath.Join(dir, "index.html")
		}

		brotli := "br"
		brotliExt := ".br"
		gzip := "gzip"
		gzipExt := ".gz"
		brotliFiles, err := filepath.Glob(dir + "/*" + brotliExt)
		if err != nil {
			log.Println(err)
		}
		gzipFiles, err := filepath.Glob(dir + "/*" + gzipExt)
		if err != nil {
			log.Println(err)
		}
		acceptedEncodings := r.Header.Get("Accept-Encoding")
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

		if len(brotliFiles) > 0 && strings.Contains(acceptedEncodings, brotli) {
			serveCompressedFile(brotli, brotliExt)
		} else if len(gzipFiles) > 0 && strings.Contains(acceptedEncodings, gzip) {
			serveCompressedFile(gzip, gzipExt)
		} else {
			http.ServeFile(w, r, reqFile)
		}
	}
}

// addHeaders adds custom headers to the response
func addHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Server", "GSS")

		for k, v := range headers {
			w.Header().Add(k, v)
		}

		h.ServeHTTP(w, r)
	}
}
