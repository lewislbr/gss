package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGSS(t *testing.T) {
	metrics := registerMetrics()

	t.Run("redirects index correctly", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/index.html", nil)

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
	})

	t.Run("serves HTML files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "html")
	})

	t.Run("serves CSS files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/static/main.68aa49f7.css", nil)

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "/static/main.68aa49f7.css", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "css")
	})

	t.Run("serves JavaScript files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/static/main.8d3db4ef.js", nil)

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "/static/main.8d3db4ef.js", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "javascript")
	})

	t.Run("serves other files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/static/main.8d3db4ef.js.LICENSE.txt", nil)

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "/static/main.8d3db4ef.js.LICENSE.txt", r.RequestURI)
	})

	t.Run("serves brotli files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		r.Header.Add("Accept-Encoding", "br")

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "br", w.Header().Get("Content-Encoding"))

		t.Run("serves brotli files under nested folders succesfully", func(t *testing.T) {
			t.Parallel()

			cfg := &config{}
			fileServer := newFileServer(cfg, metrics).init()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/static/main.8d3db4ef.js", nil)

			r.Header.Add("Accept-Encoding", "br")

			fileServer.Server.Handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
		})
	})

	t.Run("serves gzip files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		r.Header.Add("Accept-Encoding", "gzip")

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		t.Run("serves gzip files under nested folders succesfully", func(t *testing.T) {
			t.Parallel()

			cfg := &config{}
			fileServer := newFileServer(cfg, metrics).init()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/static/main.8d3db4ef.js", nil)

			r.Header.Add("Accept-Encoding", "br")

			fileServer.Server.Handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
		})
	})

	t.Run("serves unexisting files without extension", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/random-page", nil)

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/random-page", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "html")
	})

	t.Run("doesn't serve unexisting files with extension", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)

		fileServer.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("serves a cached response for a fresh resource", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		fileServer := newFileServer(cfg, metrics).init()

		t.Run("HTML files should have Cache-Control: no-cache", func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			fileServer.Server.Handler.ServeHTTP(w, r)

			last := w.Header().Get("Last-Modified")

			w = httptest.NewRecorder()
			r = httptest.NewRequest(http.MethodGet, "/", nil)

			r.Header.Set("If-Modified-Since", last)

			fileServer.Server.Handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusNotModified, w.Code)
			assert.Equal(t, w.Header().Get("Cache-Control"), "no-cache")
		})

		t.Run("other files should have Cache-Control: public, max-age=31536000, immutable", func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/static/main.8d3db4ef.js", nil)

			fileServer.Server.Handler.ServeHTTP(w, r)

			last := w.Header().Get("Last-Modified")

			w = httptest.NewRecorder()
			r = httptest.NewRequest(http.MethodGet, "/static/main.8d3db4ef.js", nil)

			r.Header.Set("If-Modified-Since", last)

			fileServer.Server.Handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusNotModified, w.Code)
			assert.Equal(t, w.Header().Get("Cache-Control"), "public, max-age=31536000, immutable")
		})
	})
}
