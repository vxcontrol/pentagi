package embeddings

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"pentagi/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbedder_New_BuildsAnAvailableEmbedderForEveryProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		provider  string
		available bool
	}{
		{"openai builds an available embedder", "openai", true},
		{"ollama builds an available embedder", "ollama", true},
		{"mistral builds an available embedder", "mistral", true},
		{"jina builds an available embedder", "jina", true},
		{"huggingface builds an available embedder", "huggingface", true},
		{"googleai builds an available embedder", "googleai", true},
		{"voyageai builds an available embedder", "voyageai", true},
		{"none builds an unavailable one", "none", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e, err := New(&config.Config{EmbeddingProvider: tc.provider, EmbeddingKey: "test-key"})
			require.NoError(t, err)
			require.NotNil(t, e)
			assert.Equal(t, tc.available, e.IsAvailable())
		})
	}
}

func TestEmbedder_New_RefusesWhatItCannotBuild(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		cfg             *config.Config
		wantErr         string
		wantUnavailable bool
	}{
		{
			name: "an unknown provider", cfg: &config.Config{EmbeddingProvider: "unknown-provider"},
			wantErr: "unsupported embedding provider", wantUnavailable: true,
		},
		{
			name: "an empty provider", cfg: &config.Config{EmbeddingProvider: ""},
			wantErr: "unsupported embedding provider", wantUnavailable: true,
		},
		{
			name: "a CA certificate that cannot be read",
			cfg: &config.Config{
				EmbeddingProvider: "openai", OpenAIKey: "test-key", ExternalSSLCAPath: "/non/existent/ca.pem",
			},
			wantErr: "failed to read external CA certificate",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e, err := New(tc.cfg)
			require.ErrorContains(t, err, tc.wantErr)
			if tc.wantUnavailable {
				require.NotNil(t, e, "a caller asks an unsupported embedder whether it is available")
				assert.False(t, e.IsAvailable())
			}
		})
	}
}

func TestEmbedder_ResolvedProviderType_CallsAnOpenAIDoorOnAnotherServerCustom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		url      string
		want     string
	}{
		{"openai without a custom url stays openai", "openai", "", "openai"},
		{"openai with a custom url becomes custom", "openai", "https://my-gateway/v1", "custom"},
		{"non-openai provider is untouched even with a url", "ollama", "https://my-gateway/v1", "ollama"},
		{"none stays none", "none", "", "none"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{EmbeddingProvider: tc.provider, EmbeddingURL: tc.url}
			assert.Equal(t, tc.want, ResolvedProviderType(cfg))
		})
	}
}

func modelOnTheWire(r *http.Request) string {
	raw, _ := io.ReadAll(r.Body)

	var body struct {
		Model string `json:"model"`
	}

	if json.Unmarshal(raw, &body) == nil && body.Model != "" {
		return body.Model
	}

	_, after, found := strings.Cut(r.URL.Path, "/models/")
	if !found {
		return ""
	}

	name, _, _ := strings.Cut(after, "/pipeline/")
	name, _, _ = strings.Cut(name, ":")

	return name
}

type rewritingTransport struct {
	target *url.URL
}

func (t rewritingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Scheme = t.target.Scheme
	r.URL.Host = t.target.Host

	return http.DefaultTransport.RoundTrip(r)
}

// embedderRequest is what the probe saw of the first request a door sent.
type embedderRequest struct {
	host, path, authorization, model string
	reportedModel                    string
}

// embedderProbe builds a door on a client that delivers every request to a local probe, and embeds one text.
func embedderProbe(t *testing.T, build constructor, cfg *config.Config) embedderRequest {
	t.Helper()

	var (
		got  embedderRequest
		hits int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits == 1 {
			got.host, got.path, got.authorization = r.Host, r.URL.Path, r.Header.Get("Authorization")
			got.model = modelOnTheWire(r)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"embedding":[0.1,0.2]}],"embeddings":[[0.1,0.2]]}`))
	}))
	t.Cleanup(srv.Close)

	target, err := url.Parse(srv.URL)
	require.NoError(t, err)

	cfg.EmbeddingBatchSize = 512
	e, err := build(cfg, &http.Client{Transport: rewritingTransport{target: target}})
	require.NoError(t, err)

	// The canned answer fails most doors; the request is already recorded, so that error is not one.
	_, _ = e.EmbedDocuments(context.Background(), []string{"probe"})
	require.Positive(t, hits, "the door never reached the probe through the client it was handed")

	w, ok := e.(*wrapper)
	require.True(t, ok, "every door is built through the observing wrapper")
	got.reportedModel = w.model

	return got
}

// TestEmbedder_EveryDoorReportsTheModelItAsksFor is keyed by door.
func TestEmbedder_EveryDoorReportsTheModelItAsksFor(t *testing.T) {
	t.Parallel()

	const atTheEndpoint, ignoringTheEndpoint = "at the configured endpoint", "and ignores the endpoint"

	doors := []struct {
		provider string
		build    constructor
		named    string
		unnamed  string
		endpoint string
		host     string
	}{
		{"openai", newOpenAI, "text-embedding-3-large", "text-embedding-ada-002", atTheEndpoint, "embedder.example"},
		{"ollama", newOllama, "nomic-embed-text", "", atTheEndpoint, "embedder.example"},
		{"mistral", newMistral, "mistral/mistral-embed", "mistral-embed", atTheEndpoint, "embedder.example"},
		{"jina", newJina, "jina-embeddings-v3", "jina-embeddings-v2-small-en", atTheEndpoint, "embedder.example"},
		{"huggingface", newHuggingface, "BAAI/bge-m3", "BAAI/bge-small-en-v1.5", atTheEndpoint, "embedder.example"},
		{
			"googleai", newGoogleAI, "text-embedding-004", "gemini-embedding-001", ignoringTheEndpoint,
			"generativelanguage.googleapis.com",
		},
		{"voyageai", newVoyageAI, "voyage-3", "voyage-4", atTheEndpoint, "embedder.example"},
	}

	for _, door := range doors {
		t.Run(door.provider+" names the model it was given "+door.endpoint, func(t *testing.T) {
			t.Parallel()

			got := embedderProbe(t, door.build, &config.Config{
				EmbeddingKey: "test-key", EmbeddingURL: "http://embedder.example/v1", EmbeddingModel: door.named,
			})

			assert.Equal(t, door.host, got.host, "the door asked another host")
			assert.Equal(t, door.named, got.model, "the configured model must reach the request")
			assert.Equal(t, got.model, got.reportedModel, "the observation names another model")
		})

		t.Run(door.provider+" names its default through the caller's client without an endpoint", func(t *testing.T) {
			t.Parallel()

			got := embedderProbe(t, door.build, &config.Config{EmbeddingKey: "test-key"})

			assert.Equal(t, door.unnamed, got.model, "an unnamed model embeds with the vendor default")
			assert.Equal(t, got.model, got.reportedModel, "the observation names another model")
		})
	}
}

// TestEmbedder_OpenAIShapedDoorsAppendTheEmbeddingsPathToTheirBase is keyed by door.
func TestEmbedder_OpenAIShapedDoorsAppendTheEmbeddingsPathToTheirBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		build    constructor
		endpoint string
		want     string
	}{
		{"openai keeps a gateway prefix", newOpenAI, "https://gw.example.com/gateway/openai/v1", "/gateway/openai/v1/embeddings"},
		{"mistral keeps a gateway prefix", newMistral, "https://gw.example.com/gateway/openai/v1", "/gateway/openai/v1/embeddings"},
		{"mistral without an endpoint asks the vendor's versioned path", newMistral, "", "/v1/embeddings"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := embedderProbe(t, tc.build, &config.Config{
				EmbeddingKey: "test-key", EmbeddingURL: tc.endpoint, EmbeddingModel: "probe-model",
			})

			assert.Equal(t, tc.want, got.path)
		})
	}
}

func TestEmbedder_EmbeddingServer_TakesTheServerAndKeyAsOnePair(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("OPENAI_API_BASE", "")
	t.Setenv("OLLAMA_HOST", "")
	t.Setenv("OLLAMA_API_KEY", "")

	openaiChat := func(cfg config.Config) *config.Config {
		cfg.OpenAIServerURL, cfg.OpenAIKey = "https://gateway.example/v1", "gateway-key"
		return &cfg
	}
	mistralChat := func(cfg config.Config) *config.Config {
		cfg.MistralServerURL, cfg.MistralAPIKey = "https://gateway.example/v1", "gateway-key"
		return &cfg
	}
	ollamaChat := func(cfg config.Config) *config.Config {
		cfg.OllamaServerURL, cfg.OllamaServerAPIKey, cfg.EmbeddingModel = "http://keyed-ollama:11434", "chat-key", "nomic-embed-text"
		return &cfg
	}

	tests := []struct {
		name          string
		build         constructor
		cfg           *config.Config
		host          string
		authorization string
	}{
		{
			name: "openai with nothing set for embeddings takes the chat door's pair", build: newOpenAI,
			cfg: openaiChat(config.Config{}), host: "gateway.example", authorization: "Bearer gateway-key",
		},
		{
			name: "openai with a key of its own goes to the vendor", build: newOpenAI,
			cfg: openaiChat(config.Config{EmbeddingKey: "direct-key"}), host: "api.openai.com", authorization: "Bearer direct-key",
		},
		{
			name: "openai on another endpoint leaves the chat door's key behind", build: newOpenAI,
			cfg: openaiChat(config.Config{EmbeddingURL: "https://embedder.example/v1"}), host: "embedder.example",
		},
		{
			name: "openai on the chat door's own server takes its key", build: newOpenAI,
			cfg: openaiChat(config.Config{EmbeddingURL: "https://gateway.example/v1"}), host: "gateway.example",
			authorization: "Bearer gateway-key",
		},
		{
			name: "mistral with nothing set for embeddings takes the chat door's pair", build: newMistral,
			cfg: mistralChat(config.Config{}), host: "gateway.example", authorization: "Bearer gateway-key",
		},
		{
			name: "mistral with a key of its own goes to the vendor", build: newMistral,
			cfg: mistralChat(config.Config{EmbeddingKey: "direct-key"}), host: "api.mistral.ai", authorization: "Bearer direct-key",
		},
		{
			name: "mistral on another endpoint leaves the chat door's key behind", build: newMistral,
			cfg: mistralChat(config.Config{EmbeddingURL: "https://embedder.example/v1"}), host: "embedder.example",
		},
		{
			name: "mistral on the chat door's own server, trailing slash and all, takes its key", build: newMistral,
			cfg: mistralChat(config.Config{EmbeddingURL: "https://gateway.example/v1/"}), host: "gateway.example",
			authorization: "Bearer gateway-key",
		},
		{
			name: "mistral with only a chat key goes to the vendor with it", build: newMistral,
			cfg: &config.Config{MistralAPIKey: "mistral-chat-key"}, host: "api.mistral.ai", authorization: "Bearer mistral-chat-key",
		},
		{
			name: "ollama with no endpoint set for embeddings takes the chat door's pair", build: newOllama,
			cfg: ollamaChat(config.Config{}), host: "keyed-ollama:11434", authorization: "Bearer chat-key",
		},
		{
			name: "ollama on another endpoint sends no key, not even one of its own", build: newOllama,
			cfg:  ollamaChat(config.Config{EmbeddingURL: "http://embedder:11434", EmbeddingKey: "left-from-another-provider"}),
			host: "embedder:11434",
		},
		{
			name: "ollama on the chat door's own server takes its key", build: newOllama,
			cfg: ollamaChat(config.Config{EmbeddingURL: "http://keyed-ollama:11434"}), host: "keyed-ollama:11434",
			authorization: "Bearer chat-key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := embedderProbe(t, tc.build, tc.cfg)

			assert.Equal(t, tc.host, got.host)
			assert.Equal(t, tc.authorization, got.authorization)
		})
	}
}

func TestEmbedder_NewMistral_NamesTheModelTheWayTheChatDoorsServerServesIt(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	gateway := func(cfg config.Config) *config.Config {
		cfg.MistralServerURL, cfg.MistralAPIKey, cfg.MistralProvider = "https://gateway.example/v1", "gateway-key", "mistral"
		return &cfg
	}

	tests := []struct {
		name  string
		cfg   *config.Config
		host  string
		model string
	}{
		{"the chat door's server", gateway(config.Config{}), "gateway.example", "mistral/mistral-embed"},
		{
			"the chat door's server named again",
			gateway(config.Config{EmbeddingURL: "https://gateway.example/v1/", EmbeddingModel: "codestral-embed"}),
			"gateway.example", "mistral/codestral-embed",
		},
		{
			"a model already named the chat door's way",
			gateway(config.Config{EmbeddingModel: "mistral/mistral-embed"}),
			"gateway.example", "mistral/mistral-embed",
		},
		{
			"the chat door's server without a prefix",
			&config.Config{MistralServerURL: "https://gateway.example/v1", MistralAPIKey: "gateway-key"},
			"gateway.example", "mistral-embed",
		},
		{"the vendor, reached with a key of its own", gateway(config.Config{EmbeddingKey: "direct-key"}), "api.mistral.ai", "mistral-embed"},
		{
			"another path on the chat door's host",
			gateway(config.Config{EmbeddingURL: "https://gateway.example/mistral/v1", EmbeddingKey: "gateway-key"}),
			"gateway.example", "mistral-embed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := embedderProbe(t, newMistral, tc.cfg)

			assert.Equal(t, tc.host, got.host)
			assert.Equal(t, tc.model, got.model)
			assert.Equal(t, got.model, got.reportedModel, "the trace names another model than the request")
		})
	}
}
