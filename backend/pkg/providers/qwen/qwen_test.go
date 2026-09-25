package qwen

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

type qwenRoute struct {
	name   string
	prefix string
	direct bool
}

var qwenRoutes = []qwenRoute{
	{name: "direct DashScope with the default empty prefix", direct: true},
	{name: "a gateway with the dashscope prefix", prefix: "dashscope"},
}

func qwenNew(t *testing.T, cfg *config.Config) provider.Provider {
	t.Helper()

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, provider.DefaultProviderNameQwen, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	return prov
}

func qwenWireBody(t *testing.T, r qwenRoute, opt pconfig.ProviderOptionsType, patch map[string]any) map[string]any {
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

	cfg := &config.Config{QwenAPIKey: "k", QwenServerURL: srv.URL, QwenProvider: r.prefix}
	if r.direct {
		cfg.QwenServerURL = "http://dashscope-us.aliyuncs.com/compatible-mode/v1"
		cfg.ProxyURL = srv.URL
	}
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
	prov, err := New(cfg, provider.DefaultProviderNameQwen, pc)
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

func TestQwen_New_BuildsTheDoorWithAPricedModelForEveryAgent(t *testing.T) {
	prov := qwenNew(t, &config.Config{
		QwenAPIKey:    "test-key",
		QwenServerURL: "https://dashscope-us.aliyuncs.com/compatible-mode/v1",
	})

	if prov.Type() != provider.ProviderQwen {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderQwen, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameQwen {
		t.Errorf("Expected provider name %v, got %v", provider.DefaultProviderNameQwen, prov.Name())
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

func TestQwen_BuildProviderConfig_UsesQwen38MaxAsTheDefaultModel(t *testing.T) {
	cfg, err := BuildProviderConfig([]byte("simple:\n  n: 1\n"))
	if err != nil {
		t.Fatalf("build provider config: %v", err)
	}

	options := llms.CallOptions{}
	for _, option := range cfg.GetOptionsForType(pconfig.OptionsTypeSimple) {
		option(&options)
	}
	if got := options.GetModel(); got != "qwen3.8-max" {
		t.Errorf("default model = %q, want qwen3.8-max", got)
	}
}

func TestQwen_DefaultProviderConfig_UsesOnlyModelsSharedByBothPlatforms(t *testing.T) {
	cfg, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("load provider config: %v", err)
	}

	shared := map[string]bool{
		"qwen3.8-max": true, "qwen3.8-flash": true,
		"qwen3.7-plus": true, "qwen3.7-max": true, "qwen3.6-flash": true,
		"deepseek-v4.1-flash": true, "deepseek-v4-pro": true,
	}
	for _, agentType := range pconfig.AllAgentTypes {
		agent := cfg.AgentConfigForType(agentType)
		if agent == nil {
			t.Errorf("%s has no configuration", agentType)
			continue
		}
		if !shared[agent.Model] {
			t.Errorf("%s model %q is not shared by Qwen Cloud and DashScope", agentType, agent.Model)
		}
	}
}

func TestQwen_New_PrefixesModelsOnlyWhenAProviderIsConfigured(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prefix string
		lead   string
	}{
		{name: "a configured provider leads every model", prefix: "tongyi", lead: "tongyi/"},
		{name: "qwen cloud leads every model", prefix: "qwen_cloud", lead: "qwen_cloud/"},
		{name: "no provider leaves every model bare"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prov := qwenNew(t, &config.Config{
				QwenAPIKey:    "test-key",
				QwenServerURL: "https://dashscope-us.aliyuncs.com/compatible-mode/v1",
				QwenProvider:  tc.prefix,
			})

			for _, agentType := range pconfig.AllAgentTypes {
				if got, want := prov.ModelWithPrefix(agentType), tc.lead+prov.Model(agentType); got != want {
					t.Errorf("Agent type %v: expected prefixed model %q, got %q", agentType, want, got)
				}
			}
		})
	}
}

func TestQwen_New_RefusesAMissingAPIKey(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	_, err = New(&config.Config{QwenServerURL: "https://dashscope-us.aliyuncs.com/compatible-mode/v1"},
		provider.DefaultProviderNameQwen, providerConfig)
	if err == nil || !strings.Contains(err.Error(), "missing API key") {
		t.Fatalf("Expected the missing API key error, got %v", err)
	}
}

func TestQwen_DefaultModels_ShipsANonEmptyPricedCatalogue(t *testing.T) {
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

func TestQwen_DefaultModels_IncludesCurrentTextModelsFromBothPlatforms(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("load models: %v", err)
	}

	want := map[string]bool{
		"qwen3.8-max": false, "qwen3.8-flash": false, "qwen3.8-omni-flash": false,
		"qwen3.7-plus": false, "qwen3.7-max": false, "qwen3.7-flash": false,
		"qwen3.6-plus": false, "qwen3.6-flash": false, "qwen3.6-27b": false,
		"qwen3.5-plus": false, "qwen3.5-flash": false, "qwen3.5-27b": false,
		"qwen-max": false, "qwen-plus": false, "qwen-flash": false,
		"qwen3-coder-flash": false, "deepseek-v4.1-flash": false,
		"deepseek-v4-pro": false, "deepseek-v4-pro-0813": false,
		"glm-5.2-fast-preview": false, "glm-5.3-prime": false, "kimi-k3": false,
	}
	for _, model := range models {
		if _, ok := want[model.Name]; ok {
			want[model.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("current text model %q is missing", name)
		}
	}
}

func TestQwen_DefaultProviderConfig_ShipsNoExtraBodyThinkingToggle(t *testing.T) {
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
	}
}

func TestQwen_DefaultProviderConfig_LeavesReasoningAtTheCrossPlatformDefault(t *testing.T) {
	pc, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	for _, opt := range pconfig.AllAgentTypes {
		ac := pc.AgentConfigForType(opt)
		if ac != nil && !ac.Reasoning.IsZero() {
			t.Errorf("%s pins reasoning controls that differ between Qwen Cloud and DashScope: %+v", opt, ac.Reasoning)
		}
	}
}

func TestQwen_DefaultProviderConfig_DisablesThinkingOnDashScopeWhenRequested(t *testing.T) {
	for _, r := range qwenRoutes {
		for _, tc := range []struct {
			name  string
			opt   pconfig.ProviderOptionsType
			patch map[string]any
		}{
			{
				name:  "an assistant switched to reasoning mode off",
				opt:   pconfig.OptionsTypeAssistant,
				patch: map[string]any{"reasoning": map[string]any{"mode": "off"}},
			},
		} {
			t.Run(tc.name+" over "+r.name, func(t *testing.T) {
				body := qwenWireBody(t, r, tc.opt, tc.patch)
				if got, ok := body["enable_thinking"].(bool); !ok || got {
					t.Errorf("enable_thinking = %v (present=%v), want false; body=%v", body["enable_thinking"], ok, body)
				}
			})
		}
	}
}
