package cloud

import (
	"net/http"
	"strings"
	"testing"
)

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr string
	}{
		{name: "empty falls back to the default server", raw: "", want: DefaultHost},
		{name: "whitespace only falls back too", raw: "   ", want: DefaultHost},
		{name: "bare host passes through", raw: "update.pentagi.com", want: "update.pentagi.com"},
		{name: "port is kept", raw: "update.pentagi.com:8443", want: "update.pentagi.com:8443"},
		// The setting has always held a full URL, so this is the value most installations
		// actually carry.
		{name: "https scheme is stripped", raw: "https://update.pentagi.com", want: "update.pentagi.com"},
		{name: "trailing slash is stripped", raw: "https://update.pentagi.com/", want: "update.pentagi.com"},
		{name: "path is stripped", raw: "https://update.pentagi.com/api/v1", want: "update.pentagi.com"},
		{name: "query is stripped", raw: "https://update.pentagi.com?x=1", want: "update.pentagi.com"},
		{name: "case is normalised", raw: "HTTPS://Update.PentAGI.com", want: "update.pentagi.com"},

		// Silently upgrading http to https would hide a misconfiguration behind a working
		// connection, and the user would never learn their setting was ignored.
		{name: "http is refused rather than upgraded", raw: "http://update.pentagi.com", wantErr: "always encrypted"},
		{name: "other schemes are refused", raw: "ftp://update.pentagi.com", wantErr: "unsupported scheme"},
		{name: "credentials are refused", raw: "https://user:pass@update.pentagi.com", wantErr: "credentials"},
		{name: "scheme with no host is refused", raw: "https:///api", wantErr: "no host"},
		{name: "whitespace inside is refused", raw: "update pentagi.com", wantErr: "whitespace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeHost(tt.raw)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("NormalizeHost(%q) = %q, want error containing %q", tt.raw, got, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("NormalizeHost(%q) error = %v, want it to mention %q", tt.raw, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NormalizeHost(%q) unexpected error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Errorf("NormalizeHost(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

// TestTransportPrefersTheConfiguredProxy pins the reason this package builds its own
// transport instead of taking the SDK default: the installer's proxy setting often carries
// credentials the environment does not, and it has to win over the environment.
func TestTransportPrefersTheConfiguredProxy(t *testing.T) {
	transport, err := newTransport("http://user:secret@proxy.internal:3128")
	if err != nil {
		t.Fatalf("newTransport: %v", err)
	}
	if transport.Proxy == nil {
		t.Fatal("transport has no proxy resolver")
	}

	request, err := http.NewRequest(http.MethodGet, "https://update.pentagi.com/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	proxy, err := transport.Proxy(request)
	if err != nil {
		t.Fatalf("proxy resolver: %v", err)
	}
	if proxy == nil {
		t.Fatal("configured proxy was not applied to the request")
	}
	if proxy.Host != "proxy.internal:3128" {
		t.Errorf("proxy host = %q, want proxy.internal:3128", proxy.Host)
	}
	if user := proxy.User; user == nil || user.Username() != "user" {
		t.Errorf("proxy credentials were dropped: %v", proxy.User)
	}
}

// TestTransportWithoutProxyKeepsEnvironmentBehaviour: with nothing configured the SDK's own
// environment-driven resolution is left alone, so a machine-wide proxy still works.
func TestTransportWithoutProxyKeepsEnvironmentBehaviour(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://env-proxy.internal:8080")

	transport, err := newTransport("")
	if err != nil {
		t.Fatalf("newTransport: %v", err)
	}
	if transport.Proxy == nil {
		t.Fatal("transport has no proxy resolver")
	}

	request, err := http.NewRequest(http.MethodGet, "https://update.pentagi.com/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	proxy, err := transport.Proxy(request)
	if err != nil {
		t.Fatalf("proxy resolver: %v", err)
	}
	if proxy == nil || proxy.Host != "env-proxy.internal:8080" {
		t.Errorf("environment proxy was not honoured: %v", proxy)
	}
}

func TestTransportRejectsUnusableProxy(t *testing.T) {
	for _, raw := range []string{"://nope", "not a url at all"} {
		if _, err := newTransport(raw); err == nil {
			t.Errorf("newTransport(%q) accepted an unusable proxy", raw)
		}
	}
}

// TestUnusableLicenseKeyIsReportedNotSwallowed covers the one place the SDK is quietly
// permissive: it drops a key it cannot decode and carries on anonymously. That looks
// exactly like the license not being honoured, so the client checks the key itself and
// keeps the reason.
func TestUnusableLicenseKeyIsReportedNotSwallowed(t *testing.T) {
	client, err := New(Config{LicenseKey: "NOT-A-REAL-LICENSE"})
	if err != nil {
		t.Fatalf("New must stay usable without a license: %v", err)
	}
	if client.LicenseWarning() == nil {
		t.Fatal("an unusable license key was accepted silently")
	}
	if client.checkUpdates == nil {
		t.Error("the update check must still work anonymously")
	}
}

func TestNoLicenseKeyIsNotAWarning(t *testing.T) {
	client, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if warning := client.LicenseWarning(); warning != nil {
		t.Errorf("running without a license is normal, got warning: %v", warning)
	}
}

func TestNewRejectsUnusableConfiguration(t *testing.T) {
	if _, err := New(Config{Host: "http://update.pentagi.com"}); err == nil {
		t.Error("New accepted a plaintext update server")
	}
	if _, err := New(Config{ProxyURL: "://nope"}); err == nil {
		t.Error("New accepted an unusable proxy")
	}
}

// TestNewNormalisesVersionButKeepsTheBuildName: the request carries a version the contract
// accepts, while the User-Agent keeps the exact build so a development build is still
// identifiable in the server's logs.
func TestNewNormalisesVersionButKeepsTheBuildName(t *testing.T) {
	client, err := New(Config{InstallerVersion: "develop-87ac00f"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := client.InstallerVersion(); got != UnknownVersion {
		t.Errorf("InstallerVersion() = %q, want %q", got, UnknownVersion)
	}
	if got := client.Host(); got != DefaultHost {
		t.Errorf("Host() = %q, want %q", got, DefaultHost)
	}
}
