package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database/converter"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

const ollamaDefaultModel = "llama3.1:8b-instruct-q8_0"

func ollamaNew(t *testing.T, cfg *config.Config) provider.Provider {
	t.Helper()

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)

	prov, err := New(cfg, provider.DefaultProviderNameOllama, providerConfig)
	require.NoError(t, err)

	return prov
}

func TestOllama_New_BuildsTheDoorOnTheConfiguredServerModel(t *testing.T) {
	cfg := &config.Config{
		OllamaServerURL:   "http://localhost:11434",
		OllamaServerModel: ollamaDefaultModel,
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)

	prov, err := New(cfg, provider.DefaultProviderNameOllama, providerConfig)
	require.NoError(t, err)

	assert.Equal(t, provider.ProviderOllama, prov.Type())
	assert.Equal(t, provider.DefaultProviderNameOllama, prov.Name())
	assert.Same(t, providerConfig, prov.GetProviderConfig())
	assert.NotEmpty(t, prov.GetRawConfig())
	for _, agent := range pconfig.AllAgentTypes {
		assert.Equal(t, ollamaDefaultModel, prov.Model(agent), "the shipped config names no model, so %s runs the server model", agent)
	}
}

func TestOllama_GetUsage_ReadsPromptAndCompletionTokens(t *testing.T) {
	prov := ollamaNew(t, &config.Config{OllamaServerURL: "http://localhost:11434"})

	usage := prov.GetUsage(map[string]any{"PromptTokens": 100, "CompletionTokens": 50})
	assert.Equal(t, int64(100), usage.Input)
	assert.Equal(t, int64(50), usage.Output)
}

func TestOllama_DefaultProviderConfig_ReadsTheConfiguredFileOverTheEmbeddedOne(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "custom.yml")
	require.NoError(t, os.WriteFile(custom, []byte("simple:\n  model: gemma3:1b\n  max_tokens: 2000\n"), 0o600))
	broken := filepath.Join(dir, "broken.yml")
	require.NoError(t, os.WriteFile(broken, []byte("simple: [\n"), 0o600))

	for _, tc := range []struct {
		name      string
		path      string
		errIs     error
		errText   string
		model     string
		maxTokens map[pconfig.ProviderOptionsType]int
		absent    []pconfig.ProviderOptionsType
	}{
		{
			// The assistant is left out: its embedded block equals the primary agent's, which it falls back to.
			name: "no configured file takes the embedded config",
			maxTokens: map[pconfig.ProviderOptionsType]int{
				pconfig.OptionsTypeSimple: 8192, pconfig.OptionsTypeSimpleJSON: 4096, pconfig.OptionsTypePrimaryAgent: 16384,
				pconfig.OptionsTypeGenerator: 20480, pconfig.OptionsTypeRefiner: 16384, pconfig.OptionsTypeAdviser: 8192,
			},
		},
		{
			name:      "a configured file replaces the embedded config",
			path:      custom,
			model:     "gemma3:1b",
			maxTokens: map[pconfig.ProviderOptionsType]int{pconfig.OptionsTypeSimple: 2000},
			absent:    []pconfig.ProviderOptionsType{pconfig.OptionsTypeGenerator},
		},
		{name: "a configured file that is missing is an error", path: filepath.Join(dir, "missing.yml"), errIs: fs.ErrNotExist},
		{name: "a configured file that does not parse is an error", path: broken, errText: "failed to parse config"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			providerConfig, err := DefaultProviderConfig(&config.Config{OllamaServerConfig: tc.path})
			if tc.errIs != nil || tc.errText != "" {
				if tc.errIs != nil {
					assert.ErrorIs(t, err, tc.errIs)
				}
				if tc.errText != "" {
					assert.ErrorContains(t, err, tc.errText)
				}
				return
			}
			require.NoError(t, err)

			for agentType, want := range tc.maxTokens {
				if agent := providerConfig.AgentConfigForType(agentType); assert.NotNil(t, agent, "agent type %s should be configured", agentType) {
					assert.Equal(t, want, agent.MaxTokens, "max_tokens of %s", agentType)
				}
			}
			for _, agentType := range tc.absent {
				assert.Nil(t, providerConfig.AgentConfigForType(agentType), "agent type %s should not be configured", agentType)
			}
			assert.Equal(t, tc.model, providerConfig.AgentConfigForType(pconfig.OptionsTypeSimple).Model)
		})
	}
}

func TestOllama_LoadAvailableModelsFromServer_TakesTheThinkingFlagFromTheDaemonAndFallsBackToTheTables(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"models":[
			{"name":"deepseek-r1:8b","capabilities":["completion","thinking"]},
			{"name":"deepseek-r1:70b","capabilities":["completion"]},
			{"name":"deepseek-r1:14b"},
			{"name":"qwen3-coder:30b"},
			{"name":"llama3.1:8b"}
		]}`)
	}))
	defer srv.Close()

	models := ollamaNew(t, &config.Config{
		OllamaServerURL:               srv.URL,
		OllamaServerModel:             "llama3.1:8b",
		OllamaServerLoadModelsEnabled: true,
	}).GetModels()

	thinking := make(map[string]*bool, len(models))
	names := make([]string, 0, len(models))
	for _, model := range models {
		thinking[model.Name] = model.Thinking
		names = append(names, model.Name)
	}

	require.Len(t, models, 5)
	require.ElementsMatch(t, []string{
		"deepseek-r1:8b", "deepseek-r1:70b", "deepseek-r1:14b", "qwen3-coder:30b", "llama3.1:8b",
	}, names, "every listed model reaches the catalogue, reasoning or not")

	require.NotNil(t, thinking["deepseek-r1:8b"])
	assert.True(t, *thinking["deepseek-r1:8b"])

	assert.Nil(t, thinking["deepseek-r1:70b"],
		"a listing that answers about capabilities outranks the tables, even for a family they call reasoning")

	require.NotNil(t, thinking["deepseek-r1:14b"], "a silent listing leaves the answer to the tables")
	assert.True(t, *thinking["deepseek-r1:14b"])

	assert.Nil(t, thinking["qwen3-coder:30b"],
		"the hint predicate calls this family reasoning and the wire predicate does not; the wire decides, or the door sends think to a daemon that answers 400")

	assert.Nil(t, thinking["llama3.1:8b"])
}

func TestOllama_CatalogFromConfig_ListsTheBaseAndAgentModelsWithTheirThinkingFlag(t *testing.T) {
	thinks := true
	for _, tc := range []struct {
		name       string
		baseModel  string
		configData []byte // nil: the shipped config, which names no model
		want       map[string]*bool
	}{
		{
			name:      "a reasoning base model alone under the shipped config",
			baseModel: "deepseek-r1:8b",
			want:      map[string]*bool{"deepseek-r1:8b": &thinks},
		},
		{
			name:      "config file agent models beside the base model",
			baseModel: "llama3.1:8b",
			configData: []byte(`
simple:
  model: "qwen3-coder:30b"
  temperature: 1
  n: 1
  max_tokens: 8192
primary_agent:
  model: "deepseek-r1:8b"
  temperature: 1
  n: 1
  max_tokens: 8192
`),
			want: map[string]*bool{"deepseek-r1:8b": &thinks, "llama3.1:8b": nil, "qwen3-coder:30b": nil},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{OllamaServerURL: "http://localhost:11434", OllamaServerModel: tc.baseModel}

			providerConfig, err := DefaultProviderConfig(cfg)
			if tc.configData != nil {
				providerConfig, err = BuildProviderConfig(cfg, tc.configData)
			}
			require.NoError(t, err)

			prov, err := New(cfg, provider.DefaultProviderNameOllama, providerConfig)
			require.NoError(t, err)

			got := make(map[string]*bool)
			for _, model := range prov.GetModels() {
				got[model.Name] = model.Thinking
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOllama_CatalogFromConfig_LetsTheModelFormOfferOffForAThinkingModel(t *testing.T) {
	prov := ollamaNew(t, &config.Config{OllamaServerURL: "http://localhost:11434", OllamaServerModel: "deepseek-r1:8b"})

	models := converter.ConvertModels(prov.GetModels(), prov.Type().ReasoningProvider())
	require.Len(t, models, 1)
	require.Equal(t, "deepseek-r1:8b", models[0].Name)
	require.NotNil(t, models[0].Reasoning)
	require.NotNil(t, models[0].Reasoning.CannotDisable)
	assert.False(t, *models[0].Reasoning.CannotDisable, "deepseek-r1:8b must offer Off")
}

func TestOllama_CatalogFromConfig_PricesAgentModelsOnlyOnTheCloud(t *testing.T) {
	configData := []byte(`
simple:
  model: "nemotron-3-nano:30b-cloud"
  n: 1
primary_agent:
  model: "kimi-k2.6:cloud"
  n: 1
searcher:
  model: "deepseek-v4.1-flash:cloud"
  n: 1
`)

	for _, tc := range []struct {
		name      string
		serverURL string
		want      map[pconfig.ProviderOptionsType]*pconfig.PriceInfo
	}{
		{
			name:      "the vendor cloud prices every agent model",
			serverURL: "https://ollama.com",
			want: map[pconfig.ProviderOptionsType]*pconfig.PriceInfo{
				pconfig.OptionsTypeSimple:       {Input: 0.06, Output: 0.24},
				pconfig.OptionsTypePrimaryAgent: {Input: 0.95, Output: 4.00, CacheRead: 0.16},
				pconfig.OptionsTypeSearcher:     {Input: 0.30, Output: 1.20, CacheRead: 0.006},
			},
		},
		{
			name:      "a local daemon prices none",
			serverURL: "http://localhost:11434",
			want: map[pconfig.ProviderOptionsType]*pconfig.PriceInfo{
				pconfig.OptionsTypeSimple:       nil,
				pconfig.OptionsTypePrimaryAgent: nil,
				pconfig.OptionsTypeSearcher:     nil,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{OllamaServerURL: tc.serverURL, OllamaServerModel: "gpt-oss:120b-cloud"}

			providerConfig, err := BuildProviderConfig(cfg, configData)
			require.NoError(t, err)

			prov, err := New(cfg, provider.DefaultProviderNameOllama, providerConfig)
			require.NoError(t, err)

			for opt, price := range tc.want {
				assert.Equal(t, price, prov.GetPriceInfo(opt), opt)
			}
		})
	}
}

func TestOllama_GetPriceInfo_FallsBackToTheAgentModelsCatalogueEntry(t *testing.T) {
	providerConfig, err := DefaultProviderConfig(&config.Config{})
	require.NoError(t, err)
	require.Nil(t, providerConfig.GetPriceInfoForType(pconfig.OptionsTypeSimple),
		"ollama ships no per-agent prices, which is what makes the fallback load-bearing")

	agent := providerConfig.AgentConfigForType(pconfig.OptionsTypeSimple)
	require.NotNil(t, agent)

	for _, tc := range []struct {
		name  string
		price *pconfig.PriceInfo
	}{
		{name: "a priced model lends the agent its price", price: &pconfig.PriceInfo{Input: 3.0, Output: 15.0, CacheRead: 0.3}},
		{name: "an unpriced model leaves the agent unpriced"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &ollamaProvider{
				model:          agent.Model,
				models:         pconfig.ModelsConfig{{Name: agent.Model, Price: tc.price}},
				providerConfig: providerConfig,
			}

			assert.Equal(t, tc.price, p.GetPriceInfo(pconfig.OptionsTypeSimple))
		})
	}
}

func ollamaChat(t *testing.T, cfg *config.Config, configData string) (map[string]any, http.Header) {
	t.Helper()

	var (
		body   map[string]any
		header http.Header
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			http.NotFound(w, r)
			return
		}
		header = r.Header.Clone()
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"model":"gpt-oss:120b","created_at":"2026-01-01T00:00:00Z",`+
			`"message":{"role":"assistant","content":"ok"},"done":true,"done_reason":"stop",`+
			`"prompt_eval_count":1,"eval_count":1}`)
	}))
	t.Cleanup(srv.Close)

	cfg.OllamaServerURL = srv.URL
	providerConfig, err := BuildProviderConfig(cfg, []byte(configData))
	require.NoError(t, err)
	prov, err := New(cfg, provider.DefaultProviderNameOllama, providerConfig)
	require.NoError(t, err)

	_, err = prov.CallEx(context.Background(), pconfig.OptionsTypeSimple,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil)
	require.NoError(t, err)

	return body, header
}

func TestOllama_New_SendsTheCloudAPIKeyAsABearerToken(t *testing.T) {
	_, header := ollamaChat(t, &config.Config{OllamaServerModel: "gpt-oss:120b", OllamaServerAPIKey: "ollama-cloud-key"},
		"simple:\n  model: gpt-oss:120b\n")

	assert.Equal(t, "Bearer ollama-cloud-key", header.Get("Authorization"))
}

func TestOllama_CallEx_NestsSamplingUnderOptions(t *testing.T) {
	body, _ := ollamaChat(t, &config.Config{OllamaServerModel: "gpt-oss:120b"},
		"simple:\n  model: gpt-oss:120b\n  max_tokens: 4096\n  temperature: 0.4\n  top_p: 0.3\n")

	options, _ := body["options"].(map[string]any)
	require.NotNil(t, options, "no options block reached the wire; body=%v", body)
	// num_predict is this door's name for the output cap.
	for field, want := range map[string]float64{"temperature": 0.4, "top_p": 0.3, "num_predict": 4096} {
		assert.Equal(t, want, options[field], "%s; options=%v", field, options)
	}
}
