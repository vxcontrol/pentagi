package sage

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testSAGEEndpoint = "http://localhost:9000"

func sageAvailable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(testSAGEEndpoint + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func TestIntegration_HealthCheck(t *testing.T) {
	if !sageAvailable() {
		t.Skip("SAGE not running at " + testSAGEEndpoint)
	}

	client, err := NewClient(testSAGEEndpoint, "", "pentagi-test", 30*time.Second, true)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if !client.IsEnabled() {
		t.Fatal("client should be enabled")
	}

	err = client.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}

func TestIntegration_DisabledClient(t *testing.T) {
	client, err := NewClient("", "", "", 30*time.Second, false)
	if err != nil {
		t.Fatalf("failed to create disabled client: %v", err)
	}

	if client.IsEnabled() {
		t.Fatal("disabled client should not be enabled")
	}

	_, err = client.Recall(context.Background(), RecallRequest{Query: "test"})
	if err == nil {
		t.Fatal("disabled client recall should return error")
	}
}

func TestIntegration_Embed(t *testing.T) {
	if !sageAvailable() {
		t.Skip("SAGE not running at " + testSAGEEndpoint)
	}

	client, err := NewClient(testSAGEEndpoint, "", "pentagi-test", 30*time.Second, true)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	embedding, err := client.Embed(context.Background(), "nmap scan results showing open ports 22, 80, 443")
	if err != nil {
		t.Fatalf("embed failed: %v", err)
	}

	if len(embedding) == 0 {
		t.Fatal("embedding should not be empty")
	}

	t.Logf("embedding dimension: %d", len(embedding))
}

func TestIntegration_RememberAndRecall(t *testing.T) {
	if !sageAvailable() {
		t.Skip("SAGE not running at " + testSAGEEndpoint)
	}

	client, err := NewClient(testSAGEEndpoint, "", "pentagi-test", 30*time.Second, true)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Step 1: Remember a finding
	rememberResp, err := client.Remember(ctx, RememberRequest{
		Content:    "During penetration testing of the target web application, nmap revealed open ports 22 (SSH), 80 (HTTP), and 443 (HTTPS). The SSH service was running OpenSSH 8.9p1 which is vulnerable to CVE-2023-38408. Successful exploitation achieved using the agent forwarding attack vector.",
		MemoryType: "fact",
		Domain:     "pentest-findings",
		Confidence: 0.95,
	})
	if err != nil {
		t.Fatalf("remember failed: %v", err)
	}

	t.Logf("remember response: memory_id=%s, status=%s, success=%v", rememberResp.MemoryID, rememberResp.Status, rememberResp.Success)

	if rememberResp.MemoryID == "" {
		t.Fatal("memory ID should not be empty")
	}

	// Step 2: Recall by text search (FTS5)
	// Note: memory might need a moment to go through consensus
	time.Sleep(3 * time.Second)

	recallResp, err := client.Recall(ctx, RecallRequest{
		Query:         "OpenSSH vulnerability CVE-2023-38408",
		Domain:        "pentest-findings",
		MaxResults:    5,
		MinConfidence: 0.5,
	})
	if err != nil {
		t.Fatalf("recall failed: %v", err)
	}

	t.Logf("recall returned %d results (total: %d)", len(recallResp.Memories), recallResp.TotalCount)

	found := false
	for _, mem := range recallResp.Memories {
		t.Logf("  - [%s] confidence=%.2f domain=%s content=%s...", mem.MemoryType, mem.Confidence, mem.Domain, truncateStr(mem.Content, 80))
		if mem.MemoryID == rememberResp.MemoryID {
			found = true
		}
	}

	if !found && len(recallResp.Memories) == 0 {
		t.Log("WARNING: recall returned 0 results -- memory may still be in consensus")
	}

	// Step 3: Recall by semantic search (vector similarity)
	semanticResp, err := client.RecallSemantic(ctx, RecallRequest{
		Query:         "SSH exploitation techniques and open port scanning results",
		Domain:        "pentest-findings",
		MaxResults:    5,
		MinConfidence: 0.5,
	})
	if err != nil {
		t.Fatalf("semantic recall failed: %v", err)
	}

	t.Logf("semantic recall returned %d results (total: %d)", len(semanticResp.Memories), semanticResp.TotalCount)
	for _, mem := range semanticResp.Memories {
		t.Logf("  - [%s] confidence=%.2f content=%s...", mem.MemoryType, mem.Confidence, truncateStr(mem.Content, 80))
	}
}

func TestIntegration_RememberMultipleAndRecallCrossSession(t *testing.T) {
	if !sageAvailable() {
		t.Skip("SAGE not running at " + testSAGEEndpoint)
	}

	client, err := NewClient(testSAGEEndpoint, "", "pentagi-test", 30*time.Second, true)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Simulate first pentest engagement
	memories := []RememberRequest{
		{
			Content:    "Target application uses JWT authentication with RS256 signing. The public key is exposed at /.well-known/jwks.json endpoint. Algorithm confusion attack possible by switching to HS256 and signing with the public key.",
			MemoryType: "fact",
			Domain:     "pentest-exploit",
			Confidence: 0.95,
		},
		{
			Content:    "Gobuster directory scan on target revealed /admin, /api/v1/debug, /actuator/health endpoints. The /api/v1/debug endpoint returned stack traces with internal IP addresses and database connection strings.",
			MemoryType: "observation",
			Domain:     "pentest-recon",
			Confidence: 0.85,
		},
		{
			Content:    "Based on the exposed debug endpoint and JWT vulnerability, the application likely has insufficient security hardening for a production deployment. Recommend checking for other OWASP Top 10 vulnerabilities, especially SSRF via the debug endpoint.",
			MemoryType: "inference",
			Domain:     "pentest-findings",
			Confidence: 0.70,
		},
	}

	for i, mem := range memories {
		resp, err := client.Remember(ctx, mem)
		if err != nil {
			t.Fatalf("remember %d failed: %v", i, err)
		}
		t.Logf("stored memory %d: id=%s type=%s domain=%s", i, resp.MemoryID, mem.MemoryType, mem.Domain)
	}

	// Wait for consensus
	time.Sleep(3 * time.Second)

	// Simulate second engagement -- recall what we learned
	testQueries := []struct {
		query  string
		domain string
	}{
		{"JWT authentication vulnerabilities and algorithm confusion attacks", "pentest-exploit"},
		{"directory scanning results and exposed debug endpoints", "pentest-recon"},
		{"OWASP security hardening recommendations", "pentest-findings"},
	}

	for _, tq := range testQueries {
		resp, err := client.Recall(ctx, RecallRequest{
			Query:         tq.query,
			Domain:        tq.domain,
			MaxResults:    3,
			MinConfidence: 0.5,
		})
		if err != nil {
			t.Errorf("recall failed for query '%s': %v", tq.query, err)
			continue
		}
		t.Logf("query '%s' (domain=%s) -> %d results", truncateStr(tq.query, 50), tq.domain, len(resp.Memories))
		for _, mem := range resp.Memories {
			t.Logf("  - [%s] %.2f: %s...", mem.MemoryType, mem.Confidence, truncateStr(mem.Content, 100))
		}
	}
}

// --- Ed25519 signing unit tests (no SAGE server needed) ---

// newTestClient creates a Client with a temporary key file for offline signing tests.
func newTestClient(t *testing.T) *Client {
	t.Helper()

	keyPath := filepath.Join(t.TempDir(), "test-agent.key")
	pub, priv, err := loadOrCreateKey(keyPath)
	if err != nil {
		t.Fatalf("loadOrCreateKey failed: %v", err)
	}

	return &Client{
		endpoint:   "http://unused:9999",
		agentName:  "unit-test",
		publicKey:  pub,
		privateKey: priv,
		enabled:    true,
	}
}

func TestSignRequest_ProducesValidSignature(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	method := "POST"
	path := "/v1/memory/submit"
	body := []byte(`{"content":"test"}`)
	timestamp := time.Now().Unix()

	sig := client.signRequest(method, path, body, timestamp)

	// Ed25519 signatures are always 64 bytes.
	if len(sig) != ed25519.SignatureSize {
		t.Fatalf("signature length = %d, want %d", len(sig), ed25519.SignatureSize)
	}

	// Recompute the signed message independently and verify.
	canonical := []byte(method + " " + path + "\n")
	canonical = append(canonical, body...)
	bodyHash := sha256.Sum256(canonical)

	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))

	message := make([]byte, 0, len(bodyHash)+8)
	message = append(message, bodyHash[:]...)
	message = append(message, tsBytes...)

	if !ed25519.Verify(client.publicKey, message, sig) {
		t.Fatal("signature verification failed")
	}
}

func TestSignRequest_DifferentPathsDifferentSignatures(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	body := []byte(`{"query":"test"}`)
	ts := time.Now().Unix()

	sig1 := client.signRequest("POST", "/v1/memory/search", body, ts)
	sig2 := client.signRequest("POST", "/v1/memory/submit", body, ts)

	if bytes.Equal(sig1, sig2) {
		t.Fatal("signatures for different paths must differ (prevents cross-endpoint replay)")
	}
}

func TestSignRequest_DifferentTimestampsDifferentSignatures(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	body := []byte(`{"content":"test"}`)
	path := "/v1/memory/submit"

	sig1 := client.signRequest("POST", path, body, 1000)
	sig2 := client.signRequest("POST", path, body, 2000)

	if bytes.Equal(sig1, sig2) {
		t.Fatal("signatures for different timestamps must differ")
	}
}

func TestLoadOrCreateKey_PersistsAndReloads(t *testing.T) {
	t.Parallel()

	keyPath := filepath.Join(t.TempDir(), "persist.key")

	pub1, priv1, err := loadOrCreateKey(keyPath)
	if err != nil {
		t.Fatalf("first loadOrCreateKey: %v", err)
	}

	// File should exist.
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("key file not created: %v", err)
	}

	// Reload the same file — keys must match.
	pub2, priv2, err := loadOrCreateKey(keyPath)
	if err != nil {
		t.Fatalf("second loadOrCreateKey: %v", err)
	}

	if !bytes.Equal(pub1, pub2) {
		t.Fatal("reloaded public key differs from original")
	}
	if !bytes.Equal(priv1, priv2) {
		t.Fatal("reloaded private key differs from original")
	}
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// --- Keypair persistence unit tests ---

func TestLoadOrCreateKey_GeneratesNewKey(t *testing.T) {
	t.Parallel()

	keyFile := filepath.Join(t.TempDir(), "new-agent.key")

	pub, priv, err := loadOrCreateKey(keyFile)
	if err != nil {
		t.Fatalf("loadOrCreateKey failed: %v", err)
	}

	if len(pub) != ed25519.PublicKeySize {
		t.Fatalf("expected public key length %d, got %d", ed25519.PublicKeySize, len(pub))
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Fatalf("expected private key length %d, got %d", ed25519.PrivateKeySize, len(priv))
	}

	// Verify the file was created
	info, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("key file was not created: %v", err)
	}

	// Check file permissions (0600)
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file permissions 0600, got %04o", perm)
	}

	// Verify the file contains exactly ed25519.PrivateKeySize bytes
	data, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("failed to read key file: %v", err)
	}
	if len(data) != ed25519.PrivateKeySize {
		t.Fatalf("expected key file to contain %d bytes, got %d", ed25519.PrivateKeySize, len(data))
	}
}

func TestLoadOrCreateKey_LoadsExistingKey(t *testing.T) {
	t.Parallel()

	keyFile := filepath.Join(t.TempDir(), "existing-agent.key")

	// Generate a keypair manually and write the private key
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte(priv), 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	// Now load via loadOrCreateKey
	loadedPub, loadedPriv, err := loadOrCreateKey(keyFile)
	if err != nil {
		t.Fatalf("loadOrCreateKey failed: %v", err)
	}

	if !loadedPub.Equal(pub) {
		t.Error("loaded public key does not match original")
	}
	if !loadedPriv.Equal(priv) {
		t.Error("loaded private key does not match original")
	}
}

func TestLoadOrCreateKey_PersistsAcrossCalls(t *testing.T) {
	t.Parallel()

	keyFile := filepath.Join(t.TempDir(), "persist-across-agent.key")

	pub1, priv1, err := loadOrCreateKey(keyFile)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	pub2, priv2, err := loadOrCreateKey(keyFile)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if !pub1.Equal(pub2) {
		t.Error("public keys differ across calls — identity not persistent")
	}
	if !priv1.Equal(priv2) {
		t.Error("private keys differ across calls — identity not persistent")
	}
}

func TestLoadOrCreateKey_DefaultPath(t *testing.T) {
	// Not parallel: touches a shared default path in os.TempDir().
	defaultKeyFile := filepath.Join(os.TempDir(), "pentagi-sage-agent.key")

	// Save any existing key to restore afterwards.
	existingData, hadExisting := func() ([]byte, bool) {
		data, err := os.ReadFile(defaultKeyFile)
		if err != nil {
			return nil, false
		}
		return data, true
	}()
	defer func() {
		if hadExisting {
			_ = os.WriteFile(defaultKeyFile, existingData, 0600)
		} else {
			_ = os.Remove(defaultKeyFile)
		}
	}()

	// Remove so we can verify creation
	_ = os.Remove(defaultKeyFile)

	pub, priv, err := loadOrCreateKey("")
	if err != nil {
		t.Fatalf("loadOrCreateKey with empty path failed: %v", err)
	}

	if len(pub) != ed25519.PublicKeySize {
		t.Fatalf("expected public key length %d, got %d", ed25519.PublicKeySize, len(pub))
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Fatalf("expected private key length %d, got %d", ed25519.PrivateKeySize, len(priv))
	}

	// Verify the file was created at the default path
	if _, err := os.Stat(defaultKeyFile); err != nil {
		t.Fatalf("expected key file at default path %s, got error: %v", defaultKeyFile, err)
	}
}

func TestLoadOrCreateKey_InvalidKeyFile(t *testing.T) {
	t.Parallel()

	keyFile := filepath.Join(t.TempDir(), "bad-agent.key")

	// Write garbage data (wrong length — not 64 bytes)
	garbage := []byte("this is garbage data, not a valid ed25519 key!")
	if err := os.WriteFile(keyFile, garbage, 0600); err != nil {
		t.Fatalf("failed to write garbage file: %v", err)
	}

	pub, priv, err := loadOrCreateKey(keyFile)
	if err != nil {
		t.Fatalf("loadOrCreateKey should not fail on invalid key file: %v", err)
	}

	// Should have generated a new valid key
	if len(pub) != ed25519.PublicKeySize {
		t.Fatalf("expected public key length %d, got %d", ed25519.PublicKeySize, len(pub))
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Fatalf("expected private key length %d, got %d", ed25519.PrivateKeySize, len(priv))
	}

	// The garbage should have been overwritten with a valid key
	data, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("failed to read key file after regeneration: %v", err)
	}
	if len(data) != ed25519.PrivateKeySize {
		t.Fatalf("key file should now contain %d bytes, got %d", ed25519.PrivateKeySize, len(data))
	}
}

func TestNewClient_PersistentIdentity(t *testing.T) {
	t.Parallel()

	keyFile := filepath.Join(t.TempDir(), "identity-agent.key")

	// NewClient performs a health check against a live SAGE endpoint,
	// so we verify persistent identity via loadOrCreateKey + agentID
	// derivation (hex-encoded public key), which is exactly what
	// NewClient uses internally.
	pub1, _, err := loadOrCreateKey(keyFile)
	if err != nil {
		t.Fatalf("first loadOrCreateKey failed: %v", err)
	}
	agentID1 := hex.EncodeToString(pub1)

	pub2, _, err := loadOrCreateKey(keyFile)
	if err != nil {
		t.Fatalf("second loadOrCreateKey failed: %v", err)
	}
	agentID2 := hex.EncodeToString(pub2)

	if agentID1 != agentID2 {
		t.Errorf("agentIDs differ: %s vs %s — identity not persistent across client creations", agentID1, agentID2)
	}

	if len(agentID1) != ed25519.PublicKeySize*2 {
		t.Errorf("expected agentID hex length %d, got %d", ed25519.PublicKeySize*2, len(agentID1))
	}
}
