package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestBedrock_DefaultProviderConfig_GivesEveryAgentAModelAndAPrice(t *testing.T) {
	t.Parallel()

	providerConfig, err := DefaultProviderConfig(&config.Config{})
	require.NoError(t, err)

	prov, err := New(&config.Config{
		BedrockRegion: "us-east-1", BedrockAccessKey: "test-key", BedrockSecretKey: "test-key",
	}, provider.DefaultProviderNameBedrock, providerConfig)
	require.NoError(t, err)

	assert.Equal(t, provider.ProviderBedrock, prov.Type())
	assert.NotEmpty(t, prov.GetRawConfig())
	assert.Same(t, providerConfig, prov.GetProviderConfig())

	for _, agentType := range pconfig.AllAgentTypes {
		assert.NotEmpty(t, prov.Model(agentType), "agent type %v has no model", agentType)

		priceInfo := prov.GetPriceInfo(agentType)
		if assert.NotNil(t, priceInfo, "agent type %v has no price", agentType) {
			assert.Positive(t, priceInfo.Input, "agent type %v input price", agentType)
			assert.Positive(t, priceInfo.Output, "agent type %v output price", agentType)
		}
	}
}

func TestBedrock_DefaultProviderConfig_ReadsAnExternalFileInsteadOfTheEmbeddedOne(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "bedrock.provider.yml")
	external := "simple:\n  model: zai.glm-4.7-flash\n  temperature: 0.5\n  n: 1\n  max_tokens: 1000\n" +
		"  price:\n    input: 0.15\n    output: 0.6\n"
	require.NoError(t, os.WriteFile(path, []byte(external), 0o600))

	cfg := &config.Config{
		BedrockRegion: "us-east-1", BedrockAccessKey: "test-key", BedrockSecretKey: "test-key", BedrockConfig: path,
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)

	prov, err := New(cfg, provider.DefaultProviderNameBedrock, providerConfig)
	require.NoError(t, err)

	assert.Equal(t, "zai.glm-4.7-flash", prov.Model(pconfig.OptionsTypeSimple))
}

func TestBedrock_DefaultModels_PricesEveryModelAndDatesItToItsBedrockLaunch(t *testing.T) {
	t.Parallel()

	const (
		whatsNew           = "https://aws.amazon.com/about-aws/whats-new/"
		docHistory         = "https://docs.aws.amazon.com/bedrock/latest/userguide/doc-history.html"
		novaLaunch         = whatsNew + "2024/12/amazon-nova-foundation-models-bedrock"
		llama3Launch       = whatsNew + "2024/04/meta-llama-3-amazon-bedrock/"
		llama31Launch      = whatsNew + "2024/07/meta-llama-3-1-generative-ai-models-amazon-bedrock"
		qwen3Launch        = whatsNew + "2025/09/qwen3-models-fully-managed-amazon-bedrock"
		openWeights18      = whatsNew + "2025/12/amazon-bedrock-fully-managed-open-weight-models"
		openWeights6       = whatsNew + "2026/02/amazon-bedrock-adds-support-six-open-weights-models"
		minimaxGLMLaunch   = whatsNew + "2026/03/amazon-bedrock-minimax-glm/"
		mistralLarge3Batch = whatsNew + "2025/12/mistral-large-3-ministral-3-family-available-amazon-bedrock"
	)

	announced := map[string]struct{ date, source string }{
		"us.amazon.nova-2-lite-v1:0": {"2025-12-02", whatsNew + "2025/12/nova-2-foundation-models-amazon-bedrock/"},
		"us.amazon.nova-pro-v1:0":    {"2024-12-03", novaLaunch},
		"us.amazon.nova-lite-v1:0":   {"2024-12-03", novaLaunch},
		"us.amazon.nova-micro-v1:0":  {"2024-12-03", novaLaunch},

		"us.anthropic.claude-opus-5":                   {"2026-07-24", whatsNew + "2026/07/claude-opus-5-aws/"},
		"us.anthropic.claude-fable-5-1":                {"2026-09-01", whatsNew + "2026/09/claude-fable-5-1-aws/"},
		"us.anthropic.claude-fable-5":                  {"2026-06-09", whatsNew + "2026/06/claude-fable-5-aws/"},
		"us.anthropic.claude-opus-4-8":                 {"2026-05-28", whatsNew + "2026/05/claude-opus-4.8-aws/"},
		"us.anthropic.claude-opus-4-7":                 {"2026-04-16", whatsNew + "2026/04/claude-opus-4.7-amazon-bedrock/"},
		"us.anthropic.claude-sonnet-5":                 {"2026-06-30", whatsNew + "2026/06/claude-sonnet-5-now-available-on-aws"},
		"us.anthropic.claude-opus-4-6-v1":              {"2026-02-05", whatsNew + "2026/2/claude-opus-4.6-available-amazon-bedrock/"},
		"us.anthropic.claude-sonnet-4-6":               {"2026-02-17", whatsNew + "2026/02/claude-sonnet-4.6-available-in-amazon-bedrock/"},
		"us.anthropic.claude-opus-4-5-20251101-v1:0":   {"2025-11-24", whatsNew + "2025/11/claude-opus-4-5-amazon-bedrock"},
		"us.anthropic.claude-haiku-4-5-20251001-v1:0":  {"2025-10-15", whatsNew + "2025/10/claude-4-5-haiku-anthropic-amazon-bedrock"},
		"us.anthropic.claude-sonnet-4-5-20250929-v1:0": {"2025-09-29", whatsNew + "2025/09/anthropics-claude-sonnet-4-5-amazon-bedrock"},

		"us.meta.llama4-maverick-17b-instruct-v1:0": {"2025-04-28", docHistory},
		"us.meta.llama4-scout-17b-instruct-v1:0":    {"2025-04-28", docHistory},
		"us.meta.llama3-3-70b-instruct-v1:0":        {"2024-12-19", whatsNew + "2024/12/metas-llama-3-3-70b-model-amazon-bedrock/"},
		"us.meta.llama3-1-70b-instruct-v1:0":        {"2024-07-23", llama31Launch},
		"us.meta.llama3-1-8b-instruct-v1:0":         {"2024-07-23", llama31Launch},
		"meta.llama3-70b-instruct-v1:0":             {"2024-04-23", llama3Launch},
		"meta.llama3-8b-instruct-v1:0":              {"2024-04-23", llama3Launch},

		"deepseek.v3.2":       {"2026-02-10", openWeights6},
		"us.deepseek.r1-v1:0": {"2025-03-10", whatsNew + "2025/03/deepseek-r1-fully-managed-amazon-bedrock"},

		"openai.gpt-oss-safeguard-120b": {"2025-12-02", openWeights18},
		"openai.gpt-oss-safeguard-20b":  {"2025-12-02", openWeights18},
		"openai.gpt-oss-120b-1:0":       {"2025-08-05", docHistory},
		"openai.gpt-oss-20b-1:0":        {"2025-08-05", docHistory},

		"qwen.qwen3-next-80b-a3b":       {"2025-12-02", openWeights18},
		"qwen.qwen3-vl-235b-a22b":       {"2025-12-02", openWeights18},
		"qwen.qwen3-32b-v1:0":           {"2025-09-18", qwen3Launch},
		"qwen.qwen3-coder-30b-a3b-v1:0": {"2025-09-18", qwen3Launch},
		"qwen.qwen3-coder-next":         {"2026-02-10", openWeights6},

		"mistral.mistral-large-3-675b-instruct": {"2025-12-02", mistralLarge3Batch},
		"mistral.devstral-2-123b":               {"2026-02-18", whatsNew + "2026/02/mistral-ai-devstral-bedrock/"},
		"mistral.magistral-small-2509":          {"2025-12-02", mistralLarge3Batch},
		"mistral.mistral-large-2402-v1:0":       {"2024-04-03", whatsNew + "2024/04/mistral-large-foundation-model-amazon-bedrock/"},

		"moonshotai.kimi-k2.5":      {"2026-02-10", openWeights6},
		"moonshot.kimi-k2-thinking": {"2025-12-02", openWeights18},

		"zai.glm-4.7":       {"2026-02-10", openWeights6},
		"zai.glm-4.7-flash": {"2026-02-10", openWeights6},
		"zai.glm-5":         {"2026-03-18", minimaxGLMLaunch},

		"minimax.minimax-m2.5": {"2026-03-18", minimaxGLMLaunch},
		"minimax.minimax-m2.1": {"2026-02-10", openWeights6},
		"minimax.minimax-m2":   {"2025-12-02", openWeights18},

		"nvidia.nemotron-nano-3-30b":   {"2025-12-23", whatsNew + "2025/12/nvidia-nemotron-3-nano-amazon-bedrock/"},
		"nvidia.nemotron-super-3-120b": {"2026-03-18", whatsNew + "2026/03/amazon-bedrock-nemotron-3-super/"},
	}

	models, err := DefaultModels()
	require.NoError(t, err, "cannot read the bundled bedrock catalogue")

	seen := map[string]bool{}
	for _, m := range models {
		if m.Price == nil {
			t.Errorf("%s carries no price", m.Name)
		} else if m.Price.Input <= 0 || m.Price.Output <= 0 {
			t.Errorf("%s prices must be positive, got input %f and output %f", m.Name, m.Price.Input, m.Price.Output)
		}

		want, ok := announced[m.Name]
		if !ok {
			t.Errorf("%s has no entry here; add the day Bedrock began serving it and the AWS record that says so", m.Name)
			continue
		}
		seen[m.Name] = true

		if m.ReleaseDate == nil {
			t.Errorf("%s carries no release_date; Bedrock began serving it on %s (%s)", m.Name, want.date, want.source)
			continue
		}
		if got := m.ReleaseDate.Format("2006-01-02"); got != want.date {
			t.Errorf("%s release_date = %s, but Bedrock began serving it on %s (%s)", m.Name, got, want.date, want.source)
		}
	}

	for name := range announced {
		if !seen[name] {
			t.Errorf("%s is no longer in the bedrock catalogue; drop it from this table", name)
		}
	}
}

func TestBedrock_GetUsage_ReadsTheTokenCountsBedrockReports(t *testing.T) {
	t.Parallel()

	prov := &bedrockProvider{}

	assert.Equal(t, pconfig.CallUsage{Input: 100, Output: 50},
		prov.GetUsage(map[string]any{"PromptTokens": int32(100), "CompletionTokens": int32(50)}))
	usage := prov.GetUsage(map[string]any{})
	assert.True(t, usage.IsZero(), "missing usage info must read as zero, got %s", usage.String())
}

// bedrockProbe is a Converse endpoint that answers every request and keeps the last one's headers and body.
type bedrockProbe struct {
	url    string
	calls  int
	header http.Header
	body   map[string]any
}

func bedrockEndpoint(t *testing.T) *bedrockProbe {
	t.Helper()

	p := &bedrockProbe{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		p.calls++
		p.header, p.body = r.Header.Clone(), nil
		_ = json.Unmarshal(raw, &p.body)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":{"message":{"role":"assistant","content":[{"text":"ok"}]}},` +
			`"stopReason":"end_turn","usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`))
	}))
	t.Cleanup(srv.Close)
	p.url = srv.URL

	return p
}

// door builds a provider on the credentials in cfg that reaches the probe.
func (p *bedrockProbe) door(t *testing.T, cfg config.Config, providerConfig *pconfig.ProviderConfig) provider.Provider {
	t.Helper()

	cfg.BedrockRegion, cfg.BedrockServerURL, cfg.ExternalSSLInsecure = "us-east-1", p.url, true
	prov, err := New(&cfg, provider.DefaultProviderNameBedrock, providerConfig)
	require.NoError(t, err)

	return prov
}

func TestBedrock_New_SignsWithTheCredentialItIsGiven(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIDENV")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "env-secret")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "")
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(t.TempDir(), "config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "credentials"))

	providerConfig, err := DefaultProviderConfig(&config.Config{})
	require.NoError(t, err)

	tests := []struct {
		name         string
		cfg          config.Config
		signer       string
		sessionToken string
	}{
		{
			name:   "static credentials sign the request",
			cfg:    config.Config{BedrockAccessKey: "test-access-key", BedrockSecretKey: "test-secret-key"},
			signer: "AWS4-HMAC-SHA256 Credential=test-access-key",
		},
		{
			name: "a session token travels with static credentials",
			cfg: config.Config{
				BedrockAccessKey: "test-access-key", BedrockSecretKey: "test-secret-key", BedrockSessionToken: "test-session-token",
			},
			signer: "AWS4-HMAC-SHA256 Credential=test-access-key", sessionToken: "test-session-token",
		},
		{
			name:   "a bearer token needs no AWS credentials",
			cfg:    config.Config{BedrockBearerToken: "test-bearer-token-value"},
			signer: "Bearer test-bearer-token-value",
		},
		{
			name: "a bearer token wins over static credentials",
			cfg: config.Config{
				BedrockBearerToken: "bearer-token", BedrockAccessKey: "access-key", BedrockSecretKey: "secret-key",
			},
			signer: "Bearer bearer-token",
		},
		{
			name: "default auth wins over every configured credential",
			cfg: config.Config{
				BedrockDefaultAuth: true, BedrockBearerToken: "bearer-token",
				BedrockAccessKey: "access-key", BedrockSecretKey: "secret-key",
			},
			signer: "AWS4-HMAC-SHA256 Credential=AKIDENV",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			probe := bedrockEndpoint(t)
			prov := probe.door(t, tc.cfg, providerConfig)

			_, err := prov.CallEx(context.Background(), pconfig.OptionsTypeSimple,
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil)
			require.NoError(t, err)

			signer, _, _ := strings.Cut(probe.header.Get("Authorization"), "/")
			assert.Equal(t, tc.signer, signer)
			assert.Equal(t, tc.sessionToken, probe.header.Get("X-Amz-Security-Token"))
		})
	}
}

func TestBedrock_New_RefusesWithoutACompleteCredential(t *testing.T) {
	t.Parallel()

	providerConfig, err := DefaultProviderConfig(&config.Config{})
	require.NoError(t, err)

	tests := []struct {
		name string
		cfg  config.Config
	}{
		{"no credential at all", config.Config{}},
		{"an access key without its secret", config.Config{BedrockAccessKey: "test-key"}},
		{"a secret without its access key", config.Config{BedrockSecretKey: "test-secret"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := tc.cfg
			cfg.BedrockRegion = "us-east-1"

			_, err := New(&cfg, provider.DefaultProviderNameBedrock, providerConfig)
			assert.EqualError(t, err, "no valid authentication method configured for Bedrock")
		})
	}
}

// TestBedrock_EveryCallPathSendsTheAgentsThinkingAndTheChainsTools is keyed by calling method.
func TestBedrock_EveryCallPathSendsTheAgentsThinkingAndTheChainsTools(t *testing.T) {
	t.Parallel()

	providerConfig, err := BuildProviderConfig([]byte(
		"simple:\n  model: us.anthropic.claude-opus-5\n  reasoning:\n    mode: adaptive\n    effort: low\n"))
	require.NoError(t, err)

	ctx := context.Background()
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "hi"),
		{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.ToolCall{
			ID: "c1", Type: "function", FunctionCall: &llms.FunctionCall{Name: "search", Arguments: `{"query":"x"}`},
		}}},
		{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{llms.ToolCallResponse{
			ToolCallID: "c1", Name: "search", Content: "found",
		}}},
	}
	tools := []llms.Tool{createToolWithSchema("noop", "draft/2020-12")}

	paths := []struct {
		name  string
		call  func(prov provider.Provider) error
		tools []string
	}{
		{"a single prompt", func(prov provider.Provider) error {
			_, err := prov.Call(ctx, pconfig.OptionsTypeSimple, "hi")
			return err
		}, nil},
		{"a chain", func(prov provider.Provider) error {
			_, err := prov.CallEx(ctx, pconfig.OptionsTypeSimple, chain, nil)
			return err
		}, []string{"search"}},
		{"a chain with tools", func(prov provider.Provider) error {
			_, err := prov.CallWithTools(ctx, pconfig.OptionsTypeSimple, chain, tools, nil)
			return err
		}, []string{"noop", "search"}},
		{"a chain with extra options", func(prov provider.Provider) error {
			_, err := prov.CallWithExtraOptions(ctx, pconfig.OptionsTypeSimple, chain, tools, nil)
			return err
		}, []string{"noop", "search"}},
	}

	for _, path := range paths {
		t.Run(path.name, func(t *testing.T) {
			t.Parallel()

			probe := bedrockEndpoint(t)
			require.NoError(t, path.call(probe.door(t, config.Config{BedrockBearerToken: "t"}, providerConfig)))
			require.Equal(t, 1, probe.calls)

			fields, _ := probe.body["additionalModelRequestFields"].(map[string]any)
			thinking, _ := fields["thinking"].(map[string]any)
			assert.Equal(t, "adaptive", thinking["type"], "body=%v", probe.body)

			toolConfig, _ := probe.body["toolConfig"].(map[string]any)
			specs, _ := toolConfig["tools"].([]any)
			var names []string
			for _, spec := range specs {
				toolSpec, _ := spec.(map[string]any)["toolSpec"].(map[string]any)
				names = append(names, fmt.Sprint(toolSpec["name"]))
				inputSchema, _ := toolSpec["inputSchema"].(map[string]any)
				assert.NotContains(t, inputSchema["json"], "$schema", "tool %v", toolSpec["name"])
			}
			assert.Equal(t, path.tools, names, "the Converse API refuses a chain with tool use but no toolConfig")
		})
	}
}

func TestBedrock_InferPropertyType_ClassifiesTheTopLevelJSONType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"nil value", nil, "null"},
		{"string", "hello", "string"},
		{"boolean", true, "boolean"},
		{"integer", 42, "number"},
		{"float", 3.14159, "number"},
		{"slice", []int{1, 2, 3}, "array"},
		{"map", map[string]string{"key": "value"}, "object"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, inferPropertyType(tc.value))
		})
	}
}

func bedrockSchema(types map[string]string) map[string]any {
	properties := make(map[string]any, len(types))
	for key, typ := range types {
		properties[key] = map[string]any{"type": typ}
	}

	return map[string]any{"type": "object", "properties": properties}
}

func TestBedrock_InferSchemaFromArguments_AggregatesThePropertiesOfEverySample(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		samples []string
		want    map[string]string
	}{
		{name: "no samples"},
		{
			name:    "one sample of every type",
			samples: []string{`{"name":"test","count":5,"active":true,"tags":["a","b"],"meta":{}}`},
			want:    map[string]string{"name": "string", "count": "number", "active": "boolean", "tags": "array", "meta": "object"},
		},
		{
			name:    "samples add up",
			samples: []string{`{"field1":"value1"}`, `{"field2":42}`, `{"field3":true}`},
			want:    map[string]string{"field1": "string", "field2": "number", "field3": "boolean"},
		},
		{
			name:    "invalid JSON is skipped",
			samples: []string{`{invalid json}`, `{"valid":"field"}`, `not json at all`},
			want:    map[string]string{"valid": "string"},
		},
		{name: "empty samples are skipped", samples: []string{"", "", `{"key":"value"}`}, want: map[string]string{"key": "string"}},
		{
			name:    "the first sample of a key decides its type",
			samples: []string{`{"field":"string_value"}`, `{"field":123}`},
			want:    map[string]string{"field": "string"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, bedrockSchema(tc.want), inferSchemaFromArguments(tc.samples))
		})
	}
}

func TestBedrock_CollectToolUsageFromChain_RecordsEveryToolTheChainUsed(t *testing.T) {
	t.Parallel()

	call := func(id, name, arguments string) llms.MessageContent {
		return llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.ToolCall{
			ID: id, Type: "function", FunctionCall: &llms.FunctionCall{Name: name, Arguments: arguments},
		}}}
	}
	response := func(id, name string) llms.MessageContent {
		return llms.MessageContent{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{llms.ToolCallResponse{
			ToolCallID: id, Name: name, Content: "done",
		}}}
	}

	tests := []struct {
		name  string
		chain []llms.MessageContent
		want  map[string][]string
	}{
		{name: "an empty chain", want: map[string][]string{}},
		{
			name:  "a tool call",
			chain: []llms.MessageContent{call("c1", "search", `{"query":"test"}`)},
			want:  map[string][]string{"search": {`{"query":"test"}`}},
		},
		{
			name:  "a tool response",
			chain: []llms.MessageContent{response("c1", "execute")},
			want:  map[string][]string{"execute": {}},
		},
		{
			name: "every call to one tool is a sample",
			chain: []llms.MessageContent{
				call("c1", "calc", `{"op":"add","a":1,"b":2}`), call("c2", "calc", `{"op":"multiply","a":3,"b":4}`),
			},
			want: map[string][]string{"calc": {`{"op":"add","a":1,"b":2}`, `{"op":"multiply","a":3,"b":4}`}},
		},
		{
			name:  "a response keeps the samples of its call",
			chain: []llms.MessageContent{call("c1", "tool1", `{}`), response("c1", "tool1"), response("c2", "tool2")},
			want:  map[string][]string{"tool1": {`{}`}, "tool2": {}},
		},
		{
			name: "a tool call without a function",
			chain: []llms.MessageContent{{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{
				llms.ToolCall{ID: "c1", Type: "function"},
			}}},
			want: map[string][]string{},
		},
		{
			name: "text only",
			chain: []llms.MessageContent{
				llms.TextParts(llms.ChatMessageTypeHuman, "hello"), llms.TextParts(llms.ChatMessageTypeAI, "hi"),
			},
			want: map[string][]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, collectToolUsageFromChain(tc.chain))
		})
	}
}

func TestBedrock_RestoreMissedToolsFromChain_DeclaresOnlyTheToolsNobodyDeclared(t *testing.T) {
	t.Parallel()

	calls := func(names ...string) []llms.MessageContent {
		parts := make([]llms.ContentPart, 0, len(names))
		for idx, name := range names {
			parts = append(parts, llms.ToolCall{
				ID: fmt.Sprintf("c%d", idx), Type: "function", FunctionCall: &llms.FunctionCall{Name: name, Arguments: `{}`},
			})
		}
		return []llms.MessageContent{{Role: llms.ChatMessageTypeAI, Parts: parts}}
	}
	restored := func(name string, types map[string]string) llms.Tool {
		return llms.Tool{Type: "function", Function: &llms.FunctionDefinition{
			Name: name, Description: "Tool: " + name, Parameters: bedrockSchema(types),
		}}
	}
	declared := func(name string) llms.Tool {
		return llms.Tool{Type: "function", Function: &llms.FunctionDefinition{Name: name}}
	}
	search := llms.Tool{Type: "function", Function: &llms.FunctionDefinition{
		Name: "search", Description: "Custom search description",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"query": map[string]any{"type": "string"}, "custom": map[string]any{"type": "boolean"},
		}},
	}}
	searchCall := []llms.MessageContent{{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.ToolCall{
		ID: "c1", Type: "function", FunctionCall: &llms.FunctionCall{Name: "search", Arguments: `{"query":"test"}`},
	}}}}
	scanCall := []llms.MessageContent{{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.ToolCall{
		ID: "c1", Type: "function", FunctionCall: &llms.FunctionCall{Name: "scan_tool", Arguments: `{"target":"10.0.0.1"}`},
	}}}}

	tests := []struct {
		name     string
		chain    []llms.MessageContent
		declared []llms.Tool
		want     []llms.Tool
	}{
		{
			name:     "a chain without tool use returns the declared tools",
			chain:    []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Hello")},
			declared: []llms.Tool{declared("tool1")},
			want:     []llms.Tool{declared("tool1")},
		},
		{
			name:  "an undeclared tool is restored with a schema inferred from its arguments",
			chain: scanCall,
			want:  []llms.Tool{restored("scan_tool", map[string]string{"target": "string"})},
		},
		{
			name:     "an empty declared list is restored like a missing one",
			chain:    scanCall,
			declared: []llms.Tool{},
			want:     []llms.Tool{restored("scan_tool", map[string]string{"target": "string"})},
		},
		{
			name:     "a declared tool is kept as declared",
			chain:    searchCall,
			declared: []llms.Tool{search},
			want:     []llms.Tool{search},
		},
		{
			name:     "declared and restored tools are merged",
			chain:    calls("tool_b", "tool_c"),
			declared: []llms.Tool{declared("tool_a"), declared("tool_b")},
			want:     []llms.Tool{declared("tool_a"), declared("tool_b"), restored("tool_c", nil)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, restoreMissedToolsFromChain(tc.chain, tc.declared))
		})
	}
}

func TestBedrock_ExtractToolsFromOptions_ReturnsTheToolsTheOptionsCarry(t *testing.T) {
	t.Parallel()

	tools := []llms.Tool{
		{Type: "function", Function: &llms.FunctionDefinition{Name: "tool1"}},
		{Type: "function", Function: &llms.FunctionDefinition{Name: "tool2"}},
	}

	assert.Nil(t, extractToolsFromOptions(nil))
	assert.Equal(t, tools, extractToolsFromOptions([]llms.CallOption{
		llms.WithModel("test-model"), llms.WithTemperature(0.7), llms.WithTools(tools), llms.WithMaxTokens(100),
	}))
}

func TestBedrock_CleanToolSchemas_DropsTheSchemaKeyWithoutTouchingTheCaller(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tools func() []llms.Tool
		want  []llms.Tool
	}{
		{name: "no tools", tools: func() []llms.Tool { return nil }},
		{
			name:  "a map schema loses its $schema key",
			tools: func() []llms.Tool { return []llms.Tool{createToolWithSchema("test_tool", "draft/2020-12")} },
			want:  []llms.Tool{createToolWithoutSchema("test_tool")},
		},
		{
			name:  "a map schema without the key is kept",
			tools: func() []llms.Tool { return []llms.Tool{createToolWithoutSchema("clean_tool")} },
			want:  []llms.Tool{createToolWithoutSchema("clean_tool")},
		},
		{
			name: "every tool of several is cleaned",
			tools: func() []llms.Tool {
				return []llms.Tool{
					createToolWithSchema("tool1", "draft/2020-12"),
					createToolWithoutSchema("tool2"),
					createToolWithSchema("tool3", "draft-07"),
				}
			},
			want: []llms.Tool{createToolWithoutSchema("tool1"), createToolWithoutSchema("tool2"), createToolWithoutSchema("tool3")},
		},
		{
			name:  "a tool without a function is kept",
			tools: func() []llms.Tool { return []llms.Tool{{Type: "function"}} },
			want:  []llms.Tool{{Type: "function"}},
		},
		{
			name: "a function without parameters is kept",
			tools: func() []llms.Tool {
				return []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{Name: "no_params"}}}
			},
			want: []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{Name: "no_params"}}},
		},
		{
			name: "parameters of another shape are kept",
			tools: func() []llms.Tool {
				return []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{
					Name: "string_params", Parameters: "not a map",
				}}}
			},
			want: []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{
				Name: "string_params", Parameters: "not a map",
			}}},
		},
		{
			name:  "a reflected schema becomes a map without the key",
			tools: func() []llms.Tool { return []llms.Tool{createToolWithJsonSchemaType("json_schema_tool")} },
			want: []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{
				Name: "json_schema_tool",
				Parameters: map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []any{"arg"},
					"properties": map[string]any{
						"arg": map[string]any{"type": "string", "description": "Test argument"},
					},
				},
			}}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := tc.tools()
			assert.Equal(t, tc.want, cleanToolSchemas(input))
			assert.Equal(t, tc.tools(), input, "cleaning must not touch the caller's tools")
		})
	}
}

func createToolWithSchema(name, schemaVersion string) llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name: name,
			Parameters: map[string]any{
				"$schema": fmt.Sprintf("https://json-schema.org/%s/schema", schemaVersion),
				"type":    "object",
				"properties": map[string]any{
					"arg": map[string]any{"type": "string"},
				},
			},
		},
	}
}

func createToolWithoutSchema(name string) llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name: name,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"arg": map[string]any{"type": "string"},
				},
			},
		},
	}
}

func createToolWithJsonSchemaType(name string) llms.Tool {
	type TestStruct struct {
		Arg string `json:"arg" jsonschema:"required,description=Test argument"`
	}

	reflector := &jsonschema.Reflector{
		DoNotReference: true,
		ExpandedStruct: true,
	}

	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:       name,
			Parameters: reflector.Reflect(&TestStruct{}),
		},
	}
}
