package deepseek

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

func deepseekNew(t *testing.T, cfg *config.Config) provider.Provider {
	t.Helper()

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, provider.DefaultProviderNameDeepSeek, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	return prov
}

func deepseekWireBody(t *testing.T, opt pconfig.ProviderOptionsType, patch map[string]any) map[string]any {
	t.Helper()

	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","choices":[{"index":0,` +
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer srv.Close()

	cfg := &config.Config{DeepSeekAPIKey: "k", DeepSeekServerURL: srv.URL, DeepSeekProvider: "deepseek"}
	pc, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if patch != nil {
		var asMap map[string]any
		if err := yaml.Unmarshal(pc.GetRawConfig(), &asMap); err != nil {
			t.Fatalf("raw: %v", err)
		}
		agent, _ := asMap[string(opt)].(map[string]any)
		for k, v := range patch {
			agent[k] = v
		}
		patched, _ := yaml.Marshal(asMap)
		if pc, err = BuildProviderConfig(patched); err != nil {
			t.Fatalf("rebuild: %v", err)
		}
	}
	prov, err := New(cfg, provider.DefaultProviderNameDeepSeek, pc)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	if _, err = prov.CallEx(context.Background(), opt,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil); err != nil {
		t.Fatalf("call: %v", err)
	}
	if body == nil {
		t.Fatal("no request reached the wire")
	}
	return body
}

func deepseekThinkingType(body map[string]any) string {
	th, ok := body["thinking"].(map[string]any)
	if !ok {
		return ""
	}
	s, _ := th["type"].(string)
	return s
}

func TestDeepseek_New_BuildsTheDoorWithAPricedModelForEveryAgent(t *testing.T) {
	prov := deepseekNew(t, &config.Config{
		DeepSeekAPIKey:    "test-key",
		DeepSeekServerURL: "https://api.deepseek.com",
	})

	if prov.Type() != provider.ProviderDeepSeek {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderDeepSeek, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameDeepSeek {
		t.Errorf("Expected provider name %v, got %v", provider.DefaultProviderNameDeepSeek, prov.Name())
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

func TestDeepseek_New_PrefixesModelsOnlyWhenAProviderIsConfigured(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prefix string
		lead   string
	}{
		{name: "a configured provider leads every model", prefix: "deepseek", lead: "deepseek/"},
		{name: "no provider leaves every model bare"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prov := deepseekNew(t, &config.Config{
				DeepSeekAPIKey:    "test-key",
				DeepSeekServerURL: "https://api.deepseek.com",
				DeepSeekProvider:  tc.prefix,
			})

			for _, agentType := range pconfig.AllAgentTypes {
				if got, want := prov.ModelWithPrefix(agentType), tc.lead+prov.Model(agentType); got != want {
					t.Errorf("Agent type %v: expected prefixed model %q, got %q", agentType, want, got)
				}
			}
		})
	}
}

func TestDeepseek_New_RefusesAMissingAPIKey(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	_, err = New(&config.Config{DeepSeekServerURL: "https://api.deepseek.com"},
		provider.DefaultProviderNameDeepSeek, providerConfig)
	if err == nil || !strings.Contains(err.Error(), "missing API key") {
		t.Fatalf("Expected the missing API key error, got %v", err)
	}
}

func TestDeepseek_DefaultModels_ShipsANonEmptyPricedCatalogue(t *testing.T) {
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

func TestDeepseek_DefaultProviderConfig_ShipsNoExtraBodyThinkingToggle(t *testing.T) {
	pc, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	for _, opt := range pconfig.AllAgentTypes {
		ac := pc.AgentConfigForType(opt)
		if ac == nil {
			continue
		}
		if _, ok := ac.ExtraBody["enable_thinking"]; ok {
			t.Errorf("%s ships enable_thinking in extra_body: %v", opt, ac.ExtraBody)
		}
		if th, ok := ac.ExtraBody["thinking"].(map[string]any); ok {
			if _, ok := th["type"]; ok {
				t.Errorf("%s ships thinking.type in extra_body: %v", opt, ac.ExtraBody)
			}
		}
	}
}

func TestDeepseek_DefaultProviderConfig_DisablesThinkingOnTheWireWhereReasoningIsOff(t *testing.T) {
	for _, tc := range []struct {
		name  string
		opt   pconfig.ProviderOptionsType
		patch map[string]any
	}{
		{name: "the shipped simple agent", opt: pconfig.OptionsTypeSimple},
		{name: "the shipped simple json agent", opt: pconfig.OptionsTypeSimpleJSON},
		{name: "the shipped reflector", opt: pconfig.OptionsTypeReflector},
		{name: "the shipped searcher", opt: pconfig.OptionsTypeSearcher},
		{name: "the shipped enricher", opt: pconfig.OptionsTypeEnricher},
		{
			name:  "an assistant switched to reasoning mode off",
			opt:   pconfig.OptionsTypeAssistant,
			patch: map[string]any{"reasoning": map[string]any{"mode": "off"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := deepseekWireBody(t, tc.opt, tc.patch)
			if got := deepseekThinkingType(body); got != "disabled" {
				t.Errorf("thinking.type = %q, want \"disabled\"; body=%v", got, body)
			}
		})
	}
}

func TestDeepseek_New_KeepsPenaltiesOffTheWire(t *testing.T) {
	body := deepseekWireBody(t, pconfig.OptionsTypeSimple, map[string]any{
		"frequency_penalty": 0.5,
		"presence_penalty":  0.5,
	})
	for _, field := range []string{"frequency_penalty", "presence_penalty"} {
		if _, ok := body[field]; ok {
			t.Errorf("%s reached the wire (%v); langchaingo drops it for DeepSeek", field, body[field])
		}
	}
}
