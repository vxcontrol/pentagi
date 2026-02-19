package embeddings

import (
	"context"
	"fmt"
	"net/http"

	"pentagi/pkg/config"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/vxcontrol/langchaingo/embeddings"
	"github.com/vxcontrol/langchaingo/embeddings/huggingface"
	"github.com/vxcontrol/langchaingo/embeddings/jina"
	"github.com/vxcontrol/langchaingo/embeddings/voyageai"
	"github.com/vxcontrol/langchaingo/llms/googleai"
	hgclient "github.com/vxcontrol/langchaingo/llms/huggingface"
	"github.com/vxcontrol/langchaingo/llms/mistral"
	"github.com/vxcontrol/langchaingo/llms/ollama"
	"github.com/vxcontrol/langchaingo/llms/openai"
)

type constructor func(cfg *config.Config, httpClient *http.Client) (embeddings.Embedder, error)

type Embedder interface {
	embeddings.Embedder
	IsAvailable() bool
}

type embedder struct {
	embeddings.Embedder
}

func (e *embedder) IsAvailable() bool {
	return e.Embedder != nil
}

func New(cfg *config.Config) (Embedder, error) {
	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	var f constructor

	switch cfg.EmbeddingProvider {
	case "openai":
		f = newOpenAI
	case "ollama":
		f = newOllama
	case "mistral":
		f = newMistral
	case "jina":
		f = newJina
	case "huggingface":
		f = newHuggingface
	case "googleai":
		f = newGoogleAI
	case "voyageai":
		f = newVoyageAI
	case "none":
		return &embedder{nil}, nil
	default:
		return &embedder{nil}, fmt.Errorf("unsupported embedding provider: %s", cfg.EmbeddingProvider)
	}

	e, err := f(cfg, httpClient)
	if err != nil {
		return &embedder{nil}, err
	}

	return &embedder{e}, nil
}

func newOpenAI(cfg *config.Config, httpClient *http.Client) (embeddings.Embedder, error) {
	model, provider := cfg.EmbeddingModel, "openai"
	if model == "" {
		model = "text-embedding-ada-002"
	}

	var opts []openai.Option
	metadata := langfuse.Metadata{
		"strip_new_lines": cfg.EmbeddingStripNewLines,
		"batch_size":      cfg.EmbeddingBatchSize,
	}
	if cfg.EmbeddingURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.EmbeddingURL))
		metadata["url"] = cfg.EmbeddingURL
	} else if cfg.OpenAIServerURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.OpenAIServerURL))
		metadata["url"] = cfg.OpenAIServerURL
	}
	if cfg.EmbeddingKey != "" {
		opts = append(opts, openai.WithToken(cfg.EmbeddingKey))
	} else if cfg.OpenAIKey != "" {
		opts = append(opts, openai.WithToken(cfg.OpenAIKey))
	}
	if cfg.EmbeddingModel != "" {
		opts = append(opts, openai.WithEmbeddingModel(cfg.EmbeddingModel))
	}
	if httpClient != nil {
		opts = append(opts, openai.WithHTTPClient(httpClient))
	}

	client, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	eopts := []embeddings.Option{
		embeddings.WithStripNewLines(cfg.EmbeddingStripNewLines),
		embeddings.WithBatchSize(cfg.EmbeddingBatchSize),
	}

	e, err := embeddings.NewEmbedder(client, eopts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return &wrapper{
		model:    model,
		provider: provider,
		metadata: metadata,
		Embedder: e,
	}, nil
}

func newOllama(cfg *config.Config, httpClient *http.Client) (embeddings.Embedder, error) {
	// EmbeddingKey is not supported for ollama
	model, provider := cfg.EmbeddingModel, "ollama"

	var opts []ollama.Option
	metadata := langfuse.Metadata{
		"strip_new_lines": cfg.EmbeddingStripNewLines,
		"batch_size":      cfg.EmbeddingBatchSize,
	}
	if cfg.EmbeddingURL != "" {
		opts = append(opts, ollama.WithServerURL(cfg.EmbeddingURL))
		metadata["url"] = cfg.EmbeddingURL
	}
	if cfg.EmbeddingModel != "" {
		opts = append(opts, ollama.WithModel(cfg.EmbeddingModel))
	}
	if httpClient != nil {
		opts = append(opts, ollama.WithHTTPClient(httpClient))
	}

	client, err := ollama.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama client: %w", err)
	}

	eopts := []embeddings.Option{
		embeddings.WithStripNewLines(cfg.EmbeddingStripNewLines),
		embeddings.WithBatchSize(cfg.EmbeddingBatchSize),
	}

	e, err := embeddings.NewEmbedder(client, eopts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return &wrapper{
		model:    model,
		provider: provider,
		metadata: metadata,
		Embedder: e,
	}, nil
}

func newMistral(cfg *config.Config, _ *http.Client) (embeddings.Embedder, error) {
	// EmbeddingModel is not supported for mistral
	// Custom HTTP client is not supported for mistral
	model, provider := "mistral-embed", "mistral"

	var opts []mistral.Option
	metadata := langfuse.Metadata{
		"strip_new_lines": cfg.EmbeddingStripNewLines,
		"batch_size":      cfg.EmbeddingBatchSize,
	}
	if cfg.EmbeddingURL != "" {
		opts = append(opts, mistral.WithEndpoint(cfg.EmbeddingURL))
		metadata["url"] = cfg.EmbeddingURL
	}
	if cfg.EmbeddingKey != "" {
		opts = append(opts, mistral.WithAPIKey(cfg.EmbeddingKey))
	}

	client, err := mistral.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create mistral client: %w", err)
	}

	eopts := []embeddings.Option{
		embeddings.WithStripNewLines(cfg.EmbeddingStripNewLines),
		embeddings.WithBatchSize(cfg.EmbeddingBatchSize),
	}

	e, err := embeddings.NewEmbedder(client, eopts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return &wrapper{
		model:    model,
		provider: provider,
		metadata: metadata,
		Embedder: e,
	}, nil
}

func newJina(cfg *config.Config, httpClient *http.Client) (embeddings.Embedder, error) {
	// Custom HTTP client is not supported for jina
	model, provider := cfg.EmbeddingModel, "jina"
	if model == "" {
		model = "jina-embeddings-v2-small-en"
	}

	var opts []jina.Option
	metadata := langfuse.Metadata{
		"strip_new_lines": cfg.EmbeddingStripNewLines,
		"batch_size":      cfg.EmbeddingBatchSize,
	}
	opts = append(opts,
		jina.WithStripNewLines(cfg.EmbeddingStripNewLines),
		jina.WithBatchSize(cfg.EmbeddingBatchSize),
	)
	if cfg.EmbeddingURL != "" {
		opts = append(opts, jina.WithAPIBaseURL(cfg.EmbeddingURL))
		metadata["url"] = cfg.EmbeddingURL
	}
	if cfg.EmbeddingKey != "" {
		opts = append(opts, jina.WithAPIKey(cfg.EmbeddingKey))
	}
	if cfg.EmbeddingModel != "" {
		opts = append(opts, jina.WithModel(cfg.EmbeddingModel))
	}

	e, err := jina.NewJina(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create jina embedder: %w", err)
	}

	return &wrapper{
		model:    model,
		provider: provider,
		metadata: metadata,
		Embedder: e,
	}, nil
}

func newHuggingface(cfg *config.Config, httpClient *http.Client) (embeddings.Embedder, error) {
	// Custom HTTP client is not supported for huggingface
	model, provider := cfg.EmbeddingModel, "huggingface"
	if model == "" {
		model = "BAAI/bge-small-en-v1.5"
	}

	var opts []hgclient.Option
	metadata := langfuse.Metadata{
		"strip_new_lines": cfg.EmbeddingStripNewLines,
		"batch_size":      cfg.EmbeddingBatchSize,
	}
	if cfg.EmbeddingURL != "" {
		opts = append(opts, hgclient.WithURL(cfg.EmbeddingURL))
		metadata["url"] = cfg.EmbeddingURL
	}
	if cfg.EmbeddingKey != "" {
		opts = append(opts, hgclient.WithToken(cfg.EmbeddingKey))
	}
	if cfg.EmbeddingModel != "" {
		opts = append(opts, hgclient.WithModel(cfg.EmbeddingModel))
	}

	client, err := hgclient.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create huggingface client: %w", err)
	} else if client == nil {
		return nil, fmt.Errorf("huggingface client is nil")
	}

	eopts := []huggingface.Option{
		huggingface.WithStripNewLines(cfg.EmbeddingStripNewLines),
		huggingface.WithBatchSize(cfg.EmbeddingBatchSize),
		huggingface.WithClient(*client),
	}
	if cfg.EmbeddingModel != "" {
		eopts = append(eopts, huggingface.WithModel(cfg.EmbeddingModel))
	}

	e, err := huggingface.NewHuggingface(eopts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create huggingface embedder: %w", err)
	}

	return &wrapper{
		model:    model,
		provider: provider,
		metadata: metadata,
		Embedder: e,
	}, nil
}

func newGoogleAI(cfg *config.Config, httpClient *http.Client) (embeddings.Embedder, error) {
	// EmbeddingURL is not supported for googleai
	model, provider := cfg.EmbeddingModel, "googleai"
	if model == "" {
		model = "embedding-001"
	}

	var opts []googleai.Option
	metadata := langfuse.Metadata{
		"strip_new_lines": cfg.EmbeddingStripNewLines,
		"batch_size":      cfg.EmbeddingBatchSize,
	}
	if cfg.EmbeddingKey != "" {
		opts = append(opts, googleai.WithAPIKey(cfg.EmbeddingKey))
	}
	if cfg.EmbeddingModel != "" {
		opts = append(opts, googleai.WithDefaultEmbeddingModel(cfg.EmbeddingModel))
	}
	if httpClient != nil {
		opts = append(opts, googleai.WithHTTPClient(httpClient))
	}

	client, err := googleai.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create googleai client: %w", err)
	}

	eopts := []embeddings.Option{
		embeddings.WithStripNewLines(cfg.EmbeddingStripNewLines),
		embeddings.WithBatchSize(cfg.EmbeddingBatchSize),
	}

	e, err := embeddings.NewEmbedder(client, eopts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return &wrapper{
		model:    model,
		provider: provider,
		metadata: metadata,
		Embedder: e,
	}, nil
}

func newVoyageAI(cfg *config.Config, httpClient *http.Client) (embeddings.Embedder, error) {
	// EmbeddingURL client is not supported for voyageai
	model, provider := cfg.EmbeddingModel, "voyageai"
	if model == "" {
		model = "voyage-4"
	}

	var opts []voyageai.Option
	metadata := langfuse.Metadata{
		"strip_new_lines": cfg.EmbeddingStripNewLines,
		"batch_size":      cfg.EmbeddingBatchSize,
	}
	opts = append(opts,
		voyageai.WithStripNewLines(cfg.EmbeddingStripNewLines),
		voyageai.WithBatchSize(cfg.EmbeddingBatchSize),
	)
	if cfg.EmbeddingKey != "" {
		opts = append(opts, voyageai.WithToken(cfg.EmbeddingKey))
	}
	if cfg.EmbeddingModel != "" {
		opts = append(opts, voyageai.WithModel(cfg.EmbeddingModel))
	}
	if httpClient != nil {
		opts = append(opts, voyageai.WithClient(*httpClient))
	}

	e, err := voyageai.NewVoyageAI(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create voyageai embedder: %w", err)
	}

	return &wrapper{
		model:    model,
		provider: provider,
		metadata: metadata,
		Embedder: e,
	}, nil
}
