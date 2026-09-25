package ollama

import (
	"testing"

	"pentagi/pkg/providers/pconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// What GET https://ollama.com/api/tags returned on 2026-09-24.
var cloudPricingCatalog = []string{
	"gpt-oss:20b", "minimax-m2.7", "kimi-k2.6", "gpt-oss:120b", "nemotron-3-ultra",
	"mistral-large-3:675b", "glm-5.3", "nemotron-3-nano:30b", "qwen3.5:397b",
	"glm-5.3-flash", "nemotron-3-super", "glm-5.1", "kimi-k2.7-code",
	"deepseek-v4-pro:0813", "deepseek-v4-flash:0731", "kimi-k3", "minimax-m3",
	"gemma4:31b", "glm-5.2", "deepseek-v4.1-flash",
}

func TestCloudPricing_CloudPriceFor_PricesEveryCloudModel(t *testing.T) {
	for _, model := range cloudPricingCatalog {
		price := cloudPriceFor(model)
		require.NotNilf(t, price, "no price for %q, so a call to it bills zero", model)
		assert.Greaterf(t, price.Input, 0.0, "input price for %q", model)
		assert.Greaterf(t, price.Output, 0.0, "output price for %q", model)
	}
}

func TestCloudPricing_CloudPriceFor_FallsBackOnlyToAnUnambiguousUntaggedPrice(t *testing.T) {
	for _, tc := range []struct {
		name  string
		model string
		want  *pconfig.PriceInfo
	}{
		{name: "the newly listed DeepSeek flash model has its exact price", model: "deepseek-v4.1-flash", want: &pconfig.PriceInfo{Input: 0.30, Output: 1.20, CacheRead: 0.006}},
		{name: "a tagged name takes its untagged price", model: "deepseek-v4-flash:0731", want: &pconfig.PriceInfo{Input: 0.44, Output: 1.32, CacheRead: 0.014}},
		{name: "a tag on a base two prices share is not priced", model: "gpt-oss:7b"},
		{name: "a tag on an unpriced base is not priced", model: "llama3.1:8b"},
		{name: "an unknown untagged name is not priced", model: "something-nobody-sells"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, cloudPriceFor(tc.model))
		})
	}
}

func TestCloudPricing_IsCloudServer_AcceptsOnlyTheVendorsHosts(t *testing.T) {
	for _, server := range []string{"https://ollama.com", "https://ollama.com/", "https://api.ollama.com"} {
		assert.Truef(t, isCloudServer(server), "%s is the paid cloud", server)
	}
	for _, server := range []string{
		"http://localhost:11434", "http://127.0.0.1:11434",
		"https://ollama.internal.example.com", "https://notollama.com", "",
	} {
		assert.Falsef(t, isCloudServer(server), "%s is not the paid cloud", server)
	}
}
