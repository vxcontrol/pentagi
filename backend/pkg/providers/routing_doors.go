package providers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/googleai"
)

const doorListingTimeout = 10 * time.Second

func servedModels(door string, e doorEndpoint, client *http.Client) (pconfig.ModelsConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), doorListingTimeout)
	defer cancel()

	switch door {
	case "anthropic":
		return anthropicServedModels(ctx, e, client)
	case "gemini":
		return geminiServedModels(ctx, e, client)
	default:
		return provider.LoadModelsFromHTTP(e.url, e.key, client, e.prefix)
	}
}

func anthropicServedModels(ctx context.Context, e doorEndpoint, client *http.Client) (pconfig.ModelsConfig, error) {
	llm, err := anthropic.New(
		anthropic.WithToken(e.key),
		anthropic.WithBaseURL(e.url),
		anthropic.WithHTTPClient(client),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build anthropic door: %w", err)
	}

	var served pconfig.ModelsConfig
	req := anthropic.ListModelsRequest{Limit: 100}
	for {
		page, err := llm.ListModels(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch models: %w", err)
		}
		for _, model := range page.Models {
			served = append(served, pconfig.ModelConfig{Name: model.ID})
		}
		if !page.HasMore || page.LastID == "" {
			return served, nil
		}
		req.AfterID = page.LastID
	}
}

func geminiServedModels(ctx context.Context, e doorEndpoint, client *http.Client) (pconfig.ModelsConfig, error) {
	transport := http.DefaultTransport
	if client != nil && client.Transport != nil {
		transport = client.Transport
	}

	llm, err := googleai.New(ctx,
		googleai.WithRest(),
		googleai.WithAPIKey(e.key),
		googleai.WithHTTPClient(&http.Client{
			Transport: &httputil.ApiKeyTransport{Transport: transport, APIKey: e.key, BaseURL: e.url},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build gemini door: %w", err)
	}

	ids, err := llm.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}

	served := make(pconfig.ModelsConfig, 0, len(ids))
	for _, id := range ids {
		served = append(served, pconfig.ModelConfig{Name: id})
	}
	return served, nil
}
