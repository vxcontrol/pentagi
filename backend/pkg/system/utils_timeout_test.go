package system

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// shortenConnectTimeouts makes the connect bounds small enough for a test to wait them out.
func shortenConnectTimeouts(t *testing.T, dial, handshake time.Duration) {
	t.Helper()

	savedDial, savedHandshake := httpDialTimeout, httpTLSHandshakeTimeout
	httpDialTimeout, httpTLSHandshakeTimeout = dial, handshake
	t.Cleanup(func() {
		httpDialTimeout, httpTLSHandshakeTimeout = savedDial, savedHandshake
	})
}

func TestGetHTTPClient_BoundsConnectingOnly(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}
	caPath := createTempFile(t, certs.rootCAPEM)

	tests := []struct {
		name     string
		insecure bool
		proxyURL string
	}{
		{name: "direct", insecure: false},
		{name: "direct insecure", insecure: true},
		{name: "through a proxy", proxyURL: "http://proxy.example.com:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := GetHTTPClient(createTestConfig(caPath, tt.insecure, tt.proxyURL))
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			transport, ok := client.Transport.(*http.Transport)
			if !ok {
				t.Fatal("expected http.Transport")
			}

			if transport.DialContext == nil {
				t.Error("expected a bounded dialer")
			}
			if transport.TLSHandshakeTimeout != httpTLSHandshakeTimeout {
				t.Errorf("expected TLS handshake timeout %v, got %v", httpTLSHandshakeTimeout, transport.TLSHandshakeTimeout)
			}
			if transport.ResponseHeaderTimeout != 0 {
				t.Errorf("a slow model answer must not be cut off, got response header timeout %v", transport.ResponseHeaderTimeout)
			}
			if client.Timeout != 600*time.Second {
				t.Errorf("expected the configured overall timeout, got %v", client.Timeout)
			}

			if transport.TLSClientConfig == nil || transport.TLSClientConfig.RootCAs == nil {
				t.Fatal("expected the custom root CA pool to be kept")
			}
			if transport.TLSClientConfig.InsecureSkipVerify != tt.insecure {
				t.Errorf("expected InsecureSkipVerify %v", tt.insecure)
			}

			if tt.proxyURL == "" {
				if transport.Proxy != nil {
					t.Error("expected no proxy")
				}
				return
			}
			proxy, err := transport.Proxy(httptest.NewRequest(http.MethodGet, "https://api.openai.com/v1/models", nil))
			if err != nil || proxy == nil || proxy.String() != tt.proxyURL {
				t.Errorf("expected proxy %s, got %v (%v)", tt.proxyURL, proxy, err)
			}
		})
	}
}

func TestGetHTTPClient_SilentServerFailsTheTLSHandshakeInTime(t *testing.T) {
	shortenConnectTimeouts(t, time.Second, 200*time.Millisecond)

	// Accepts the connection and never answers the ClientHello.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	t.Cleanup(func() { listener.Close() })
	go func() {
		var held []net.Conn
		defer func() {
			for _, conn := range held {
				conn.Close()
			}
		}()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			held = append(held, conn)
		}
	}()

	client, err := GetHTTPClient(createTestConfig("", false, ""))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// The guard only keeps a missing bound from hanging the test for minutes.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+listener.Addr().String(), nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected the handshake to fail")
	}
	if !strings.Contains(err.Error(), "TLS handshake timeout") {
		t.Errorf("expected a TLS handshake timeout, got: %v", err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("expected to give up within the handshake timeout, took %v", elapsed)
	}
}

func TestGetHTTPClient_SlowAnswerIsNotCutOff(t *testing.T) {
	shortenConnectTimeouts(t, 100*time.Millisecond, 100*time.Millisecond)

	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(time.Second) // ten times both connect bounds, like a model thinking before it answers
		_, _ = io.WriteString(w, "OK")
	}))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{{
		Certificate: [][]byte{certs.serverCert.Raw, certs.intermediate.Raw},
		PrivateKey:  certs.serverKey,
	}}}
	server.StartTLS()
	t.Cleanup(server.Close)

	client, err := GetHTTPClient(createTestConfig(createTempFile(t, certs.rootCAPEM), false, ""))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("a slow answer must not be cut off: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil || string(body) != "OK" {
		t.Errorf("expected body OK, got %q (%v)", body, err)
	}
}

func TestGetHTTPClient_RequestStillGoesThroughTheProxy(t *testing.T) {
	shortenConnectTimeouts(t, time.Second, time.Second)

	var seen string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = io.WriteString(w, "via proxy")
	}))
	t.Cleanup(proxy.Close)

	client, err := GetHTTPClient(createTestConfig("", false, proxy.URL))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	resp, err := client.Get("http://llm.example.invalid/v1/models")
	if err != nil {
		t.Fatalf("expected the proxy to answer, got: %v", err)
	}
	defer resp.Body.Close()

	if seen != "http://llm.example.invalid/v1/models" {
		t.Errorf("expected the proxy to receive the request, got %q", seen)
	}
}
