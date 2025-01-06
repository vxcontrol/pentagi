package internal

import (
	"fmt"
	"net/http"
	"net/url"
)

// HTTPClient is an interface for a subset of the *http.Client.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// ResolveBaseURL resolves the base URL from the given arguments,
// preferring the first non-empty value.
func ResolveBaseURL(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// EncodeURL encodes the given arguments into the URL, escaping
// values as needed.
func EncodeURL(urlFormat string, args ...interface{}) string {
	escapedArgs := make([]interface{}, 0, len(args))
	for _, arg := range args {
		escapedArgs = append(escapedArgs, url.PathEscape(fmt.Sprintf("%v", arg)))
	}
	return fmt.Sprintf(urlFormat, escapedArgs...)
}

// MergeHeaders merges the given headers together, where the right
// takes precedence over the left.
func MergeHeaders(left, right http.Header) http.Header {
	for key, values := range right {
		if len(values) > 1 {
			left[key] = values
			continue
		}
		if value := right.Get(key); value != "" {
			left.Set(key, value)
		}
	}
	return left
}
