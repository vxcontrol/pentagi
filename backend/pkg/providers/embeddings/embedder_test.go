package embeddings

import (
	"testing"

	"pentagi/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_AllProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				EmbeddingProvider: tt.provider,
				EmbeddingKey:      "test-key",
			}

			e, err := New(cfg)
			require.NoError(t, err)
			require.NotNil(t, e)
			assert.Equal(t, tt.available, e.IsAvailable())
		})
	}
}

func TestNew_UnsupportedProvider(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "unknown-provider",
	}

	e, err := New(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported embedding provider")
	require.NotNil(t, e)
	assert.False(t, e.IsAvailable())
}

func TestNew_OpenAI_DefaultModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "openai",
		OpenAIKey:         "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_OpenAI_CustomURL(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "openai",
		EmbeddingURL:      "https://custom-openai.example.com",
		EmbeddingKey:      "custom-key",
		EmbeddingModel:    "text-embedding-3-small",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_OpenAI_FallbackToOpenAIServerURL(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "openai",
		OpenAIKey:         "test-key",
		OpenAIServerURL:   "https://api.openai.com/v1",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_OpenAI_KeyPriority(t *testing.T) {
	t.Parallel()

	t.Run("EmbeddingKey takes priority", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			EmbeddingProvider: "openai",
			EmbeddingKey:      "embedding-specific-key",
			OpenAIKey:         "generic-key",
		}

		e, err := New(cfg)
		require.NoError(t, err)
		require.NotNil(t, e)
		assert.True(t, e.IsAvailable())
	})

	t.Run("Falls back to OpenAIKey", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			EmbeddingProvider: "openai",
			OpenAIKey:         "generic-key",
		}

		e, err := New(cfg)
		require.NoError(t, err)
		require.NotNil(t, e)
		assert.True(t, e.IsAvailable())
	})
}

func TestNew_Jina_DefaultModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "jina",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Huggingface_DefaultModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "huggingface",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_GoogleAI_DefaultModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "googleai",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_VoyageAI_DefaultModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "voyageai",
		EmbeddingKey:      "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_WithBatchSizeAndStripNewLines(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider:      "openai",
		OpenAIKey:              "test-key",
		EmbeddingBatchSize:     100,
		EmbeddingStripNewLines: true,
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_HTTPClientError(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider:  "openai",
		OpenAIKey:          "test-key",
		ExternalSSLCAPath:  "/non/existent/ca.pem",
		EmbeddingBatchSize: 512,
	}

	_, err := New(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read external CA certificate")
}

func TestIsAvailable_NilEmbedder(t *testing.T) {
	t.Parallel()

	e := &embedder{nil}
	assert.False(t, e.IsAvailable())
}

func TestIsAvailable_ValidEmbedder(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "openai",
		OpenAIKey:         "test-key",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Ollama_WithCustomModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "ollama",
		EmbeddingURL:      "http://localhost:11434",
		EmbeddingModel:    "nomic-embed-text",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Mistral_WithCustomURL(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "mistral",
		EmbeddingKey:      "test-key",
		EmbeddingURL:      "https://api.mistral.ai",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Jina_WithCustomModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "jina",
		EmbeddingKey:      "test-key",
		EmbeddingModel:    "jina-embeddings-v2-base-en",
		EmbeddingURL:      "https://api.jina.ai/v1",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_Huggingface_WithCustomModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "huggingface",
		EmbeddingKey:      "test-key",
		EmbeddingModel:    "sentence-transformers/all-MiniLM-L6-v2",
		EmbeddingURL:      "https://api-inference.huggingface.co",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_GoogleAI_WithCustomModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "googleai",
		EmbeddingKey:      "test-key",
		EmbeddingModel:    "text-embedding-004",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_VoyageAI_WithCustomModel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "voyageai",
		EmbeddingKey:      "test-key",
		EmbeddingModel:    "voyage-code-3",
	}

	e, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.True(t, e.IsAvailable())
}

func TestNew_DifferentBatchSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		batchSize int
	}{
		{"default", 0},
		{"small batch", 10},
		{"medium batch", 100},
		{"large batch", 512},
		{"very large batch", 2048},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				EmbeddingProvider:  "openai",
				OpenAIKey:          "test-key",
				EmbeddingBatchSize: tt.batchSize,
			}

			e, err := New(cfg)
			require.NoError(t, err)
			require.NotNil(t, e)
			assert.True(t, e.IsAvailable())
		})
	}
}

func TestNew_StripNewLinesVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stripNewLines bool
	}{
		{"strip enabled", true},
		{"strip disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				EmbeddingProvider:      "openai",
				OpenAIKey:              "test-key",
				EmbeddingStripNewLines: tt.stripNewLines,
			}

			e, err := New(cfg)
			require.NoError(t, err)
			require.NotNil(t, e)
			assert.True(t, e.IsAvailable())
		})
	}
}

func TestNew_EmptyProvider(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EmbeddingProvider: "",
	}

	e, err := New(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported embedding provider")
	require.NotNil(t, e)
	assert.False(t, e.IsAvailable())
}
