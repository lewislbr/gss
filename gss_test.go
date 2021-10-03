package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGSS(t *testing.T) {
	t.Run("uses the provided directory", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()

		require.Equal(t, "test/web/dist", app.Config.Dir)
	})

	t.Run("uses the provided port", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir:  "test/web/dist",
			Port: "8080",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()

		require.Equal(t, ":8080", app.Server.Addr)
	})

	t.Run("uses the provided headers", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
			Headers: map[string]string{
				"X-Test": "test",
			},
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, "test", w.Header().Get("X-Test"))
	})

	t.Run("redirects index correctly", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/index.html", nil)

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusMovedPermanently, w.Result().StatusCode)
	})

	t.Run("serves HTML files succesfully", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "/", r.RequestURI)
		require.Contains(t, w.Header().Get("Content-Type"), "html")
	})

	t.Run("serves CSS files succesfully", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.css", nil)

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "/main.css", r.RequestURI)
		require.Contains(t, w.Header().Get("Content-Type"), "css")
	})

	t.Run("serves JavaScript files succesfully", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.js", nil)

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "/main.js", r.RequestURI)
		require.Contains(t, w.Header().Get("Content-Type"), "javascript")
	})

	t.Run("serves other files succesfully", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.js.LICENSE.txt", nil)

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "/main.js.LICENSE.txt", r.RequestURI)
	})

	t.Run("serves brotli files succesfully", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.js", nil)

		r.Header.Add("Accept-Encoding", "br")

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "br", w.Header().Get("Content-Encoding"))
	})

	t.Run("serves gzip files succesfully", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/main.js", nil)

		r.Header.Add("Accept-Encoding", "gzip")

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	})

	t.Run("serves unexisting files without extension", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/random-page", nil)

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "/random-page", r.RequestURI)
		require.Contains(t, w.Header().Get("Content-Type"), "html")
	})

	t.Run("doesn't serve unexisting files with extension", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Dir: "test/web/dist",
		}
		config, err := config.Validate()

		require.NoError(t, err)

		app := NewApp(config).Init()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)

		app.Server.Handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})
}
