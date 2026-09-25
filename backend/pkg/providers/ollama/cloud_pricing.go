package ollama

import (
	"net/url"
	"strings"

	"pentagi/pkg/providers/pconfig"
)

const cloudHost = "ollama.com"

// Ollama Cloud's published per-million-token rates as of 2026-09-24, keyed by the
// name its pricing page uses. The DeepSeek rows are the peak rate, charged between
// 12:00 and 18:00 UTC Monday to Friday and double the rate the rest of the week.
var cloudPrices = map[string]pconfig.PriceInfo{
	"deepseek-v4.1-flash": {Input: 0.30, Output: 1.20, CacheRead: 0.006},
	"deepseek-v4-flash":   {Input: 0.44, Output: 1.32, CacheRead: 0.014},
	"deepseek-v4-pro":     {Input: 1.32, Output: 3.96, CacheRead: 0.044},
	"gemma4":              {Input: 0.14, Output: 0.40, CacheRead: 0.05},
	"glm-5.3":             {Input: 1.40, Output: 4.40, CacheRead: 0.26},
	"glm-5.3-flash":       {Input: 0.15, Output: 0.50, CacheRead: 0.03},
	"glm-5.2":             {Input: 1.40, Output: 4.40, CacheRead: 0.26},
	"glm-5.1":             {Input: 1.00, Output: 3.20, CacheRead: 0.20},
	"gpt-oss:120b":        {Input: 0.15, Output: 0.60, CacheRead: 0.014},
	"gpt-oss:20b":         {Input: 0.07, Output: 0.30, CacheRead: 0.035},
	"kimi-k3":             {Input: 3.00, Output: 15.00, CacheRead: 0.30},
	"kimi-k2.7-code":      {Input: 0.95, Output: 4.00, CacheRead: 0.19},
	"kimi-k2.6":           {Input: 0.95, Output: 4.00, CacheRead: 0.16},
	"minimax-m3":          {Input: 0.60, Output: 2.40, CacheRead: 0.12},
	"minimax-m2.7":        {Input: 0.30, Output: 1.20, CacheRead: 0.06},
	"mistral-large-3":     {Input: 0.50, Output: 1.50},
	"nemotron-3-nano":     {Input: 0.06, Output: 0.24},
	"nemotron-3-super":    {Input: 0.015, Output: 0.60, CacheRead: 0.015},
	"nemotron-3-ultra":    {Input: 0.10, Output: 3.00, CacheRead: 0.10},
	"qwen3.5:397b":        {Input: 0.60, Output: 3.60},
}

func isCloudServer(serverURL string) bool {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return false
	}
	host := parsed.Hostname()

	return host == cloudHost || strings.HasSuffix(host, "."+cloudHost)
}

func cloudPriceFor(model string) *pconfig.PriceInfo {
	if price, ok := cloudPrices[model]; ok {
		return &price
	}

	base, _, tagged := strings.Cut(model, ":")
	if !tagged {
		return nil
	}

	var found *pconfig.PriceInfo
	for name, price := range cloudPrices {
		if candidate, _, _ := strings.Cut(name, ":"); candidate == base {
			if found != nil {
				return nil
			}
			found = &price
		}
	}

	return found
}
