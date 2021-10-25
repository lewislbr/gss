package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGSS(t *testing.T) {
	t.Run("uses the provided directory", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()

		assert.Equal(t, "test/web/dist", app.Config.Dir)
	})

	t.Run("uses the provided port", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir:  "test/web/dist",
			Port: "8080",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()

		assert.Equal(t, ":8080", app.Server.Addr)
	})

	t.Run("uses the provided headers", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
			Headers: map[string]string{
				"X-Test": "test",
			},
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, "test", w.Header().Get("X-Test"))
	})

	t.Run("redirects index correctly", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/index.html", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
	})

	t.Run("serves HTML files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "html")
	})

	t.Run("serves CSS files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.css", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/main.css", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "css")
	})

	t.Run("serves JavaScript files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.js", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/main.js", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "javascript")
	})

	t.Run("serves other files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.js.LICENSE.txt", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/main.js.LICENSE.txt", r.RequestURI)
	})

	t.Run("serves brotli files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.js", nil)

		r.Header.Add("Accept-Encoding", "br")

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
	})

	t.Run("serves gzip files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.js", nil)

		r.Header.Add("Accept-Encoding", "gzip")

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	})

	t.Run("serves unexisting files without extension", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/random-page", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/random-page", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "html")
	})

	t.Run("doesn't serve unexisting files with extension", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Dir: "test/web/dist",
		}
		cfg, err := cfg.validate()

		assert.NoError(t, err)

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
