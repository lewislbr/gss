package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGSS(t *testing.T) {
	t.Run("uses the provided headers", func(t *testing.T) {
		t.Parallel()

		cfg := &config{
			Headers: map[string]string{
				"X-Test": "test",
			},
		}

		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, "test", w.Header().Get("X-Test"))
	})

	t.Run("redirects index correctly", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/index.html", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
	})

	t.Run("serves HTML files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
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

		cfg := &config{}
		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.68aa49f7.css", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "/main.68aa49f7.css", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "css")
	})

	t.Run("serves JavaScript files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.8d3db4ef.js", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "/main.8d3db4ef.js", r.RequestURI)
		assert.Contains(t, w.Header().Get("Content-Type"), "javascript")
	})

	t.Run("serves other files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.8d3db4ef.js.LICENSE.txt", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "/main.8d3db4ef.js.LICENSE.txt", r.RequestURI)
	})

	t.Run("serves brotli files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.8d3db4ef.js", nil)

		r.Header.Add("Accept-Encoding", "br")

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
	})

	t.Run("serves gzip files succesfully", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.8d3db4ef.js", nil)

		r.Header.Add("Accept-Encoding", "gzip")

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	})

	t.Run("serves unexisting files without extension", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
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

		cfg := &config{}
		app := newApp(cfg).init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)

		app.Server.Handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("serves a cached response for a fresh resource", func(t *testing.T) {
		t.Parallel()

		cfg := &config{}
		app := newApp(cfg).init()

		t.Run("HTML files should have Cache-Control: no-cache", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			app.Server.Handler.ServeHTTP(w, r)

			last := w.Header().Get("Last-Modified")

			w = httptest.NewRecorder()
			r = httptest.NewRequest(http.MethodGet, "/", nil)

			r.Header.Set("If-Modified-Since", last)

			app.Server.Handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusNotModified, w.Code)
			assert.Equal(t, w.Header().Get("Cache-Control"), "no-cache")
		})

		t.Run("other files should have Cache-Control: public, max-age=31536000, immutable", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/main.8d3db4ef.js", nil)

			app.Server.Handler.ServeHTTP(w, r)

			last := w.Header().Get("Last-Modified")

			w = httptest.NewRecorder()
			r = httptest.NewRequest(http.MethodGet, "/main.8d3db4ef.js", nil)

			r.Header.Set("If-Modified-Since", last)

			app.Server.Handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusNotModified, w.Code)
			assert.Equal(t, w.Header().Get("Cache-Control"), "public, max-age=31536000, immutable")
		})
	})
}
