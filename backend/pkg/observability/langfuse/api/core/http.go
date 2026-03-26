package core

import "net/http"

// HTTPClient is an interface for a subset of the *http.Client.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Response is an HTTP response from an HTTP client.
type Response[T any] struct {
	StatusCode int
	Header     http.Header
	Body       T
}
