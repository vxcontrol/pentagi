package embeddings

import (
	"testing"

	"pentagi/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_OpenAI(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "openai",
		OpenAIKey:         "test-key",
		OpenAIServerURL:   "https://api.openai.com/v1",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_OpenAI_WithCustomURL(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "openai",
		EmbeddingURL:      "https://custom-openai.example.com",
		EmbeddingKey:      "custom-key",
		EmbeddingModel:    "text-embedding-3-small",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Ollama(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "ollama",
		EmbeddingURL:      "http://localhost:11434",
		EmbeddingModel:    "nomic-embed-text",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Mistral(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "mistral",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Jina(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "jina",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Huggingface(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "huggingface",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_GoogleAI(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "googleai",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_VoyageAI(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "voyageai",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_None(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "none",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.False(t, e.IsAvailable())
}

func TestNew_UnsupportedProvider(t *testing.T) {
	cfg := &config.Config{
		EmbeddingProvider: "unknown-provider",
	}

	e, err := New(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported embedding provider")
	assert.NotNil(t, e)
	assert.False(t, e.IsAvailable())
}

func TestNew_AllProviders_TableDriven(t *testing.T) {
	providers := []struct {
		name      string
		provider  string
		available bool
	}{
		{"openai", "openai", true},
		{"ollama", "ollama", true},
		{"mistral", "mistral", true},
		{"jina", "jina", true},
		{"huggingface", "huggingface", true},
		{"googleai", "googleai", true},
		{"voyageai", "voyageai", true},
		{"none", "none", false},
	}

	for _, tt := range providers {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				EmbeddingProvider: tt.provider,
				EmbeddingKey:      "test-key",
			}

			e, err := New(cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.available, e.IsAvailable())
		})
	}
}

func TestIsAvailable_NilEmbedder(t *testing.T) {
	e := &embedder{nil}
	assert.False(t, e.IsAvailable())
}
