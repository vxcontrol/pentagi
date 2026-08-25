package oauth

import (
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

func newTestClient(opts ...ClientOption) OAuthClient {
	return NewOAuthClient("test", &oauth2.Config{
		ClientID:    "client-id",
		RedirectURL: "https://pentagi.example.com/api/v1/auth/login-callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://idp.example.com/auth",
			TokenURL: "https://idp.example.com/token",
		},
	}, nil, opts...)
}

func TestAuthCodeOptionsDefaultsToPlainCodeFlow(t *testing.T) {
	client := newTestClient()

	authURL, err := url.Parse(client.AuthCodeURL("state", client.AuthCodeOptions("nonce-value")...))
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	query := authURL.Query()
	if got := query.Get("nonce"); got != "nonce-value" {
		t.Errorf("nonce = %q, want %q", got, "nonce-value")
	}
	if got := query.Get("response_mode"); got != "" {
		t.Errorf("response_mode = %q, want empty for a plain code flow", got)
	}
	if got := query.Get("response_type"); got != "code" {
		t.Errorf("response_type = %q, want %q", got, "code")
	}
}

func TestAuthCodeOptionsWithFormPostCallback(t *testing.T) {
	client := newTestClient(WithFormPostCallback())

	authURL, err := url.Parse(client.AuthCodeURL("state", client.AuthCodeOptions("nonce-value")...))
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	query := authURL.Query()
	if got := query.Get("response_mode"); got != "form_post" {
		t.Errorf("response_mode = %q, want %q", got, "form_post")
	}
	if got := query.Get("response_type"); got != "code id_token" {
		t.Errorf("response_type = %q, want %q", got, "code id_token")
	}
}

func TestCallbackSameSite(t *testing.T) {
	tests := []struct {
		name string
		opts []ClientOption
		want http.SameSite
	}{
		{
			name: "get callback keeps cookies with Lax",
			want: http.SameSiteLaxMode,
		},
		{
			name: "cross-site form post requires None",
			opts: []ClientOption{WithFormPostCallback()},
			want: http.SameSiteNoneMode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := newTestClient(tc.opts...).CallbackSameSite(); got != tc.want {
				t.Errorf("CallbackSameSite() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGoogleClientUsesFormPostCallback(t *testing.T) {
	client := NewGoogleOAuthClient("client-id", "client-secret", "https://pentagi.example.com/callback")

	if got := client.CallbackSameSite(); got != http.SameSiteNoneMode {
		t.Errorf("google CallbackSameSite() = %v, want %v", got, http.SameSiteNoneMode)
	}
}

func TestGithubClientUsesGetCallback(t *testing.T) {
	client := NewGithubOAuthClient("client-id", "client-secret", "https://pentagi.example.com/callback")

	if got := client.CallbackSameSite(); got != http.SameSiteLaxMode {
		t.Errorf("github CallbackSameSite() = %v, want %v", got, http.SameSiteLaxMode)
	}
}
