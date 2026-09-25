package glm

import (
	"context"
	"encoding/json"
	"errors"
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
	"gopkg.in/yaml.v3"
)

func glmNew(t *testing.T, cfg *config.Config) provider.Provider {
	t.Helper()

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, provider.DefaultProviderNameGLM, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	return prov
}

func glmWireBody(t *testing.T, opt pconfig.ProviderOptionsType, patch map[string]any) map[string]any {
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

	cfg := &config.Config{GLMAPIKey: "k", GLMServerURL: srv.URL, GLMProvider: "zai"}
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
	prov, err := New(cfg, provider.DefaultProviderNameGLM, pc)
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

func TestGlm_New_BuildsTheDoorWithAPricedModelForEveryAgent(t *testing.T) {
	prov := glmNew(t, &config.Config{
		GLMAPIKey:    "test-key",
		GLMServerURL: "https://api.z.ai/api/paas/v4",
	})

	if prov.Type() != provider.ProviderGLM {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderGLM, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameGLM {
		t.Errorf("Expected provider name %v, got %v", provider.DefaultProviderNameGLM, prov.Name())
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
		} else if priceInfo.Input < 0 || priceInfo.Output < 0 {
			t.Errorf("Agent type %v should have non-negative input (%f) and output (%f) prices",
				agentType, priceInfo.Input, priceInfo.Output)
		}
	}
}

func TestGlm_New_PrefixesModelsOnlyWhenAProviderIsConfigured(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prefix string
		lead   string
	}{
		{name: "a configured provider leads every model", prefix: "zhipu", lead: "zhipu/"},
		{name: "no provider leaves every model bare"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prov := glmNew(t, &config.Config{
				GLMAPIKey:    "test-key",
				GLMServerURL: "https://api.z.ai/api/paas/v4",
				GLMProvider:  tc.prefix,
			})

			for _, agentType := range pconfig.AllAgentTypes {
				if got, want := prov.ModelWithPrefix(agentType), tc.lead+prov.Model(agentType); got != want {
					t.Errorf("Agent type %v: expected prefixed model %q, got %q", agentType, want, got)
				}
			}
		})
	}
}

func TestGlm_New_RefusesAMissingAPIKey(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	_, err = New(&config.Config{GLMServerURL: "https://api.z.ai/api/paas/v4"},
		provider.DefaultProviderNameGLM, providerConfig)
	if err == nil || !strings.Contains(err.Error(), "missing API key") {
		t.Fatalf("Expected the missing API key error, got %v", err)
	}
}

func TestGlm_DefaultModels_ShipsANonEmptyPricedCatalogue(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("Models list should not be empty")
	}

	current := map[string]bool{
		"glm-5.3":        false,
		"glm-5.3-flash":  false,
		"glm-5.3-flashx": false,
	}
	for _, model := range models {
		if _, ok := current[model.Name]; ok {
			current[model.Name] = true
		}
		if model.Name == "" {
			t.Error("Model name should not be empty")
		}

		if model.Price == nil {
			t.Errorf("Model %s should have price information", model.Name)
			continue
		}

		if model.Price.Input < 0 {
			t.Errorf("Model %s should have non-negative input price", model.Name)
		}

		if model.Price.Output < 0 {
			t.Errorf("Model %s should have non-negative output price", model.Name)
		}
	}
	for model, found := range current {
		if !found {
			t.Errorf("current model %s is missing", model)
		}
	}
}

func TestGlm_DefaultProviderConfig_ShipsNoExtraBodyThinkingToggle(t *testing.T) {
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

func TestGlm_DefaultProviderConfig_SendsLowEffortForUtilityAgents(t *testing.T) {
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
			body := glmWireBody(t, tc.opt, nil)
			if got := body["reasoning_effort"]; got != "low" {
				t.Errorf("reasoning_effort = %v, want \"low\"; body=%v", got, body)
			}
		})
	}
}

func TestGlm_DefaultProviderConfig_PreservesThinkingAcrossTurnsForEveryAgent(t *testing.T) {
	pc, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	var probed int
	for _, opt := range pconfig.AllAgentTypes {
		ac := pc.AgentConfigForType(opt)
		if ac == nil {
			continue
		}
		shipped, ok := ac.ExtraBody["thinking"].(map[string]any)
		if !ok {
			continue
		}
		probed++
		t.Run("the shipped "+string(opt), func(t *testing.T) {
			body := glmWireBody(t, opt, nil)
			wire, _ := body["thinking"].(map[string]any)
			for k, v := range shipped {
				if wire[k] != v {
					t.Errorf("thinking.%s = %v on the wire, want the shipped %v; body=%v", k, wire[k], v, body)
				}
			}
		})
	}
	if probed == 0 {
		t.Fatal("no shipped agent carries an extra_body thinking object")
	}
}

func TestGlm_DefaultProviderConfig_RefusesReasoningOffOnAnAlwaysThinkingModelBeforeTheWire(t *testing.T) {
	var reached bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","choices":[{"index":0,` +
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer srv.Close()

	cfg := &config.Config{GLMAPIKey: "k", GLMServerURL: srv.URL, GLMProvider: "zai"}
	pc, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	var asMap map[string]any
	if err := yaml.Unmarshal(pc.GetRawConfig(), &asMap); err != nil {
		t.Fatalf("raw: %v", err)
	}
	simple := asMap[string(pconfig.OptionsTypeSimple)].(map[string]any)
	simple["model"] = "glm-5.3-flash"
	simple["reasoning"] = map[string]any{"mode": "off"}
	patched, _ := yaml.Marshal(asMap)
	if pc, err = BuildProviderConfig(patched); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	prov, err := New(cfg, provider.DefaultProviderNameGLM, pc)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	_, err = prov.CallEx(context.Background(), pconfig.OptionsTypeSimple,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil)
	var offErr *reasoning.ErrReasoningOffUnsupported
	if !errors.As(err, &offErr) || reached {
		t.Fatalf("want a typed reasoning-off error for glm-5.3-flash before the wire, got %v (reached wire=%v)", err, reached)
	}
}
