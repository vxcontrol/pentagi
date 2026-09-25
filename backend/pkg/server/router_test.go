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

// routerStaticEngine serves a throwaway dist dir holding one hashed asset and an index.html.
func routerStaticEngine(t *testing.T) *gin.Engine {
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

func routerGetStatic(t *testing.T, engine *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

	return rec
}

func TestRouter_RegisterStaticFileServer_ServesTheSPAWithItsCachePolicy(t *testing.T) {
	engine := routerStaticEngine(t)

	t.Run("existing hashed asset is served immutable", func(t *testing.T) {
		rec := routerGetStatic(t, engine, "/assets/app-abc123.js")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	})

	t.Run("missing asset is 404 and never cached as a permanent negative", func(t *testing.T) {
		rec := routerGetStatic(t, engine, "/assets/missing-deadbeef.js")

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
		assert.NotContains(t, rec.Header().Get("Cache-Control"), "immutable")
	})

	t.Run("index.html is served revalidated", func(t *testing.T) {
		rec := routerGetStatic(t, engine, "/")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("SPA deep-link falls back to index.html with no-cache", func(t *testing.T) {
		rec := routerGetStatic(t, engine, "/templates")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.Contains(t, rec.Body.String(), "<!doctype html>")
	})

	t.Run("unknown non-asset path redirects to root", func(t *testing.T) {
		rec := routerGetStatic(t, engine, "/favicon-not-there.ico")

		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/", rec.Header().Get("Location"))
	})

	t.Run("api paths are untouched by the static cache policy", func(t *testing.T) {
		rec := routerGetStatic(t, engine, baseURL+"/anything")

		assert.Empty(t, rec.Header().Get("Cache-Control"))
	})
}

const (
	routerTokenTierStatus = 598
	routerUserTierStatus  = 599
)

func routerTierEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	tokenTier := engine.Group("/")
	tokenTier.Use(func(c *gin.Context) { c.AbortWithStatus(routerTokenTierStatus) })

	userTier := engine.Group("/")
	userTier.Use(func(c *gin.Context) { c.AbortWithStatus(routerUserTierStatus) })

	registerPrivateRoutes(tokenTier, userTier, privateServices{})

	return engine
}

func TestRouter_RegisterPrivateRoutes_KeepsAccountAdministrationOffTheTokenTier(t *testing.T) {
	engine := routerTierEngine()

	cases := []struct {
		name   string
		method string
		path   string
		tier   int
	}{
		{"users list", http.MethodGet, "/users/", routerUserTierStatus},
		{"user create", http.MethodPost, "/users/", routerUserTierStatus},
		{"user delete", http.MethodDelete, "/users/somehash", routerUserTierStatus},
		{"roles list", http.MethodGet, "/roles/", routerUserTierStatus},
		{"tokens list", http.MethodGet, "/tokens/", routerUserTierStatus},
		{"token create", http.MethodPost, "/tokens/", routerUserTierStatus},
		{"token delete", http.MethodDelete, "/tokens/sometoken", routerUserTierStatus},
		{"current user stays reachable with an api token", http.MethodGet, "/user/", routerTokenTierStatus},
		{"flows stay reachable with an api token", http.MethodPost, "/flows/", routerTokenTierStatus},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))

			assert.Equal(t, tc.tier, rec.Code)
		})
	}
}

func TestRouter_RegisterAuthRoutes_RefusesALogoutByGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	const reached = 597
	authGroup := engine.Group("/auth")
	authGroup.Use(func(c *gin.Context) { c.AbortWithStatus(reached) })
	registerAuthRoutes(authGroup, nil)

	get := httptest.NewRecorder()
	engine.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/auth/logout", nil))
	assert.Equal(t, http.StatusNotFound, get.Code,
		"a cross-site top-level navigation carries the Lax session cookie, so a GET must not end sessions")

	post := httptest.NewRecorder()
	engine.ServeHTTP(post, httptest.NewRequest(http.MethodPost, "/auth/logout", nil))
	assert.Equal(t, reached, post.Code)
}

func TestRouter_SetKnowledgeGroup_RefusesAMissingPrivilege(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		priv   string
	}{
		{"list", http.MethodGet, "/knowledge/", "knowledge.view"},
		{"read one", http.MethodGet, "/knowledge/doc-1", "knowledge.view"},
		{"create", http.MethodPost, "/knowledge/", "knowledge.create"},
		{"search", http.MethodPost, "/knowledge/search", "knowledge.search"},
		{"update", http.MethodPut, "/knowledge/doc-1", "knowledge.edit"},
		{"delete", http.MethodDelete, "/knowledge/doc-1", "knowledge.delete"},
	}

	every := []string{"knowledge.view", "knowledge.create", "knowledge.search", "knowledge.edit", "knowledge.delete"}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			granted := []string{}
			for _, priv := range every {
				if priv != tc.priv {
					granted = append(granted, priv)
				}
			}

			gin.SetMode(gin.TestMode)
			engine := gin.New()
			engine.Use(gin.Recovery(), func(c *gin.Context) { c.Set("prm", granted) })
			setKnowledgeGroup(engine.Group(""), nil)

			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))

			assert.Equal(t, http.StatusForbidden, rec.Code,
				"a caller holding every knowledge privilege but %s reached %s %s", tc.priv, tc.method, tc.path)
		})
	}
}
