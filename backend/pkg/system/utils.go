package system

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"pentagi/pkg/config"
)

const (
	// defaultHTTPClientTimeout is the fallback timeout when no config is provided.
	defaultHTTPClientTimeout = 10 * time.Minute
)

// Without them an unreachable host holds every attempt for the OS SYN timeout
// (about two minutes on Linux). They bound connecting only: a slow model answer
// is still limited by HTTPClientTimeout alone.
var (
	httpDialTimeout         = 30 * time.Second
	httpDialKeepAlive       = 30 * time.Second
	httpTLSHandshakeTimeout = 10 * time.Second
)

func getHostname() string {
	hn, err := os.Hostname()
	if err != nil {
		return ""
	}

	return hn
}

func GetSystemCertPool(cfg *config.Config) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get system cert pool: %w", err)
	}

	if cfg.ExternalSSLCAPath != "" {
		ca, err := os.ReadFile(cfg.ExternalSSLCAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read external CA certificate: %w", err)
		}

		if !pool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("failed to append external CA certificate to pool")
		}
	}

	return pool, nil
}

func GetHTTPClient(cfg *config.Config) (*http.Client, error) {
	var httpClient *http.Client

	if cfg == nil {
		return &http.Client{
			Timeout: defaultHTTPClientTimeout,
		}, nil
	}

	rootCAPool, err := GetSystemCertPool(cfg)
	if err != nil {
		return nil, err
	}

	// Convert timeout from config (in seconds) to time.Duration
	// 0 = no timeout (unlimited), >0 = timeout in seconds
	// Default value (600) is automatically set in config.go via envDefault:"600" tag
	// when HTTP_CLIENT_TIMEOUT environment variable is not set
	timeout := max(time.Duration(cfg.HTTPClientTimeout)*time.Second, 0)

	if cfg.ProxyURL != "" {
		httpClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(cfg.ProxyURL)
				},
				DialContext:         newHTTPDialer().DialContext,
				TLSHandshakeTimeout: httpTLSHandshakeTimeout,
				TLSClientConfig: &tls.Config{
					RootCAs:            rootCAPool,
					InsecureSkipVerify: cfg.ExternalSSLInsecure,
				},
			},
		}
	} else {
		httpClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext:         newHTTPDialer().DialContext,
				TLSHandshakeTimeout: httpTLSHandshakeTimeout,
				TLSClientConfig: &tls.Config{
					RootCAs:            rootCAPool,
					InsecureSkipVerify: cfg.ExternalSSLInsecure,
				},
			},
		}
	}

	return httpClient, nil
}

func newHTTPDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   httpDialTimeout,
		KeepAlive: httpDialKeepAlive,
	}
}
