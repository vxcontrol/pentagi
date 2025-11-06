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
