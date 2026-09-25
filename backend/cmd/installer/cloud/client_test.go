package cloud

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/vxcontrol/cloud/models"
)

// TestClient_UninitializedCallsFailCleanly: a failed New leaves a nil client; subtests are keyed by call.
func TestClient_UninitializedCallsFailCleanly(t *testing.T) {
	var client *Client
	ctx := context.Background()

	calls := []struct {
		name string
		call func() error
	}{
		{"check updates reports it", func() error {
			_, err := client.CheckUpdates(ctx, models.CheckUpdatesRequest{})
			return err
		}},
		{"package info reports it", func() error {
			_, err := client.PackageInfo(ctx, models.PackageInfoRequest{})
			return err
		}},
		{"download package reports it", func() error {
			return client.DownloadPackage(ctx, models.DownloadPackageRequest{}, models.PackageInfoResponse{}, io.Discard)
		}},
	}

	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil || !strings.Contains(err.Error(), "update client is not initialized") {
				t.Errorf("error = %v, want the client reported uninitialized", err)
			}
		})
	}
}

func TestClient_NormalizeHost_ReducesAnAddressToHostAndPort(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr string
	}{
		{name: "empty falls back to the default server", raw: "", want: "update.pentagi.com"},
		{name: "whitespace only falls back too", raw: "   ", want: "update.pentagi.com"},
		{name: "bare host passes through", raw: "update.pentagi.com", want: "update.pentagi.com"},
		{name: "port is kept", raw: "update.pentagi.com:8443", want: "update.pentagi.com:8443"},
		// The setting has always held a full URL, so this is what most installations carry.
		{name: "https scheme is stripped", raw: "https://update.pentagi.com", want: "update.pentagi.com"},
		{name: "trailing slash is stripped", raw: "https://update.pentagi.com/", want: "update.pentagi.com"},
		{name: "path is stripped", raw: "https://update.pentagi.com/api/v1", want: "update.pentagi.com"},
		{name: "query is stripped", raw: "https://update.pentagi.com?x=1", want: "update.pentagi.com"},
		{name: "case is normalised", raw: "HTTPS://Update.PentAGI.com", want: "update.pentagi.com"},

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
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("NormalizeHost(%q) = %q, %v; want an error mentioning %q", tt.raw, got, err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("NormalizeHost(%q) = %q, %v; want %q", tt.raw, got, err, tt.want)
			}
		})
	}
}

// The installer's proxy often carries credentials the environment does not, so it has to win.
func TestClient_NewTransport_PrefersTheConfiguredProxyOverTheEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		proxy    string
		wantHost string
		wantUser string
		wantErr  string
	}{
		{name: "a configured proxy wins and keeps its credentials", proxy: "http://user:secret@proxy.internal:3128", wantHost: "proxy.internal:3128", wantUser: "user"},
		{name: "without one the environment proxy is honoured", proxy: "", wantHost: "env-proxy.internal:8080"},
		{name: "an unparsable proxy is refused", proxy: "://nope", wantErr: "not a valid URL"},
		{name: "a proxy without a host is refused", proxy: "not a url at all", wantErr: "has no host"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HTTPS_PROXY", "http://env-proxy.internal:8080")

			transport, err := newTransport(tt.proxy)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("newTransport(%q) error = %v, want it to mention %q", tt.proxy, err, tt.wantErr)
				}
				return
			}
			if err != nil || transport.Proxy == nil {
				t.Fatalf("newTransport(%q) = %v, %v; want a proxy resolver", tt.proxy, transport, err)
			}

			request, err := http.NewRequest(http.MethodGet, "https://update.pentagi.com/", nil)
			if err != nil {
				t.Fatalf("NewRequest: %v", err)
			}
			proxy, err := transport.Proxy(request)
			if err != nil || proxy == nil || proxy.Host != tt.wantHost {
				t.Fatalf("proxy = %v, %v; want host %q", proxy, err, tt.wantHost)
			}
			if tt.wantUser != "" && (proxy.User == nil || proxy.User.Username() != tt.wantUser) {
				t.Errorf("proxy credentials were dropped: %v", proxy.User)
			}
		})
	}
}

// An unusable license key is not fatal: the update check still works anonymously and says why.
func TestClient_New_BuildsAClientOrRefusesItsConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		wantErr     string
		wantWarning string
		wantVersion string
	}{
		{name: "a zero config talks to the default server anonymously", cfg: Config{}, wantVersion: "0.0.0"},
		{name: "a development build is sent as the unknown version", cfg: Config{InstallerVersion: "develop-87ac00f"}, wantVersion: "0.0.0"},
		{name: "an unusable license key is reported, not fatal", cfg: Config{LicenseKey: "NOT-A-REAL-LICENSE"}, wantWarning: "license key is not usable", wantVersion: "0.0.0"},
		{name: "a plaintext update server is refused", cfg: Config{Host: "http://update.pentagi.com"}, wantErr: "always encrypted"},
		{name: "an unusable proxy is refused", cfg: Config{ProxyURL: "://nope"}, wantErr: "not a valid URL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.cfg)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("New() error = %v, want it to mention %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			warning := client.LicenseWarning()
			if (warning == nil) != (tt.wantWarning == "") || (warning != nil && !strings.Contains(warning.Error(), tt.wantWarning)) {
				t.Errorf("LicenseWarning() = %v, want %q", warning, tt.wantWarning)
			}
			if client.checkUpdates == nil || client.packageInfo == nil || client.downloadPackage == nil {
				t.Error("New() left a call unbuilt")
			}
			if client.Host() != "update.pentagi.com" || client.InstallerVersion() != tt.wantVersion {
				t.Errorf("Host(), InstallerVersion() = %q, %q; want update.pentagi.com, %q", client.Host(), client.InstallerVersion(), tt.wantVersion)
			}
		})
	}
}
