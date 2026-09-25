package router

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// RequestTimeout must stay shorter than the deadline the browser client uses:
// the caller is meant to see this server's answer, not an abort of its own on a
// request still being served here.
const RequestTimeout = 30 * time.Second

// unboundedRoutes are left alone: moving a file takes as long as the file is
// large, and the GraphQL door bounds each operation itself.
var unboundedRoutes = map[string]struct{}{
	"GET " + baseURL + "/graphql":                           {},
	"POST " + baseURL + "/graphql":                          {},
	"POST " + baseURL + "/flows/:flowID/files/":             {},
	"POST " + baseURL + "/flows/:flowID/files/pull":         {},
	"POST " + baseURL + "/flows/:flowID/files/resources":    {},
	"POST " + baseURL + "/flows/:flowID/files/to-resources": {},
	"GET " + baseURL + "/flows/:flowID/files/container":     {},
	"GET " + baseURL + "/flows/:flowID/files/download":      {},
	"POST " + baseURL + "/resources/":                       {},
	"GET " + baseURL + "/resources/download":                {},
}

func isUnboundedRoute(c *gin.Context) bool {
	_, unbounded := unboundedRoutes[c.Request.Method+" "+c.FullPath()]

	return unbounded
}

func isWebsocketUpgrade(c *gin.Context) bool {
	return strings.EqualFold(c.GetHeader("Upgrade"), "websocket")
}

func requestDeadline(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isUnboundedRoute(c) || isWebsocketUpgrade(c) {
			c.Next()

			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if errors.Is(ctx.Err(), context.DeadlineExceeded) && !c.Writer.Written() {
			response.Error(c, response.ErrRequestTimeout, ctx.Err())
		}
	}
}

func localUserRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.IsAborted() {
			return
		}

		session := sessions.Default(c)
		tid, ok := session.Get("tid").(string)

		if !ok || tid != models.UserTypeLocal.String() {
			response.Error(c, response.ErrLocalUserRequired, nil)
			return
		}

		c.Next()
	}
}

var alreadyCompressedExtensions = []string{
	".br", ".gif", ".gz", ".jpeg", ".jpg", ".png", ".webp", ".woff", ".woff2", ".zip",
}

func shouldCompressResponse(c *gin.Context) bool {
	req := c.Request

	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		req.Header.Get("Range") != "" {
		return false
	}

	return !slices.Contains(alreadyCompressedExtensions, strings.ToLower(filepath.Ext(req.URL.Path)))
}

func compressionMiddleware() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression, gzip.WithCustomShouldCompressFn(shouldCompressResponse))
}

func noCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1
		c.Header("Pragma", "no-cache")                                   // HTTP 1.0
		c.Header("Expires", "0")                                         // prevents caching at the proxy server
		c.Next()
	}
}

// staticCacheMiddleware sets cache policy for the locally-served SPA build:
// content-hashed /assets/* are immutable; everything else resolving to the SPA
// (index.html, client routes) is no-cache, so a redeploy is picked up instead of
// replaying a stale index.html that imports chunks the deploy already deleted.
func staticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet && !strings.HasPrefix(c.Request.URL.Path, baseURL) {
			if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				c.Header("Cache-Control", "no-cache")
			}
		}

		c.Next()
	}
}
