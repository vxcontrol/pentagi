package searchers

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
)

// testProxy is a MITM HTTP/HTTPS proxy that routes targetDomain to a mock server, so an engine runs
// its real client against a scripted vendor API. Point config.ProxyURL and ExternalSSLCAPath at it.
type testProxy struct {
	proxyServer  *http.Server
	mockServer   *http.Server
	proxyURL     string
	targetDomain string
	caCert       *x509.Certificate
	caKey        *rsa.PrivateKey
	caFilePath   string
	certCache    sync.Map // host -> *tls.Certificate
}

// newTestProxy starts a proxy that sends targetDomain to mockHandler and answers any other host
// with 502, so a test never reaches the network. It shuts down when the test ends.
func newTestProxy(t *testing.T, targetDomain string, mockHandler http.Handler) *testProxy {
	t.Helper()

	caCert, caKey, caPEM, err := generateCA()
	if err != nil {
		t.Fatalf("failed to generate the proxy CA: %v", err)
	}
	caFilePath := filepath.Join(t.TempDir(), "proxy-ca.pem")
	if err := os.WriteFile(caFilePath, caPEM, 0o600); err != nil {
		t.Fatalf("failed to write the proxy CA: %v", err)
	}

	p := &testProxy{
		targetDomain: strings.ToLower(strings.TrimSpace(targetDomain)),
		caCert:       caCert,
		caKey:        caKey,
		caFilePath:   caFilePath,
	}

	mockListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create the mock listener: %v", err)
	}
	p.mockServer = &http.Server{Handler: mockHandler, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	mockURL := &url.URL{Scheme: "http", Host: mockListener.Addr().String()}
	go searchersServe(p.mockServer, mockListener)

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		_ = p.mockServer.Close()
		t.Fatalf("failed to create the proxy listener: %v", err)
	}
	p.proxyServer = &http.Server{Handler: p.handler(mockURL), ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	p.proxyURL = "http://" + proxyListener.Addr().String()
	go searchersServe(p.proxyServer, proxyListener)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = p.proxyServer.Shutdown(ctx)
		_ = p.mockServer.Shutdown(ctx)
	})
	return p
}

func searchersServe(srv *http.Server, l net.Listener) {
	if err := srv.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("test proxy server error: %v", err))
	}
}

// URL is the value for config.ProxyURL.
func (p *testProxy) URL() string {
	return p.proxyURL
}

// CACertPath is the value for config.ExternalSSLCAPath, so the engine trusts the MITM certificates.
func (p *testProxy) CACertPath() string {
	return p.caFilePath
}

func (p *testProxy) intercepts(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.HasPrefix(strings.ToLower(host), p.targetDomain)
}

func (p *testProxy) handler(mockURL *url.URL) http.Handler {
	reverseProxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.Out.URL.Scheme = mockURL.Scheme
			pr.Out.URL.Host = mockURL.Host
			pr.Out.Host = mockURL.Host
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.URL.Host
		if host == "" {
			host = r.Host
		}
		switch {
		case !p.intercepts(host):
			http.Error(w, "test proxy does not intercept "+host, http.StatusBadGateway)
		case r.Method == http.MethodConnect:
			p.handleConnect(w, r, mockURL)
		default:
			reverseProxy.ServeHTTP(w, r)
		}
	})
}

// handleConnect terminates the client's TLS with a certificate for the requested host, signed
// by the proxy CA, and replays the one request it carries against the mock server over HTTP.
func (p *testProxy) handleConnect(w http.ResponseWriter, r *http.Request, mockURL *url.URL) {
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

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

	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	tlsCert, err := p.generateCertForHost(host)
	if err != nil {
		return
	}
	tlsConn := tls.Server(clientConn, &tls.Config{Certificates: []tls.Certificate{*tlsCert}})
	defer tlsConn.Close()
	if err := tlsConn.Handshake(); err != nil {
		return
	}

	req, err := http.ReadRequest(bufio.NewReader(tlsConn))
	if err != nil {
		return
	}
	mockReq, err := http.NewRequest(req.Method, mockURL.String()+req.URL.Path, req.Body)
	if err != nil {
		return
	}
	mockReq.Header = req.Header.Clone()
	mockReq.URL.RawQuery = req.URL.RawQuery

	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(mockReq)
	if err != nil {
		errorResp := &http.Response{
			StatusCode: http.StatusBadGateway,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("proxy error: %v", err))),
		}
		_ = errorResp.Write(tlsConn)
		return
	}
	defer resp.Body.Close()

	_ = resp.Write(tlsConn)
}

func generateCA() (*x509.Certificate, *rsa.PrivateKey, []byte, error) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	caTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Proxy CA"},
			CommonName:   "Test Proxy CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(24 * time.Hour),
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

	return caCert, caKey, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}), nil
}

// generateCertForHost caches per host: an engine that retries opens a tunnel for every attempt.
func (p *testProxy) generateCertForHost(host string) (*tls.Certificate, error) {
	if cached, ok := p.certCache.Load(host); ok {
		return cached.(*tls.Certificate), nil
	}

	certKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Proxy"},
			CommonName:   host,
		},
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(24 * time.Hour),
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

	p.certCache.Store(host, &tlsCert)
	return &tlsCert, nil
}

// searchersUpstream plays a vendor API behind the proxy: it answers every request with status
// (200 when unset) and body, and keeps what the engine sent last.
type searchersUpstream struct {
	mu      sync.Mutex
	status  int
	body    string
	hits    int
	method  string
	path    string
	header  http.Header
	query   url.Values
	payload []byte
}

func (u *searchersUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, _ := io.ReadAll(r.Body)

	u.mu.Lock()
	defer u.mu.Unlock()
	u.hits++
	u.method, u.path, u.header, u.query, u.payload = r.Method, r.URL.Path, r.Header.Clone(), r.URL.Query(), payload
	if u.status != 0 {
		w.WriteHeader(u.status)
	}
	_, _ = io.WriteString(w, u.body)
}

// reset sets the answer to the next requests and forgets the earlier ones, for a table whose
// rows share one upstream.
func (u *searchersUpstream) reset(status int, body string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.status, u.body = status, body
	u.hits, u.method, u.path, u.header, u.query, u.payload = 0, "", "", nil, nil, nil
}
