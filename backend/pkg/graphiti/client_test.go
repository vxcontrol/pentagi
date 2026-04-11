package graphiti

import (
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

	err = client.AddMessages(t.Context(), AddMessagesRequest{})
	assert.NoError(t, err)
}

func TestSearchMethods_Disabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(t *testing.T, client *Client)
	}{
		{
			"TemporalWindowSearch",
			func(t *testing.T, client *Client) {
				resp, err := client.TemporalWindowSearch(t.Context(), TemporalSearchRequest{})
				assert.Nil(t, resp, "response must be nil when disabled")
				require.Error(t, err)
				require.ErrorContains(t, err, "graphiti is not enabled")
			},
		},
		{
			"EntityRelationshipsSearch",
			func(t *testing.T, client *Client) {
				resp, err := client.EntityRelationshipsSearch(t.Context(), EntityRelationshipSearchRequest{})
				assert.Nil(t, resp, "response must be nil when disabled")
				require.Error(t, err)
				require.ErrorContains(t, err, "graphiti is not enabled")
			},
		},
		{
			"DiverseResultsSearch",
			func(t *testing.T, client *Client) {
				resp, err := client.DiverseResultsSearch(t.Context(), DiverseSearchRequest{})
				assert.Nil(t, resp, "response must be nil when disabled")
				require.Error(t, err)
				require.ErrorContains(t, err, "graphiti is not enabled")
			},
		},
		{
			"EpisodeContextSearch",
			func(t *testing.T, client *Client) {
				resp, err := client.EpisodeContextSearch(t.Context(), EpisodeContextSearchRequest{})
				assert.Nil(t, resp, "response must be nil when disabled")
				require.Error(t, err)
				require.ErrorContains(t, err, "graphiti is not enabled")
			},
		},
		{
			"SuccessfulToolsSearch",
			func(t *testing.T, client *Client) {
				resp, err := client.SuccessfulToolsSearch(t.Context(), SuccessfulToolsSearchRequest{})
				assert.Nil(t, resp, "response must be nil when disabled")
				require.Error(t, err)
				require.ErrorContains(t, err, "graphiti is not enabled")
			},
		},
		{
			"RecentContextSearch",
			func(t *testing.T, client *Client) {
				resp, err := client.RecentContextSearch(t.Context(), RecentContextSearchRequest{})
				assert.Nil(t, resp, "response must be nil when disabled")
				require.Error(t, err)
				require.ErrorContains(t, err, "graphiti is not enabled")
			},
		},
		{
			"EntityByLabelSearch",
			func(t *testing.T, client *Client) {
				resp, err := client.EntityByLabelSearch(t.Context(), EntityByLabelSearchRequest{})
				assert.Nil(t, resp, "response must be nil when disabled")
				require.Error(t, err)
				require.ErrorContains(t, err, "graphiti is not enabled")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create fresh disabled client per subtest to avoid shared state
			client, err := NewClient("", 0, false)
			require.NoError(t, err)
			tt.fn(t, client)
		})
	}
}
