package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
)

const (
	sploitusAPIURL         = "https://sploitus.com/search"
	sploitusDefaultSort    = "default"
	defaultSploitusLimit   = 10
	maxSploitusLimit       = 25
	defaultSploitusType    = "exploits"
	sploitusRequestTimeout = 30 * time.Second

	// Hard limits to prevent memory overflow and excessive response sizes
	maxSourceSize       = 50 * 1024 // 50 KB max per source field
	maxTotalResultSize  = 80 * 1024 // 80 KB total output limit
	truncationMsgBuffer = 500       // Reserve space for truncation message
)

// sploitus represents the Sploitus exploit search tool
type sploitus struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64
	enabled   bool
	proxyURL  string
	slp       SearchLogProvider
}

// NewSploitusTool creates a new Sploitus search tool instance
func NewSploitusTool(
	flowID int64,
	taskID, subtaskID *int64,
	enabled bool,
	proxyURL string,
	slp SearchLogProvider,
) Tool {
	return &sploitus{
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		enabled:   enabled,
		proxyURL:  proxyURL,
		slp:       slp,
	}
}

// IsAvailable returns true if the Sploitus tool is enabled and configured
func (s *sploitus) IsAvailable() bool {
	return s.enabled && s.slp != nil
}

// Handle processes a Sploitus exploit search request from an AI agent
func (s *sploitus) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action SploitusAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(s.flowID, s.taskID, s.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal sploitus search action")
		return "", fmt.Errorf("failed to unmarshal %s search action arguments: %w", name, err)
	}

	// Normalise exploit type
	exploitType := strings.ToLower(strings.TrimSpace(action.ExploitType))
	if exploitType == "" {
		exploitType = defaultSploitusType
	}

	// Normalise sort order
	sort := strings.ToLower(strings.TrimSpace(action.Sort))
	if sort == "" {
		sort = sploitusDefaultSort
	}

	// Clamp max results
	limit := action.MaxResults.Int()
	if limit < 1 || limit > maxSploitusLimit {
		limit = defaultSploitusLimit
	}

	logger = logger.WithFields(logrus.Fields{
		"query":        action.Query[:min(len(action.Query), 1000)],
		"exploit_type": exploitType,
		"sort":         sort,
		"limit":        limit,
	})

	result, err := s.search(ctx, action.Query, exploitType, sort, limit)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("sploitus search error swallowed"),
			langfuse.WithEventInput(action.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name":    SploitusToolName,
				"engine":       "sploitus",
				"query":        action.Query,
				"exploit_type": exploitType,
				"sort":         sort,
				"limit":        limit,
				"error":        err.Error(),
			}),
		)

		logger.WithError(err).Error("failed to search in Sploitus")
		return fmt.Sprintf("failed to search in Sploitus: %v", err), nil
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = s.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeSploitus,
			action.Query,
			result,
			s.taskID,
			s.subtaskID,
		)
	}

	return result, nil
}

// sploitusRequest is the JSON body sent to the Sploitus search API
type sploitusRequest struct {
	Query  string `json:"query"`
	Type   string `json:"type"`
	Sort   string `json:"sort"`
	Title  bool   `json:"title"`
	Offset int    `json:"offset"`
}

// sploitusExploit represents a single exploit record returned by Sploitus
// The API returns the same structure for both exploits and tools
type sploitusExploit struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Type      string  `json:"type"`
	Href      string  `json:"href"`
	Download  string  `json:"download,omitempty"`  // Only present for tools
	Score     float64 `json:"score,omitempty"`     // CVSS score, only for exploits
	Published string  `json:"published,omitempty"` // Publication date, only for exploits
	Source    string  `json:"source,omitempty"`    // Source code/description, only for exploits
	Language  string  `json:"language,omitempty"`  // Programming language, only for exploits
}

// sploitusResponse is the top-level JSON response from the Sploitus API
type sploitusResponse struct {
	Exploits      []sploitusExploit `json:"exploits"`
	ExploitsTotal int               `json:"exploits_total"`
}

// search calls the Sploitus API and returns a formatted markdown result string
func (s *sploitus) search(ctx context.Context, query, exploitType, sort string, limit int) (string, error) {
	reqBody := sploitusRequest{
		Query:  query,
		Type:   exploitType,
		Sort:   sort,
		Title:  false, // search only for titles
		Offset: 0,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Build HTTP client, optionally routed through a proxy
	httpClient := &http.Client{Timeout: sploitusRequestTimeout}
	if s.proxyURL != "" {
		proxyParsed, parseErr := url.Parse(s.proxyURL)
		if parseErr != nil {
			return "", fmt.Errorf("invalid proxy URL: %w", parseErr)
		}
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyParsed)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sploitusAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Build referer with query to mimic browser behavior
	referer := fmt.Sprintf("https://sploitus.com/?query=%s", url.QueryEscape(query))

	// Mimic Chrome browser headers to bypass Cloudflare protection
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://sploitus.com")
	req.Header.Set("Referer", referer)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("DNT", "1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to Sploitus failed: %w", err)
	}
	defer resp.Body.Close()

	// Sploitus API returns 499 when rate limit is temporarily exceeded
	if resp.StatusCode == 499 || resp.StatusCode == 422 {
		return "", fmt.Errorf("Sploitus API rate limit exceeded (HTTP %d), please try again later", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Sploitus API returned HTTP %d", resp.StatusCode)
	}

	var apiResp sploitusResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode Sploitus response: %w", err)
	}

	return formatSploitusResults(query, exploitType, limit, apiResp), nil
}

// formatSploitusResults converts a sploitusResponse into a human-readable markdown string
func formatSploitusResults(query, exploitType string, limit int, resp sploitusResponse) string {
	var sb strings.Builder

	sb.WriteString("# Sploitus Search Results\n\n")
	sb.WriteString(fmt.Sprintf("**Query:** `%s`  \n", query))
	sb.WriteString(fmt.Sprintf("**Type:** %s  \n", exploitType))
	sb.WriteString(fmt.Sprintf("**Total matches on Sploitus:** %d\n\n", resp.ExploitsTotal))
	sb.WriteString("---\n\n")

	// Ensure limit is positive
	if limit < 1 {
		limit = defaultSploitusLimit
	}

	results := resp.Exploits
	if len(results) > limit {
		results = results[:limit]
	}

	if len(results) == 0 {
		switch strings.ToLower(exploitType) {
		case "tools":
			sb.WriteString("No security tools were found for the given query.\n")
		default:
			sb.WriteString("No exploits were found for the given query.\n")
		}
		return sb.String()
	}

	// Track total size to enforce hard limit
	currentSize := len(sb.String())
	actualShown := 0
	truncatedBySize := false

	switch strings.ToLower(exploitType) {
	case "tools":
		sb.WriteString(fmt.Sprintf("## Security Tools (showing up to %d)\n\n", len(results)))
		currentSize = len(sb.String())

		for i, item := range results {
			// Check if we're approaching the size limit (reserve space for truncation message)
			if currentSize >= maxTotalResultSize-truncationMsgBuffer {
				truncatedBySize = true
				break
			}

			var itemBuilder strings.Builder
			itemBuilder.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, item.Title))
			if item.Href != "" {
				itemBuilder.WriteString(fmt.Sprintf("**URL:** %s  \n", item.Href))
			}
			if item.Download != "" {
				itemBuilder.WriteString(fmt.Sprintf("**Download:** %s  \n", item.Download))
			}
			if item.Type != "" {
				itemBuilder.WriteString(fmt.Sprintf("**Source Type:** %s  \n", item.Type))
			}
			if item.ID != "" {
				itemBuilder.WriteString(fmt.Sprintf("**ID:** %s  \n", item.ID))
			}
			itemBuilder.WriteString("\n---\n\n")

			itemContent := itemBuilder.String()
			// Check if adding this item would exceed limit (with buffer for truncation msg)
			if currentSize+len(itemContent) > maxTotalResultSize-truncationMsgBuffer {
				truncatedBySize = true
				break
			}

			sb.WriteString(itemContent)
			currentSize += len(itemContent)
			actualShown++
		}

	default: // "exploits" or anything else
		sb.WriteString(fmt.Sprintf("## Exploits (showing up to %d)\n\n", len(results)))
		currentSize = len(sb.String())

		for i, item := range results {
			// Check if we're approaching the size limit (reserve space for truncation message)
			if currentSize >= maxTotalResultSize-truncationMsgBuffer {
				truncatedBySize = true
				break
			}

			var itemBuilder strings.Builder
			itemBuilder.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, item.Title))
			if item.Href != "" {
				itemBuilder.WriteString(fmt.Sprintf("**URL:** %s  \n", item.Href))
			}
			if item.Score > 0 {
				itemBuilder.WriteString(fmt.Sprintf("**CVSS Score:** %.1f  \n", item.Score))
			}
			if item.Type != "" {
				itemBuilder.WriteString(fmt.Sprintf("**Type:** %s  \n", item.Type))
			}
			if item.Published != "" {
				itemBuilder.WriteString(fmt.Sprintf("**Published:** %s  \n", item.Published))
			}
			if item.ID != "" {
				itemBuilder.WriteString(fmt.Sprintf("**ID:** %s  \n", item.ID))
			}
			if item.Language != "" {
				itemBuilder.WriteString(fmt.Sprintf("**Language:** %s  \n", item.Language))
			}

			// Truncate source if it's too large (hard limit: 50 KB)
			if item.Source != "" {
				sourcePreview := item.Source
				if len(sourcePreview) > maxSourceSize {
					sourcePreview = sourcePreview[:maxSourceSize] + "\n... [source truncated, exceeded 50 KB limit]"
				}
				itemBuilder.WriteString(fmt.Sprintf("\n**Source Preview:**\n```\n%s\n```\n", sourcePreview))
			}
			itemBuilder.WriteString("\n---\n\n")

			itemContent := itemBuilder.String()
			// Check if adding this item would exceed limit (with buffer for truncation msg)
			if currentSize+len(itemContent) > maxTotalResultSize-truncationMsgBuffer {
				truncatedBySize = true
				break
			}

			sb.WriteString(itemContent)
			currentSize += len(itemContent)
			actualShown++
		}
	}

	// Add warning if results were truncated due to size limit
	if truncatedBySize {
		sb.WriteString(fmt.Sprintf(
			"\n\n**⚠️ Note:** Results truncated after %d items due to %d bytes size limit. Total shown: %d of %d available.\n",
			actualShown, maxTotalResultSize, actualShown, len(results),
		))
	}

	return sb.String()
}
