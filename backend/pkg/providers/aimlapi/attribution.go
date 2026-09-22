package aimlapi

import (
	"net/http"
	"net/url"
	"strings"
)

// aimlapiDomain is the only domain that may receive the attribution headers below.
const aimlapiDomain = "aimlapi.com"

// Attribution headers for AI/ML API's partner programme.
//
// HTTP-Referer / X-Title are the OpenRouter convention and name the *calling*
// application — PentAGI — not the gateway. X-AIMLAPI-Source and
// X-AIMLAPI-Partner-ID are AI/ML API's own channel attribution; a malformed
// partner id is silently treated as untagged traffic rather than rejected,
// which is why its shape is asserted in a test instead of at runtime.
const (
	attributionReferer   = "https://github.com/vxcontrol/pentagi"
	attributionTitle     = "PentAGI"
	attributionSource    = "agent/pentagi"
	attributionPartnerID = "part_6bffrRIYBS8OtYbQhsEPi0SS"
)

// attributionHeaders builds a fresh map per call so no caller can mutate a
// package-level map shared by every provider instance.
func attributionHeaders() map[string]string {
	return map[string]string{
		"HTTP-Referer":         attributionReferer,
		"X-Title":              attributionTitle,
		"X-AIMLAPI-Source":     attributionSource,
		"X-AIMLAPI-Partner-ID": attributionPartnerID,
	}
}

// attributionHost returns the host allowed to receive attribution headers for
// baseURL, or "" to disable attribution entirely.
//
// Only aimlapi.com and its subdomains qualify. AIMLAPI_SERVER_URL can legally
// point at a LiteLLM proxy or a self-hosted gateway that fronts AI/ML API, and
// tagging those would attribute another operator's traffic to this integration,
// so anything else disables the headers rather than forwarding them.
func attributionHost(baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	host := strings.ToLower(parsed.Hostname())
	if host == aimlapiDomain || strings.HasSuffix(host, "."+aimlapiDomain) {
		return host
	}

	return ""
}

// attributionTransport tags requests bound for host with the attribution headers.
type attributionTransport struct {
	base    http.RoundTripper
	host    string
	headers map[string]string
}

func (t *attributionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	// Re-checked per request rather than once at construction: a redirect can
	// move a request off our origin, and the headers must not follow it.
	if t.host == "" || !strings.EqualFold(req.URL.Hostname(), t.host) {
		return base.RoundTrip(req)
	}

	// RoundTrip must not modify the request it was handed, so tag a clone. Values
	// already present win, so an operator-supplied header is never overwritten.
	tagged := req.Clone(req.Context())
	for key, value := range t.headers {
		if tagged.Header.Get(key) == "" {
			tagged.Header.Set(key, value)
		}
	}

	return base.RoundTrip(tagged)
}

// withAttribution returns a client that tags AI/ML API traffic. The client it is
// given is copied, never mutated: it is the process-wide client built from proxy
// and TLS settings, and every other provider is handed the same shape.
func withAttribution(client *http.Client, baseURL string) *http.Client {
	host := attributionHost(baseURL)
	if client == nil || host == "" {
		return client
	}

	tagged := *client
	tagged.Transport = &attributionTransport{
		base:    client.Transport,
		host:    host,
		headers: attributionHeaders(),
	}

	return &tagged
}
