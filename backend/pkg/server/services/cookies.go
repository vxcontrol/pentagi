package services

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
)

// A forged scheme header can only mark the sender's own cookie Secure, so the
// headers are honoured regardless of TRUSTED_PROXIES.
func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if strings.EqualFold(firstListElement(r.Header.Get("X-Forwarded-Proto")), "https") {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Ssl")), "on") {
		return true
	}
	return strings.EqualFold(forwardedProto(r.Header.Get("Forwarded")), "https")
}

// Chained proxies append to the list, so the first element is the hop the browser reached.
func firstListElement(v string) string {
	first, _, _ := strings.Cut(v, ",")
	return strings.TrimSpace(first)
}

func forwardedProto(v string) string {
	for _, pair := range strings.Split(firstListElement(v), ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if ok && strings.EqualFold(strings.TrimSpace(key), "proto") {
			return strings.Trim(strings.TrimSpace(value), `"`)
		}
	}
	return ""
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
