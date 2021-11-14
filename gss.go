package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

func main() {
	setUpLogger()

	cfg := newConfig().withYAML()
	err := newApp(cfg).init().run()
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

type config struct {
	Headers map[string]string `yaml:"headers,omitempty"`
}

func newConfig() *config {
	return &config{}
}

func (c *config) withYAML() *config {
	file := "gss.yaml"

	if _, err := os.Stat(file); os.IsNotExist(err) {
		// If no file is found we assume config via YAML is not used
		return c
	}

	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal().Msgf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Fatal().Msgf("Error unmarshalling file data: %v", err)
	}

	return c
}

type app struct {
	Config config
	Server *http.Server
}

func newApp(cfg *config) *app {
	return &app{
		Config: *cfg,
		Server: &http.Server{
			Addr:         ":8080",
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (a *app) init() *app {
	a.Server.Handler = a.setHeaders((a.serveSPA()))

	return a
}

func (a *app) run() error {
	return a.Server.ListenAndServe()
}

func (a *app) serveSPA() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dir := "dist"
		file := filepath.Join(dir, filepath.Clean(r.URL.Path))

		// Send the index if the root path is requested.
		if filepath.Clean(r.URL.Path) == "/" {
			file = filepath.Join(dir, "index.html")
		}

		// Send a 404 if a file with extension is not found, and the index if it has no extension,
		// as it will likely be a SPA route.
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if filepath.Ext(file) != "" {
				w.WriteHeader(http.StatusNotFound)

				return
			}

			file = filepath.Join(dir, "index.html")
		}

		serveFile := func(mimeType string) {
			files, err := filepath.Glob(dir + "/*")
			if err != nil {
				log.Error().Msgf("Error getting files to serve: %v", err)
			}

			encodings := r.Header.Get("Accept-Encoding")
			brotli := "br"
			brotliExt := ".br"
			gzip := "gzip"
			gzipExt := ".gz"
			serveCompressed := func(encoding, extension string) {
				w.Header().Set("Content-Encoding", encoding)
				w.Header().Set("Content-Type", mimeType)

				http.ServeFile(w, r, file+extension)
			}

			if strings.Contains(encodings, brotli) {
				for _, f := range files {
					if f == file+brotliExt {
						serveCompressed(brotli, brotliExt)

						return
					}
				}
			}

			if strings.Contains(encodings, gzip) {
				for _, f := range files {
					if f == file+gzipExt {
						serveCompressed(gzip, gzipExt)

						return
					}
				}
			}

			// If the request does not accept compressed files, or the directory does not contain compressed files,
			// serve the file as is.
			http.ServeFile(w, r, file)
		}

		switch filepath.Ext(file) {
		case ".html":
			w.Header().Set("Cache-Control", "no-cache")

			serveFile("text/html")
		case ".css":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

			serveFile("text/css")
		case ".js":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

			serveFile("application/javascript")
		case ".svg":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

			serveFile("image/svg+xml")
		default:
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

			http.ServeFile(w, r, file)
		}
	}
}

func (a *app) setHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")

		for k, v := range a.Config.Headers {
			w.Header().Set(k, v)
		}

		h.ServeHTTP(w, r)
	}
}
