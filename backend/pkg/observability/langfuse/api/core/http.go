package core

import "net/http"

// HTTPClient is an interface for a subset of the *http.Client.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}
