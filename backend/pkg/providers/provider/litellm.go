package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/providers/pconfig"
)

// ApplyModelPrefix adds provider prefix to model name if prefix is not empty.
// Returns "prefix/modelName" when prefix is set, otherwise returns modelName unchanged.
func ApplyModelPrefix(modelName, prefix string) string {
	if prefix == "" {
		return modelName
	}
	return prefix + "/" + modelName
}

// RemoveModelPrefix strips provider prefix from model name if present.
// Returns modelName without "prefix/" when it has that prefix, otherwise returns unchanged.
func RemoveModelPrefix(modelName, prefix string) string {
	if prefix == "" {
		return modelName
	}
	return strings.TrimPrefix(modelName, prefix+"/")
}

// modelsResponse represents the response from /models API endpoint
type modelsResponse struct {
	Data []modelInfo `json:"data"`
}

// modelInfo represents a single model from the API
type modelInfo struct {
	ID                  string       `json:"id"`
	Created             *int64       `json:"created,omitempty"`
	Description         string       `json:"description,omitempty"`
	SupportedParameters []string     `json:"supported_parameters,omitempty"`
	Pricing             *pricingInfo `json:"pricing,omitempty"`
}

// fallbackModelInfo represents simplified model structure for fallback parsing
type fallbackModelInfo struct {
	ID string `json:"id"`
}

// fallbackModelsResponse represents simplified API response structure
type fallbackModelsResponse struct {
	Data []fallbackModelInfo `json:"data"`
}

// pricingInfo represents pricing information from the API
type pricingInfo struct {
	Prompt     string `json:"prompt,omitempty"`
	Completion string `json:"completion,omitempty"`
}

// LoadModelsFromHTTP loads models from HTTP /models endpoint with optional prefix filtering.
// When prefix is set, it:
// - Filters models to include only those with "prefix/" in their ID
// - Strips the prefix from model names in the returned config
// This enables transparent LiteLLM proxy integration where models are namespaced.
func LoadModelsFromHTTP(baseURL, apiKey string, httpClient *http.Client, prefix string) (pconfig.ModelsConfig, error) {
	modelsURL := strings.TrimRight(baseURL, "/") + "/models"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse with full structure first
	var response modelsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// Fallback to simplified structure if main parsing fails
		var fallbackResponse fallbackModelsResponse
		if err := json.Unmarshal(body, &fallbackResponse); err != nil {
			return nil, fmt.Errorf("failed to parse models response: %w", err)
		}

		return parseFallbackModels(fallbackResponse.Data, prefix), nil
	}

	return parseFullModels(response.Data, prefix), nil
}

// parseFallbackModels parses simplified model structure with prefix filtering
func parseFallbackModels(models []fallbackModelInfo, prefix string) pconfig.ModelsConfig {
	var result pconfig.ModelsConfig

	for _, model := range models {
		// Filter by prefix if set
		if prefix != "" && !strings.HasPrefix(model.ID, prefix+"/") {
			continue
		}

		// Strip prefix from name
		modelName := model.ID
		if prefix != "" {
			modelName = strings.TrimPrefix(model.ID, prefix+"/")
		}

		result = append(result, pconfig.ModelConfig{
			Name: modelName,
		})
	}

	return result
}

// parseFullModels parses full model structure with all metadata and prefix filtering
func parseFullModels(models []modelInfo, prefix string) pconfig.ModelsConfig {
	var result pconfig.ModelsConfig

	for _, model := range models {
		// Filter by prefix if set
		if prefix != "" && !strings.HasPrefix(model.ID, prefix+"/") {
			continue
		}

		// Strip prefix from name
		modelName := model.ID
		if prefix != "" {
			modelName = strings.TrimPrefix(model.ID, prefix+"/")
		}

		modelConfig := pconfig.ModelConfig{
			Name: modelName,
		}

		// Parse description if available
		if model.Description != "" {
			modelConfig.Description = &model.Description
		}

		// Parse created timestamp to release_date if available
		if model.Created != nil && *model.Created > 0 {
			releaseDate := time.Unix(*model.Created, 0).UTC()
			modelConfig.ReleaseDate = &releaseDate
		}

		// Check for reasoning support in supported_parameters
		if len(model.SupportedParameters) > 0 {
			thinking := slices.Contains(model.SupportedParameters, "reasoning")
			modelConfig.Thinking = &thinking
		}

		// Check for tool support - skip models without tool/structured output support
		if len(model.SupportedParameters) > 0 {
			hasTools := slices.Contains(model.SupportedParameters, "tools")
			hasStructuredOutputs := slices.Contains(model.SupportedParameters, "structured_outputs")
			if !hasTools && !hasStructuredOutputs {
				continue
			}
		}

		// Parse pricing if available
		if model.Pricing != nil {
			if input, err := strconv.ParseFloat(model.Pricing.Prompt, 64); err == nil {
				if output, err := strconv.ParseFloat(model.Pricing.Completion, 64); err == nil {
					// Convert per-token prices to per-million-token if needed
					if input < 0.001 && output < 0.001 {
						input = input * 1000000
						output = output * 1000000
					}

					modelConfig.Price = &pconfig.PriceInfo{
						Input:  input,
						Output: output,
					}
				}
			}
		}

		result = append(result, modelConfig)
	}

	return result
}
