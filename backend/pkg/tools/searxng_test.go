package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pentagi/pkg/database"
)

// MockSearchLogProvider implements the SearchLogProvider interface for testing
type MockSearchLogProvider struct{}

func (m *MockSearchLogProvider) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	engine database.SearchengineType,
	query string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	// Mock implementation - just return a dummy ID
	return 1, nil
}

// MockSummarizer implements a simple mock summarizer
func MockSummarizer(ctx context.Context, result string) (string, error) {
	return "Mock summarized: " + result, nil
}

func TestNewSearxngTool(t *testing.T) {
	flowID := int64(123)
	taskID := int64(456)
	subtaskID := int64(789)
	baseURL := "https://searxng.example.com"
	categories := "general"
	language := "en"
	safeSearch := "0"
	timeRange := ""
	proxyURL := ""
	timeout := 30
	searchLog := &MockSearchLogProvider{}
	summarizer := MockSummarizer

	tool := NewSearxngTool(
		flowID,
		&taskID,
		&subtaskID,
		baseURL,
		categories,
		language,
		safeSearch,
		timeRange,
		proxyURL,
		timeout,
		searchLog,
		summarizer,
	)

	if tool == nil {
		t.Fatal("NewSearxngTool returned nil")
	}

	if tool.flowID != flowID {
		t.Errorf("Expected flowID %d, got %d", flowID, tool.flowID)
	}

	if tool.taskID == nil || *tool.taskID != taskID {
		t.Errorf("Expected taskID %d, got %v", taskID, tool.taskID)
	}

	if tool.subtaskID == nil || *tool.subtaskID != subtaskID {
		t.Errorf("Expected subtaskID %d, got %v", subtaskID, tool.subtaskID)
	}

	if tool.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, tool.baseURL)
	}

	if tool.categories != categories {
		t.Errorf("Expected categories %s, got %s", categories, tool.categories)
	}

	if tool.language != language {
		t.Errorf("Expected language %s, got %s", language, tool.language)
	}

	if tool.safeSearch != safeSearch {
		t.Errorf("Expected safeSearch %s, got %s", safeSearch, tool.safeSearch)
	}

	if tool.timeRange != timeRange {
		t.Errorf("Expected timeRange %s, got %s", timeRange, tool.timeRange)
	}

	if tool.proxyURL != proxyURL {
		t.Errorf("Expected proxyURL %s, got %s", proxyURL, tool.proxyURL)
	}

	if tool.timeout != time.Duration(timeout)*time.Second {
		t.Errorf("Expected timeout %v, got %v", time.Duration(timeout)*time.Second, tool.timeout)
	}

	if tool.searchLog != searchLog {
		t.Error("Expected searchLog to be set")
	}

	if tool.summarizer == nil {
		t.Error("Expected summarizer to be set")
	}
}

func TestSearxngToolIsAvailable(t *testing.T) {
	// Test with URL and searchLog available
	tool1 := &SearxngTool{
		baseURL:   "https://searxng.example.com",
		searchLog: &MockSearchLogProvider{},
	}
	if !tool1.IsAvailable() {
		t.Error("Expected tool to be available when URL and searchLog are set")
	}

	// Test with URL empty
	tool2 := &SearxngTool{
		baseURL: "",
	}
	if tool2.IsAvailable() {
		t.Error("Expected tool to be unavailable when URL is empty")
	}

	// Test with searchLog nil
	tool3 := &SearxngTool{
		baseURL:   "https://searxng.example.com",
		searchLog: nil,
	}
	if tool3.IsAvailable() {
		t.Error("Expected tool to be unavailable when searchLog is nil")
	}
}

func TestSearxngToolHandleWithInvalidArgs(t *testing.T) {
	tool := &SearxngTool{
		baseURL:   "https://searxng.example.com",
		searchLog: &MockSearchLogProvider{},
	}

	// Test with invalid JSON args
	_, err := tool.Handle(context.Background(), SearxngToolName, []byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON args")
	}

	// Test with empty query
	args := SearchAction{
		Query: "",
	}
	argsJSON, _ := json.Marshal(args)
	_, err = tool.Handle(context.Background(), SearxngToolName, argsJSON)
	if err == nil {
		t.Error("Expected error for empty query")
	}
}

func TestSearxngToolHandleWithValidArgs(t *testing.T) {
	// Create a mock server that returns valid searxng response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := SearxngResponse{
			Results: []SearxngResult{
				{
					Title:   "Test Result",
					URL:     "https://example.com/test",
					Content: "This is a test result",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	tool := &SearxngTool{
		baseURL:   mockServer.URL,
		searchLog: &MockSearchLogProvider{},
		summarizer: func(ctx context.Context, result string) (string, error) {
			return result, nil // Don't modify result for this test
		},
	}

	args := SearchAction{
		Query:      "test query",
		MaxResults: 5,
	}
	argsJSON, _ := json.Marshal(args)

	result, err := tool.Handle(context.Background(), SearxngToolName, argsJSON)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Check if result contains expected content
	if !contains(result, "Test Result") {
		t.Error("Expected result to contain 'Test Result'")
	}
	if !contains(result, "https://example.com/test") {
		t.Error("Expected result to contain URL")
	}
	if !contains(result, "This is a test result") {
		t.Error("Expected result to contain content")
	}
}

func TestSearxngToolHandleWithServerError(t *testing.T) {
	// Create a mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	tool := &SearxngTool{
		baseURL:   mockServer.URL,
		searchLog: &MockSearchLogProvider{},
		summarizer: func(ctx context.Context, result string) (string, error) {
			return result, nil
		},
	}

	args := SearchAction{
		Query:      "test query",
		MaxResults: 5,
	}
	argsJSON, _ := json.Marshal(args)

	_, err := tool.Handle(context.Background(), SearxngToolName, argsJSON)
	if err == nil {
		t.Error("Expected error for server error response")
	}
}

func TestSearxngToolHandleWithInvalidURL(t *testing.T) {
	tool := &SearxngTool{
		baseURL:   "invalid-url",
		searchLog: &MockSearchLogProvider{},
		summarizer: func(ctx context.Context, result string) (string, error) {
			return result, nil
		},
	}

	args := SearchAction{
		Query:      "test query",
		MaxResults: 5,
	}
	argsJSON, _ := json.Marshal(args)

	_, err := tool.Handle(context.Background(), SearxngToolName, argsJSON)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestSearxngToolHandleWithNoResults(t *testing.T) {
	// Create a mock server that returns no results
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := SearxngResponse{
			Results: []SearxngResult{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	tool := &SearxngTool{
		baseURL:   mockServer.URL,
		searchLog: &MockSearchLogProvider{},
		summarizer: func(ctx context.Context, result string) (string, error) {
			return result, nil
		},
	}

	args := SearchAction{
		Query:      "test query",
		MaxResults: 5,
	}
	argsJSON, _ := json.Marshal(args)

	result, err := tool.Handle(context.Background(), SearxngToolName, argsJSON)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Check if result contains expected "No Results Found" message
	if !contains(result, "No Results Found") {
		t.Error("Expected result to contain 'No Results Found'")
	}
	if !contains(result, "test query") {
		t.Error("Expected result to contain the query")
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr)))
}

// Helper function to check substring in a more robust way
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
