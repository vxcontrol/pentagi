package sage

import (
	"context"
	"net/http"
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

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
