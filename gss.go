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

func main() {
	setUpLogger()

	config, err := NewConfig().GetYAML().GetCLI().Validate()
	if err != nil {
		log.Fatal().Msgf("Error validating config: %v", err)
	}

	err = NewApp(config).Init().ListenAndServe()
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

type Config struct {
	Dir     string            `yaml:"directory,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Port    string            `yaml:"port,omitempty"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) GetYAML() *Config {
	configFile := "gss.yaml"

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return c
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal().Msgf("Error reading YAML config: %v", err)
	}

	err = yaml.Unmarshal([]byte(content), &c)
	if err != nil {
		log.Fatal().Msgf("Error unmarshalling YAML data: %v", err)
	}

	return c
}

func (c *Config) GetCLI() *Config {
	dir := flag.String("d", c.Dir, "Path to the directory to serve.")
	port := flag.String("p", c.Port, "Port where to run the server.")

	flag.Parse()

	if *dir != "" {
		c.Dir = *dir
	}
	if *port != "" {
		c.Port = *port
	}

	return c
}

func (c *Config) Validate() (*Config, error) {
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

type App struct {
	Config Config
	Server *http.Server
}

func NewApp(config *Config) *App {
	return &App{
		Config: *config,
		Server: &http.Server{
			Addr:         ":" + config.Port,
			IdleTimeout:  120 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (a *App) Init() *App {
	a.Server.Handler = a.AddHeaders((a.ServeSPA()))

	return a
}

func (a *App) ListenAndServe() error {
	log.Info().Msgf("Serving directory %q on port %v", a.Config.Dir, a.Config.Port)

	return a.Server.ListenAndServe()
}

func (a *App) ServeSPA() http.HandlerFunc {
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

		acceptedEncodings := r.Header.Get("Accept-Encoding")
		files, err := filepath.Glob(a.Config.Dir + "/*")
		if err != nil {
			log.Error().Msgf("Error getting files to serve: %v", err)
		}

		brotli := "br"
		brotliExt := ".br"

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

func (a *App) AddHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Server", "GSS")

		for k, v := range a.Config.Headers {
			w.Header().Add(k, v)
		}

		h.ServeHTTP(w, r)
	}
}
