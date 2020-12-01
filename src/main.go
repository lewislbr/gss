package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

// start initializes the server
func start() error {
	httpServer := customHTTPServer(
		responseMiddleware(serveSPA("dist")),
	)

	fmt.Println("GSS ready ✅")

	err := httpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

// customHTTPServer configures a basic HTTP server
func customHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":80",
		Handler:      handler,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// serveSPA serves files from a directory, defaulting to the index if the root
// is requested or a file is not found, leaving it for the SPA to handle
func serveSPA(directory string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestedPath := filepath.Join(directory, filepath.Clean(r.URL.Path))

		if filepath.Clean(r.URL.Path) == "/" {
			requestedPath = requestedPath + "/index.html"
		}
		if _, err := os.Stat(requestedPath); os.IsNotExist(err) {
			requestedPath = filepath.Join(directory, "index.html")
		}

		http.ServeFile(w, r, requestedPath)
	}
}

// responseMiddleware adds custom headers to the response
func responseMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Server", "GSS")

		h.ServeHTTP(w, r)
	}
}
