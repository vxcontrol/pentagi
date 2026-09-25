package providers

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database/converter"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/providers/deepseek"
	"pentagi/pkg/providers/glm"
	"pentagi/pkg/providers/kimi"
	"pentagi/pkg/providers/minimax"
	"pentagi/pkg/providers/mistral"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/qwen"
	"pentagi/pkg/providers/xai"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestRegistry_ProviderRegistry_MatchesAllProviderTypes(t *testing.T) {
	registryTypes := make(map[provider.ProviderType]struct{}, len(providerRegistry))
	for _, e := range providerRegistry {
		if _, dup := registryTypes[e.Type]; dup {
			t.Errorf("duplicate providerRegistry entry for %q", e.Type)
		}
		registryTypes[e.Type] = struct{}{}
		assert.Equal(t, provider.ProviderName(e.Type), e.Name,
			"registry Name for %q must be the canonical provider name", e.Type)
	}

	allTypes := make(map[provider.ProviderType]struct{}, len(provider.AllProviderTypes))
	for _, pt := range provider.AllProviderTypes {
		allTypes[pt] = struct{}{}
	}

	for pt := range allTypes {
		_, ok := registryTypes[pt]
		assert.Truef(t, ok, "%q is in AllProviderTypes but missing from providerRegistry", pt)
	}
	for pt := range registryTypes {
		_, ok := allTypes[pt]
		assert.Truef(t, ok, "%q is in providerRegistry but missing from AllProviderTypes", pt)
	}
}

func TestRegistry_EnabledDefaultProviderTypes_ReportsOnlyTypesWhoseGateIsSatisfied(t *testing.T) {
	t.Run("nothing configured", func(t *testing.T) {
		assert.Empty(t, EnabledDefaultProviderTypes(&config.Config{}))
	})

	t.Run("only credentialed types are reported", func(t *testing.T) {
		cfg := &config.Config{
			OpenAIKey:       "key",
			AnthropicAPIKey: "key",
			KimiAPIKey:      "key",
		}

		types := EnabledDefaultProviderTypes(cfg)

		assert.ElementsMatch(t, []string{
			string(provider.ProviderOpenAI),
			string(provider.ProviderAnthropic),
			string(provider.ProviderKimi),
		}, types)
	})

	t.Run("custom needs a URL, a key, and a model or config path", func(t *testing.T) {
		assert.Empty(t, EnabledDefaultProviderTypes(&config.Config{LLMServerURL: "http://llm", LLMServerKey: "key"}))
		assert.Empty(t, EnabledDefaultProviderTypes(&config.Config{
			LLMServerURL:   "http://llm",
			LLMServerModel: "some-model",
		}))

		types := EnabledDefaultProviderTypes(&config.Config{
			LLMServerURL:   "http://llm",
			LLMServerKey:   "key",
			LLMServerModel: "some-model",
		})
		assert.Equal(t, []string{string(provider.ProviderCustom)}, types)
	})

	t.Run("without a key anthropic needs a complete federation config", func(t *testing.T) {
		types := EnabledDefaultProviderTypes(&config.Config{
			AnthropicFederationRuleID:  "fdrl_1",
			AnthropicOrganizationID:    "org",
			AnthropicServiceAccountID:  "svac_1",
			AnthropicIdentityTokenFile: "/var/run/secrets/anthropic.com/token",
		})
		assert.Equal(t, []string{string(provider.ProviderAnthropic)}, types)

		assert.Empty(t, EnabledDefaultProviderTypes(&config.Config{AnthropicFederationRuleID: "fdrl_1"}))
	})
}

func TestRegistry_ProviderRegistry_CustomEntryHandsTheDoorAnEnrichedCatalogue(t *testing.T) {
	const name = "claude-sonnet-5"

	srv, _ := gatewayServing(t, name)
	cfg := &config.Config{LLMServerKey: "k", LLMServerURL: srv.URL, LLMServerModel: name}

	entry, ok := entryForType(provider.ProviderCustom)
	if !ok {
		t.Fatal("no custom entry in the provider registry")
	}

	providerConfig, err := entry.NewConfig(cfg)
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	prov, err := entry.New(cfg, entry.Name, providerConfig)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}

	models := prov.GetModels()
	if len(models) != 1 {
		t.Fatalf("got %d models, want 1", len(models))
	}
	if models[0].Reasoning == nil {
		t.Fatal("the registry handed the transport a catalogue with no reasoning metadata")
	}
	if models[0].Reasoning.Mode != pconfig.ModelReasoningAdaptiveOnly {
		t.Errorf("mode = %q, want the bundled adaptive-only", models[0].Reasoning.Mode)
	}
}

func TestRegistry_ProviderRegistry_CustomEntryKeepsSharedSettingsOnTheWireForAReasoningOnlyBlock(t *testing.T) {
	const name = "pentagi-shared-settings-model"

	entry, ok := entryForType(provider.ProviderCustom)
	require.True(t, ok, "no custom entry in the provider registry")

	for _, tc := range []struct {
		label string
		block map[string]any
	}{
		{"reasoning turned off", map[string]any{"reasoning": map[string]any{"mode": "off"}}},
		{"a reasoning effort", map[string]any{"reasoning": map[string]any{"mode": "effort", "effort": "low"}}},
		{"a reasoning budget", map[string]any{"reasoning": map[string]any{"mode": "budget", "max_tokens": 2048}}},
	} {
		t.Run(tc.label, func(t *testing.T) {
			prov, body := customDoor(t, name, tc.block, entry)

			_, err := prov.CallEx(context.Background(), pconfig.OptionsTypeSimple, hiChain(), nil)
			if err != nil {
				t.Fatalf("call: %v", err)
			}

			for key, want := range map[string]any{
				"model":                 name,
				"temperature":           1.0,
				"top_p":                 1.0,
				"n":                     1.0,
				"max_completion_tokens": 16384.0,
			} {
				if got := (*body)[key]; got != want {
					t.Errorf("%s on the wire = %v, want %v; body=%v", key, got, want, *body)
				}
			}
		})
	}
}

// The card is saved through the converter and rebuilt through the registry, as GetProviders rebuilds it.
func TestRegistry_ProviderRegistry_AFieldClearedOnTheAgentCardFallsBackToTheDoorDefault(t *testing.T) {
	resolve := func(t *testing.T, ptype provider.ProviderType, card *model.AgentConfig) llms.CallOptions {
		t.Helper()

		entry, ok := entryForType(ptype)
		require.True(t, ok, "no registry entry for %s", ptype)

		raw, err := json.Marshal(converter.ConvertAgentsConfigFromGqlModel(&model.AgentsConfig{Simple: card}))
		require.NoError(t, err)

		pc, err := entry.BuildConfig(&config.Config{}, raw)
		require.NoError(t, err)

		var resolved llms.CallOptions
		for _, option := range pc.GetOptionsForType(pconfig.OptionsTypeSimple) {
			option(&resolved)
		}
		return resolved
	}

	temperature, maxTokens := 0.3, 2048

	t.Run("a value set on the card goes to the wire", func(t *testing.T) {
		got := resolve(t, provider.ProviderAnthropic,
			&model.AgentConfig{Model: "claude-sonnet-5", Temperature: &temperature, MaxTokens: &maxTokens})

		require.NotNil(t, got.Temperature)
		require.NotNil(t, got.MaxTokens)
		assert.Equal(t, temperature, *got.Temperature)
		assert.Equal(t, maxTokens, *got.MaxTokens)
	})

	t.Run("a cleared value takes the door default", func(t *testing.T) {
		got := resolve(t, provider.ProviderAnthropic, &model.AgentConfig{Model: "claude-sonnet-5"})

		require.NotNil(t, got.Temperature)
		require.NotNil(t, got.MaxTokens)
		assert.Equal(t, 1.0, *got.Temperature)
		assert.Equal(t, 4000, *got.MaxTokens)
		assert.Equal(t, "claude-sonnet-5", got.GetModel())
	})

	t.Run("a cleared value the door has no default for stays off the wire", func(t *testing.T) {
		got := resolve(t, provider.ProviderDeepSeek, &model.AgentConfig{Model: "deepseek-v4-flash"})

		assert.Nil(t, got.Temperature)
		require.NotNil(t, got.MaxTokens)
		assert.Equal(t, 4000, *got.MaxTokens)
	})
}

// registryDoorSource is one provider subpackage's non-test Go files.
type registryDoorSource struct {
	pkg   string
	files []*ast.File
}

func registryParseDoorSources(t *testing.T) (*token.FileSet, []registryDoorSource) {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read provider packages: %v", err)
	}

	fset := token.NewFileSet()
	var sources []registryDoorSource
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		files, err := os.ReadDir(entry.Name())
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}

		source := registryDoorSource{pkg: entry.Name()}
		for _, file := range files {
			name := file.Name()
			if file.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			parsed, err := parser.ParseFile(fset, filepath.Join(entry.Name(), name), nil, 0)
			if err != nil {
				t.Fatalf("parse %s/%s: %v", entry.Name(), name, err)
			}
			source.files = append(source.files, parsed)
		}
		sources = append(sources, source)
	}

	return fset, sources
}

type doorFacts struct {
	isDoor          bool
	takesClientArg  bool
	usesSharedNew   bool
	buildsOwnClient []string
}

func inspectDoor(fset *token.FileSet, files []*ast.File) doorFacts {
	var facts doorFacts
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Name.Name != "New" || node.Recv != nil {
					return true
				}
				for _, param := range node.Type.Params.List {
					switch typ := param.Type.(type) {
					case *ast.SelectorExpr:
						if typ.Sel.Name == "ProviderName" {
							facts.isDoor = true
						}
					case *ast.StarExpr:
						if sel, ok := typ.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "Client" {
							if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "http" {
								facts.takesClientArg = true
							}
						}
					}
				}
			case *ast.CallExpr:
				sel, ok := node.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if ident.Name == "system" && sel.Sel.Name == "GetHTTPClient" {
					facts.usesSharedNew = true
				}
			case *ast.CompositeLit:
				sel, ok := node.Type.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if ok && ident.Name == "http" && sel.Sel.Name == "Client" {
					facts.buildsOwnClient = append(facts.buildsOwnClient, fset.Position(node.Pos()).String())
				}
			}
			return true
		})
	}

	return facts
}

func adaptiveCallPaths(files []*ast.File) (found int, missing []string) {
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !strings.HasPrefix(fn.Name.Name, "Call") {
				continue
			}
			found++
			wired := false
			ast.Inspect(fn, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok &&
					strings.HasPrefix(sel.Sel.Name, "WrapGenerate") {
					wired = true
				}
				return true
			})
			if !wired {
				missing = append(missing, fn.Name.Name)
			}
		}
	}

	return found, missing
}

func TestRegistry_DoorPackagesUseTheSharedPlumbing(t *testing.T) {
	fset, sources := registryParseDoorSources(t)

	t.Run("every door takes the shared http client", func(t *testing.T) {
		doors := 0
		for _, source := range sources {
			facts := inspectDoor(fset, source.files)
			if !facts.isDoor {
				continue
			}
			doors++

			if len(facts.buildsOwnClient) > 0 {
				t.Errorf("door %q builds its own http.Client at %s: the configured timeout, CA and insecure flag never reach it",
					source.pkg, strings.Join(facts.buildsOwnClient, ", "))
			}

			if !facts.usesSharedNew && !facts.takesClientArg {
				t.Errorf("door %q neither calls system.GetHTTPClient nor accepts one from its caller", source.pkg)
			}
		}

		if doors < 10 {
			t.Fatalf("found only %d provider doors, so this check stopped covering them", doors)
		}
	})

	t.Run("every call method reaches the shared generate wrapper", func(t *testing.T) {
		checked := 0
		for _, source := range sources {
			// The provider package holds the wrapper itself, and its failover only delegates.
			if source.pkg == "provider" {
				continue
			}
			found, missing := adaptiveCallPaths(source.files)
			if found == 0 {
				continue
			}
			checked++
			if len(missing) > 0 {
				t.Errorf("%s: %v do not reach provider.Wrap*, so the adaptive option never joins those calls",
					source.pkg, missing)
			}
		}
		if checked == 0 {
			t.Error("no provider package declares a Call* method, so this gate guards nothing")
		}
	})
}

var entryPoints = []struct {
	name string
	call func(provider.Provider) error
}{
	{"Call", func(p provider.Provider) error {
		_, err := p.Call(context.Background(), pconfig.OptionsTypeSimple, "hi")
		return err
	}},
	{"CallEx", func(p provider.Provider) error {
		_, err := p.CallEx(context.Background(), pconfig.OptionsTypeSimple, hiChain(), nil)
		return err
	}},
	{"CallWithTools", func(p provider.Provider) error {
		_, err := p.CallWithTools(context.Background(), pconfig.OptionsTypeSimple, hiChain(), nil, nil)
		return err
	}},
	{"CallWithExtraOptions", func(p provider.Provider) error {
		_, err := p.CallWithExtraOptions(context.Background(), pconfig.OptionsTypeSimple, hiChain(), nil, nil)
		return err
	}},
}

func hiChain() []llms.MessageContent {
	return []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}
}

func TestRegistry_ProviderRegistry_OpenAICompatDoorsSendTheOutputCeilingUnderTheVendorsName(t *testing.T) {
	doors := []struct {
		ptype provider.ProviderType
		field string
		model string
		point func(cfg *config.Config, url string)
	}{
		{provider.ProviderDeepSeek, "max_tokens", deepseek.DeepSeekAgentModel,
			func(c *config.Config, u string) { c.DeepSeekAPIKey, c.DeepSeekServerURL = "k", u }},
		{provider.ProviderQwen, "max_tokens", qwen.QwenAgentModel,
			func(c *config.Config, u string) { c.QwenAPIKey, c.QwenServerURL = "k", u }},
		{provider.ProviderXAI, "max_tokens", xai.XAIAgentModel,
			func(c *config.Config, u string) { c.XAIAPIKey, c.XAIServerURL = "k", u }},
		{provider.ProviderMistral, "max_tokens", mistral.MistralAgentModel,
			func(c *config.Config, u string) { c.MistralAPIKey, c.MistralServerURL = "k", u }},
		{provider.ProviderGLM, "max_tokens", glm.GLMAgentModel,
			func(c *config.Config, u string) { c.GLMAPIKey, c.GLMServerURL = "k", u }},
		{provider.ProviderKimi, "max_completion_tokens", kimi.KimiAgentModel,
			func(c *config.Config, u string) { c.KimiAPIKey, c.KimiServerURL = "k", u }},
		{provider.ProviderMiniMax, "max_completion_tokens", minimax.MiniMaxAgentModel,
			func(c *config.Config, u string) { c.MiniMaxAPIKey, c.MiniMaxServerURL = "k", u }},
	}

	for _, door := range doors {
		t.Run("the "+door.ptype.String()+" door", func(t *testing.T) {
			other := map[string]string{
				"max_tokens":            "max_completion_tokens",
				"max_completion_tokens": "max_tokens",
			}[door.field]

			entry, ok := entryForType(door.ptype)
			require.True(t, ok, "no registry entry for %s", door.ptype)

			shipped, err := entry.NewConfig(&config.Config{})
			require.NoError(t, err)
			require.NotZero(t, shipped.Simple.MaxTokens)

			cleared, err := entry.BuildConfig(&config.Config{}, []byte("simple:\n  model: "+door.model+"\n"))
			require.NoError(t, err)

			for _, tc := range []struct {
				label string
				pc    *pconfig.ProviderConfig
				want  float64
			}{
				{"the shipped block", shipped, float64(shipped.Simple.MaxTokens)},
				{"a block with max tokens cleared", cleared, 4000},
			} {
				for _, call := range entryPoints {
					t.Run(tc.label+" through "+call.name, func(t *testing.T) {
						srv, body := gatewayServing(t, door.model)
						cfg := &config.Config{}
						door.point(cfg, srv.URL)

						prov, err := entry.New(cfg, provider.ProviderName(door.ptype), tc.pc)
						require.NoError(t, err)

						require.NoError(t, call.call(prov))

						assert.Equal(t, tc.want, (*body)[door.field], "body=%v", *body)
						assert.NotContains(t, *body, other)
						assert.NotContains(t, *body, "metadata")
					})
				}
			}
		})
	}
}
