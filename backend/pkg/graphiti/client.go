package graphiti

import (
	"context"
	"fmt"
	"time"

	graphiti "github.com/vxcontrol/graphiti-go-client"
)

// Re-export types from the graphiti-go-client package for convenience
type (
	Message            = graphiti.Message
	AddMessagesRequest = graphiti.AddMessagesRequest

	// Search request/response types
	TemporalSearchRequest            = graphiti.TemporalSearchRequest
	TemporalSearchResponse           = graphiti.TemporalSearchResponse
	EntityRelationshipSearchRequest  = graphiti.EntityRelationshipSearchRequest
	EntityRelationshipSearchResponse = graphiti.EntityRelationshipSearchResponse
	DiverseSearchRequest             = graphiti.DiverseSearchRequest
	DiverseSearchResponse            = graphiti.DiverseSearchResponse
	EpisodeContextSearchRequest      = graphiti.EpisodeContextSearchRequest
	EpisodeContextSearchResponse     = graphiti.EpisodeContextSearchResponse
	SuccessfulToolsSearchRequest     = graphiti.SuccessfulToolsSearchRequest
	SuccessfulToolsSearchResponse    = graphiti.SuccessfulToolsSearchResponse
	RecentContextSearchRequest       = graphiti.RecentContextSearchRequest
	RecentContextSearchResponse      = graphiti.RecentContextSearchResponse
	EntityByLabelSearchRequest       = graphiti.EntityByLabelSearchRequest
	EntityByLabelSearchResponse      = graphiti.EntityByLabelSearchResponse

	// Common types used in search responses
	NodeResult      = graphiti.NodeResult
	EdgeResult      = graphiti.EdgeResult
	EpisodeResult   = graphiti.EpisodeResult
	CommunityResult = graphiti.CommunityResult
	TimeWindow      = graphiti.TimeWindow
)

// Client wraps the Graphiti client with Pentagi-specific functionality
type Client struct {
	client  *graphiti.Client
	enabled bool
	timeout time.Duration
}

// NewClient creates a new Graphiti client wrapper
func NewClient(url string, timeout time.Duration, enabled bool) (*Client, error) {
	if !enabled {
		return &Client{enabled: false}, nil
	}

	client := graphiti.NewClient(url, graphiti.WithTimeout(timeout))

	_, err := client.HealthCheck()
	if err != nil {
		return nil, fmt.Errorf("graphiti health check failed: %w", err)
	}

	return &Client{
		client:  client,
		enabled: true,
		timeout: timeout,
	}, nil
}

// IsEnabled returns whether Graphiti integration is active
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// GetTimeout returns the configured timeout duration
func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}

// AddMessages adds messages to Graphiti (no-op if disabled)
func (c *Client) AddMessages(ctx context.Context, req graphiti.AddMessagesRequest) error {
	if !c.enabled {
		return nil
	}

	_, err := c.client.AddMessages(req)
	return err
}

// TemporalWindowSearch searches within a time window
func (c *Client) TemporalWindowSearch(ctx context.Context, req TemporalSearchRequest) (*TemporalSearchResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("graphiti is not enabled")
	}
	return c.client.TemporalWindowSearch(req)
}

// EntityRelationshipsSearch finds relationships from a center node
func (c *Client) EntityRelationshipsSearch(ctx context.Context, req EntityRelationshipSearchRequest) (*EntityRelationshipSearchResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("graphiti is not enabled")
	}
	return c.client.EntityRelationshipsSearch(req)
}

// DiverseResultsSearch gets diverse, non-redundant results
func (c *Client) DiverseResultsSearch(ctx context.Context, req DiverseSearchRequest) (*DiverseSearchResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("graphiti is not enabled")
	}
	return c.client.DiverseResultsSearch(req)
}

// EpisodeContextSearch searches through agent responses and tool execution records
func (c *Client) EpisodeContextSearch(ctx context.Context, req EpisodeContextSearchRequest) (*EpisodeContextSearchResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("graphiti is not enabled")
	}
	return c.client.EpisodeContextSearch(req)
}

// SuccessfulToolsSearch finds successful tool executions and attack patterns
func (c *Client) SuccessfulToolsSearch(ctx context.Context, req SuccessfulToolsSearchRequest) (*SuccessfulToolsSearchResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("graphiti is not enabled")
	}
	return c.client.SuccessfulToolsSearch(req)
}

// RecentContextSearch retrieves recent relevant context
func (c *Client) RecentContextSearch(ctx context.Context, req RecentContextSearchRequest) (*RecentContextSearchResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("graphiti is not enabled")
	}
	return c.client.RecentContextSearch(req)
}

// EntityByLabelSearch searches for entities by label/type
func (c *Client) EntityByLabelSearch(ctx context.Context, req EntityByLabelSearchRequest) (*EntityByLabelSearchResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("graphiti is not enabled")
	}
	return c.client.EntityByLabelSearch(req)
}
