package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

func main() {
	setUpLogger()

	metrics := registerMetrics()
	internalServer := newInternalServer(metrics)

	go func() {
		err := internalServer.run()
		if err != nil {
			log.Fatal().Msgf("Error starting internal server: %v", err)
		}
	}()

	cfg := newConfig().withYAML()
	err := newFileServer(cfg, metrics).init().run()
	if err != nil {
		log.Fatal().Msgf("Error starting file server: %v", err)
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
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
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

type fileServer struct {
	Config  config
	Metrics metrics
	Server  *http.Server
}

func newFileServer(cfg *config, metrics *metrics) *fileServer {
	return &fileServer{
		Config:  *cfg,
		Metrics: *metrics,
		Server: &http.Server{
			Addr:         ":8080",
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (f *fileServer) init() *fileServer {
	f.Server.Handler = metricsMiddleware(&f.Metrics)(f.setHeaders((f.serveSPA())))

	return f
}

func (f *fileServer) run() error {
	return f.Server.ListenAndServe()
}

func (f *fileServer) serveSPA() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dir := "dist"
		file := filepath.Join(dir, filepath.Clean(r.URL.Path))

		// Send the index if the root path is requested.
		if filepath.Clean(r.URL.Path) == "/" {
			file = filepath.Join(dir, "index.html")
		}

		// Send a 404 if a file with extension is not found, and the index if it has no extension,
		// as it will likely be a SPA route.
		_, err := os.Stat(file)
		if os.IsNotExist(err) {
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

func (f *fileServer) setHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")

		for k, v := range f.Config.Headers {
			w.Header().Set(k, v)
		}

		h.ServeHTTP(w, r)
	}
}

type internalServer struct {
	Server *http.Server
}

func newInternalServer(metrics *metrics) *internalServer {
	http.HandleFunc("/metrics", metrics.Default().ServeHTTP)

	s := &http.Server{}
	s.Addr = ":9090"

	return &internalServer{
		Server: s,
	}
}

func (i *internalServer) run() error {
	return i.Server.ListenAndServe()
}

type metrics struct {
	requestsReceived *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	bytesWritten     prometheus.Counter
}

func registerMetrics() *metrics {
	const labelCode = "code"

	reqReceived := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "http",
			Name:      "requests_total",
			Help:      "Total number of requests received.",
		},
		[]string{labelCode},
	)
	reqDuration := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "http",
			Name:      "request_duration_seconds",
			Help:      "Duration of a request in seconds.",
		},
		[]string{labelCode},
	)
	bytesWritten := promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "http",
			Name:      "bytes_written_total",
			Help:      "Total number of bytes written.",
		},
	)

	return &metrics{
		requestsReceived: reqReceived,
		requestDuration:  reqDuration,
		bytesWritten:     bytesWritten,
	}
}

func (m *metrics) Default() http.Handler {
	return promhttp.Handler()
}

func (m *metrics) IncRequests(code int) {
	m.requestsReceived.WithLabelValues(strconv.Itoa(code)).Inc()
}

func (m *metrics) ObsDuration(code int, duration float64) {
	m.requestDuration.WithLabelValues(strconv.Itoa(code)).Observe(duration)
}

func (m *metrics) AddBytes(bytes float64) {
	m.bytesWritten.Add(bytes)
}

func metricsMiddleware(metrics *metrics) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			snoop := httpsnoop.CaptureMetrics(h, w, r)

			metrics.IncRequests(snoop.Code)
			metrics.ObsDuration(snoop.Code, snoop.Duration.Seconds())
			metrics.AddBytes(float64(snoop.Written))
		})
	}
}
