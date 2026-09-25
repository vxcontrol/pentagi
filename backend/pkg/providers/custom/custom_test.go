package custom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

// customServer is an OpenAI-compatible endpoint listing its catalogue to a keyed GET /models and keeping the last chat request.
type customServer struct {
	url      string
	listings int
	path     string
	query    string
	header   http.Header
	body     map[string]any
}

func customEndpoint(t *testing.T, catalogue string) *customServer {
	t.Helper()

	s := &customServer{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			s.listings++
			if catalogue == "" || r.Header.Get("Authorization") != "Bearer k" {
				http.NotFound(w, r)
				return
			}
			fmt.Fprint(w, catalogue)
			return
		}

		raw, _ := io.ReadAll(r.Body)
		s.path, s.query, s.header, s.body = r.URL.Path, r.URL.RawQuery, r.Header.Clone(), nil
		_ = json.Unmarshal(raw, &s.body)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"x","object":"chat.completion","choices":[{"index":0,`+
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
	}))
	t.Cleanup(srv.Close)
	s.url = srv.URL

	return s
}

func TestCustom_BuildProviderConfig_DefaultsEveryAgentToTheServerModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configData string
		named      map[pconfig.ProviderOptionsType]string
		wantErr    string
	}{
		{name: "an empty object", configData: "{}"},
		{name: "the default empty config", configData: pconfig.EmptyProviderConfigRaw},
		{
			name:       "an agent naming its own model",
			configData: `{"simple": {"model": "custom-model", "temperature": 0.5}}`,
			// simple_json takes the simple agent's settings when it has none of its own.
			named: map[pconfig.ProviderOptionsType]string{
				pconfig.OptionsTypeSimple: "custom-model", pconfig.OptionsTypeSimpleJSON: "custom-model",
			},
		},
		{
			name:       "invalid json",
			configData: `{"simple": {"model": "test", "temperature": invalid}}`,
			wantErr:    "failed to parse config: invalid character",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			providerConfig, err := BuildProviderConfig(&config.Config{LLMServerModel: "test-model"}, []byte(tc.configData))
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			for _, agent := range pconfig.AllAgentTypes {
				want, named := tc.named[agent]
				if !named {
					want = "test-model"
				}

				var applied llms.CallOptions
				for _, option := range providerConfig.GetOptionsForType(agent) {
					option(&applied)
				}
				assert.Equal(t, want, applied.GetModel(), "agent %s", agent)
			}
		})
	}
}

type customAgent struct {
	model     string
	maxTokens int
}

func TestCustom_DefaultProviderConfig_ReadsTheAgentsFromTheFileOrTheEnvironment(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(file, []byte(`{
		"simple": {"model": "gpt-3.5-turbo", "temperature": 0.2, "max_tokens": 2000},
		"agent": {"model": "gpt-4", "temperature": 0.8}
	}`), 0o600))

	every := func(model string) map[pconfig.ProviderOptionsType]customAgent {
		want := make(map[pconfig.ProviderOptionsType]customAgent, len(pconfig.AllAgentTypes))
		for _, agent := range pconfig.AllAgentTypes {
			want[agent] = customAgent{model: model, maxTokens: 16384}
		}
		return want
	}

	tests := []struct {
		name    string
		cfg     config.Config
		want    map[pconfig.ProviderOptionsType]customAgent
		wantErr error
	}{
		{
			name: "the environment alone gives every agent the server model",
			cfg:  config.Config{LLMServerModel: "gpt-4o-mini"},
			want: every("gpt-4o-mini"),
		},
		{
			name: "a file overrides the environment for the agents it names",
			cfg:  config.Config{LLMServerModel: "gpt-4o-mini", LLMServerConfig: file},
			want: map[pconfig.ProviderOptionsType]customAgent{
				pconfig.OptionsTypeSimple:       {model: "gpt-3.5-turbo", maxTokens: 2000},
				pconfig.OptionsTypePrimaryAgent: {model: "gpt-4", maxTokens: 16384},
				pconfig.OptionsTypeCoder:        {model: "gpt-4o-mini", maxTokens: 16384},
			},
		},
		{name: "no model anywhere leaves the choice to the server", want: every("")},
		{
			name:    "a file that is not there",
			cfg:     config.Config{LLMServerConfig: "/nonexistent/config.yml"},
			wantErr: fs.ErrNotExist,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := tc.cfg
			cfg.LLMServerKey, cfg.LLMServerURL = "k", customEndpoint(t, "").url

			providerConfig, err := DefaultProviderConfig(&cfg)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			prov, err := New(&cfg, provider.DefaultProviderNameCustom, providerConfig, nil)
			require.NoError(t, err, "a catalogue the server does not list must not stop the door")

			assert.Equal(t, provider.ProviderCustom, prov.Type())
			assert.Equal(t, provider.DefaultProviderNameCustom, prov.Name())
			assert.Same(t, providerConfig, prov.GetProviderConfig())
			assert.NotEmpty(t, prov.GetRawConfig())

			for agent, want := range tc.want {
				maxTokens := 0
				var applied llms.CallOptions
				for _, option := range providerConfig.GetOptionsForType(agent) {
					option(&applied)
				}
				if applied.MaxTokens != nil {
					maxTokens = *applied.MaxTokens
				}

				assert.Equal(t, want, customAgent{model: prov.Model(agent), maxTokens: maxTokens}, "agent %s", agent)
			}
		})
	}
}

func TestCustom_New_RefusesAnAddressWithoutAKey(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LLMServerURL:   "https://llm.example.net/v1",
		LLMServerModel: "gpt-5.4-mini",
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)

	_, err = New(cfg, provider.DefaultProviderNameCustom, providerConfig,
		func(m pconfig.ModelsConfig) pconfig.ModelsConfig { return m })
	require.Error(t, err, "a gateway address with no key must fail at startup, not on every call")
	assert.Contains(t, err.Error(), "LLM_SERVER_KEY", "the error must name the variable the operator left empty")
}

func TestCustom_New_LoadsTheCatalogueOnlyForAnOpenAIDoor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		apiType      string
		want         pconfig.ModelsConfig
		wantListings int
	}{
		{"an unset api type lists the models with its key", "", pconfig.ModelsConfig{{Name: "gpt-5.6-terra"}}, 1},
		{"the openai api type lists the models with its key", "openai", pconfig.ModelsConfig{{Name: "gpt-5.6-terra"}}, 1},
		{"azure calls deployments by name and lists nothing", "azure", pconfig.ModelsConfig{}, 0},
		{"azure_ad calls deployments by name and lists nothing", "azure_ad", pconfig.ModelsConfig{}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := customEndpoint(t, `{"object":"list","data":[{"id":"gpt-5.6-terra","object":"model"}]}`)
			cfg := &config.Config{
				LLMServerURL:        server.url + "/",
				LLMServerKey:        "k",
				LLMServerModel:      "gpt-5.6-terra",
				LLMServerAPIType:    tc.apiType,
				LLMServerAPIVersion: "2024-10-21",
			}
			providerConfig, err := DefaultProviderConfig(cfg)
			require.NoError(t, err)

			prov, err := New(cfg, provider.DefaultProviderNameCustom, providerConfig, nil)
			require.NoError(t, err)

			assert.Equal(t, tc.want, prov.GetModels())
			assert.Equal(t, tc.wantListings, server.listings)
		})
	}
}

// Not parallel: it swaps the hooks of the global logger.
func TestCustom_New_WarnsWithTheGatewayAndTheCauseOnlyWhenTheListingFails(t *testing.T) {
	hook := new(logtest.Hook)
	previous := logrus.StandardLogger().ReplaceHooks(logrus.LevelHooks{})
	logrus.AddHook(hook)
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(previous) })

	tests := []struct {
		name       string
		gateway    func(t *testing.T) string
		wantCause  error
		wantText   string
		wantModels pconfig.ModelsConfig
	}{
		{
			name: "a gateway that refuses the connection",
			gateway: func(t *testing.T) string {
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				require.NoError(t, err)
				require.NoError(t, listener.Close())
				return "http://" + listener.Addr().String() + "/v1"
			},
			wantCause:  syscall.ECONNREFUSED,
			wantModels: pconfig.ModelsConfig{},
		},
		{
			name:       "a gateway that answers the listing with an error status",
			gateway:    func(t *testing.T) string { return customEndpoint(t, "").url },
			wantText:   "unexpected status code: 404",
			wantModels: pconfig.ModelsConfig{},
		},
		{
			name: "a gateway that lists its models",
			gateway: func(t *testing.T) string {
				return customEndpoint(t, `{"object":"list","data":[{"id":"gpt-5.6-terra","object":"model"}]}`).url
			},
			wantModels: pconfig.ModelsConfig{{Name: "gpt-5.6-terra"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook.Reset()
			address := tc.gateway(t)

			cfg := &config.Config{LLMServerKey: "k", LLMServerURL: address, LLMServerModel: "some-model"}
			providerConfig, err := DefaultProviderConfig(cfg)
			require.NoError(t, err)

			prov, err := New(cfg, "my-gateway", providerConfig, nil)
			require.NoError(t, err, "a gateway that does not list its models still builds a provider")
			assert.Equal(t, tc.wantModels, prov.GetModels())

			var warnings []*logrus.Entry
			for _, entry := range hook.AllEntries() {
				if entry.Level == logrus.WarnLevel {
					warnings = append(warnings, entry)
				}
			}
			if tc.wantCause == nil && tc.wantText == "" {
				assert.Empty(t, warnings, "a listing that succeeds has nothing to warn about")
				return
			}

			require.Len(t, warnings, 1)
			assert.Equal(t, provider.ProviderName("my-gateway"), warnings[0].Data["provider"])
			assert.Equal(t, address, warnings[0].Data["url"])
			cause, _ := warnings[0].Data[logrus.ErrorKey].(error)
			require.Error(t, cause, "the warning carries no cause")
			if tc.wantCause != nil {
				assert.ErrorIs(t, cause, tc.wantCause)
			} else {
				assert.EqualError(t, cause, tc.wantText)
			}
		})
	}
}

func TestCustom_CustomAPI_AddressesTheDeploymentTheAPITypeNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		apiType       string
		path          string
		query         string
		apiKey        string
		authorization string
		wantErr       string
	}{
		{
			name: "azure sends the deployment path, the api-version and its own key header", apiType: "azure",
			path: "/openai/deployments/gpt-5.6-terra/chat/completions", query: "api-version=2024-10-21", apiKey: "k",
		},
		{
			name: "azure_ad keeps the bearer token", apiType: "azure_ad",
			path: "/openai/deployments/gpt-5.6-terra/chat/completions", query: "api-version=2024-10-21",
			authorization: "Bearer k",
		},
		{name: "an unset api type keeps the plain openai address", path: "/chat/completions", authorization: "Bearer k"},
		{name: "an unknown api type is refused instead of falling back", apiType: "azrue", wantErr: `"azrue"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := customEndpoint(t, "")
			cfg := &config.Config{
				LLMServerURL:        server.url,
				LLMServerKey:        "k",
				LLMServerModel:      "gpt-5.6-terra",
				LLMServerAPIType:    tc.apiType,
				LLMServerAPIVersion: "2024-10-21",
			}
			providerConfig, err := DefaultProviderConfig(cfg)
			require.NoError(t, err)

			prov, err := New(cfg, provider.DefaultProviderNameCustom, providerConfig, nil)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr, "a typo must not silently build a plain OpenAI door")
				return
			}
			require.NoError(t, err)

			_, err = prov.Call(context.Background(), pconfig.OptionsTypeSimple, "hi")
			require.NoError(t, err)

			assert.Equal(t, tc.path, server.path)
			assert.Equal(t, tc.query, server.query)
			assert.Equal(t, tc.apiKey, server.header.Get("api-key"))
			assert.Equal(t, tc.authorization, server.header.Get("Authorization"))
		})
	}
}

func customDoor(t *testing.T, server *customServer, model, configData string) provider.Provider {
	t.Helper()

	cfg := &config.Config{LLMServerKey: "k", LLMServerURL: server.url, LLMServerModel: model}
	providerConfig, err := BuildProviderConfig(cfg, []byte(configData))
	require.NoError(t, err)

	prov, err := New(cfg, provider.DefaultProviderNameCustom, providerConfig, nil)
	require.NoError(t, err)

	return prov
}

func TestCustom_CallWithExtraOptions_SendsTheCallsSamplingOverTheAgentsUnlessThinkingRuns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		model   string
		agent   string
		extra   []llms.CallOption
		want    map[string]any
		without []string
	}{
		{
			// A model name no reasoning table matches, so the library leaves its sampling alone.
			name:  "the call's sampling wins over the agent's",
			model: "wire-test-model",
			agent: `{"simple": {"max_tokens": 100, "temperature": 0.9, "top_p": 0.1}}`,
			extra: []llms.CallOption{llms.WithMaxTokens(16), llms.WithTemperature(0.25), llms.WithTopP(0.5)},
			want:  map[string]any{"max_completion_tokens": 16.0, "temperature": 0.25, "top_p": 0.5},
		},
		{
			name:    "a thinking model drops the temperature while thinking runs",
			model:   "gpt-5.5",
			agent:   `{"simple": {}}`,
			extra:   []llms.CallOption{llms.WithTemperature(0.25)},
			without: []string{"temperature"},
		},
		{
			name:  "a thinking model keeps the temperature with thinking off",
			model: "gpt-5.5",
			agent: `{"simple": {"reasoning": {"mode": "off"}}}`,
			extra: []llms.CallOption{llms.WithTemperature(0.25)},
			want:  map[string]any{"temperature": 0.25},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := customEndpoint(t, "")
			prov := customDoor(t, server, tc.model, tc.agent)

			_, err := prov.CallWithExtraOptions(context.Background(), pconfig.OptionsTypeSimple,
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil, nil, tc.extra...)
			require.NoError(t, err)

			for field, want := range tc.want {
				assert.Equal(t, want, server.body[field], "%s; body=%v", field, server.body)
			}
			for _, field := range tc.without {
				assert.NotContains(t, server.body, field)
			}
		})
	}
}

func TestCustom_CallWithExtraOptions_SendsTheToolChoiceBesideAPrefilledAssistantTurn(t *testing.T) {
	t.Parallel()

	server := customEndpoint(t, "")
	prov := customDoor(t, server, "wire-test-model", `{"simple": {}}`)

	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "hi"),
		llms.TextParts(llms.ChatMessageTypeAI, "The answer is"),
	}
	tools := []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{
		Name: "noop", Description: "does nothing",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{}},
	}}}

	_, err := prov.CallWithExtraOptions(context.Background(), pconfig.OptionsTypeSimple, chain, tools, nil,
		llms.WithToolChoice("required"))
	require.NoError(t, err)

	assert.Equal(t, "required", server.body["tool_choice"])

	messages, _ := server.body["messages"].([]any)
	require.NotEmpty(t, messages, "body=%v", server.body)
	assert.Equal(t, map[string]any{"role": "assistant", "content": "The answer is"}, messages[len(messages)-1],
		"a prefilled turn must reach the wire last, neither dropped nor merged")
}

// TestCustom_EveryCallPathPutsTheAgentsAdaptiveEffortOnTheWire is keyed by calling method.
func TestCustom_EveryCallPathPutsTheAgentsAdaptiveEffortOnTheWire(t *testing.T) {
	t.Parallel()

	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}
	paths := []struct {
		name string
		call func(prov provider.Provider) error
	}{
		{"a single prompt", func(prov provider.Provider) error {
			_, err := prov.Call(context.Background(), pconfig.OptionsTypeSimple, "hi")
			return err
		}},
		{"a chain", func(prov provider.Provider) error {
			_, err := prov.CallEx(context.Background(), pconfig.OptionsTypeSimple, chain, nil)
			return err
		}},
		{"a chain with tools", func(prov provider.Provider) error {
			_, err := prov.CallWithTools(context.Background(), pconfig.OptionsTypeSimple, chain, nil, nil)
			return err
		}},
		{"a chain with extra options", func(prov provider.Provider) error {
			_, err := prov.CallWithExtraOptions(context.Background(), pconfig.OptionsTypeSimple, chain, nil, nil)
			return err
		}},
	}

	for _, path := range paths {
		t.Run(path.name, func(t *testing.T) {
			t.Parallel()

			server := customEndpoint(t, "")
			prov := customDoor(t, server, "claude-sonnet-5", `{"simple": {"reasoning": {"mode": "adaptive", "effort": "xhigh"}}}`)
			require.NoError(t, path.call(prov))

			effort := server.body["reasoning_effort"]
			if modern, ok := server.body["reasoning"].(map[string]any); ok {
				effort = modern["effort"]
			}
			assert.Equal(t, "xhigh", effort, "body=%v", server.body)
		})
	}
}
