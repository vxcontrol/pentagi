package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newStaticTestServer builds a gin engine wired exactly like the production
// local-serving branch (registerStaticFileServer) over a throwaway dist dir
// holding one hashed asset and an index.html.
func newStaticTestServer(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html><title>app</title>"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "assets"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "assets", "app-abc123.js"), []byte("export const x = 1;\n"), 0o600))

	engine := gin.New()
	registerStaticFileServer(engine, dir)

	return engine
}

func getStatic(t *testing.T, engine *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

	return rec
}

func TestStaticFileServer(t *testing.T) {
	engine := newStaticTestServer(t)

	t.Run("existing hashed asset is served immutable", func(t *testing.T) {
		rec := getStatic(t, engine, "/assets/app-abc123.js")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	})

	t.Run("missing asset is 404 and never cached as a permanent negative", func(t *testing.T) {
		rec := getStatic(t, engine, "/assets/missing-deadbeef.js")

		// 404 (not 301->HTML) so the browser module loader fails cleanly and
		// the SPA reloads; no-store (not immutable) so a transient rolling
		// deploy isn't cached as a permanent miss.
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
		assert.NotContains(t, rec.Header().Get("Cache-Control"), "immutable")
	})

	t.Run("index.html is served revalidated", func(t *testing.T) {
		rec := getStatic(t, engine, "/")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("SPA deep-link falls back to index.html with no-cache", func(t *testing.T) {
		rec := getStatic(t, engine, "/templates")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.Contains(t, rec.Body.String(), "<!doctype html>")
	})

	t.Run("unknown non-asset path redirects to root", func(t *testing.T) {
		rec := getStatic(t, engine, "/favicon-not-there.ico")

		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/", rec.Header().Get("Location"))
	})

	t.Run("api paths are untouched by the static cache policy", func(t *testing.T) {
		rec := getStatic(t, engine, baseURL+"/anything")

		assert.Empty(t, rec.Header().Get("Cache-Control"))
	})
}
