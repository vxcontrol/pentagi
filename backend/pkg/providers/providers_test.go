package providers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

type stubProvidersQuerier struct {
	database.Querier
	rows []database.Provider
}

func (s stubProvidersQuerier) GetUserProviders(context.Context, int64) ([]database.Provider, error) {
	return s.rows, nil
}

func TestProviders_GetProviders_SkipsAnUnusableUserRowAndKeepsTheRest(t *testing.T) {
	pc := &providerController{
		cfg: &config.Config{},
		db: stubProvidersQuerier{rows: []database.Provider{
			// ollama.New needs no key and makes no call, so this row clears the real NewProvider.
			{Name: "ollama-user", Type: "ollama"},
			{Name: "stale-minimax", Type: "minimax"}, // type absent from ListTypes()
			{Name: "unreadable-ollama", Type: "ollama", Config: []byte("{not a config")},
		}},
		Providers: provider.Providers{
			"ollama-default": stubTypedProvider{ptype: provider.ProviderOllama},
		},
	}

	got, err := pc.GetProviders(context.Background(), 1)

	require.NoError(t, err, "one unusable user provider must not fail the whole query")
	assert.Contains(t, got, provider.ProviderName("ollama-default"), "enabled providers stay reachable")
	assert.Contains(t, got, provider.ProviderName("ollama-user"), "a valid user provider survives an unusable sibling")
	assert.NotContains(t, got, provider.ProviderName("stale-minimax"), "the provider of an unavailable type is skipped")
	assert.NotContains(t, got, provider.ProviderName("unreadable-ollama"),
		"the provider whose stored config no longer parses is skipped")
}

func TestProviders_BuildDefaultConfigs_ToleratesABadConfigPathOnlyForADisabledType(t *testing.T) {
	for _, tc := range []struct {
		name    string
		token   string
		refused string
	}{
		{name: "a disabled type is skipped with its reason kept"},
		{name: "an enabled type stops startup", token: "token", refused: "failed to create bedrock provider config"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				BedrockConfig:      filepath.Join(t.TempDir(), "missing.yml"),
				BedrockBearerToken: tc.token,
			}

			configs, skipReasons, err := buildDefaultConfigs(cfg)

			if tc.refused != "" {
				require.ErrorContains(t, err, tc.refused)
				return
			}
			require.NoError(t, err)
			_, hasBedrock := configs[provider.ProviderBedrock]
			assert.False(t, hasBedrock, "disabled provider with an unreadable config path is skipped")
			assert.ErrorContains(t, skipReasons[provider.ProviderBedrock], "missing.yml",
				"the skip reason must be recorded, not just swallowed")
			_, hasOpenAI := configs[provider.ProviderOpenAI]
			assert.True(t, hasOpenAI, "providers with a readable (embedded) default config still load")
		})
	}
}

func TestProviders_CallWithSetupRetries_AsksAgainOnlyWhileTheCallMayStillPass(t *testing.T) {
	background := func(*testing.T) context.Context { return context.Background() }

	for _, tc := range []struct {
		name      string
		prv       fakeCallProvider
		ctx       func(t *testing.T) context.Context
		want      string
		wantErr   error
		wantCalls int
		// within bounds how long the call may take; zero leaves it unchecked.
		within time.Duration
		why    string
	}{
		{
			name: "the first answer is returned",
			prv:  fakeCallProvider{result: "kali-linux"}, ctx: background,
			want: "kali-linux", wantCalls: 1,
			why: "a first-try success must not retry",
		},
		{
			// Waits out the real backoff once.
			name: "a transient gateway error is asked again",
			prv: fakeCallProvider{
				failTimes: 1,
				err:       fmt.Errorf("API returned unexpected status code: 502: bad gateway"),
				result:    "kali-linux",
			},
			ctx:  background,
			want: "kali-linux", wantCalls: 2,
			why: "must retry exactly once after the transient failure",
		},
		{
			name: "a context cancelled during the backoff returns at once",
			prv:  fakeCallProvider{failTimes: maxRetriesToCallSimpleChain, err: errors.New("connection refused")},
			ctx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				time.AfterFunc(20*time.Millisecond, cancel)
				return ctx
			},
			wantErr: context.Canceled, wantCalls: 1, within: 2 * time.Second,
			why: "canceling the context mid-wait must abort at once, not wait out the 5s backoff",
		},
		{
			name: "a cancelled call is not asked again",
			prv:  fakeCallProvider{failTimes: maxRetriesToCallSimpleChain, err: context.Canceled},
			ctx: func(*testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			wantErr: context.Canceled, wantCalls: 1,
			why: "a context.Canceled error from Call must stop retrying immediately",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prv := tc.prv
			ctx := tc.ctx(t)

			start := time.Now()
			got, err := callWithSetupRetries(ctx, &prv, pconfig.OptionsTypeSimple, "prompt")
			elapsed := time.Since(start)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantCalls, prv.callCount, tc.why)
			if tc.within > 0 {
				assert.Less(t, elapsed, tc.within, tc.why)
			}
		})
	}
}

// providersStoredQuerier reads any row back as ptype, the type UpdateProvider validates, and fails writes on a done context as the real one does.
type providersStoredQuerier struct {
	recordingProvidersQuerier
	ptype provider.ProviderType
}

func (q *providersStoredQuerier) GetUserProvider(
	_ context.Context, arg database.GetUserProviderParams,
) (database.Provider, error) {
	return database.Provider{ID: arg.ID, UserID: arg.UserID, Name: "stored", Type: database.ProviderType(q.ptype)}, nil
}

func (q *providersStoredQuerier) CreateProvider(
	ctx context.Context, params database.CreateProviderParams,
) (database.Provider, error) {
	if err := ctx.Err(); err != nil {
		return database.Provider{}, err
	}
	return q.recordingProvidersQuerier.CreateProvider(ctx, params)
}

func (q *providersStoredQuerier) UpdateUserProvider(
	ctx context.Context, params database.UpdateUserProviderParams,
) (database.Provider, error) {
	if err := ctx.Err(); err != nil {
		return database.Provider{}, err
	}
	return q.recordingProvidersQuerier.UpdateUserProvider(ctx, params)
}

// providersSaving builds the controller a save goes through, offering only ollama when unavailable is set.
func providersSaving(
	t *testing.T, cfg *config.Config, ptype provider.ProviderType, unavailable bool,
) (*providerController, *providersStoredQuerier) {
	t.Helper()

	defaultConfigs, skipReasons, err := buildDefaultConfigs(cfg)
	require.NoError(t, err)

	offered := ptype
	if unavailable {
		offered = provider.ProviderOllama
	}

	db := &providersStoredQuerier{ptype: ptype}

	return &providerController{
		cfg:                 cfg,
		db:                  db,
		defaultConfigs:      defaultConfigs,
		defaultConfigErrors: skipReasons,
		Providers:           provider.Providers{"default": stubTypedProvider{ptype: offered}},
	}, db
}

// providersSave saves through CreateProvider, or UpdateProvider when update is set, and returns the rows written.
func providersSave(
	pc *providerController, db *providersStoredQuerier, update bool,
	prvname provider.ProviderName, cfg *pconfig.ProviderConfig,
) (int, error) {
	if update {
		_, err := pc.UpdateProvider(context.Background(), 1, 7, prvname, cfg)
		return db.updated, err
	}

	_, err := pc.CreateProvider(context.Background(), 1, prvname, db.ptype, cfg)
	return db.created, err
}

// providersRequireStored wants exactly one row written, or, when refused is set, that error and none written.
func providersRequireStored(t *testing.T, written int, err error, refused string) {
	t.Helper()

	if refused != "" {
		require.ErrorContains(t, err, refused)
		require.Zero(t, written, "a refused config still reached the database")
		return
	}
	require.NoError(t, err)
	require.Equal(t, 1, written)
}

func TestProviders_ValidateProviderName_StoresTheTrimmedNameAndRefusesAnUnusableOne(t *testing.T) {
	for _, tc := range []struct {
		name    string
		prvname provider.ProviderName
		stored  string
		refused string
	}{
		{name: "an ordinary name", prvname: "my-ollama", stored: "my-ollama"},
		{name: "a padded name is stored trimmed", prvname: "  my-ollama  ", stored: "my-ollama"},
		{name: "an empty name", prvname: "", refused: "provider name is required"},
		{name: "a name of whitespace only", prvname: " \t\n ", refused: "provider name is required"},
		{
			name:    "a name at the length limit",
			prvname: provider.ProviderName(strings.Repeat("a", maxProviderNameLen)),
			stored:  strings.Repeat("a", maxProviderNameLen),
		},
		{
			name:    "a name one over the length limit",
			prvname: provider.ProviderName(strings.Repeat("a", maxProviderNameLen+1)),
			refused: "provider name must not exceed 50 characters",
		},
		{
			name:    "a multibyte name at the limit counts characters",
			prvname: provider.ProviderName(strings.Repeat("я", maxProviderNameLen)),
			stored:  strings.Repeat("я", maxProviderNameLen),
		},
		{
			name:    "a multibyte name one over the limit",
			prvname: provider.ProviderName(strings.Repeat("я", maxProviderNameLen+1)),
			refused: "provider name must not exceed 50 characters",
		},
	} {
		for _, update := range []bool{false, true} {
			t.Run(map[bool]string{false: "create ", true: "update "}[update]+tc.name, func(t *testing.T) {
				pc, db := providersSaving(t, &config.Config{}, provider.ProviderOllama, false)

				written, err := providersSave(pc, db, update, tc.prvname, nil)

				providersRequireStored(t, written, err, tc.refused)
				if tc.refused == "" {
					assert.Equal(t, tc.stored, db.lastName, "the stored name keeps its padding")
				}
			})
		}
	}
}

func TestProviders_StoresAConfigOnlyWhenEveryGateAcceptsIt(t *testing.T) {
	agent := func(slot pconfig.ProviderOptionsType, ac pconfig.AgentConfig) func() *pconfig.ProviderConfig {
		return func() *pconfig.ProviderConfig {
			c := ac
			pc := &pconfig.ProviderConfig{}
			switch slot {
			case pconfig.OptionsTypeAdviser:
				pc.Adviser = &c
			case pconfig.OptionsTypeCoder:
				pc.Coder = &c
			case pconfig.OptionsTypePentester:
				pc.Pentester = &c
			case pconfig.OptionsTypePrimaryAgent:
				pc.PrimaryAgent = &c
			default:
				panic("no slot for " + slot)
			}
			return pc
		}
	}
	none := func() *pconfig.ProviderConfig { return nil }
	budget := pconfig.ReasoningConfig{Mode: pconfig.ReasoningModeBudget, MaxTokens: 2048}
	disabledByExtraBody := pconfig.AgentConfig{
		Model:     "claude-sonnet-4-5",
		Reasoning: budget,
		ExtraBody: map[string]any{"thinking": map[string]any{"type": "disabled"}},
	}
	glmBudget := pconfig.AgentConfig{Model: "glm-5-turbo", Reasoning: budget}
	off := pconfig.ReasoningConfig{Mode: pconfig.ReasoningModeOff}

	for _, tc := range []struct {
		name  string
		ptype provider.ProviderType
		env   config.Config
		cfg   func() *pconfig.ProviderConfig
		// refused is the error a refused save names; empty when the save goes through.
		refused string
		// unavailable offers no provider of ptype; only creating checks it.
		unavailable bool
	}{
		{name: "an available type with no config of its own", ptype: provider.ProviderOllama, cfg: none},
		{
			name: "a type without credentials", ptype: provider.ProviderMiniMax, cfg: none, unavailable: true,
			refused: "provider type 'minimax' is not available",
		},
		{
			name: "a type whose default config failed to load at startup", ptype: provider.ProviderBedrock,
			env: config.Config{BedrockConfig: filepath.Join(t.TempDir(), "missing.yml")}, cfg: none,
			refused: "failed to load at startup",
		},

		{
			name: "a named agent model", ptype: provider.ProviderOllama,
			cfg: agent(pconfig.OptionsTypeCoder, pconfig.AgentConfig{Model: "gpt-4o"}),
		},
		{
			name: "an agent with no model of its own", ptype: provider.ProviderOllama,
			cfg: agent(pconfig.OptionsTypeCoder, pconfig.AgentConfig{Model: ""}),
		},
		{
			name: "an agent model of whitespace only", ptype: provider.ProviderOllama,
			cfg:     agent(pconfig.OptionsTypeCoder, pconfig.AgentConfig{Model: " \t\n "}),
			refused: "coder: model must not consist of whitespace only",
		},

		{
			name: "extra_body enables thinking while reasoning is off", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{
				Model: "gpt-5.2", Reasoning: off, ExtraBody: map[string]any{"enable_thinking": true},
			}),
			refused: "adviser: extra_body enables thinking while reasoning.mode is",
		},
		{
			name: "extra_body selects adaptive thinking while reasoning is off", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{
				Model: "gpt-5.2", Reasoning: off, ExtraBody: map[string]any{"thinking": map[string]any{"type": "adaptive"}},
			}),
			refused: "adviser: extra_body enables thinking while reasoning.mode is",
		},
		{
			name: "extra_body disables thinking while reasoning asks for it", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{
				Model:     "gpt-5.2",
				Reasoning: pconfig.ReasoningConfig{Effort: llms.ReasoningHigh},
				ExtraBody: map[string]any{"enable_thinking": false},
			}),
			refused: "adviser: extra_body disables thinking while the reasoning block requests it",
		},
		{
			name: "reasoning off with an agreeing extra_body disable", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{
				Model: "gpt-5.2", Reasoning: off, ExtraBody: map[string]any{"thinking": map[string]any{"type": "disabled"}},
			}),
		},
		{
			name: "reasoning off with a leftover effort and an agreeing extra_body disable", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{
				Model:     "gpt-5.2",
				Reasoning: pconfig.ReasoningConfig{Mode: pconfig.ReasoningModeOff, Effort: llms.ReasoningHigh},
				ExtraBody: map[string]any{"thinking": map[string]any{"type": "disabled"}},
			}),
		},
		{
			name: "reasoning off beside a vendor key in the extra_body thinking object", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{
				Model: "gpt-5.2", Reasoning: off, ExtraBody: map[string]any{"thinking": map[string]any{"clear_thinking": false}},
			}),
		},

		// Only a door that merges extra_body over its own fields can send the contradiction.
		{
			name: "a contradicting extra_body on the anthropic door", ptype: provider.ProviderAnthropic,
			cfg: agent(pconfig.OptionsTypeCoder, disabledByExtraBody),
		},
		{
			name: "a contradicting extra_body on the gemini door", ptype: provider.ProviderGemini,
			cfg: agent(pconfig.OptionsTypeCoder, disabledByExtraBody),
		},
		{
			name: "a contradicting extra_body on the bedrock door", ptype: provider.ProviderBedrock,
			cfg: agent(pconfig.OptionsTypeCoder, disabledByExtraBody),
		},
		{
			name: "a contradicting extra_body on the ollama door", ptype: provider.ProviderOllama,
			cfg: agent(pconfig.OptionsTypeCoder, disabledByExtraBody),
		},
		{
			name: "a contradicting extra_body on the custom door", ptype: provider.ProviderCustom,
			cfg:     agent(pconfig.OptionsTypeCoder, disabledByExtraBody),
			refused: "coder: extra_body disables thinking while the reasoning block requests it",
		},
		{
			name: "a contradicting extra_body on the glm door", ptype: provider.ProviderGLM,
			cfg:     agent(pconfig.OptionsTypeCoder, disabledByExtraBody),
			refused: "coder: extra_body disables thinking while the reasoning block requests it",
		},

		{
			name: "a budget the model takes none of, sent with no prefix", ptype: provider.ProviderGLM,
			cfg:     agent(pconfig.OptionsTypePrimaryAgent, glmBudget),
			refused: `primary_agent: model "glm-5-turbo" takes no thinking budget`,
		},
		{
			name: "a budget the model takes none of, sent behind the zai prefix", ptype: provider.ProviderGLM,
			env:     config.Config{GLMProvider: "zai"},
			cfg:     agent(pconfig.OptionsTypePrimaryAgent, glmBudget),
			refused: `primary_agent: model "zai/glm-5-turbo" takes no thinking budget`,
		},
		{
			name: "the same budget behind a prefix whose model takes one", ptype: provider.ProviderGLM,
			env: config.Config{GLMProvider: "dashscope"},
			cfg: agent(pconfig.OptionsTypePrimaryAgent, glmBudget),
		},

		{
			name: "a level on a model that refuses one with tools", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeCoder, pconfig.AgentConfig{
				Model: "gpt-5.6-sol", Reasoning: pconfig.ReasoningConfig{Effort: llms.ReasoningMedium},
			}),
			refused: `coder: model "gpt-5.6-sol" refuses reasoning effort "medium" on requests with function tools`,
		},
		{
			name: "a budget on a model that refuses a level with tools", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypePentester, pconfig.AgentConfig{
				Model: "gpt-5.4-mini", Reasoning: pconfig.ReasoningConfig{MaxTokens: 2048},
			}),
			refused: `pentester: model "gpt-5.4-mini" refuses a reasoning budget of 2048 tokens on requests with function tools`,
		},
		{
			name: "a level on the adviser, whose calls carry no tools", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{
				Model: "gpt-5.6-sol", Reasoning: pconfig.ReasoningConfig{Effort: llms.ReasoningXHigh},
			}),
		},
		{
			name: "thinking turned off on a model that refuses a level with tools", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeCoder, pconfig.AgentConfig{Model: "gpt-5.6-sol", Reasoning: off}),
		},
		{
			name: "a level on a model that takes one with tools", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeCoder, pconfig.AgentConfig{
				Model: "gpt-5.2", Reasoning: pconfig.ReasoningConfig{Effort: llms.ReasoningHigh},
			}),
		},

		{
			name: "off on a model that cannot disable thinking", ptype: provider.ProviderOpenAI,
			cfg:     agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{Model: "o3", Reasoning: off}),
			refused: `adviser: model "o3" cannot turn reasoning off`,
		},
		{
			name: "no mode on a model that cannot disable thinking", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{Model: "o3"}),
		},
		{
			name: "off on a model that can disable thinking", ptype: provider.ProviderOpenAI,
			cfg: agent(pconfig.OptionsTypeAdviser, pconfig.AgentConfig{Model: "gpt-5.2", Reasoning: off}),
		},
	} {
		for _, update := range []bool{false, true} {
			if update && tc.unavailable {
				continue
			}
			t.Run(map[bool]string{false: "create ", true: "update "}[update]+tc.name, func(t *testing.T) {
				env := tc.env
				pc, db := providersSaving(t, &env, tc.ptype, tc.unavailable)

				written, err := providersSave(pc, db, update, "my-provider", tc.cfg())

				providersRequireStored(t, written, err, tc.refused)
			})
		}
	}
}

func TestProviders_StoresEveryConfigTheProjectShips(t *testing.T) {
	type shippedConfig struct {
		name  string
		ptype provider.ProviderType
		env   *config.Config
	}

	var shipped []shippedConfig
	for _, e := range providerRegistry {
		shipped = append(shipped, shippedConfig{"the default " + e.Type.String() + " config", e.Type, &config.Config{}})
	}

	paths, err := filepath.Glob(filepath.Join("..", "..", "..", "examples", "configs", "*.provider.yml"))
	require.NoError(t, err)
	require.NotEmpty(t, paths)
	for _, path := range paths {
		name := filepath.Base(path)
		ptype, env := provider.ProviderCustom, &config.Config{LLMServerConfig: path}
		switch {
		case strings.HasPrefix(name, "ollama-"):
			ptype, env = provider.ProviderOllama, &config.Config{OllamaServerConfig: path}
		case strings.HasPrefix(name, "bedrock"):
			ptype, env = provider.ProviderBedrock, &config.Config{BedrockConfig: path}
		}
		shipped = append(shipped, shippedConfig{"the example " + name, ptype, env})
	}

	for _, sc := range shipped {
		entry, ok := entryForType(sc.ptype)
		require.True(t, ok, "no registry entry for %s", sc.ptype)

		for _, update := range []bool{false, true} {
			t.Run(map[bool]string{false: "create ", true: "update "}[update]+sc.name, func(t *testing.T) {
				cfg, err := entry.NewConfig(sc.env)
				require.NoError(t, err)
				pc, db := providersSaving(t, &config.Config{}, sc.ptype, false)

				written, err := providersSave(pc, db, update, "shipped", cfg)

				providersRequireStored(t, written, err, "")
			})
		}
	}
}

// providersSeedQuerier answers the seed's lookup and records its writes, failing them on a done context as the real one does.
type providersSeedQuerier struct {
	database.Querier

	stored    database.Provider
	lookupErr error
	lookups   []database.GetUserProviderByNameParams
	created   []database.CreateProviderParams
	updated   []database.UpdateUserProviderParams
}

func (q *providersSeedQuerier) GetUserProviderByName(
	_ context.Context, arg database.GetUserProviderByNameParams,
) (database.Provider, error) {
	q.lookups = append(q.lookups, arg)
	return q.stored, q.lookupErr
}

func (q *providersSeedQuerier) CreateProvider(
	ctx context.Context, params database.CreateProviderParams,
) (database.Provider, error) {
	if err := ctx.Err(); err != nil {
		return database.Provider{}, err
	}
	q.created = append(q.created, params)
	return database.Provider{Name: params.Name}, nil
}

func (q *providersSeedQuerier) UpdateUserProvider(
	ctx context.Context, params database.UpdateUserProviderParams,
) (database.Provider, error) {
	if err := ctx.Err(); err != nil {
		return database.Provider{}, err
	}
	q.updated = append(q.updated, params)
	return database.Provider{ID: params.ID, Name: params.Name}, nil
}

func TestProviders_SeedDefaultProviders_WritesTheBedrockConfigIntoTheUsersProviders(t *testing.T) {
	named := provider.ProvidersConfig{provider.ProviderBedrock: &pconfig.ProviderConfig{Name: "team-bedrock"}}
	unnamed := provider.ProvidersConfig{provider.ProviderBedrock: &pconfig.ProviderConfig{}}
	withToken := config.Config{BedrockConfig: "bedrock.yml", BedrockBearerToken: "token"}

	for _, tc := range []struct {
		name      string
		cfg       config.Config
		configs   provider.ProvidersConfig
		stored    database.Provider
		lookupErr error
		// want is the write expected: "create", "update", or none.
		want string
		// wantLookup is the name looked up; empty when the seed must not touch the database.
		wantLookup string
		wantName   string
		wantConfig string
		refused    string
	}{
		{
			name: "no config path seeds nothing", cfg: config.Config{BedrockBearerToken: "token"},
			configs: named, lookupErr: sql.ErrNoRows,
		},
		{
			name: "a config path without credentials seeds nothing", cfg: config.Config{BedrockConfig: "bedrock.yml"},
			configs: named, lookupErr: sql.ErrNoRows,
		},
		{
			name:    "an access key without its secret seeds nothing",
			cfg:     config.Config{BedrockConfig: "bedrock.yml", BedrockAccessKey: "key"},
			configs: named, lookupErr: sql.ErrNoRows,
		},
		{
			name: "a config that failed to load at startup seeds nothing", cfg: withToken,
			configs: provider.ProvidersConfig{}, lookupErr: sql.ErrNoRows,
		},
		{
			name: "a user without the provider gets it created", cfg: withToken,
			configs: named, lookupErr: sql.ErrNoRows,
			want: "create", wantLookup: "team-bedrock", wantName: "team-bedrock", wantConfig: `{"name":"team-bedrock"}`,
		},
		{
			name:    "static credentials are enough to seed",
			cfg:     config.Config{BedrockConfig: "bedrock.yml", BedrockAccessKey: "key", BedrockSecretKey: "secret"},
			configs: named, lookupErr: sql.ErrNoRows,
			want: "create", wantLookup: "team-bedrock", wantName: "team-bedrock", wantConfig: `{"name":"team-bedrock"}`,
		},
		{
			name: "an unnamed config is seeded under the door's name", cfg: withToken,
			configs: unnamed, lookupErr: sql.ErrNoRows,
			want: "create", wantLookup: "bedrock", wantName: "bedrock", wantConfig: `{}`,
		},
		{
			name: "a user who has the provider gets it updated in place", cfg: withToken,
			configs: named, stored: database.Provider{ID: 9, Name: "team-bedrock"},
			want: "update", wantLookup: "team-bedrock", wantName: "team-bedrock", wantConfig: `{"name":"team-bedrock"}`,
		},
		{
			name: "a failed lookup writes nothing", cfg: withToken,
			configs: named, lookupErr: errors.New("connection reset"), wantLookup: "team-bedrock",
			refused: "failed to get provider 'team-bedrock' from database: connection reset",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			db := &providersSeedQuerier{stored: tc.stored, lookupErr: tc.lookupErr}
			pc := &providerController{cfg: &cfg, db: db, defaultConfigs: tc.configs}

			err := pc.SeedDefaultProviders(context.Background(), 42)

			if tc.refused != "" {
				require.ErrorContains(t, err, tc.refused)
			} else {
				require.NoError(t, err)
			}

			if tc.wantLookup == "" {
				assert.Empty(t, db.lookups, "a seed its gate stops must not touch the database")
			} else {
				assert.Equal(t, []database.GetUserProviderByNameParams{{Name: tc.wantLookup, UserID: 42}}, db.lookups)
			}

			switch tc.want {
			case "create":
				require.Len(t, db.created, 1)
				assert.Empty(t, db.updated)
				assert.Equal(t, int64(42), db.created[0].UserID)
				assert.Equal(t, database.ProviderType(provider.ProviderBedrock), db.created[0].Type)
				assert.Equal(t, tc.wantName, db.created[0].Name)
				assert.JSONEq(t, tc.wantConfig, string(db.created[0].Config))
			case "update":
				require.Len(t, db.updated, 1)
				assert.Empty(t, db.created)
				assert.Equal(t, int64(9), db.updated[0].ID)
				assert.Equal(t, int64(42), db.updated[0].UserID)
				assert.Equal(t, tc.wantName, db.updated[0].Name)
				assert.JSONEq(t, tc.wantConfig, string(db.updated[0].Config))
			default:
				assert.Empty(t, db.created, "nothing may be seeded")
				assert.Empty(t, db.updated, "nothing may be seeded")
			}
		})
	}
}

type noUsers struct{ database.Querier }

func (noUsers) GetUsers(context.Context) ([]database.GetUsersRow, error) { return nil, nil }

func TestProviders_NewProviderController_LeavesAnUnusableDoorOffAndSaysWhy(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		door    provider.ProviderName
		warning string
	}{
		{
			name:    "a custom provider without a key",
			cfg:     &config.Config{LLMServerURL: "http://127.0.0.1:1", LLMServerModel: "some-model"},
			door:    provider.DefaultProviderNameCustom,
			warning: "custom provider disabled: LLM_SERVER_URL is set but LLM_SERVER_KEY is empty",
		},
		{
			name: "an anthropic federation config without its organisation and service account",
			cfg: &config.Config{
				AnthropicFederationRuleID:  "fdrl_1",
				AnthropicIdentityTokenFile: "/var/run/secrets/anthropic.com/token",
			},
			door:    provider.DefaultProviderNameAnthropic,
			warning: "anthropic provider disabled: federation is missing ANTHROPIC_ORGANIZATION_ID, ANTHROPIC_SERVICE_ACCOUNT_ID",
		},
	}

	hook := logtest.NewGlobal()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()

			pc, err := NewProviderController(tt.cfg, noUsers{}, nil)
			require.NoError(t, err, "an unusable door must not stop startup")

			_, wired := pc.(*providerController).Providers[tt.door]
			assert.False(t, wired)

			var warnings []string
			for _, entry := range hook.AllEntries() {
				if entry.Level == logrus.WarnLevel {
					warnings = append(warnings, entry.Message)
				}
			}
			assert.Contains(t, warnings, tt.warning)
		})
	}
}

func TestProviders_TestAgent_RunsTheSimpleAgentOnTheFormsModel(t *testing.T) {
	const defaultModel, formModel = "vendor-a", "vendor-b"

	var (
		mu     sync.Mutex
		models = map[string]int{}
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			fmt.Fprintf(w, `{"data":[{"id":%q},{"id":%q}]}`, defaultModel, formModel)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		var body struct {
			Model string `json:"model"`
		}
		_ = json.Unmarshal(raw, &body)
		mu.Lock()
		models[body.Model]++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"x","object":"chat.completion","choices":[{"index":0,`+
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
	}))
	t.Cleanup(srv.Close)

	cfg := &config.Config{LLMServerKey: "k", LLMServerURL: srv.URL, LLMServerModel: defaultModel}
	defaultConfigs, _, err := buildDefaultConfigs(cfg)
	require.NoError(t, err)
	pc := &providerController{cfg: cfg, defaultConfigs: defaultConfigs}

	var coder pconfig.AgentConfig
	require.NoError(t, json.Unmarshal([]byte(`{"model":"`+formModel+`"}`), &coder))

	_, err = pc.TestAgent(context.Background(), provider.ProviderCustom, pconfig.OptionsTypeCoder, &coder, &coder)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.NotZero(t, models[formModel], "requests per model: %v", models)
	require.Zero(t, models[defaultModel], "requests per model: %v", models)
}

// stubTypedProvider is a provider that reports only its type, all that ListTypes reads.
type stubTypedProvider struct {
	provider.Provider
	ptype provider.ProviderType
}

func (s stubTypedProvider) Type() provider.ProviderType { return s.ptype }

type recordingProvidersQuerier struct {
	database.Querier
	created  int
	updated  int
	lastName string
}

func (r *recordingProvidersQuerier) GetUserProvider(
	_ context.Context, arg database.GetUserProviderParams,
) (database.Provider, error) {
	return database.Provider{ID: arg.ID, UserID: arg.UserID, Name: "stored", Type: database.ProviderType(provider.ProviderOllama)}, nil
}

func (r *recordingProvidersQuerier) UpdateUserProvider(
	_ context.Context, params database.UpdateUserProviderParams,
) (database.Provider, error) {
	r.updated++
	r.lastName = params.Name
	return database.Provider{ID: params.ID, Name: params.Name}, nil
}

func (r *recordingProvidersQuerier) CreateProvider(
	_ context.Context, params database.CreateProviderParams,
) (database.Provider, error) {
	r.created++
	r.lastName = params.Name
	return database.Provider{Name: params.Name}, nil
}
