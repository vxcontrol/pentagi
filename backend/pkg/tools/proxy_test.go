package tools

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"
)

var _ SummarizeHandler = testSummarizerHandler

// testSummarizerHandler implements a simple mock summarizer
func testSummarizerHandler(ctx context.Context, result string) (string, error) {
	return "test summarized: " + result, nil
}

var _ SearchLogProvider = &searchLogProviderMock{}

type searchLogProviderMock struct {
	calls      int64
	engine     database.SearchengineType
	query      string
	result     string
	taskID     *int64
	subtaskID  *int64
	parentType database.MsgchainType
	currType   database.MsgchainType
}

func (m *searchLogProviderMock) PutLog(
	_ context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	engine database.SearchengineType,
	query string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	m.calls++
	m.parentType = initiator
	m.currType = executor
	m.engine = engine
	m.query = query
	m.result = result
	m.taskID = taskID
	m.subtaskID = subtaskID
	return m.calls, nil
}

// testProxy is a MITM HTTP/HTTPS proxy server for unit testing that intercepts
// requests to a specific domain and redirects them to a mock HTTP server.
type testProxy struct {
	proxyServer  *http.Server
	mockServer   *http.Server
	proxyURL     string
	mockURL      string
	targetDomain string
	caCert       *x509.Certificate
	caKey        *rsa.PrivateKey
	caPEM        []byte
	caFilePath   string
	certCache    sync.Map // map[string]*tls.Certificate - cache of generated certificates by host
	mu           sync.Mutex
	closed       bool
}

// newTestProxy creates a new test proxy server that intercepts requests to targetDomain
// and redirects them to a mock server with the provided handler.
// Both servers run on random available ports.
// The proxy supports both HTTP and HTTPS (via MITM with generated CA certificate).
func newTestProxy(targetDomain string, mockHandler http.Handler) (*testProxy, error) {
	targetDomain = strings.ToLower(strings.TrimSpace(targetDomain))
	if targetDomain == "" {
		return nil, errors.New("target domain cannot be empty")
	}
	if mockHandler == nil {
		return nil, errors.New("mock handler cannot be nil")
	}

	// Generate CA certificate for MITM
	caCert, caKey, caPEM, err := generateCA()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA: %w", err)
	}

	// Write CA cert to temporary file
	tempDir := os.TempDir()
	caFilePath := filepath.Join(tempDir, fmt.Sprintf("test-proxy-ca-%d.pem", time.Now().UnixNano()))
	if err := os.WriteFile(caFilePath, caPEM, 0644); err != nil {
		return nil, fmt.Errorf("failed to write CA cert to temp file: %w", err)
	}

	proxy := &testProxy{
		targetDomain: targetDomain,
		caCert:       caCert,
		caKey:        caKey,
		caPEM:        caPEM,
		caFilePath:   caFilePath,
	}

	// Start mock server on random port
	mockListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create mock listener: %w", err)
	}

	proxy.mockServer = &http.Server{
		Handler:      mockHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	proxy.mockURL = fmt.Sprintf("http://%s", mockListener.Addr().String())

	go func() {
		if err := proxy.mockServer.Serve(mockListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("mock server error: %v", err))
		}
	}()

	// Start proxy server on random port
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		proxy.mockServer.Close()
		return nil, fmt.Errorf("failed to create proxy listener: %w", err)
	}

	proxyHandler := proxy.createProxyHandler()
	proxy.proxyServer = &http.Server{
		Handler:      proxyHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	proxy.proxyURL = fmt.Sprintf("http://%s", proxyListener.Addr().String())

	go func() {
		if err := proxy.proxyServer.Serve(proxyListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("proxy server error: %v", err))
		}
	}()

	// Wait for servers to be ready
	time.Sleep(100 * time.Millisecond)

	return proxy, nil
}

// URL returns the proxy server URL that can be used in HTTP client configuration.
func (p *testProxy) URL() string {
	return p.proxyURL
}

// MockURL returns the mock server URL for internal testing purposes.
func (p *testProxy) MockURL() string {
	return p.mockURL
}

// CACertPEM returns the CA certificate in PEM format.
// This can be used to configure HTTP clients to trust the proxy's MITM certificate.
func (p *testProxy) CACertPEM() []byte {
	return p.caPEM
}

// CACertPath returns the path to the CA certificate file.
// This can be used with config.ExternalSSLCAPath.
// The file is automatically cleaned up when Close() is called.
func (p *testProxy) CACertPath() string {
	return p.caFilePath
}

// Close shuts down both proxy and mock servers and cleans up the CA certificate file.
func (p *testProxy) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var errs []error
	if p.proxyServer != nil {
		if err := p.proxyServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("proxy server shutdown: %w", err))
		}
	}
	if p.mockServer != nil {
		if err := p.mockServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("mock server shutdown: %w", err))
		}
	}

	// Clean up CA certificate file
	if p.caFilePath != "" {
		if err := os.Remove(p.caFilePath); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("failed to remove CA cert file: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// createProxyHandler creates the proxy handler that intercepts requests to targetDomain.
func (p *testProxy) createProxyHandler() http.Handler {
	mockURL, _ := url.Parse(p.mockURL)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a CONNECT request (for HTTPS)
		if r.Method == http.MethodConnect {
			p.handleConnect(w, r, mockURL)
			return
		}

		// Check if request is for target domain
		host := r.URL.Hostname()
		if host == "" {
			host = r.Host
		}
		// Normalize host to lowercase for case-insensitive matching
		hostLower := strings.ToLower(host)
		if strings.HasPrefix(hostLower, p.targetDomain) || hostLower == p.targetDomain {
			// Redirect to mock server
			reverseProxy := httputil.NewSingleHostReverseProxy(mockURL)
			reverseProxy.Director = func(req *http.Request) {
				req.URL.Scheme = mockURL.Scheme
				req.URL.Host = mockURL.Host
				req.Host = mockURL.Host
			}
			reverseProxy.ServeHTTP(w, r)
			return
		}

		// For non-intercepted domains, forward the request as-is
		p.forwardRequest(w, r)
	})
}

// handleConnect handles CONNECT requests for HTTPS tunneling with MITM
func (p *testProxy) handleConnect(w http.ResponseWriter, r *http.Request, mockURL *url.URL) {
	// Extract target host
	host := r.Host
	if host == "" {
		http.Error(w, "no host in CONNECT request", http.StatusBadRequest)
		return
	}

	// Check if this host should be intercepted
	hostWithoutPort := host
	if colonPos := strings.Index(host, ":"); colonPos != -1 {
		hostWithoutPort = host[:colonPos]
	}
	hostLower := strings.ToLower(hostWithoutPort)
	shouldIntercept := strings.HasPrefix(hostLower, p.targetDomain) || hostLower == p.targetDomain

	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	if shouldIntercept {
		// Perform MITM: wrap connection with TLS using dynamically generated certificate
		tlsCert, err := p.generateCertForHost(hostWithoutPort)
		if err != nil {
			return
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{*tlsCert},
		}

		tlsConn := tls.Server(clientConn, tlsConfig)
		defer tlsConn.Close()

		if err := tlsConn.Handshake(); err != nil {
			return
		}

		// Read the actual HTTPS request
		reader := bufio.NewReader(tlsConn)
		req, err := http.ReadRequest(reader)
		if err != nil {
			return
		}

		// Build full URL
		req.URL.Scheme = "https"
		req.URL.Host = host

		// Forward to mock server (convert HTTPS to HTTP)
		mockReq, err := http.NewRequest(req.Method, mockURL.String()+req.URL.Path, req.Body)
		if err != nil {
			return
		}
		mockReq.Header = req.Header.Clone()
		if req.URL.RawQuery != "" {
			mockReq.URL.RawQuery = req.URL.RawQuery
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(mockReq)
		if err != nil {
			errorResp := &http.Response{
				StatusCode: http.StatusBadGateway,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("proxy error: %v", err))),
			}
			errorResp.Write(tlsConn)
			return
		}
		defer resp.Body.Close()

		// Write response back to client
		resp.Write(tlsConn)
	} else {
		// For non-intercepted domains, just tunnel the connection
		targetConn, err := net.Dial("tcp", host)
		if err != nil {
			return
		}
		defer targetConn.Close()

		// Bidirectional copy
		go io.Copy(targetConn, clientConn)
		io.Copy(clientConn, targetConn)
	}
}

// forwardRequest forwards non-intercepted requests to their original destination
func (p *testProxy) forwardRequest(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Create new request
	outReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusBadGateway)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			outReq.Header.Add(key, value)
		}
	}

	// Send request
	resp, err := client.Do(outReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// Copy body
	io.Copy(w, resp.Body)
}

// generateCA generates a CA certificate and private key for MITM.
func generateCA() (*x509.Certificate, *rsa.PrivateKey, []byte, error) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	caTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Proxy CA"},
			CommonName:   "Test Proxy CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	return caCert, caKey, caPEM, nil
}

// generateCertForHost generates a certificate for a specific host signed by the CA.
// Certificates are cached to avoid regenerating them for the same host.
func (p *testProxy) generateCertForHost(host string) (*tls.Certificate, error) {
	// Check cache first
	if cached, ok := p.certCache.Load(host); ok {
		return cached.(*tls.Certificate), nil
	}

	// Generate new certificate
	certKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate key: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Proxy"},
			CommonName:   host,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, p.caCert, &certKey.PublicKey, p.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certKey)})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS certificate: %w", err)
	}

	// Store in cache
	p.certCache.Store(host, &tlsCert)

	return &tlsCert, nil
}

// Tests for testProxy

func TestNewTestProxy_InvalidInput(t *testing.T) {
	testCases := []struct {
		name         string
		targetDomain string
		mockHandler  http.Handler
		wantErr      string
	}{
		{
			name:         "empty domain",
			targetDomain: "",
			mockHandler:  http.NewServeMux(),
			wantErr:      "target domain cannot be empty",
		},
		{
			name:         "nil handler",
			targetDomain: "example.com",
			mockHandler:  nil,
			wantErr:      "mock handler cannot be nil",
		},
		{
			name:         "whitespace domain",
			targetDomain: "   ",
			mockHandler:  http.NewServeMux(),
			wantErr:      "target domain cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proxy, err := newTestProxy(tc.targetDomain, tc.mockHandler)
			if err == nil {
				defer proxy.Close()
				t.Fatal("expected error but got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestTestProxy_BasicHTTPInterception(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mocked response"))
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	client := newProxiedHTTPClient(proxy)

	resp, err := client.Get("http://example.com/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "mocked response" {
		t.Errorf("body = %q, want %q", string(body), "mocked response")
	}
}

func TestTestProxy_HTTPSInterception(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/secure", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("secure mocked response"))
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	client := newProxiedHTTPClient(proxy)

	resp, err := client.Get("https://example.com/secure")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "secure mocked response" {
		t.Errorf("body = %q, want %q", string(body), "secure mocked response")
	}
}

func TestTestProxy_HTTPSWithCACertFile(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	proxy, err := newTestProxy("api.example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	// Load CA cert from file path (automatically created)
	caPEM, err := os.ReadFile(proxy.CACertPath())
	if err != nil {
		t.Fatalf("failed to read CA file: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		t.Fatal("failed to append CA cert to pool")
	}

	proxyURL, _ := url.Parse(proxy.URL())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://api.example.com/api")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", string(body), `{"status":"ok"}`)
	}
}

func TestTestProxy_RequestHeaders(t *testing.T) {
	var receivedHeaders http.Header
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL())
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com/headers", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-Test-Header", "test-value")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if receivedHeaders.Get("X-Test-Header") != "test-value" {
		t.Errorf("header X-Test-Header = %q, want %q",
			receivedHeaders.Get("X-Test-Header"), "test-value")
	}
}

func TestTestProxy_RequestBody(t *testing.T) {
	var receivedBody string
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/body", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL())
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	testBody := "test request body"
	resp, err := client.Post("http://example.com/body", "text/plain", strings.NewReader(testBody))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if receivedBody != testBody {
		t.Errorf("received body = %q, want %q", receivedBody, testBody)
	}
}

func TestTestProxy_NonInterceptedDomain(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not see this"))
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL())
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}

	// Create a test server for non-intercepted domain
	realMux := http.NewServeMux()
	realMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("real response"))
	})
	realServer := &http.Server{Handler: realMux}
	realListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create real server: %v", err)
	}
	defer realServer.Close()

	go realServer.Serve(realListener)
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	// Request to non-intercepted domain should go to real server
	resp, err := client.Get(fmt.Sprintf("http://%s/", realListener.Addr().String()))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "real response" {
		t.Errorf("body = %q, want %q (request was intercepted when it shouldn't be)",
			string(body), "real response")
	}
}

func TestTestProxy_HTTPMethods(t *testing.T) {
	testCases := []struct {
		name   string
		method string
	}{
		{"GET", http.MethodGet},
		{"POST", http.MethodPost},
		{"PUT", http.MethodPut},
		{"DELETE", http.MethodDelete},
		{"PATCH", http.MethodPatch},
		{"HEAD", http.MethodHead},
		{"OPTIONS", http.MethodOptions},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var receivedMethod string
			mockMux := http.NewServeMux()
			mockMux.HandleFunc("/method", func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				w.WriteHeader(http.StatusOK)
			})

			proxy, err := newTestProxy("example.com", mockMux)
			if err != nil {
				t.Fatalf("failed to create proxy: %v", err)
			}
			defer proxy.Close()

			proxyURL, err := url.Parse(proxy.URL())
			if err != nil {
				t.Fatalf("failed to parse proxy URL: %v", err)
			}

			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
				Timeout: 5 * time.Second,
			}

			req, err := http.NewRequest(tc.method, "http://example.com/method", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()

			if receivedMethod != tc.method {
				t.Errorf("received method = %q, want %q", receivedMethod, tc.method)
			}
		})
	}
}

func TestTestProxy_QueryParameters(t *testing.T) {
	var receivedQuery url.Values
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL())
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://example.com/query?foo=bar&baz=qux")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if receivedQuery.Get("foo") != "bar" {
		t.Errorf("query param foo = %q, want %q", receivedQuery.Get("foo"), "bar")
	}
	if receivedQuery.Get("baz") != "qux" {
		t.Errorf("query param baz = %q, want %q", receivedQuery.Get("baz"), "qux")
	}
}

func TestTestProxy_ConcurrentRequests(t *testing.T) {
	requestCount := 0
	var mu sync.Mutex

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/concurrent", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL())
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			resp, err := client.Get("http://example.com/concurrent")
			if err != nil {
				t.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if requestCount != concurrency {
		t.Errorf("request count = %d, want %d", requestCount, concurrency)
	}
}

func TestTestProxy_Close(t *testing.T) {
	mockMux := http.NewServeMux()
	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	// First close should succeed
	if err := proxy.Close(); err != nil {
		t.Errorf("first Close() failed: %v", err)
	}

	// Second close should be idempotent
	if err := proxy.Close(); err != nil {
		t.Errorf("second Close() failed: %v", err)
	}

	// Requests after close should fail
	proxyURL, _ := url.Parse(proxy.URL())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 1 * time.Second,
	}

	_, err = client.Get("http://example.com/")
	if err == nil {
		t.Error("expected request to fail after Close(), but it succeeded")
	}
}

func TestTestProxy_DomainCaseInsensitive(t *testing.T) {
	testCases := []struct {
		name         string
		targetDomain string
		requestURL   string
	}{
		{"lowercase to uppercase", "example.com", "http://EXAMPLE.COM/test"},
		{"uppercase to lowercase", "EXAMPLE.COM", "http://example.com/test"},
		{"mixed case", "ExAmPlE.CoM", "http://eXaMpLe.cOm/test"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			intercepted := false
			mockMux := http.NewServeMux()
			mockMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				intercepted = true
				w.WriteHeader(http.StatusOK)
			})

			proxy, err := newTestProxy(tc.targetDomain, mockMux)
			if err != nil {
				t.Fatalf("failed to create proxy: %v", err)
			}
			defer proxy.Close()

			proxyURL, err := url.Parse(proxy.URL())
			if err != nil {
				t.Fatalf("failed to parse proxy URL: %v", err)
			}

			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
				Timeout: 5 * time.Second,
			}

			resp, err := client.Get(tc.requestURL)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()

			if !intercepted {
				t.Error("request was not intercepted (domain matching should be case-insensitive)")
			}
		})
	}
}

func TestTestProxy_MockServerReachable(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/direct", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("direct access"))
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	// Direct access to mock server (without proxy)
	resp, err := http.Get(proxy.MockURL() + "/direct")
	if err != nil {
		t.Fatalf("direct request to mock server failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "direct access" {
		t.Errorf("body = %q, want %q", string(body), "direct access")
	}
}

func TestTestProxy_ReverseProxyIntegration(t *testing.T) {
	// This test verifies integration with httputil.ReverseProxy pattern
	// similar to how it's used in the reference implementation

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Mock-Server", "true")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	})

	proxy, err := newTestProxy("api.example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL())
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://api.example.com/api/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Mock-Server") != "true" {
		t.Error("expected X-Mock-Server header from mock server")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != `{"success":true}` {
		t.Errorf("body = %q, want %q", string(body), `{"success":true}`)
	}
}

func TestTestProxy_StatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"Bad Request", http.StatusBadRequest},
		{"Not Found", http.StatusNotFound},
		{"Internal Server Error", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMux := http.NewServeMux()
			mockMux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			})

			proxy, err := newTestProxy("example.com", mockMux)
			if err != nil {
				t.Fatalf("failed to create proxy: %v", err)
			}
			defer proxy.Close()

			proxyURL, err := url.Parse(proxy.URL())
			if err != nil {
				t.Fatalf("failed to parse proxy URL: %v", err)
			}

			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
				Timeout: 5 * time.Second,
			}

			resp, err := client.Get("http://example.com/status")
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != tc.statusCode {
				t.Errorf("status code = %d, want %d", resp.StatusCode, tc.statusCode)
			}
		})
	}
}

func TestTestProxy_ResponseHeaders(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/response-headers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	proxy, err := newTestProxy("example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL())
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://example.com/response-headers")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("X-Custom-Header = %q, want %q",
			resp.Header.Get("X-Custom-Header"), "custom-value")
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q",
			resp.Header.Get("Content-Type"), "application/json")
	}
}

func TestTestProxy_CertificateCaching(t *testing.T) {
	requestCount := 0
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write(fmt.Appendf(nil, "request #%d", requestCount))
	})

	proxy, err := newTestProxy("secure.example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	client := newProxiedHTTPClient(proxy)

	// Make multiple HTTPS requests to the same host
	for i := 1; i <= 5; i++ {
		resp, err := client.Get("https://secure.example.com/test")
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		expected := fmt.Sprintf("request #%d", i)
		if string(body) != expected {
			t.Errorf("request %d: body = %q, want %q", i, string(body), expected)
		}
	}

	// Verify that certificate was generated only once (cached for subsequent requests)
	// We can't directly count cert generations, but we can verify the cache has the entry
	cached, ok := proxy.certCache.Load("secure.example.com")
	if !ok {
		t.Error("certificate was not cached")
	}
	if cached == nil {
		t.Error("cached certificate is nil")
	}

	if requestCount != 5 {
		t.Errorf("request count = %d, want 5", requestCount)
	}
}

// Example usage demonstrating reverse proxy pattern from reference implementation
func Example_newTestProxy() {
	// Create mock backend server
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/v1/predict", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"response_text":"mocked answer"}}`))
	})

	// Create proxy that intercepts api.example.com
	proxy, err := newTestProxy("api.example.com", mockMux)
	if err != nil {
		panic(err)
	}
	defer proxy.Close()

	// Configure HTTP client to use proxy
	proxyURL, _ := url.Parse(proxy.URL())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	// Make request to intercepted domain
	resp, err := client.Get("http://api.example.com/v1/predict")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	// Output: {"data":{"response_text":"mocked answer"}}
}

// newProxiedHTTPClient creates an HTTP client configured to use the proxy.
// For HTTPS requests, the client will trust the proxy's CA certificate.
func newProxiedHTTPClient(proxy *testProxy) *http.Client {
	proxyURL, _ := url.Parse(proxy.URL())

	// Create cert pool with proxy's CA
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(proxy.CACertPEM())

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
		Timeout: 10 * time.Second,
	}
}

// newProxiedHTTPClientInsecure creates an HTTP client configured to use the proxy
// with InsecureSkipVerify enabled (not recommended for production, useful for testing).
func newProxiedHTTPClientInsecure(proxyURL string) *http.Client {
	proxy, _ := url.Parse(proxyURL)
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 10 * time.Second,
	}
}
