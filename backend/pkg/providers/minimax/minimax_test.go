package minimax

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func minimaxNew(t *testing.T, cfg *config.Config) provider.Provider {
	t.Helper()

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, provider.DefaultProviderNameMiniMax, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	return prov
}

func TestMinimax_New_BuildsTheDoorWithAPricedModelForEveryAgent(t *testing.T) {
	prov := minimaxNew(t, &config.Config{
		MiniMaxAPIKey:    "test-key",
		MiniMaxServerURL: "https://api.minimax.io/v1",
	})

	if prov.Type() != provider.ProviderMiniMax {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderMiniMax, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameMiniMax {
		t.Errorf("Expected provider name %v, got %v", provider.DefaultProviderNameMiniMax, prov.Name())
	}
	if len(prov.GetRawConfig()) == 0 {
		t.Fatal("Raw config should not be empty")
	}
	if prov.GetProviderConfig() == nil {
		t.Fatal("Provider config should not be nil")
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

func TestMinimax_New_PrefixesModelsOnlyWhenAProviderIsConfigured(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prefix string
		lead   string
	}{
		{name: "a configured provider leads every model", prefix: "minimax", lead: "minimax/"},
		{name: "no provider leaves every model bare"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prov := minimaxNew(t, &config.Config{
				MiniMaxAPIKey:    "test-key",
				MiniMaxServerURL: "https://api.minimax.io/v1",
				MiniMaxProvider:  tc.prefix,
			})

			for _, agentType := range pconfig.AllAgentTypes {
				if got, want := prov.ModelWithPrefix(agentType), tc.lead+prov.Model(agentType); got != want {
					t.Errorf("Agent type %v: expected prefixed model %q, got %q", agentType, want, got)
				}
			}
		})
	}
}

func TestMinimax_New_RefusesAMissingAPIKey(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	_, err = New(&config.Config{MiniMaxServerURL: "https://api.minimax.io/v1"},
		provider.DefaultProviderNameMiniMax, providerConfig)
	if err == nil || !strings.Contains(err.Error(), "missing API key") {
		t.Fatalf("Expected the missing API key error, got %v", err)
	}
}

func TestMinimax_New_ReplaysAssistantReasoningIntoHistory(t *testing.T) {
	const hidden = "the hidden reasoning"

	var raw []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","choices":[{"index":0,` +
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer srv.Close()

	prov := minimaxNew(t, &config.Config{MiniMaxAPIKey: "k", MiniMaxServerURL: srv.URL, MiniMaxProvider: "minimax"})

	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "hi"),
		{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{
			llms.TextPartWithReasoning("visible answer", &reasoning.ContentReasoning{Content: hidden}),
		}},
		llms.TextParts(llms.ChatMessageTypeHuman, "continue"),
	}
	if _, err := prov.CallEx(context.Background(), pconfig.OptionsTypeSimple, chain, nil); err != nil {
		t.Fatalf("call: %v", err)
	}
	if !strings.Contains(string(raw), hidden) {
		t.Errorf("the stored assistant turn's reasoning did not reach the wire; without PreserveReasoning MiniMax drops it. body=%s", raw)
	}
}

func TestMinimax_DefaultModels_ShipsTheCurrentPricedModelsWithTheDefaultFirst(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("Models list should not be empty")
	}

	wantModels := map[string]bool{
		"MiniMax-M3":             false,
		"MiniMax-M2.7":           false,
		"MiniMax-M2.7-highspeed": false,
	}

	for _, model := range models {
		if model.Name == "" {
			t.Error("Model name should not be empty")
		}

		if model.Name == "MiniMax-M2.5" || model.Name == "MiniMax-M2.5-highspeed" ||
			model.Name == "MiniMax-M2.1" || model.Name == "MiniMax-M2" || model.Name == "MiniMax-M1" {
			t.Errorf("Legacy model %q should not be in the model list", model.Name)
		}

		if _, ok := wantModels[model.Name]; ok {
			wantModels[model.Name] = true
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

	for name, found := range wantModels {
		if !found {
			t.Errorf("Expected model %q in models list, but it was not found", name)
		}
	}

	if models[0].Name != "MiniMax-M3" {
		t.Errorf("Expected first model to be MiniMax-M3 (default), got %q", models[0].Name)
	}
}
