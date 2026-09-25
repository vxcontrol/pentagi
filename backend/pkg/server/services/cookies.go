package services

import (
	"net/http"

	"github.com/gin-contrib/sessions"
)

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func sessionOptions(r *http.Request, basePath string, maxAge int) sessions.Options {
	return sessions.Options{
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		Path:     basePath,
		MaxAge:   maxAge,
	}
}
