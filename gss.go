package main

import (
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

func main() {
	setUpLogger()

	cfg, err := newConfig().withYAML().validate()
	if err != nil {
		log.Fatal().Msgf("Error validating config: %v", err)
	}

	err = newApp(cfg).init().run()
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
	Dir     string            `yaml:"directory,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Port    string            `yaml:"port,omitempty"`
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

func (c *config) validate() (*config, error) {
	if c.Dir == "" {
		c.Dir = "dist"
	}
	if c.Port == "" {
		c.Port = "80"
	}
	if _, err := os.Stat(c.Dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory %q not found", c.Dir)
	}

	return c, nil
}

type app struct {
	Config config
	Server *http.Server
}

func newApp(cfg *config) *app {
	return &app{
		Config: *cfg,
		Server: &http.Server{
			Addr:         ":" + cfg.Port,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (a *app) init() *app {
	a.Server.Handler = a.setHeaders((a.serveSPA()))

	return a
}

func (a *app) run() error {
	log.Info().Msgf("Serving directory %q on port %v", a.Config.Dir, a.Config.Port)

	return a.Server.ListenAndServe()
}

func (a *app) serveSPA() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqFile := filepath.Join(a.Config.Dir, filepath.Clean(r.URL.Path))

		// Send the index if the root path is requested.
		if filepath.Clean(r.URL.Path) == "/" {
			reqFile = filepath.Join(a.Config.Dir, "index.html")
		}

		// Send a 404 if a file with extension is not found, and the index if it has no extension,
		// as it will likely be a SPA route.
		if _, err := os.Stat(reqFile); os.IsNotExist(err) {
			if filepath.Ext(reqFile) != "" {
				w.WriteHeader(http.StatusNotFound)

				return
			}

			reqFile = filepath.Join(a.Config.Dir, "index.html")
		}

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

		encodings := r.Header.Get("Accept-Encoding")
		files, err := filepath.Glob(a.Config.Dir + "/*")
		if err != nil {
			log.Error().Msgf("Error getting files to serve: %v", err)
		}

		brotli := "br"
		brotliExt := ".br"

		if strings.Contains(encodings, brotli) {
			for _, f := range files {
				if f == reqFile+brotliExt {
					serveCompressedFile(brotli, brotliExt)

					return
				}
			}
		}

		gzip := "gzip"
		gzipExt := ".gz"

		if strings.Contains(encodings, gzip) {
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

func (a *app) setHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Server", "GSS")

		for k, v := range a.Config.Headers {
			w.Header().Add(k, v)
		}

		h.ServeHTTP(w, r)
	}
}
