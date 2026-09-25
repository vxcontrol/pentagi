package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
)

const openaiAPIChangelog = "https://developers.openai.com/api/docs/changelog"

func TestOpenai_New_BuildsTheDoorWithAPricedModelForEveryAgent(t *testing.T) {
	cfg := &config.Config{
		OpenAIKey:       "test-key",
		OpenAIServerURL: "https://api.openai.com/v1",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, provider.DefaultProviderNameOpenAI, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if prov.Type() != provider.ProviderOpenAI {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderOpenAI, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameOpenAI {
		t.Errorf("Expected provider name %v, got %v", provider.DefaultProviderNameOpenAI, prov.Name())
	}
	if len(prov.GetRawConfig()) == 0 {
		t.Fatal("Raw config should not be empty")
	}
	if prov.GetProviderConfig() != providerConfig {
		t.Fatal("Provider config should be the one New was given")
	}

	for _, agentType := range pconfig.AllAgentTypes {
		if prov.Model(agentType) == "" {
			t.Errorf("Agent type %v should have a model assigned", agentType)
		}

		priceInfo := prov.GetPriceInfo(agentType)
		if priceInfo == nil {
			t.Errorf("Agent type %v should have price information", agentType)
		} else if priceInfo.Input <= 0 || priceInfo.Output <= 0 {
			t.Errorf("Agent type %v should have positive input (%f) and output (%f) prices",
				agentType, priceInfo.Input, priceInfo.Output)
		}
	}
}

func TestOpenai_DefaultModels_ShipsANonEmptyPricedCatalogue(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("Models list should not be empty")
	}

	for _, model := range models {
		if model.Name == "" {
			t.Error("Model name should not be empty")
		}

		if model.Price == nil {
			t.Errorf("Model %s should have price information", model.Name)
			continue
		}

		if model.Price.Input <= 0 {
			t.Errorf("Model %s should have positive input price", model.Name)
		}

		if model.Price.Output <= 0 {
			t.Errorf("Model %s should have positive output price", model.Name)
		}
	}
}

func TestOpenai_DefaultModels_DatesEachModelTheDayTheAPIBeganServingIt(t *testing.T) {
	released := map[string]string{
		"gpt-6-sol":     "2026-09-22",
		"gpt-6-luna":    "2026-09-22",
		"gpt-6-astra":   "2026-09-03",
		"daybreak-red":  "2026-08-07",
		"daybreak-blue": "2026-08-07",
		"gpt-5.6-cyber": "2026-08-07",
		"gpt-5.6-sol":   "2026-07-09",
		"gpt-5.6-terra": "2026-07-09",
		"gpt-5.6-luna":  "2026-07-09",
		"gpt-5.5":       "2026-04-24",
		"gpt-5.4":       "2026-03-05",
		"gpt-5.4-mini":  "2026-03-17",
		"gpt-5.4-nano":  "2026-03-17",
		"gpt-5.2":       "2025-12-11",
		"gpt-5.1":       "2025-11-13",
		"gpt-5":         "2025-08-07",
		"gpt-5-mini":    "2025-08-07",
		"gpt-5-nano":    "2025-08-07",
		"gpt-4.1":       "2025-04-14",
		"gpt-4.1-mini":  "2025-04-14",
		"gpt-4.1-nano":  "2025-04-14",
		"o3":            "2025-04-16",
		"o4-mini":       "2025-04-16",
		"o3-mini":       "2025-01-31",
		"o1":            "2024-12-17",
		"gpt-4o-mini":   "2024-07-18",
		"gpt-4o":        "2024-05-13",
	}

	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("cannot read the bundled openai catalogue: %v", err)
	}

	seen := map[string]bool{}
	for _, m := range models {
		want, ok := released[m.Name]
		if !ok {
			t.Errorf("%s has no entry here; add the day %s says the API released it", m.Name, openaiAPIChangelog)
			continue
		}
		seen[m.Name] = true

		if m.ReleaseDate == nil {
			t.Errorf("%s carries no release_date; the API released it on %s (%s)", m.Name, want, openaiAPIChangelog)
			continue
		}
		if got := m.ReleaseDate.Format("2006-01-02"); got != want {
			t.Errorf("%s release_date = %s, but the API released it on %s (%s)", m.Name, got, want, openaiAPIChangelog)
		}
	}

	for name := range released {
		if !seen[name] {
			t.Errorf("%s is no longer in the openai catalogue; drop it from this table", name)
		}
	}
}

func TestOpenai_CallEx_SendsTheEffortWithoutTopP(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"x","object":"chat.completion","choices":[{"index":0,`+
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
	}))
	t.Cleanup(srv.Close)

	providerConfig, err := BuildProviderConfig([]byte("simple:\n  model: gpt-5.5\n  top_p: 0.3\n  reasoning:\n    effort: high\n"))
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	prov, err := New(&config.Config{OpenAIKey: "k", OpenAIServerURL: srv.URL}, provider.DefaultProviderNameOpenAI, providerConfig)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	if _, err := prov.CallEx(context.Background(), pconfig.OptionsTypeSimple,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil); err != nil {
		t.Fatalf("call: %v", err)
	}

	effort, ok := body["reasoning_effort"]
	if !ok {
		nested, _ := body["reasoning"].(map[string]any)
		effort = nested["effort"]
	}
	if effort != "high" {
		t.Errorf("effort on the wire = %v, want high; body=%v", effort, body)
	}
	if got, present := body["top_p"]; present {
		t.Errorf("top_p = %v: a request carrying an effort must not also carry top_p", got)
	}
}
