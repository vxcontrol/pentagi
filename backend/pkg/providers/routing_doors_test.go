package providers

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"pentagi/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedRequest struct {
	path    string
	version string
}

func recordingDoor(t *testing.T, body string) (*httptest.Server, func() recordedRequest) {
	t.Helper()

	var (
		mu   sync.Mutex
		last recordedRequest
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		last = recordedRequest{path: r.URL.Path, version: r.Header.Get("anthropic-version")}
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	return srv, func() recordedRequest {
		mu.Lock()
		defer mu.Unlock()
		return last
	}
}

// The two vendors whose listing is not the OpenAI shape each get their own request.
func TestRoutingDoors_ServedModels_AsksEachDoorInItsOwnShape(t *testing.T) {
	for _, tc := range []struct {
		name  string
		door  string
		body  string
		cfg   func(url string) *config.Config
		check func(t *testing.T, asked recordedRequest, door RoutingCheck)
	}{
		{
			name: "anthropic is asked with its version header and lists dated snapshots",
			door: "anthropic",
			body: `{"data":[{"type":"model","id":"claude-haiku-4-5-20251001"}],"has_more":false}`,
			cfg: func(url string) *config.Config {
				return &config.Config{AnthropicServerURL: url, AnthropicAPIKey: "k"}
			},
			check: func(t *testing.T, asked recordedRequest, door RoutingCheck) {
				assert.Equal(t, "2023-06-01", asked.version, "anthropic-version header, want the version the door requires")
				assert.NotContains(t, door.Missing, "claude-haiku-4-5",
					"the door serves the dated snapshot, so the alias the catalogue ships is served too")
			},
		},
		{
			name: "gemini is asked at its versioned path",
			door: "gemini",
			body: `{"models":[{"name":"models/gemini-3.8-flash"}]}`,
			cfg: func(url string) *config.Config {
				return &config.Config{GeminiServerURL: url, GeminiAPIKey: "k"}
			},
			check: func(t *testing.T, asked recordedRequest, _ RoutingCheck) {
				assert.Contains(t, asked.path, "/v1beta/models", "gemini listing path, want the versioned models path")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, recorded := recordingDoor(t, tc.body)

			door := checkFor(t, tc.door, CheckRouting(tc.cfg(srv.URL), srv.Client()))

			require.True(t, door.Checked, "the %s door answered but was not checked: %+v", tc.door, door)
			tc.check(t, recorded(), door)
		})
	}
}
