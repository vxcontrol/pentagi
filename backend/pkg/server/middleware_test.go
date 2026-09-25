package router

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/server/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func middlewareDeadlineEngine(t *testing.T, timeout time.Duration, register func(*gin.RouterGroup)) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	engine := newEngine(nil)
	api := engine.Group(baseURL)
	api.Use(requestDeadline(timeout))
	register(api)

	return engine
}

func TestMiddleware_RequestDeadline_AnswersATimeoutWhenTheHandlerOutlivesIt(t *testing.T) {
	engine := middlewareDeadlineEngine(t, 100*time.Millisecond, func(api *gin.RouterGroup) {
		api.GET("/slow", func(c *gin.Context) {
			<-c.Done()
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, baseURL+"/slow", nil)

	done := make(chan struct{})
	go func() {
		engine.ServeHTTP(w, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("a request with no deadline of its own never ended")
	}

	require.Equal(t, http.StatusGatewayTimeout, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "RequestTimeout",
		"the caller must get a code it can show instead of a dropped connection")
}

func TestMiddleware_RequestDeadline_BoundsOnlyOrdinaryRequests(t *testing.T) {
	type deadlineCase struct {
		name      string
		method    string
		route     string
		websocket bool
		bounded   bool
	}

	cases := []deadlineCase{
		{name: "an ordinary request", method: http.MethodGet, route: baseURL + "/bounded", bounded: true},
		{name: "a websocket upgrade on a bounded route", method: http.MethodGet, route: baseURL + "/bounded", websocket: true},
	}
	for route := range unboundedRoutes {
		method, path, found := strings.Cut(route, " ")
		require.True(t, found, "an unbounded route is written as method and path")
		cases = append(cases, deadlineCase{name: "the exempt route " + route, method: method, route: path})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var deadline time.Time
			var hasDeadline bool

			engine := middlewareDeadlineEngine(t, RequestTimeout, func(api *gin.RouterGroup) {
				api.Handle(tc.method, strings.TrimPrefix(tc.route, baseURL), func(c *gin.Context) {
					deadline, hasDeadline = c.Deadline()
					c.Status(http.StatusOK)
				})
			})

			req := httptest.NewRequest(tc.method, strings.ReplaceAll(tc.route, ":flowID", "42"), nil)
			if tc.websocket {
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Upgrade", "websocket")
			}

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, tc.bounded, hasDeadline)
			if tc.bounded {
				assert.WithinDuration(t, time.Now().Add(RequestTimeout), deadline, 5*time.Second)
			}
		})
	}
}

func TestMiddleware_UnboundedRoutes_AreAllServedByTheRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := newEngine(nil)
	api := engine.Group(baseURL)

	setFlowFilesGroup(api, (*services.FlowFileService)(nil))
	setResourcesGroup(api, (*services.ResourceService)(nil))
	setGraphqlGroup(api, (*services.GraphqlService)(nil))

	served := make(map[string]struct{})
	for _, route := range engine.Routes() {
		served[route.Method+" "+route.Path] = struct{}{}
	}

	for route := range unboundedRoutes {
		assert.Contains(t, served, route, "an exempt route that nothing serves exempts nothing")
	}
}

func TestMiddleware_RequestTimeout_MatchesTheGraphQLDoorAndOutlastsAStatement(t *testing.T) {
	assert.Equal(t, RequestTimeout, services.OperationTimeout,
		"a caller must not have to know which door it came through to know how long it may wait")
	assert.Less(t, database.StatementTimeout, RequestTimeout,
		"a query outliving its request leaves the caller with nothing to be told")
}

const middlewareRangedBody = "0123456789abcdefghij"

// The download path keeps the production shape ending in /file: an image suffix would be excluded by extension.
func middlewareCompressionEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	fixture := filepath.Join(dir, "shot")
	require.NoError(t, os.WriteFile(fixture, []byte(middlewareRangedBody), 0o600))

	archive := filepath.Join(dir, "bundle.zip")
	require.NoError(t, os.WriteFile(archive, []byte(strings.Repeat("z", 4096)), 0o600))

	engine := gin.New()
	engine.Use(compressionMiddleware())

	api := engine.Group(baseURL)

	setGraphqlGroup(api, services.NewGraphqlService(
		nil, &config.Config{}, baseURL, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	))

	api.GET("/flows/:flowID/screenshots/:screenshotID/file", func(c *gin.Context) {
		c.File(fixture)
	})

	api.GET("/flows", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"flows": strings.Repeat("a flow row ", 200)})
	})

	engine.GET("/assets/bundle.zip", func(c *gin.Context) {
		c.File(archive)
	})

	return engine
}

type middlewareUnhijackableRecorder struct {
	*httptest.ResponseRecorder
}

func (r *middlewareUnhijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack unsupported in tests")
}

func middlewarePostGraphql(t *testing.T, engine *gin.Engine, acceptEncoding string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, baseURL+"/graphql", strings.NewReader(`{"query":"{__typename}"}`))
	req.Header.Set("Content-Type", "application/json")
	if acceptEncoding != "" {
		req.Header.Set("Accept-Encoding", acceptEncoding)
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	return rec
}

func TestMiddleware_CompressionMiddleware_GzipsAnswersButNotUpgradesRangesOrArchives(t *testing.T) {
	engine := middlewareCompressionEngine(t)

	t.Run("gzipped body decompresses to the identity response", func(t *testing.T) {
		plain := middlewarePostGraphql(t, engine, "")
		compressed := middlewarePostGraphql(t, engine, "gzip")

		require.Equal(t, http.StatusOK, plain.Code)
		require.Equal(t, http.StatusOK, compressed.Code)
		assert.Empty(t, plain.Header().Get("Content-Encoding"))
		assert.Equal(t, "gzip", compressed.Header().Get("Content-Encoding"))

		reader, err := gzip.NewReader(compressed.Body)
		require.NoError(t, err)
		inflated, err := io.ReadAll(reader)
		require.NoError(t, err)

		assert.Equal(t, plain.Body.String(), string(inflated))
	})

	t.Run("subscription upgrade is never wrapped", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, baseURL+"/graphql", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

		rec := &middlewareUnhijackableRecorder{ResponseRecorder: httptest.NewRecorder()}
		engine.ServeHTTP(rec, req)

		assert.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("a REST response is compressed too", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, baseURL+"/flows", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

		compressed := rec.Body.Len()

		reader, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)
		inflated, err := io.ReadAll(reader)
		require.NoError(t, err)

		assert.Contains(t, string(inflated), "a flow row")
		assert.Less(t, compressed, len(inflated))
	})

	t.Run("an already-compressed asset is served as it is", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/bundle.zip", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("sibling ranged download keeps its byte range uncompressed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, baseURL+"/flows/1/screenshots/2/file", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Range", "bytes=0-9")

		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		require.Equal(t, http.StatusPartialContent, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "bytes 0-9/20", rec.Header().Get("Content-Range"))
		assert.Equal(t, middlewareRangedBody[:10], rec.Body.String())
	})
}
