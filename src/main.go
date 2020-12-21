package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var dir, port string

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

// start initializes the server
func start() error {
	setUpCLI()

	httpServer := customHTTPServer(port, addHeaders(serveSPA(dir)))

	fmt.Println("GSS ready ✅")

	err := httpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

// setUpCLI enables configuration via CLI
func setUpCLI() {
	flag.StringVar(
		&dir, "d", "dist", "Container path to the directory to serve.",
	)
	flag.StringVar(&port, "p", "80", "Port where to run the server.")
	flag.Parse()
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
			fmt.Println(err)
		}
		gzipFiles, err := filepath.Glob(dir + "/*" + gzipExt)
		if err != nil {
			fmt.Println(err)
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
		} else if len(gzipFiles) > 0 && strings.Contains(
			acceptedEncodings, gzip,
		) {
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

		h.ServeHTTP(w, r)
	}
}
