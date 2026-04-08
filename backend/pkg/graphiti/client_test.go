package graphiti

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Disabled(t *testing.T) {
	t.Parallel()

	client, err := NewClient("", 0, false)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.False(t, client.IsEnabled())
}

func TestIsEnabled(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver returns false", func(t *testing.T) {
		t.Parallel()
		var c *Client
		assert.False(t, c.IsEnabled())
	})

	t.Run("disabled client returns false", func(t *testing.T) {
		t.Parallel()
		c := &Client{enabled: false}
		assert.False(t, c.IsEnabled())
	})

	t.Run("enabled client returns true", func(t *testing.T) {
		t.Parallel()
		c := &Client{enabled: true}
		assert.True(t, c.IsEnabled())
	})
}

func TestGetTimeout(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver returns zero", func(t *testing.T) {
		t.Parallel()
		var c *Client
		assert.Equal(t, time.Duration(0), c.GetTimeout())
	})

	t.Run("returns configured timeout", func(t *testing.T) {
		t.Parallel()
		c := &Client{timeout: 30 * time.Second}
		assert.Equal(t, 30*time.Second, c.GetTimeout())
	})
}

func TestAddMessages_Disabled(t *testing.T) {
	t.Parallel()

	client, err := NewClient("", 0, false)
	require.NoError(t, err)

	err = client.AddMessages(context.Background(), AddMessagesRequest{})
	assert.NoError(t, err)
}

func TestSearchMethods_Disabled(t *testing.T) {
	t.Parallel()

	client, err := NewClient("", 0, false)
	require.NoError(t, err)

	tests := []struct {
		name string
		fn   func() (any, error)
	}{
		{
			"TemporalWindowSearch",
			func() (any, error) {
				return client.TemporalWindowSearch(context.Background(), TemporalSearchRequest{})
			},
		},
		{
			"EntityRelationshipsSearch",
			func() (any, error) {
				return client.EntityRelationshipsSearch(context.Background(), EntityRelationshipSearchRequest{})
			},
		},
		{
			"DiverseResultsSearch",
			func() (any, error) {
				return client.DiverseResultsSearch(context.Background(), DiverseSearchRequest{})
			},
		},
		{
			"EpisodeContextSearch",
			func() (any, error) {
				return client.EpisodeContextSearch(context.Background(), EpisodeContextSearchRequest{})
			},
		},
		{
			"SuccessfulToolsSearch",
			func() (any, error) {
				return client.SuccessfulToolsSearch(context.Background(), SuccessfulToolsSearchRequest{})
			},
		},
		{
			"RecentContextSearch",
			func() (any, error) {
				return client.RecentContextSearch(context.Background(), RecentContextSearchRequest{})
			},
		},
		{
			"EntityByLabelSearch",
			func() (any, error) {
				return client.EntityByLabelSearch(context.Background(), EntityByLabelSearchRequest{})
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp, err := tt.fn()
			assert.Nil(t, resp)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "graphiti is not enabled")
		})
	}
}
