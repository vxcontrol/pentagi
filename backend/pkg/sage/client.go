package sage

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Client wraps the SAGE REST API with PentAGI-specific functionality.
// Authentication uses Ed25519 signed requests per SAGE protocol.
type Client struct {
	endpoint   string
	agentName  string
	agentID    string // hex-encoded Ed25519 public key
	keyPath    string
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	enabled    bool
	timeout    time.Duration
	httpClient *http.Client
}

// loadOrCreateKey loads an Ed25519 keypair from disk, or generates and
// persists a new one. This ensures the agent keeps the same identity
// across restarts so SAGE domain ownership is preserved.
func loadOrCreateKey(keyPath string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	if keyPath == "" {
		keyPath = filepath.Join(os.TempDir(), "pentagi-sage-agent.key")
	}

	// Try to load existing key file.
	data, err := os.ReadFile(keyPath)
	if err == nil && len(data) == ed25519.PrivateKeySize {
		priv := ed25519.PrivateKey(data)
		pub := priv.Public().(ed25519.PublicKey)
		return pub, priv, nil
	}

	// Generate a fresh keypair and persist it.
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 keypair: %w", err)
	}

	// Ensure parent directory exists.
	if dir := filepath.Dir(keyPath); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, nil, fmt.Errorf("failed to create key directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(keyPath, []byte(priv), 0600); err != nil {
		return nil, nil, fmt.Errorf("failed to write key file %s: %w", keyPath, err)
	}

	return pub, priv, nil
}

// NewClient creates a new SAGE client wrapper.
// It loads (or generates) a persistent Ed25519 keypair for request signing
// and verifies the SAGE node is reachable via health check.
func NewClient(endpoint, keyPath, agentName string, timeout time.Duration, enabled bool) (*Client, error) {
	if !enabled {
		return &Client{enabled: false}, nil
	}

	pub, priv, err := loadOrCreateKey(keyPath)
	if err != nil {
		return nil, err
	}

	client := &Client{
		endpoint:   endpoint,
		agentName:  agentName,
		agentID:    hex.EncodeToString(pub),
		keyPath:    keyPath,
		publicKey:  pub,
		privateKey: priv,
		enabled:    true,
		timeout:    timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}

	if err := client.HealthCheck(context.Background()); err != nil {
		return nil, fmt.Errorf("sage health check failed: %w", err)
	}

	return client, nil
}

// IsEnabled returns whether SAGE integration is active.
func (c *Client) IsEnabled() bool {
	return c != nil && c.enabled
}

// GetTimeout returns the configured timeout duration.
func (c *Client) GetTimeout() time.Duration {
	if c == nil {
		return 0
	}
	return c.timeout
}

// HealthCheck verifies the SAGE node is reachable (no auth needed).
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sage health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sage health check returned status %d", resp.StatusCode)
	}

	return nil
}

// Embed generates an embedding vector for the given text.
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("sage is not enabled")
	}

	body := map[string]string{"text": text}
	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := c.doJSON(ctx, http.MethodPost, "/v1/embed", body, &result); err != nil {
		return nil, fmt.Errorf("sage embed failed: %w", err)
	}

	return result.Embedding, nil
}

// Remember stores a memory in SAGE. It generates an embedding and submits
// the memory for BFT consensus validation.
func (c *Client) Remember(ctx context.Context, req RememberRequest) (*RememberResponse, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("sage is not enabled")
	}

	// Generate embedding for the content
	embedding, err := c.Embed(ctx, req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for memory: %w", err)
	}

	submitReq := submitMemoryRequest{
		Content:         req.Content,
		MemoryType:      req.MemoryType,
		DomainTag:       req.Domain,
		ConfidenceScore: req.Confidence,
		Embedding:       embedding,
		Provider:        c.agentName,
	}

	var resp RememberResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/memory/submit", submitReq, &resp); err != nil {
		return &RememberResponse{Success: false, Message: err.Error()}, nil
	}

	resp.Success = true
	if resp.Message == "" {
		resp.Message = resp.Status
	}

	return &resp, nil
}

// Recall searches SAGE memories using FTS5 full-text search (no embedding needed).
func (c *Client) Recall(ctx context.Context, req RecallRequest) (*RecallResponse, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("sage is not enabled")
	}

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}

	minConf := req.MinConfidence
	if minConf <= 0 {
		minConf = 0.6
	}

	searchReq := searchMemoryRequest{
		Query:         req.Query,
		DomainTag:     req.Domain,
		TopK:          maxResults,
		StatusFilter:  "committed",
		MinConfidence: minConf,
	}

	var resp RecallResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/memory/search", searchReq, &resp); err != nil {
		return nil, fmt.Errorf("sage recall failed: %w", err)
	}

	return &resp, nil
}

// RecallSemantic searches SAGE memories using vector similarity (generates embedding first).
func (c *Client) RecallSemantic(ctx context.Context, req RecallRequest) (*RecallResponse, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("sage is not enabled")
	}

	embedding, err := c.Embed(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}

	minConf := req.MinConfidence
	if minConf <= 0 {
		minConf = 0.6
	}

	queryReq := queryMemoryRequest{
		Embedding:     embedding,
		DomainTag:     req.Domain,
		Provider:      c.agentName,
		TopK:          maxResults,
		StatusFilter:  "committed",
		MinConfidence: minConf,
	}

	var resp RecallResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/memory/query", queryReq, &resp); err != nil {
		return nil, fmt.Errorf("sage semantic recall failed: %w", err)
	}

	return &resp, nil
}

// --- Ed25519 request signing --------------------------------------------------

// signRequest produces the Ed25519 signature for a SAGE API request.
// Signed message format: SHA-256(method + " " + path + "\n" + body) || BigEndian(timestamp)
func (c *Client) signRequest(method, path string, body []byte, timestamp int64) []byte {
	canonical := []byte(method + " " + path + "\n")
	canonical = append(canonical, body...)
	bodyHash := sha256.Sum256(canonical)

	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))

	message := make([]byte, 0, len(bodyHash)+8)
	message = append(message, bodyHash[:]...)
	message = append(message, tsBytes...)

	return ed25519.Sign(c.privateKey, message)
}

// --- Internal HTTP helper -----------------------------------------------------

func (c *Client) doJSON(ctx context.Context, method, path string, reqBody, respBody any) error {
	var bodyBytes []byte
	if reqBody != nil {
		var err error
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// SAGE Ed25519 authentication headers
	timestamp := time.Now().Unix()
	sig := c.signRequest(method, path, bodyBytes, timestamp)
	req.Header.Set("X-Agent-ID", c.agentID)
	req.Header.Set("X-Signature", hex.EncodeToString(sig))
	req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("SAGE API returned status %d: %s", resp.StatusCode, string(respData))
	}

	if respBody != nil {
		if err := json.Unmarshal(respData, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// --- Request / Response types -------------------------------------------------

// RememberRequest is the input for storing a memory in SAGE.
type RememberRequest struct {
	Content    string  `json:"content"`
	MemoryType string  `json:"memory_type"` // fact, observation, inference
	Domain     string  `json:"domain"`
	Confidence float64 `json:"confidence"`
}

// RememberResponse is returned after submitting a memory.
type RememberResponse struct {
	MemoryID string `json:"memory_id"`
	TxHash   string `json:"tx_hash"`
	Status   string `json:"status"`
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
}

// RecallRequest is the input for searching SAGE memories.
type RecallRequest struct {
	Query         string  `json:"query"`
	Domain        string  `json:"domain,omitempty"`
	MaxResults    int     `json:"max_results,omitempty"`
	MinConfidence float64 `json:"min_confidence,omitempty"`
}

// RecallResponse is returned from a memory search.
type RecallResponse struct {
	Memories   []MemoryResult `json:"results"`
	TotalCount int            `json:"total_count"`
}

// MemoryResult is a single memory record from SAGE.
type MemoryResult struct {
	MemoryID   string  `json:"memory_id"`
	Content    string  `json:"content"`
	MemoryType string  `json:"memory_type"`
	Domain     string  `json:"domain_tag"`
	Confidence float64 `json:"confidence_score"`
	StoredAt   string  `json:"created_at"`
}

// --- Internal API request types -----------------------------------------------

type submitMemoryRequest struct {
	Content         string    `json:"content"`
	MemoryType      string    `json:"memory_type"`
	DomainTag       string    `json:"domain_tag"`
	ConfidenceScore float64   `json:"confidence_score"`
	Embedding       []float32 `json:"embedding,omitempty"`
	Provider        string    `json:"provider,omitempty"`
}

type searchMemoryRequest struct {
	Query         string  `json:"query"`
	DomainTag     string  `json:"domain_tag,omitempty"`
	TopK          int     `json:"top_k,omitempty"`
	StatusFilter  string  `json:"status_filter,omitempty"`
	MinConfidence float64 `json:"min_confidence,omitempty"`
}

type queryMemoryRequest struct {
	Embedding     []float32 `json:"embedding"`
	DomainTag     string    `json:"domain_tag,omitempty"`
	Provider      string    `json:"provider,omitempty"`
	TopK          int       `json:"top_k,omitempty"`
	StatusFilter  string    `json:"status_filter,omitempty"`
	MinConfidence float64   `json:"min_confidence,omitempty"`
}
