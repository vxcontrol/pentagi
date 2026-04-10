package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"pentagi/pkg/sage"
)

// mockSageClient implements SageSearcher for unit tests
type mockSageClient struct {
	enabled    bool
	memories   []sage.MemoryResult
	rememberID string
}

func (m *mockSageClient) IsEnabled() bool { return m.enabled }

func (m *mockSageClient) Recall(ctx context.Context, req sage.RecallRequest) (*sage.RecallResponse, error) {
	return &sage.RecallResponse{
		Memories:   m.memories,
		TotalCount: len(m.memories),
	}, nil
}

func (m *mockSageClient) Remember(ctx context.Context, req sage.RememberRequest) (*sage.RememberResponse, error) {
	return &sage.RememberResponse{
		MemoryID: m.rememberID,
		Status:   "proposed",
		Success:  true,
		Message:  "Memory submitted for consensus",
	}, nil
}

func TestSageTool_RecallWithResults(t *testing.T) {
	mock := &mockSageClient{
		enabled: true,
		memories: []sage.MemoryResult{
			{
				MemoryID:   "mem-001",
				Content:    "nmap scan showed SSH on port 22 with OpenSSH 8.9",
				MemoryType: "fact",
				Domain:     "pentest-recon",
				Confidence: 0.92,
				StoredAt:   "2024-01-15T10:30:00Z",
			},
			{
				MemoryID:   "mem-002",
				Content:    "gobuster found /admin and /api/debug endpoints",
				MemoryType: "observation",
				Domain:     "pentest-recon",
				Confidence: 0.85,
				StoredAt:   "2024-01-15T11:00:00Z",
			},
		},
	}

	tool := NewSageSearchTool(1, nil, nil, mock)

	if !tool.IsAvailable() {
		t.Fatal("tool should be available")
	}

	action := SageRecallAction{
		Query:   "SSH vulnerabilities from past scans",
		Domain:  "pentest-recon",
		Message: "Searching for SSH-related findings",
	}
	args, _ := json.Marshal(action)

	result, err := tool.Handle(context.Background(), SageRecallToolName, args)
	if err != nil {
		t.Fatalf("recall failed: %v", err)
	}

	t.Logf("result:\n%s", result)

	// Verify output contains expected content
	if !strings.Contains(result, "SAGE Cross-Session Memory Results") {
		t.Error("result should contain header")
	}
	if !strings.Contains(result, "nmap scan showed SSH") {
		t.Error("result should contain first memory content")
	}
	if !strings.Contains(result, "confidence: 0.92") {
		t.Error("result should contain confidence score")
	}
	if !strings.Contains(result, "pentest-recon") {
		t.Error("result should contain domain")
	}
}

func TestSageTool_RecallNoResults(t *testing.T) {
	mock := &mockSageClient{
		enabled:  true,
		memories: []sage.MemoryResult{},
	}

	tool := NewSageSearchTool(1, nil, nil, mock)
	action := SageRecallAction{
		Query:   "quantum computing attacks",
		Message: "Searching for quantum computing attacks",
	}
	args, _ := json.Marshal(action)

	result, err := tool.Handle(context.Background(), SageRecallToolName, args)
	if err != nil {
		t.Fatalf("recall failed: %v", err)
	}

	if !strings.Contains(result, "No cross-session memories found") {
		t.Errorf("expected no-results message, got: %s", result)
	}
}

func TestSageTool_Remember(t *testing.T) {
	mock := &mockSageClient{
		enabled:    true,
		rememberID: "mem-new-001",
	}

	tool := NewSageSearchTool(1, nil, nil, mock)
	action := SageRememberAction{
		Content:    "Target application vulnerable to SQL injection via the login form. Parameter: username. Payload: ' OR 1=1--",
		MemoryType: "fact",
		Domain:     "pentest-exploit",
		Confidence: 0.95,
		Message:    "Storing SQL injection finding",
	}
	args, _ := json.Marshal(action)

	result, err := tool.Handle(context.Background(), SageRememberToolName, args)
	if err != nil {
		t.Fatalf("remember failed: %v", err)
	}

	t.Logf("result:\n%s", result)

	if !strings.Contains(result, "SAGE Memory Stored") {
		t.Error("result should contain stored header")
	}
	if !strings.Contains(result, "mem-new-001") {
		t.Error("result should contain memory ID")
	}
}

func TestSageTool_NotAvailable(t *testing.T) {
	mock := &mockSageClient{enabled: false}
	tool := NewSageSearchTool(1, nil, nil, mock)

	if tool.IsAvailable() {
		t.Fatal("tool should not be available when disabled")
	}

	action := SageRecallAction{
		Query:   "test",
		Message: "test",
	}
	args, _ := json.Marshal(action)

	result, err := tool.Handle(context.Background(), SageRecallToolName, args)
	if err != nil {
		t.Fatalf("should not error when disabled: %v", err)
	}

	if !strings.Contains(result, "not enabled") {
		t.Errorf("expected not-enabled message, got: %s", result)
	}
}

func TestSageTool_NilClient(t *testing.T) {
	tool := NewSageSearchTool(1, nil, nil, nil)

	if tool.IsAvailable() {
		t.Fatal("tool should not be available with nil client")
	}
}

func TestSageTool_RememberValidation(t *testing.T) {
	mock := &mockSageClient{enabled: true, rememberID: "mem-x"}
	tool := NewSageSearchTool(1, nil, nil, mock)

	// Empty content should fail
	action := SageRememberAction{
		Content:    "",
		MemoryType: "fact",
		Domain:     "pentest-general",
		Confidence: 0.9,
		Message:    "test",
	}
	args, _ := json.Marshal(action)
	_, err := tool.Handle(context.Background(), SageRememberToolName, args)
	if err == nil {
		t.Error("empty content should fail")
	}

	// Empty query for recall should fail
	recallAction := SageRecallAction{
		Query:   "",
		Message: "test",
	}
	args, _ = json.Marshal(recallAction)
	_, err = tool.Handle(context.Background(), SageRecallToolName, args)
	if err == nil {
		t.Error("empty query should fail")
	}
}

func TestSageTool_RememberMissingMemoryType(t *testing.T) {
	mock := &mockSageClient{enabled: true, rememberID: "mem-x"}
	tool := NewSageSearchTool(1, nil, nil, mock)

	action := SageRememberAction{
		Content:    "some valid content",
		MemoryType: "",
		Domain:     "pentest-general",
		Confidence: 0.9,
		Message:    "test",
	}
	args, _ := json.Marshal(action)
	_, err := tool.Handle(context.Background(), SageRememberToolName, args)
	if err == nil {
		t.Error("empty memory_type should fail")
	}
}

func TestSageTool_RememberMissingDomain(t *testing.T) {
	mock := &mockSageClient{enabled: true, rememberID: "mem-x"}
	tool := NewSageSearchTool(1, nil, nil, mock)

	action := SageRememberAction{
		Content:    "some valid content",
		MemoryType: "fact",
		Domain:     "",
		Confidence: 0.9,
		Message:    "test",
	}
	args, _ := json.Marshal(action)
	_, err := tool.Handle(context.Background(), SageRememberToolName, args)
	if err == nil {
		t.Error("empty domain should fail")
	}
}

func TestSageTool_UnknownToolName(t *testing.T) {
	mock := &mockSageClient{enabled: true}
	tool := NewSageSearchTool(1, nil, nil, mock)

	args, _ := json.Marshal(map[string]string{"query": "test"})
	_, err := tool.Handle(context.Background(), "sage_unknown", args)
	if err == nil {
		t.Error("unknown tool name should return error")
	}
}
