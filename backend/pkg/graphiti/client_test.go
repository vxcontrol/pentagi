package graphiti

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_NewClient_EnablesOnlyAfterAHealthyProbe(t *testing.T) {
	t.Parallel()

	t.Run("a disabled client probes nothing", func(t *testing.T) {
		t.Parallel()
		client, err := NewClient("", 0, false)
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.False(t, client.IsEnabled())
	})

	tests := []struct {
		name       string
		failures   int32
		wantErr    string
		wantProbes int32
	}{
		{name: "a flaky probe is retried", failures: 1, wantProbes: 2},
		{name: "every probe failing gives up", failures: 3, wantErr: "graphiti health check failed after 3 attempts", wantProbes: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var probes atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if probes.Add(1) <= tt.failures {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			defer server.Close()

			client, err := NewClient(server.URL, 5*time.Second, true)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.True(t, client.IsEnabled())
				assert.Equal(t, 5*time.Second, client.GetTimeout())
			}
			assert.Equal(t, tt.wantProbes, probes.Load())
		})
	}
}

// Covers IsEnabled and GetTimeout on a nil receiver, which callers holding no client rely on.
func TestClient_NilClientIsDisabledWithoutATimeout(t *testing.T) {
	t.Parallel()

	var c *Client
	assert.False(t, c.IsEnabled())
	assert.Equal(t, time.Duration(0), c.GetTimeout())
}

// Subtests are keyed by method.
func TestClient_DisabledClientCallsNothing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		call    func(context.Context, *Client) (any, error)
		wantErr bool
	}{
		{
			name: "adding messages is a silent no-op",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.AddMessages(ctx, AddMessagesRequest{})
			},
		},
		{
			name: "a temporal window search refuses",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.TemporalWindowSearch(ctx, TemporalSearchRequest{})
			},
			wantErr: true,
		},
		{
			name: "an entity relationships search refuses",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.EntityRelationshipsSearch(ctx, EntityRelationshipSearchRequest{})
			},
			wantErr: true,
		},
		{
			name: "a diverse results search refuses",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.DiverseResultsSearch(ctx, DiverseSearchRequest{})
			},
			wantErr: true,
		},
		{
			name: "an episode context search refuses",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.EpisodeContextSearch(ctx, EpisodeContextSearchRequest{})
			},
			wantErr: true,
		},
		{
			name: "a successful tools search refuses",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SuccessfulToolsSearch(ctx, SuccessfulToolsSearchRequest{})
			},
			wantErr: true,
		},
		{
			name: "a recent context search refuses",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RecentContextSearch(ctx, RecentContextSearchRequest{})
			},
			wantErr: true,
		},
		{
			name: "an entity by label search refuses",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.EntityByLabelSearch(ctx, EntityByLabelSearchRequest{})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, err := NewClient("", 0, false)
			require.NoError(t, err)

			resp, err := tt.call(t.Context(), client)
			assert.Nil(t, resp, "a disabled client must answer nothing")
			if tt.wantErr {
				assert.ErrorContains(t, err, "graphiti is not enabled")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_AddMessages_ReturnsWhenTheCallerIsCancelled(t *testing.T) {
	t.Parallel()

	var enteredOnce sync.Once
	entered := make(chan struct{})
	released := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthcheck" {
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		enteredOnce.Do(func() { close(entered) })
		<-released
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(released) })

	client, err := NewClient(server.URL, time.Minute, true)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	answered := make(chan error, 1)
	go func() { answered <- client.AddMessages(ctx, AddMessagesRequest{}) }()

	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("the request never reached the knowledge graph")
	}
	cancel()

	select {
	case err := <-answered:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("the caller kept waiting for a knowledge-graph request after it was cancelled")
	}
}

func TestClient_AwaitCall_KeepsWhatAFinishedCallReturned(t *testing.T) {
	t.Parallel()

	value, err := awaitCall(context.Background(), func() (int, error) { return 7, nil })
	require.NoError(t, err)
	assert.Equal(t, 7, value)

	unreachable := errors.New("graph is unreachable")
	_, err = awaitCall(context.Background(), func() (int, error) { return 0, unreachable })
	assert.ErrorIs(t, err, unreachable)
}
