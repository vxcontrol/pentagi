package system

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pentagi/pkg/config"
)

// testCerts holds generated test certificates
type testCerts struct {
	rootCA          *x509.Certificate
	rootCAKey       *rsa.PrivateKey
	rootCAPEM       []byte
	intermediate    *x509.Certificate
	intermediateKey *rsa.PrivateKey
	intermediatePEM []byte
	serverCert      *x509.Certificate
	serverKey       *rsa.PrivateKey
	serverPEM       []byte
	serverKeyPEM    []byte
}

// generateRSAKey generates a new RSA private key
func generateRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// generateSerialNumber generates a random serial number for certificates
func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}

// createCertificate creates a certificate from template and signs it
func createCertificate(template, parent *x509.Certificate, pub, priv interface{}) (*x509.Certificate, []byte, error) {
	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return cert, certPEM, nil
}

// generateTestCerts creates a complete certificate chain for testing
func generateTestCerts() (*testCerts, error) {
	certs := &testCerts{}

	// generate root CA private key
	rootKey, err := generateRSAKey()
	if err != nil {
		return nil, err
	}
	certs.rootCAKey = rootKey

	// create root CA certificate template
	rootSerial, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	rootTemplate := &x509.Certificate{
		SerialNumber: rootSerial,
		Subject: pkix.Name{
			CommonName:   "Test Root CA",
			Organization: []string{"PentAGI Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	// self-sign root CA
	rootCert, rootPEM, err := createCertificate(rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}
	certs.rootCA = rootCert
	certs.rootCAPEM = rootPEM

	// generate intermediate CA private key
	intermediateKey, err := generateRSAKey()
	if err != nil {
		return nil, err
	}
	certs.intermediateKey = intermediateKey

	// create intermediate CA certificate template
	intermediateSerial, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	intermediateTemplate := &x509.Certificate{
		SerialNumber: intermediateSerial,
		Subject: pkix.Name{
			CommonName:   "Test Intermediate CA",
			Organization: []string{"PentAGI Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	// sign intermediate CA with root CA
	intermediateCert, intermediatePEM, err := createCertificate(intermediateTemplate, rootCert, &intermediateKey.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}
	certs.intermediate = intermediateCert
	certs.intermediatePEM = intermediatePEM

	// generate server private key
	serverKey, err := generateRSAKey()
	if err != nil {
		return nil, err
	}
	certs.serverKey = serverKey

	// create server certificate template
	serverSerial, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: serverSerial,
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"PentAGI Test Server"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// sign server certificate with intermediate CA
	serverCert, serverPEM, err := createCertificate(serverTemplate, intermediateCert, &serverKey.PublicKey, intermediateKey)
	if err != nil {
		return nil, err
	}
	certs.serverCert = serverCert
	certs.serverPEM = serverPEM

	// encode server private key
	serverKeyDER, err := x509.MarshalPKCS8PrivateKey(serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal server private key: %w", err)
	}
	certs.serverKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: serverKeyDER,
	})

	return certs, nil
}

// createTempFile creates a temporary file with given content
func createTempFile(t *testing.T, content []byte) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("cert-%d.pem", time.Now().UnixNano()))

	err := os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return tmpFile
}

// createTestConfig creates a test config with given CA path
func createTestConfig(caPath string, insecure bool, proxyURL string) *config.Config {
	return &config.Config{
		ExternalSSLCAPath:   caPath,
		ExternalSSLInsecure: insecure,
		ProxyURL:            proxyURL,
	}
}

// createTLSTestServer creates a test HTTPS server with the given certificates
func createTLSTestServer(t *testing.T, certs *testCerts, includeIntermediateInChain bool) *httptest.Server {
	t.Helper()

	// prepare certificate chain
	var certChain []tls.Certificate
	serverCertBytes := certs.serverPEM
	if includeIntermediateInChain {
		// append intermediate certificate to chain
		serverCertBytes = append(serverCertBytes, certs.intermediatePEM...)
	}

	cert, err := tls.X509KeyPair(serverCertBytes, certs.serverKeyPEM)
	if err != nil {
		t.Fatalf("failed to load server certificate: %v", err)
	}
	certChain = append(certChain, cert)

	// create TLS config for server
	tlsConfig := &tls.Config{
		Certificates: certChain,
		MinVersion:   tls.VersionTLS12,
	}

	// create test server
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	server.TLS = tlsConfig
	server.StartTLS()

	return server
}

func TestGetSystemCertPool_EmptyPath(t *testing.T) {
	cfg := createTestConfig("", false, "")

	pool, err := GetSystemCertPool(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if pool == nil {
		t.Fatal("expected non-nil cert pool")
	}
}

func TestGetSystemCertPool_NonExistentFile(t *testing.T) {
	cfg := createTestConfig("/non/existent/path/ca.pem", false, "")

	_, err := GetSystemCertPool(cfg)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestGetSystemCertPool_InvalidPEM(t *testing.T) {
	invalidPEM := []byte("this is not a valid PEM file")
	tmpFile := createTempFile(t, invalidPEM)

	cfg := createTestConfig(tmpFile, false, "")

	_, err := GetSystemCertPool(cfg)
	if err == nil {
		t.Fatal("expected error for invalid PEM content")
	}
}

func TestGetSystemCertPool_SingleRootCA(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}

	tmpFile := createTempFile(t, certs.rootCAPEM)
	cfg := createTestConfig(tmpFile, false, "")

	pool, err := GetSystemCertPool(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if pool == nil {
		t.Fatal("expected non-nil cert pool")
	}

	// verify that certificate was added by trying to verify a cert signed by it
	opts := x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// create a test chain with intermediate
	intermediates := x509.NewCertPool()
	intermediates.AddCert(certs.intermediate)
	opts.Intermediates = intermediates

	_, err = certs.serverCert.Verify(opts)
	if err != nil {
		t.Errorf("failed to verify certificate with custom root CA: %v", err)
	}
}

func TestGetSystemCertPool_MultipleRootCAs(t *testing.T) {
	// generate first certificate chain
	certs1, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate first test certs: %v", err)
	}

	// generate second certificate chain
	certs2, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate second test certs: %v", err)
	}

	// combine both root CAs in one file
	multipleCAs := append(certs1.rootCAPEM, certs2.rootCAPEM...)
	tmpFile := createTempFile(t, multipleCAs)

	cfg := createTestConfig(tmpFile, false, "")

	pool, err := GetSystemCertPool(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if pool == nil {
		t.Fatal("expected non-nil cert pool")
	}

	// verify that both CAs were added by checking certificates from both chains
	verifyChain := func(certs *testCerts, name string) {
		opts := x509.VerifyOptions{
			Roots:     pool,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}

		intermediates := x509.NewCertPool()
		intermediates.AddCert(certs.intermediate)
		opts.Intermediates = intermediates

		_, err := certs.serverCert.Verify(opts)
		if err != nil {
			t.Errorf("failed to verify certificate from %s with multiple root CAs: %v", name, err)
		}
	}

	verifyChain(certs1, "first chain")
	verifyChain(certs2, "second chain")
}

func TestGetSystemCertPool_WithIntermediateCerts(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}

	// create file with both root and intermediate certificates
	combined := append(certs.rootCAPEM, certs.intermediatePEM...)
	tmpFile := createTempFile(t, combined)

	cfg := createTestConfig(tmpFile, false, "")

	pool, err := GetSystemCertPool(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if pool == nil {
		t.Fatal("expected non-nil cert pool")
	}

	// note: when intermediate is in root pool, verification should still work
	// but this is not the correct PKI setup
	opts := x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	_, err = certs.serverCert.Verify(opts)
	if err != nil {
		t.Errorf("failed to verify certificate with intermediate in root pool: %v", err)
	}
}

func TestGetHTTPClient_NoProxy(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}

	tmpFile := createTempFile(t, certs.rootCAPEM)
	cfg := createTestConfig(tmpFile, false, "")

	client, err := GetHTTPClient(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil HTTP client")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("expected non-nil TLS config")
	}

	if transport.TLSClientConfig.RootCAs == nil {
		t.Fatal("expected non-nil root CA pool")
	}
}

func TestGetHTTPClient_WithProxy(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}

	tmpFile := createTempFile(t, certs.rootCAPEM)
	cfg := createTestConfig(tmpFile, false, "http://proxy.example.com:8080")

	client, err := GetHTTPClient(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil HTTP client")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected http.Transport")
	}

	if transport.Proxy == nil {
		t.Fatal("expected non-nil proxy function")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("expected non-nil TLS config")
	}
}

func TestGetHTTPClient_InsecureSkipVerify(t *testing.T) {
	cfg := createTestConfig("", true, "")

	client, err := GetHTTPClient(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected http.Transport")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be true")
	}
}

func TestHTTPClient_RealConnection_WithIntermediateInChain(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}

	// create HTTPS server with intermediate cert in chain
	server := createTLSTestServer(t, certs, true)
	defer server.Close()

	// create HTTP client with only root CA (proper PKI setup)
	tmpFile := createTempFile(t, certs.rootCAPEM)
	cfg := createTestConfig(tmpFile, false, "")

	client, err := GetHTTPClient(cfg)
	if err != nil {
		t.Fatalf("failed to create HTTP client: %v", err)
	}

	// make request to HTTPS server
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make HTTPS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != "OK" {
		t.Errorf("expected body 'OK', got '%s'", string(body))
	}
}

func TestHTTPClient_RealConnection_WithoutIntermediateInChain(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}

	// create HTTPS server WITHOUT intermediate cert in chain
	server := createTLSTestServer(t, certs, false)
	defer server.Close()

	// create HTTP client with only root CA
	tmpFile := createTempFile(t, certs.rootCAPEM)
	cfg := createTestConfig(tmpFile, false, "")

	client, err := GetHTTPClient(cfg)
	if err != nil {
		t.Fatalf("failed to create HTTP client: %v", err)
	}

	// this should fail because server doesn't provide intermediate cert
	_, err = client.Get(server.URL)
	if err == nil {
		t.Fatal("expected error when server doesn't provide intermediate certificate")
	}
}

func TestHTTPClient_RealConnection_WithIntermediateInRootPool(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}

	// create HTTPS server WITHOUT intermediate cert in chain
	server := createTLSTestServer(t, certs, false)
	defer server.Close()

	// create HTTP client with both root and intermediate in CA pool
	// this is not proper PKI setup, but it works around server misconfiguration
	combined := append(certs.rootCAPEM, certs.intermediatePEM...)
	tmpFile := createTempFile(t, combined)
	cfg := createTestConfig(tmpFile, false, "")

	client, err := GetHTTPClient(cfg)
	if err != nil {
		t.Fatalf("failed to create HTTP client: %v", err)
	}

	// this should succeed because intermediate is in root pool
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make HTTPS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_RealConnection_MultipleRootCAs(t *testing.T) {
	// generate two separate certificate chains
	certs1, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate first test certs: %v", err)
	}

	certs2, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate second test certs: %v", err)
	}

	// create two HTTPS servers with different certificate chains
	server1 := createTLSTestServer(t, certs1, true)
	defer server1.Close()

	server2 := createTLSTestServer(t, certs2, true)
	defer server2.Close()

	// create HTTP client with both root CAs
	multipleCAs := append(certs1.rootCAPEM, certs2.rootCAPEM...)
	tmpFile := createTempFile(t, multipleCAs)
	cfg := createTestConfig(tmpFile, false, "")

	client, err := GetHTTPClient(cfg)
	if err != nil {
		t.Fatalf("failed to create HTTP client: %v", err)
	}

	// test connection to first server
	resp1, err := client.Get(server1.URL)
	if err != nil {
		t.Fatalf("failed to connect to first server: %v", err)
	}
	resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 from first server, got %d", resp1.StatusCode)
	}

	// test connection to second server
	resp2, err := client.Get(server2.URL)
	if err != nil {
		t.Fatalf("failed to connect to second server: %v", err)
	}
	resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 from second server, got %d", resp2.StatusCode)
	}
}

func TestHTTPClient_RealConnection_InsecureMode(t *testing.T) {
	certs, err := generateTestCerts()
	if err != nil {
		t.Fatalf("failed to generate test certs: %v", err)
	}

	// create HTTPS server
	server := createTLSTestServer(t, certs, true)
	defer server.Close()

	// create HTTP client with InsecureSkipVerify=true and no CA file
	cfg := createTestConfig("", true, "")

	client, err := GetHTTPClient(cfg)
	if err != nil {
		t.Fatalf("failed to create HTTP client: %v", err)
	}

	// this should succeed because we skip verification
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make HTTPS request in insecure mode: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
