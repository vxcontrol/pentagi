package kimi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

func kimiNew(t *testing.T, cfg *config.Config) provider.Provider {
	t.Helper()

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, provider.DefaultProviderNameKimi, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	return prov
}

func kimiWireBody(t *testing.T, opt pconfig.ProviderOptionsType, patch map[string]any) map[string]any {
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

	cfg := &config.Config{KimiAPIKey: "k", KimiServerURL: srv.URL, KimiProvider: "moonshot"}
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
	prov, err := New(cfg, provider.DefaultProviderNameKimi, pc)
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

func kimiThinkingType(body map[string]any) string {
	th, ok := body["thinking"].(map[string]any)
	if !ok {
		return ""
	}
	s, _ := th["type"].(string)
	return s
}

func TestKimi_New_BuildsTheDoorWithAPricedModelForEveryAgent(t *testing.T) {
	prov := kimiNew(t, &config.Config{
		KimiAPIKey:    "test-key",
		KimiServerURL: "https://api.moonshot.ai/v1",
	})

	if prov.Type() != provider.ProviderKimi {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderKimi, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameKimi {
		t.Errorf("Expected provider name %v, got %v", provider.DefaultProviderNameKimi, prov.Name())
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

func TestKimi_New_PrefixesModelsOnlyWhenAProviderIsConfigured(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prefix string
		lead   string
	}{
		{name: "a configured provider leads every model", prefix: "moonshot", lead: "moonshot/"},
		{name: "no provider leaves every model bare"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prov := kimiNew(t, &config.Config{
				KimiAPIKey:    "test-key",
				KimiServerURL: "https://api.moonshot.ai/v1",
				KimiProvider:  tc.prefix,
			})

			for _, agentType := range pconfig.AllAgentTypes {
				if got, want := prov.ModelWithPrefix(agentType), tc.lead+prov.Model(agentType); got != want {
					t.Errorf("Agent type %v: expected prefixed model %q, got %q", agentType, want, got)
				}
			}
		})
	}
}

func TestKimi_New_RefusesAMissingAPIKey(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	_, err = New(&config.Config{KimiServerURL: "https://api.moonshot.ai/v1"},
		provider.DefaultProviderNameKimi, providerConfig)
	if err == nil || !strings.Contains(err.Error(), "missing API key") {
		t.Fatalf("Expected the missing API key error, got %v", err)
	}
}

func TestKimi_DefaultModels_ShipsANonEmptyPricedCatalogue(t *testing.T) {
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

func TestKimi_DefaultProviderConfig_ShipsNoExtraBodyThinkingToggle(t *testing.T) {
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

func TestKimi_DefaultProviderConfig_DisablesThinkingOnTheWireWhereReasoningIsOff(t *testing.T) {
	off := map[string]any{"mode": "off"}
	for _, tc := range []struct {
		name  string
		opt   pconfig.ProviderOptionsType
		patch map[string]any
		keep  string // the thinking.keep a configured extra_body must still carry
	}{
		{name: "the shipped simple agent", opt: pconfig.OptionsTypeSimple},
		{name: "the shipped simple json agent", opt: pconfig.OptionsTypeSimpleJSON},
		{name: "the shipped reflector", opt: pconfig.OptionsTypeReflector},
		{name: "the shipped searcher", opt: pconfig.OptionsTypeSearcher},
		{name: "the shipped enricher", opt: pconfig.OptionsTypeEnricher},
		{
			name:  "an installer switched to reasoning mode off",
			opt:   pconfig.OptionsTypeInstaller,
			patch: map[string]any{"reasoning": off},
		},
		{
			name: "an installer switched to reasoning mode off beside a vendor thinking key",
			opt:  pconfig.OptionsTypeInstaller,
			patch: map[string]any{
				"reasoning":  off,
				"extra_body": map[string]any{"thinking": map[string]any{"keep": "all"}},
			},
			keep: "all",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := kimiWireBody(t, tc.opt, tc.patch)
			if got := kimiThinkingType(body); got != "disabled" {
				t.Errorf("thinking.type = %q, want \"disabled\"; body=%v", got, body)
			}
			if tc.keep == "" {
				return
			}
			if wire, _ := body["thinking"].(map[string]any); wire["keep"] != tc.keep {
				t.Errorf("thinking.keep = %v on the wire, want the configured %s; body=%v", wire["keep"], tc.keep, body)
			}
		})
	}
}

func TestKimi_DefaultProviderConfig_KeepsTemperatureOffTheWireForNonThinkingAgents(t *testing.T) {
	for _, tc := range []struct {
		name string
		opt  pconfig.ProviderOptionsType
	}{
		{name: "the shipped simple agent", opt: pconfig.OptionsTypeSimple},
		{name: "the shipped simple json agent", opt: pconfig.OptionsTypeSimpleJSON},
		{name: "the shipped reflector", opt: pconfig.OptionsTypeReflector},
		{name: "the shipped searcher", opt: pconfig.OptionsTypeSearcher},
		{name: "the shipped enricher", opt: pconfig.OptionsTypeEnricher},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := kimiWireBody(t, tc.opt, nil)
			if _, ok := body["temperature"]; ok {
				t.Errorf("temperature reached the wire (%v); the library drops it for kimi-k2.6, extra_body must not re-add it",
					body["temperature"])
			}
		})
	}
}

func TestKimi_DefaultProviderConfig_SendsADocumentedThinkingShapeFromK27CodeAgents(t *testing.T) {
	pc, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	var probed int
	for _, opt := range pconfig.AllAgentTypes {
		ac := pc.AgentConfigForType(opt)
		if ac == nil || !strings.HasPrefix(ac.Model, "kimi-k2.7-code") {
			continue
		}
		probed++
		t.Run("the shipped "+string(opt)+" on kimi-k2.7-code", func(t *testing.T) {
			body := kimiWireBody(t, opt, nil)
			th, sent := body["thinking"]
			if !sent {
				return
			}
			if want := map[string]any{"type": "enabled", "keep": "all"}; !reflect.DeepEqual(th, want) {
				t.Errorf("thinking = %v, want it omitted or exactly %v; body=%v", th, want, body)
			}
		})
	}
	if probed == 0 {
		t.Fatal("no shipped agent runs kimi-k2.7-code")
	}
}
