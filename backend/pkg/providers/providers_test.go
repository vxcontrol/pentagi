package providers

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/bedrock"
	"pentagi/pkg/providers/deepseek"
	"pentagi/pkg/providers/gemini"
	"pentagi/pkg/providers/glm"
	"pentagi/pkg/providers/kimi"
	"pentagi/pkg/providers/minimax"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/qwen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubTypedProvider reports only its type — enough for Providers.ListTypes(),
// which is all GetProviders touches before it rejects an unavailable type.
type stubTypedProvider struct {
	provider.Provider
	ptype provider.ProviderType
}

func (s stubTypedProvider) Type() provider.ProviderType { return s.ptype }

type stubProvidersQuerier struct {
	database.Querier
	rows []database.Provider
}

func (s stubProvidersQuerier) GetUserProviders(context.Context, int64) ([]database.Provider, error) {
	return s.rows, nil
}

func TestGetProviders_SkipsUserProviderOfDisabledType(t *testing.T) {
	pc := &providerController{
		cfg: &config.Config{},
		db: stubProvidersQuerier{rows: []database.Provider{
			{Name: "stale-minimax", Type: "minimax"}, // not in ListTypes() below
		}},
		Providers: provider.Providers{
			"openai-default": stubTypedProvider{ptype: provider.ProviderOpenAI},
		},
	}

	got, err := pc.GetProviders(context.Background(), 1)

	require.NoError(t, err, "one user provider of a disabled type must not fail the whole query")
	assert.Contains(t, got, provider.ProviderName("openai-default"), "enabled providers stay reachable")
	assert.NotContains(t, got, provider.ProviderName("stale-minimax"), "the unavailable provider is skipped")
}

func TestGetProviders_StaleUserRowSpansValidSibling(t *testing.T) {
	pc := &providerController{
		cfg: &config.Config{},
		db: stubProvidersQuerier{rows: []database.Provider{
			// ollama.New builds with no API key and makes no network call (pull/load
			// off under the empty config), so this row clears the real NewProvider.
			{Name: "ollama-user", Type: "ollama"},
			{Name: "stale-minimax", Type: "minimax"}, // type absent from ListTypes()
		}},
		Providers: provider.Providers{
			"ollama-default": stubTypedProvider{ptype: provider.ProviderOllama},
		},
	}

	got, err := pc.GetProviders(context.Background(), 1)

	require.NoError(t, err, "a stale sibling row must not fail the whole query")
	assert.Contains(t, got, provider.ProviderName("ollama-user"), "a valid user provider survives a stale sibling")
	assert.NotContains(t, got, provider.ProviderName("stale-minimax"), "the unbuildable sibling is skipped")
}

func TestBuildDefaultConfigs_DisabledProviderBadPathIsNotFatal(t *testing.T) {
	cfg := &config.Config{
		BedrockConfig: filepath.Join(t.TempDir(), "missing.yml"),
	}

	configs, skipReasons, err := buildDefaultConfigs(cfg)

	require.NoError(t, err)
	_, hasBedrock := configs[provider.ProviderBedrock]
	assert.False(t, hasBedrock, "disabled provider with an unreadable config path is skipped")
	assert.Error(t, skipReasons[provider.ProviderBedrock], "the skip reason must be recorded, not just swallowed")
	_, hasOpenAI := configs[provider.ProviderOpenAI]
	assert.True(t, hasOpenAI, "providers with a readable (embedded) default config still load")
}

func TestBuildDefaultConfigs_EnabledProviderBadPathIsFatal(t *testing.T) {
	cfg := &config.Config{
		BedrockConfig:      filepath.Join(t.TempDir(), "missing.yml"),
		BedrockBearerToken: "token",
	}

	_, _, err := buildDefaultConfigs(cfg)

	require.Error(t, err)
}

// A provider type whose default config failed to load at startup (e.g. an
// unreadable BEDROCK_CONFIG_PATH while Bedrock is otherwise disabled) must
// report that specific cause from patchProviderConfig, not the generic "not
// found" message that reads identically to "this type does not exist".
func TestPatchProviderConfig_SurfacesSkipReasonInsteadOfAmbiguousNotFound(t *testing.T) {
	cfg := &config.Config{
		BedrockConfig: filepath.Join(t.TempDir(), "missing.yml"),
	}
	configs, skipReasons, err := buildDefaultConfigs(cfg)
	require.NoError(t, err)

	pc := &providerController{
		cfg:                 cfg,
		defaultConfigs:      configs,
		defaultConfigErrors: skipReasons,
	}

	_, err = pc.patchProviderConfig(provider.ProviderBedrock, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load at startup", "must surface the real cause, not just 'not found'")
}

func TestOpenAICompatProvidersDoNotUseAdaptiveThinking(t *testing.T) {
	t.Parallel()

	providers := []struct {
		name   string
		config func() (*pconfig.ProviderConfig, error)
		models func() (pconfig.ModelsConfig, error)
	}{
		{"qwen", qwen.DefaultProviderConfig, qwen.DefaultModels},
		{"glm", glm.DefaultProviderConfig, glm.DefaultModels},
		{"deepseek", deepseek.DefaultProviderConfig, deepseek.DefaultModels},
		{"kimi", kimi.DefaultProviderConfig, kimi.DefaultModels},
		{"minimax", minimax.DefaultProviderConfig, minimax.DefaultModels},
	}

	for _, p := range providers {
		cfg, err := p.config()
		if err != nil {
			t.Fatalf("%s: DefaultProviderConfig() error: %v", p.name, err)
		}
		models, err := p.models()
		if err != nil {
			t.Fatalf("%s: DefaultModels() error: %v", p.name, err)
		}

		for _, opt := range pconfig.AllAgentTypes {
			if cfg.UsesAdaptiveThinking(models, opt) {
				t.Errorf("%s/%s resolves to adaptive thinking, but the openaicompat transport never emits it (M7): adaptive thinking is Anthropic/Bedrock-only", p.name, opt)
			}
		}

		for _, m := range models {
			if m.Reasoning == nil {
				continue
			}
			if mode := m.Reasoning.Mode; mode == pconfig.ModelReasoningAdaptive || mode == pconfig.ModelReasoningAdaptiveOnly {
				t.Errorf("%s catalog model %q declares reasoning mode %q, but the openaicompat transport never emits adaptive thinking (M7): adaptive thinking is Anthropic/Bedrock-only", p.name, m.Name, mode)
			}
		}
	}
}

// The default agent config (config.yml) carries its own per-agent price, and
// GetPriceInfoForType returns it verbatim (no catalog fallback), so a drift from
// the model catalog silently mis-prices cost telemetry for the default config.
func TestAgentConfigPricesMatchCatalog(t *testing.T) {
	t.Parallel()

	providers := []struct {
		name   string
		config func() (*pconfig.ProviderConfig, error)
		models func() (pconfig.ModelsConfig, error)
	}{
		{"anthropic", anthropic.DefaultProviderConfig, anthropic.DefaultModels},
		{"bedrock",
			func() (*pconfig.ProviderConfig, error) { return bedrock.DefaultProviderConfig(&config.Config{}) },
			func() (pconfig.ModelsConfig, error) { return bedrock.DefaultModels() }},
		{"gemini", gemini.DefaultProviderConfig, gemini.DefaultModels},
		{"deepseek", deepseek.DefaultProviderConfig, deepseek.DefaultModels},
		{"glm", glm.DefaultProviderConfig, glm.DefaultModels},
		{"kimi", kimi.DefaultProviderConfig, kimi.DefaultModels},
		{"minimax", minimax.DefaultProviderConfig, minimax.DefaultModels},
		{"openai", openai.DefaultProviderConfig, openai.DefaultModels},
		{"qwen", qwen.DefaultProviderConfig, qwen.DefaultModels},
	}

	for _, p := range providers {
		cfg, err := p.config()
		if err != nil {
			t.Fatalf("%s: DefaultProviderConfig() error: %v", p.name, err)
		}
		models, err := p.models()
		if err != nil {
			t.Fatalf("%s: DefaultModels() error: %v", p.name, err)
		}

		catalog := make(map[string]*pconfig.PriceInfo, len(models))
		for i := range models {
			catalog[models[i].Name] = models[i].Price
		}

		for _, opt := range pconfig.AllAgentTypes {
			ac := cfg.AgentConfigForType(opt)
			if ac == nil || ac.Model == "" || ac.Price == nil {
				continue
			}
			cat, ok := catalog[ac.Model]
			if !ok || cat == nil {
				continue
			}
			if *ac.Price != *cat {
				t.Errorf("%s/%s model %q: config.yml price %+v != catalog price %+v", p.name, opt, ac.Model, *ac.Price, *cat)
			}
		}
	}
}

// fakeCallProvider overrides only Call(); embedding provider.Provider means
// every other interface method is unimplemented and would panic on use,
// which is fine since callWithSetupRetries only ever calls Call().
type fakeCallProvider struct {
	provider.Provider
	callCount int
	failTimes int
	err       error
	result    string
}

func (f *fakeCallProvider) Call(ctx context.Context, opt pconfig.ProviderOptionsType, prompt string) (string, error) {
	f.callCount++
	if f.callCount <= f.failTimes {
		return "", f.err
	}
	return f.result, nil
}

func TestCallWithSetupRetries_SucceedsImmediately(t *testing.T) {
	prv := &fakeCallProvider{result: "kali-linux"}

	got, err := callWithSetupRetries(context.Background(), prv, pconfig.OptionsTypeSimple, "prompt")

	require.NoError(t, err)
	assert.Equal(t, "kali-linux", got)
	assert.Equal(t, 1, prv.callCount, "a first-try success must not retry")
}

func TestCallWithSetupRetries_RetriesOnTransientErrorThenSucceeds(t *testing.T) {
	// Exercises the real backoff once (a single delayBetweenRetries wait), so this
	// test genuinely proves a transient 5xx from the LLM gateway self-heals
	// instead of failing flow/assistant creation outright.
	prv := &fakeCallProvider{
		failTimes: 1,
		err:       fmt.Errorf("API returned unexpected status code: 502: bad gateway"),
		result:    "kali-linux",
	}

	got, err := callWithSetupRetries(context.Background(), prv, pconfig.OptionsTypeSimple, "prompt")

	require.NoError(t, err)
	assert.Equal(t, "kali-linux", got)
	assert.Equal(t, 2, prv.callCount, "must retry exactly once after the transient failure")
}

func TestCallWithSetupRetries_ContextCanceledDuringWait_ReturnsWithoutFullBackoff(t *testing.T) {
	prv := &fakeCallProvider{
		failTimes: maxRetriesToCallSimpleChain, // always fails within the retry budget
		err:       errors.New("connection refused"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := callWithSetupRetries(ctx, prv, pconfig.OptionsTypeSimple, "prompt")
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, delayBetweenRetries, "canceling context mid-wait must abort immediately, not wait out the full backoff")
}

func TestCallWithSetupRetries_ContextAlreadyCanceled_StopsWithoutRetrying(t *testing.T) {
	prv := &fakeCallProvider{
		failTimes: maxRetriesToCallSimpleChain,
		err:       context.Canceled,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := callWithSetupRetries(ctx, prv, pconfig.OptionsTypeSimple, "prompt")

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, prv.callCount, "a context.Canceled error from Call must stop retrying immediately")
}
