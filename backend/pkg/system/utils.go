package system

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	"pentagi/pkg/config"
)

func getHostname() string {
	hn, err := os.Hostname()
	if err != nil {
		return ""
	}

	return hn
}

func getIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ips = append(ips, addr.String())
		}
	}

	return ips
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

	rootCAPool, err := GetSystemCertPool(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.ProxyURL != "" {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(cfg.ProxyURL)
				},
				TLSClientConfig: &tls.Config{
					RootCAs:            rootCAPool,
					InsecureSkipVerify: cfg.ExternalSSLInsecure,
				},
			},
		}
	} else {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:            rootCAPool,
					InsecureSkipVerify: cfg.ExternalSSLInsecure,
				},
			},
		}
	}

	return httpClient, nil
}
