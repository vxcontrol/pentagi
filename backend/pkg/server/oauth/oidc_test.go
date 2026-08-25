package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newDiscoveryServer serves a minimal OpenID Connect discovery document, which
// is all NewOIDCOAuthClient needs to configure itself.
func newDiscoveryServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                server.URL,
			"authorization_endpoint":                server.URL + "/protocol/openid-connect/auth",
			"token_endpoint":                        server.URL + "/protocol/openid-connect/token",
			"userinfo_endpoint":                     server.URL + "/protocol/openid-connect/userinfo",
			"jwks_uri":                              server.URL + "/protocol/openid-connect/certs",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})

	t.Cleanup(server.Close)

	return server
}

func TestNewOIDCOAuthClientUsesDiscoveredEndpoints(t *testing.T) {
	server := newDiscoveryServer(t)

	client, err := NewOIDCOAuthClient(
		context.Background(),
		server.URL,
		"pentagi",
		"client-secret",
		"https://pentagi.example.com/api/v1/auth/login-callback",
		nil,
	)
	if err != nil {
		t.Fatalf("NewOIDCOAuthClient() failed: %v", err)
	}

	if got := client.ProviderName(); got != OIDCProviderName {
		t.Errorf("ProviderName() = %q, want %q", got, OIDCProviderName)
	}

	// A generic provider answers on a GET callback, so Lax cookies survive it.
	if got := client.CallbackSameSite(); got != http.SameSiteLaxMode {
		t.Errorf("CallbackSameSite() = %v, want %v", got, http.SameSiteLaxMode)
	}

	authURL := client.AuthCodeURL("state", client.AuthCodeOptions("nonce-value")...)
	parsed, err := parseQuery(authURL)
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	if want := server.URL + "/protocol/openid-connect/auth"; !strings.HasPrefix(authURL, want) {
		t.Errorf("authorization URL = %q, want prefix %q", authURL, want)
	}
	if got := parsed.Get("nonce"); got != "nonce-value" {
		t.Errorf("nonce = %q, want %q", got, "nonce-value")
	}
	if got := parsed.Get("response_type"); got != "code" {
		t.Errorf("response_type = %q, want %q", got, "code")
	}
	if got := parsed.Get("scope"); got != "openid email profile" {
		t.Errorf("scope = %q, want %q", got, "openid email profile")
	}
}

func TestNewOIDCOAuthClientHonoursCustomScopes(t *testing.T) {
	server := newDiscoveryServer(t)

	client, err := NewOIDCOAuthClient(
		context.Background(),
		server.URL,
		"pentagi",
		"client-secret",
		"https://pentagi.example.com/api/v1/auth/login-callback",
		[]string{"openid", "email"},
	)
	if err != nil {
		t.Fatalf("NewOIDCOAuthClient() failed: %v", err)
	}

	parsed, err := parseQuery(client.AuthCodeURL("state"))
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	if got := parsed.Get("scope"); got != "openid email" {
		t.Errorf("scope = %q, want %q", got, "openid email")
	}
}

func TestNewOIDCOAuthClientFailsOnUnreachableIssuer(t *testing.T) {
	server := newDiscoveryServer(t)
	issuer := server.URL
	server.Close()

	_, err := NewOIDCOAuthClient(
		context.Background(),
		issuer,
		"pentagi",
		"client-secret",
		"https://pentagi.example.com/api/v1/auth/login-callback",
		nil,
	)
	if err == nil {
		t.Fatal("NewOIDCOAuthClient() succeeded for an unreachable issuer, want error")
	}
}

func parseQuery(rawURL string) (url.Values, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return parsed.Query(), nil
}
