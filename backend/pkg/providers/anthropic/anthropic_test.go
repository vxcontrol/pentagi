package anthropic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

const anthropicContextWindowsPage = "https://platform.claude.com/docs/en/build-with-claude/context-windows"

func TestAnthropic_New_BuildsTheDoorWithAPricedModelForEveryAgent(t *testing.T) {
	cfg := &config.Config{
		AnthropicAPIKey:    "test-key",
		AnthropicServerURL: "https://api.anthropic.com",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, provider.DefaultProviderNameAnthropic, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if prov.Type() != provider.ProviderAnthropic {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderAnthropic, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameAnthropic {
		t.Errorf("Expected provider name %v, got %v", provider.DefaultProviderNameAnthropic, prov.Name())
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

func TestAnthropic_DefaultModels_ShipsANonEmptyPricedCatalogue(t *testing.T) {
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

func TestAnthropic_DefaultModels_GivesEachModelTheWindowServedWithoutABetaHeader(t *testing.T) {
	millionTokens := map[string]bool{
		"claude-fable-5-1":  true,
		"claude-fable-5":    true,
		"claude-opus-5-5":   true,
		"claude-opus-5":     true,
		"claude-opus-4-8":   true,
		"claude-opus-4-7":   true,
		"claude-opus-4-6":   true,
		"claude-sonnet-5":   true,
		"claude-sonnet-4-6": true,
	}

	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("cannot read the bundled anthropic catalogue: %v", err)
	}

	for _, m := range models {
		want := 200000
		if millionTokens[m.Name] {
			want = 1000000
		}

		if m.ContextWindow == nil {
			t.Errorf("%s carries no context_window; %s gives it %d", m.Name, anthropicContextWindowsPage, want)
			continue
		}
		if *m.ContextWindow != want {
			t.Errorf("%s context_window = %d, but %s gives it %d", m.Name, *m.ContextWindow, anthropicContextWindowsPage, want)
		}
	}
}

func anthropicDoor(t *testing.T, upstream *anthropicUpstream, configData string) provider.Provider {
	t.Helper()

	providerConfig, err := BuildProviderConfig([]byte(configData))
	require.NoError(t, err)

	prov, err := New(&config.Config{AnthropicAPIKey: "k", AnthropicServerURL: upstream.url},
		provider.DefaultProviderNameAnthropic, providerConfig)
	require.NoError(t, err)

	return prov
}

// anthropicCallEx sends one CallEx for the simple agent and returns the one request body it made.
func anthropicCallEx(t *testing.T, configData string, chain []llms.MessageContent) map[string]any {
	t.Helper()

	upstream := anthropicServe(t)
	_, err := anthropicDoor(t, upstream, configData).CallEx(context.Background(), pconfig.OptionsTypeSimple, chain, nil)
	require.NoError(t, err)

	bodies := upstream.take()
	require.Len(t, bodies, 1)
	return bodies[0]
}

// Subtests are keyed by the call method they drive.
func TestAnthropic_SendsAnAdaptiveAgentsThinkingOnEveryCallPath(t *testing.T) {
	upstream := anthropicServe(t)
	prov := anthropicDoor(t, upstream, "simple:\n  model: claude-sonnet-5\n  reasoning:\n    mode: adaptive\n    effort: low\n")

	ctx := context.Background()
	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}
	tools := []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{
		Name: "noop", Description: "does nothing",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{}},
	}}}

	for _, path := range []struct {
		name string
		call func() error
	}{
		{"a single prompt", func() error {
			_, err := prov.Call(ctx, pconfig.OptionsTypeSimple, "hi")
			return err
		}},
		{"a chain", func() error {
			_, err := prov.CallEx(ctx, pconfig.OptionsTypeSimple, chain, nil)
			return err
		}},
		{"a chain with tools", func() error {
			_, err := prov.CallWithTools(ctx, pconfig.OptionsTypeSimple, chain, tools, nil)
			return err
		}},
		{"a chain with tools through the extra-options path", func() error {
			_, err := prov.CallWithExtraOptions(ctx, pconfig.OptionsTypeSimple, chain, tools, nil)
			return err
		}},
	} {
		t.Run(path.name, func(t *testing.T) {
			require.NoError(t, path.call())
			bodies := upstream.take()
			require.Len(t, bodies, 1)

			thinking, _ := bodies[0]["thinking"].(map[string]any)
			assert.Equal(t, "adaptive", thinking["type"], "body=%v", bodies[0])
		})
	}
}

func TestAnthropic_CallEx_PutsReplayedThinkingFirstWithItsSignature(t *testing.T) {
	body := anthropicCallEx(t, "simple:\n  model: claude-haiku-4-5\n", []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "hi"),
		{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextPartWithReasoning("Half an answer",
			&reasoning.ContentReasoning{Content: "counted on my fingers", Signature: []byte("sig-from-an-earlier-turn")})}},
		llms.TextParts(llms.ChatMessageTypeHuman, "go on"),
	})

	messages, _ := body["messages"].([]any)
	var assistant map[string]any
	for _, raw := range messages {
		if msg, isMap := raw.(map[string]any); isMap && msg["role"] == "assistant" {
			assistant = msg
			break
		}
	}
	require.NotNil(t, assistant, "the assistant turn never reached the wire; body=%v", body)

	blocks, _ := assistant["content"].([]any)
	require.NotEmpty(t, blocks, "assistant turn carries no content blocks: %v", assistant)
	first, _ := blocks[0].(map[string]any)
	assert.Equal(t, "thinking", first["type"], "the API rejects a turn whose thinking does not lead it: %v", blocks)
	assert.Equal(t, "counted on my fingers", first["thinking"])
	assert.Equal(t, "sig-from-an-earlier-turn", first["signature"], "without its signature the vendor rejects the turn")
}

func TestAnthropic_CallEx_KeepsTheThinkingBudgetUnderTheOutputCap(t *testing.T) {
	for _, tc := range []struct {
		name        string
		configured  string
		asked       float64
		wantAsIs    bool
		wantClamped bool
	}{
		{name: "a budget under an explicit cap", configured: "  max_tokens: 2048\n", asked: 1024, wantAsIs: true},
		{name: "a budget over an explicit cap", configured: "  max_tokens: 2048\n", asked: 4096, wantClamped: true},
		{name: "no cap configured at all", asked: 4096},
		{name: "a budget larger than any model takes", asked: 32000, wantClamped: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := anthropicCallEx(t, fmt.Sprintf(
				"simple:\n  model: claude-sonnet-4-5\n%s  reasoning:\n    mode: budget\n    max_tokens: %d\n",
				tc.configured, int(tc.asked)), []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

			thinking, _ := body["thinking"].(map[string]any)
			require.Equal(t, "enabled", thinking["type"], "body=%v", body)
			budget, _ := thinking["budget_tokens"].(float64)
			limit, hasLimit := body["max_tokens"].(float64)
			require.True(t, hasLimit, "no output cap in the request; body=%v", body)

			assert.Positive(t, budget, "a zero budget asks for no thinking at all")
			assert.Less(t, budget, limit, "the API rejects a budget that is not below the output cap")
			if tc.configured != "" {
				assert.GreaterOrEqual(t, limit, 2048.0, "the output cap fell below the one the operator configured")
			}
			if tc.wantAsIs {
				assert.Equal(t, tc.asked, budget, "a budget that fits under the cap goes untouched")
			}
			if tc.wantClamped {
				assert.Less(t, budget, tc.asked, "a budget that does not fit must be cut")
			}
		})
	}
}

func TestAnthropic_New_AuthenticatesWithTheConfiguredCredential(t *testing.T) {
	exchange := func(assertion string) map[string]any {
		return map[string]any{
			"grant_type":         "urn:ietf:params:oauth:grant-type:jwt-bearer",
			"assertion":          assertion,
			"federation_rule_id": "fdrl_1",
			"organization_id":    "00000000-0000-0000-0000-000000000000",
			"service_account_id": "svac_1",
			"workspace_id":       "wrkspc_1",
		}
	}
	minted := anthropicAuth{authorization: "Bearer minted-1"}

	tests := []struct {
		name          string
		adjust        func(cfg *config.Config, dir string)
		calls         int
		wantErr       string // {dir} stands for the directory holding the token files
		wantExchanges []map[string]any
		wantAuth      []anthropicAuth
	}{
		{
			name:          "a literal identity token is exchanged once and the minted token serves every call",
			calls:         2,
			wantExchanges: []map[string]any{exchange("literal.jwt.token")},
			wantAuth:      []anthropicAuth{minted, minted},
		},
		{
			name: "the token file wins over the literal token",
			adjust: func(cfg *config.Config, dir string) {
				cfg.AnthropicIdentityTokenFile = filepath.Join(dir, "token")
			},
			wantExchanges: []map[string]any{exchange("projected.jwt.token")},
			wantAuth:      []anthropicAuth{minted},
		},
		{
			name:   "no workspace leaves it out of the exchange",
			adjust: func(cfg *config.Config, _ string) { cfg.AnthropicWorkspaceID = "" },
			wantExchanges: []map[string]any{{
				"grant_type":         "urn:ietf:params:oauth:grant-type:jwt-bearer",
				"assertion":          "literal.jwt.token",
				"federation_rule_id": "fdrl_1",
				"organization_id":    "00000000-0000-0000-0000-000000000000",
				"service_account_id": "svac_1",
			}},
			wantAuth: []anthropicAuth{minted},
		},
		{
			name:     "a set API key shadows federation",
			adjust:   func(cfg *config.Config, _ string) { cfg.AnthropicAPIKey = "sk-ant-key" },
			wantAuth: []anthropicAuth{{apiKey: "sk-ant-key"}},
		},
		{
			name: "a partial federation config is refused before any request",
			adjust: func(cfg *config.Config, _ string) {
				cfg.AnthropicServiceAccountID, cfg.AnthropicIdentityToken = "", ""
			},
			wantErr: "anthropic: set ANTHROPIC_API_KEY or complete the federation config " +
				"(missing ANTHROPIC_SERVICE_ACCOUNT_ID, ANTHROPIC_IDENTITY_TOKEN_FILE or ANTHROPIC_IDENTITY_TOKEN)",
		},
		{
			name: "a token file that is not there fails the call and names it",
			adjust: func(cfg *config.Config, dir string) {
				cfg.AnthropicIdentityTokenFile = filepath.Join(dir, "absent")
			},
			wantErr: "anthropic federation: read ANTHROPIC_IDENTITY_TOKEN_FILE: open {dir}/absent",
		},
		{
			name: "an empty token file fails the call before any exchange",
			adjust: func(cfg *config.Config, dir string) {
				cfg.AnthropicIdentityTokenFile = filepath.Join(dir, "empty")
			},
			wantErr: "anthropic federation: ANTHROPIC_IDENTITY_TOKEN_FILE {dir}/empty holds no token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			upstream := anthropicServe(t)
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("projected.jwt.token\n"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "empty"), []byte("\n"), 0o600))

			cfg := federatedConfig(upstream.url)
			if tc.adjust != nil {
				tc.adjust(cfg, dir)
			}

			providerConfig, err := DefaultProviderConfig()
			require.NoError(t, err)

			prov, err := New(cfg, provider.DefaultProviderNameAnthropic, providerConfig)
			for call := 0; err == nil && call < max(tc.calls, 1); call++ {
				_, err = prov.Call(context.Background(), pconfig.OptionsTypeSimple, "hi")
			}
			if tc.wantErr != "" {
				require.ErrorContains(t, err, strings.ReplaceAll(tc.wantErr, "{dir}", dir))
			} else {
				require.NoError(t, err)
			}

			exchanges, auth := upstream.authentication()
			assert.Equal(t, tc.wantExchanges, exchanges)
			assert.Equal(t, tc.wantAuth, auth)
		})
	}
}

func TestAnthropic_Configured_NeedsAKeyOrEveryRequiredFederationVariable(t *testing.T) {
	absent := filepath.Join(t.TempDir(), "absent")

	tests := []struct {
		name   string
		adjust func(cfg *config.Config)
		want   bool
	}{
		{name: "a complete federation config", want: true},
		{
			name:   "an API key alone",
			adjust: func(cfg *config.Config) { *cfg = config.Config{AnthropicAPIKey: "sk-ant-key"} },
			want:   true,
		},
		{name: "nothing at all", adjust: func(cfg *config.Config) { *cfg = config.Config{} }},
		{name: "no workspace", adjust: func(cfg *config.Config) { cfg.AnthropicWorkspaceID = "" }, want: true},
		{
			name: "a token file that does not exist yet",
			adjust: func(cfg *config.Config) {
				cfg.AnthropicIdentityToken, cfg.AnthropicIdentityTokenFile = "", absent
			},
			want: true,
		},
		{name: "no federation rule", adjust: func(cfg *config.Config) { cfg.AnthropicFederationRuleID = "" }},
		{name: "no organization", adjust: func(cfg *config.Config) { cfg.AnthropicOrganizationID = "" }},
		{name: "no service account", adjust: func(cfg *config.Config) { cfg.AnthropicServiceAccountID = "" }},
		{
			name:   "no identity token of either kind",
			adjust: func(cfg *config.Config) { cfg.AnthropicIdentityToken = "" },
		},
		{
			name:   "a literal token of whitespace only",
			adjust: func(cfg *config.Config) { cfg.AnthropicIdentityToken = " \n" },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := federatedConfig("https://api.anthropic.com/v1")
			if tc.adjust != nil {
				tc.adjust(cfg)
			}

			assert.Equal(t, tc.want, Configured(cfg))
		})
	}
}

func TestAnthropic_CredentialsNotice_ExplainsAFederationConfigThatWillNotWork(t *testing.T) {
	const (
		key      = "sk-ant-key"
		shadowed = "anthropic: ANTHROPIC_API_KEY is set, so the federation variables are ignored; unset the key to federate"
		rotate   = "; point ANTHROPIC_IDENTITY_TOKEN_FILE at a token your platform rotates to run longer than its lifetime"
		keptOn   = "; the provider stays enabled and reads the file again on every exchange"
	)

	stubClock(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("projected.jwt.token\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "empty"), []byte("\n"), 0o600))
	federated := func(adjust func(cfg *config.Config)) config.Config {
		cfg := federatedConfig("https://api.anthropic.com/v1")
		adjust(cfg)
		return *cfg
	}
	literal := func(token string) config.Config {
		return federated(func(cfg *config.Config) { cfg.AnthropicIdentityToken = token })
	}
	thirteen := time.Date(2026, 9, 24, 13, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		cfg  config.Config
		want string // {dir} stands for the directory holding the token files
	}{
		{name: "a key alone", cfg: config.Config{AnthropicAPIKey: key}},
		{
			name: "a readable token file, which wins over the literal token",
			cfg:  federated(func(cfg *config.Config) { cfg.AnthropicIdentityTokenFile = filepath.Join(dir, "token") }),
		},
		{
			name: "a token file that cannot be read yet",
			cfg:  federated(func(cfg *config.Config) { cfg.AnthropicIdentityTokenFile = filepath.Join(dir, "absent") }),
			want: "anthropic: read ANTHROPIC_IDENTITY_TOKEN_FILE: open {dir}/absent: no such file or directory" + keptOn,
		},
		{
			name: "an empty token file",
			cfg:  federated(func(cfg *config.Config) { cfg.AnthropicIdentityTokenFile = filepath.Join(dir, "empty") }),
			want: "anthropic: ANTHROPIC_IDENTITY_TOKEN_FILE {dir}/empty holds no token" + keptOn,
		},
		{
			name: "a literal token that is not a JWT",
			cfg:  literal("literal.jwt.token"),
			want: "anthropic: ANTHROPIC_IDENTITY_TOKEN cannot be refreshed once it expires" + rotate,
		},
		{
			name: "a literal JWT names its expiry",
			cfg:  literal(identityJWT(thirteen, "")),
			want: "anthropic: ANTHROPIC_IDENTITY_TOKEN cannot be refreshed once it expires at 2026-09-24T13:00:00Z" + rotate,
		},
		{
			name: "a single-use literal JWT says it is exchanged once",
			cfg:  literal(identityJWT(thirteen, "jti-1")),
			want: "anthropic: ANTHROPIC_IDENTITY_TOKEN cannot be refreshed once it expires at 2026-09-24T13:00:00Z; " +
				"it carries a jti claim, so it can be exchanged only once" + rotate,
		},
		{
			name: "an expired literal JWT",
			cfg:  literal(identityJWT(time.Date(2026, 9, 24, 11, 0, 0, 0, time.UTC), "")),
			want: "anthropic: ANTHROPIC_IDENTITY_TOKEN expired at 2026-09-24T11:00:00Z and will not be exchanged" + rotate,
		},
		{
			name: "a key beside a complete federation config",
			cfg: func() config.Config {
				cfg := *federatedConfig("https://api.anthropic.com/v1")
				cfg.AnthropicAPIKey = key
				return cfg
			}(),
			want: shadowed,
		},
		{
			name: "a key beside a rule",
			cfg:  config.Config{AnthropicAPIKey: key, AnthropicFederationRuleID: "fdrl_1"},
			want: shadowed,
		},
		{
			name: "a key beside an organization",
			cfg:  config.Config{AnthropicAPIKey: key, AnthropicOrganizationID: "00000000-0000-0000-0000-000000000000"},
			want: shadowed,
		},
		{
			name: "a key beside a service account",
			cfg:  config.Config{AnthropicAPIKey: key, AnthropicServiceAccountID: "svac_1"},
			want: shadowed,
		},
		{
			name: "a key beside a token file",
			cfg:  config.Config{AnthropicAPIKey: key, AnthropicIdentityTokenFile: "/token"},
			want: shadowed,
		},
		{
			name: "a key beside a literal token",
			cfg:  config.Config{AnthropicAPIKey: key, AnthropicIdentityToken: "literal.jwt.token"},
			want: shadowed,
		},
		{
			name: "a partial config names only what it lacks",
			cfg: config.Config{
				AnthropicFederationRuleID: "fdrl_1",
				AnthropicOrganizationID:   "00000000-0000-0000-0000-000000000000",
				AnthropicWorkspaceID:      "wrkspc_1",
			},
			want: "anthropic provider disabled: federation is missing " +
				"ANTHROPIC_SERVICE_ACCOUNT_ID, ANTHROPIC_IDENTITY_TOKEN_FILE or ANTHROPIC_IDENTITY_TOKEN",
		},
		{
			name: "a lone token file names the three ids",
			cfg:  config.Config{AnthropicIdentityTokenFile: "/token"},
			want: "anthropic provider disabled: federation is missing " +
				"ANTHROPIC_FEDERATION_RULE_ID, ANTHROPIC_ORGANIZATION_ID, ANTHROPIC_SERVICE_ACCOUNT_ID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, strings.ReplaceAll(tc.want, "{dir}", dir), CredentialsNotice(&tc.cfg))
		})
	}
}
