// Package cloud is the installer's client for the PentAGI Cloud API: update checks and
// package downloads.
//
// It is a thin layer over the published SDK. Wire types come from the SDK's models package
// rather than being redeclared here, so a contract change is a compile error instead of a
// silent mismatch, and the whole package is free of installer state — it takes a Config,
// returns answers, and knows nothing about the wizard, the checker or the processor. Tests
// substitute the call functions directly; nothing here needs a server to exercise.
package cloud

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vxcontrol/cloud/sdk"
	"github.com/vxcontrol/cloud/system"
)

// DefaultHost is the update server used when none is configured.
const DefaultHost = "update.pentagi.com"

// Route names and paths are part of the API contract: the server dispatches on the name,
// and a name it does not know is refused outright rather than reported as a missing page.
const (
	routeUpdatesCheck     = "updates_check"
	routePackagesInfo     = "packages_info"
	routePackagesDownload = "packages_download"

	pathUpdatesCheck     = "/api/v1/proxy/updates/check"
	pathPackagesInfo     = "/api/v1/proxy/packages/info"
	pathPackagesDownload = "/api/v1/proxy/packages/download"
)

const (
	// clientName identifies the installer in the User-Agent, alongside its build version.
	clientName = "PentAGI-Installer"

	// powTimeout is the budget for the proof-of-work challenge each call has to solve.
	// It is raised well above the SDK default because the SDK does not retry a timed-out
	// challenge, and a laptop or a throttled VM can genuinely need the extra seconds.
	powTimeout = 60 * time.Second

	maxRetries = 3
)

// Config describes how to reach the update server. Every field is optional; the zero
// Config talks to DefaultHost anonymously, with no proxy.
type Config struct {
	// Host is the update server, with or without a scheme (UPDATE_SERVER_HOST).
	Host string
	// ProxyURL is the installer's own proxy (PROXY_URL), used for every call. It takes
	// precedence over the environment, which is what the user configured it for.
	ProxyURL string
	// LicenseKey is the PentAGI license key (LICENSE_KEY). An unusable key is dropped
	// rather than sent, and reported through LicenseWarning.
	LicenseKey string
	// InstallerVersion is this build's version, as displayed. It is normalised for the
	// request and sent verbatim in the User-Agent.
	InstallerVersion string
	// Logger receives the SDK's own diagnostics. Nil keeps them silent.
	Logger sdk.Logger
}

// Client talks to the update server. Build one with New.
type Client struct {
	checkUpdates    sdk.CallReqBytesRespBytes
	packageInfo     sdk.CallReqQueryRespBytes
	downloadPackage sdk.CallReqQueryRespWriter

	host           string
	version        string
	licenseWarning error
}

// New builds a client for cfg.
//
// It fails only on configuration that cannot produce working calls — an unusable host or
// proxy. A license key that does not hold up is deliberately not fatal: the update check is
// the one thing that still works without a license, and refusing to run would leave the
// user with no way to learn that their key is the problem. The key is dropped and the
// reason is available from LicenseWarning.
func New(cfg Config) (*Client, error) {
	host, err := NormalizeHost(cfg.Host)
	if err != nil {
		return nil, err
	}

	transport, err := newTransport(cfg.ProxyURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		host:    host,
		version: NormalizeVersion(cfg.InstallerVersion),
	}

	userAgentVersion := strings.TrimSpace(cfg.InstallerVersion)
	if userAgentVersion == "" {
		userAgentVersion = UnknownVersion
	}

	options := []sdk.Option{
		sdk.WithTransport(transport),
		// The installation ID identifies this machine to the server across calls, and the
		// contract requires it on every one of them.
		sdk.WithInstallationID(system.GetInstallationID()),
		sdk.WithClient(clientName, userAgentVersion),
		sdk.WithPowTimeout(powTimeout),
		sdk.WithMaxRetries(maxRetries),
	}
	if cfg.Logger != nil {
		options = append(options, sdk.WithLogger(cfg.Logger))
	}

	// Validate before handing the key over: the SDK drops a key it cannot decode without
	// saying so, and the call then runs anonymously — which looks, from the outside, like
	// the license simply not being honoured.
	if key := strings.TrimSpace(cfg.LicenseKey); key != "" {
		if _, err := sdk.IntrospectLicenseKey(key); err != nil {
			client.licenseWarning = fmt.Errorf("license key is not usable, continuing without it: %w", err)
		} else {
			options = append(options, sdk.WithLicenseKey(key))
		}
	}

	configs := []sdk.CallConfig{
		{
			Calls:  []any{&client.checkUpdates},
			Host:   host,
			Name:   routeUpdatesCheck,
			Path:   pathUpdatesCheck,
			Method: sdk.CallMethodPOST,
		},
		{
			Calls:  []any{&client.packageInfo},
			Host:   host,
			Name:   routePackagesInfo,
			Path:   pathPackagesInfo,
			Method: sdk.CallMethodGET,
		},
		{
			Calls:  []any{&client.downloadPackage},
			Host:   host,
			Name:   routePackagesDownload,
			Path:   pathPackagesDownload,
			Method: sdk.CallMethodGET,
		},
	}

	if err := sdk.Build(configs, options...); err != nil {
		return nil, fmt.Errorf("failed to build update client: %w", err)
	}

	return client, nil
}

// Host reports the normalised update server this client talks to.
func (c *Client) Host() string { return c.host }

// InstallerVersion reports the version this client sends, after normalisation. It differs
// from the displayed build version whenever that one cannot be expressed as a version
// number.
func (c *Client) InstallerVersion() string { return c.version }

// LicenseWarning reports why the configured license key was not used, or nil when it was
// used or none was configured. Calls still work — anonymously.
func (c *Client) LicenseWarning() error { return c.licenseWarning }

// NormalizeHost reduces a configured update server address to the host[:port] form the
// client needs.
//
// The setting has historically held a full URL, so a scheme and a trailing path are
// tolerated and stripped. The transport is always TLS, so an explicit "http://" is an
// error rather than something to silently upgrade — a user who wrote it meant it, and
// quietly doing the opposite would hide a misconfiguration.
func NormalizeHost(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return DefaultHost, nil
	}

	if scheme, rest, found := strings.Cut(value, "://"); found {
		switch strings.ToLower(scheme) {
		case "https":
			value = rest
		case "http":
			return "", fmt.Errorf("update server %q uses http, but the connection is always encrypted: use https", raw)
		default:
			return "", fmt.Errorf("update server %q uses unsupported scheme %q", raw, scheme)
		}
	}

	// Drop anything after the authority: a path, a query, a fragment.
	value, _, _ = strings.Cut(value, "/")
	value, _, _ = strings.Cut(value, "?")
	value, _, _ = strings.Cut(value, "#")
	value = strings.ToLower(strings.TrimSpace(value))

	switch {
	case value == "":
		return "", fmt.Errorf("update server %q has no host", raw)
	case strings.Contains(value, "@"):
		return "", fmt.Errorf("update server %q carries credentials, which are not supported", raw)
	case strings.ContainsAny(value, " \t"):
		return "", fmt.Errorf("update server %q contains whitespace", raw)
	}

	return value, nil
}

// newTransport builds the transport every call uses.
//
// The SDK's default reads the proxy from the environment. The installer keeps its own
// proxy setting — often with credentials the environment does not carry — and that setting
// has to win. With none configured the environment behaviour is left as it is.
func newTransport(proxyURL string) (*http.Transport, error) {
	transport := sdk.DefaultTransport()

	value := strings.TrimSpace(proxyURL)
	if value == "" {
		return transport, nil
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("proxy %q is not a valid URL: %w", proxyURL, err)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("proxy %q has no host", proxyURL)
	}

	transport.Proxy = http.ProxyURL(parsed)

	return transport, nil
}
