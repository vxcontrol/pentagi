package providers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database/converter"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/providers/custom"
	"pentagi/pkg/providers/gemini"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/qwen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// firstModelDeclaringEfforts fails rather than skips: every enrichment row reads its declaration.
func firstModelDeclaringEfforts(t *testing.T) (string, *pconfig.ModelReasoningInfo) {
	t.Helper()

	models, err := openai.DefaultModels()
	if err != nil {
		t.Fatalf("cannot read the bundled openai catalog: %v", err)
	}
	for _, m := range models {
		if m.Reasoning != nil && len(m.Reasoning.Efforts) > 3 {
			return m.Name, m.Reasoning
		}
	}
	t.Fatal("no bundled openai model declares more than the low/medium/high triple: point the rows at another catalogue")

	return "", nil
}

func TestCatalog_EnrichCatalogCapabilities_FillsOnlyWhatTheGatewayLeftUnsaid(t *testing.T) {
	name, declared := firstModelDeclaringEfforts(t)
	own := func() *pconfig.ModelReasoningInfo {
		return &pconfig.ModelReasoningInfo{Efforts: []llms.ReasoningEffort{llms.ReasoningLow}}
	}
	yes, no := true, false

	if known, ok := declaredModelCapabilities()["gpt-4o"]; !ok || known.Thinking == nil || *known.Thinking {
		t.Fatal("gpt-4o is no longer indexed as thinking:false — point the contradiction row at another one")
	}

	for _, tc := range []struct {
		label         string
		model         string
		reasoning     *pconfig.ModelReasoningInfo
		thinking      *bool
		wantReasoning *pconfig.ModelReasoningInfo
		wantThinking  *bool
		why           string
	}{
		{
			label:         "a bare name the gateway says nothing about",
			model:         name,
			wantReasoning: declared,
			wantThinking:  &yes,
			why:           "the reasoning stays absent, so the UI falls back to low/medium/high",
		},
		{
			label:         "a vendor-prefixed name the gateway says nothing about",
			model:         "openai/" + name,
			wantReasoning: declared,
			wantThinking:  &yes,
			why:           "the reasoning stays absent, so the UI falls back to low/medium/high",
		},
		{
			label:         "a name behind two prefixes",
			model:         "openrouter/openai/" + name,
			wantReasoning: declared,
			wantThinking:  &yes,
			why:           "the reasoning stays absent, so the UI falls back to low/medium/high",
		},
		{
			label:         "the gateway says the model reasons",
			model:         "openai/" + name,
			thinking:      &yes,
			wantReasoning: declared,
			wantThinking:  &yes,
			why:           "a gateway agreeing with the catalogue still gets the declared efforts",
		},
		{
			label:        "the gateway says the model does not reason",
			model:        "openai/" + name,
			thinking:     &no,
			wantThinking: &no,
			why:          "the gateway turned thinking off, so no efforts may be offered",
		},
		{
			label:        "the gateway contradicts a catalogue that says the model does not reason",
			model:        "openai/gpt-4o",
			thinking:     &yes,
			wantThinking: &yes,
			why:          "the gateway's word on thinking stands, and the catalogue declares no efforts to add",
		},
		{
			label:         "the gateway describes the reasoning and turns thinking off",
			model:         "openai/" + name,
			reasoning:     own(),
			thinking:      &no,
			wantReasoning: own(),
			wantThinking:  &no,
			why:           "a gateway that describes the model was overwritten",
		},
		{
			label:         "the gateway describes the reasoning and says nothing about thinking",
			model:         "openai/" + name,
			reasoning:     own(),
			wantReasoning: own(),
			why:           "a gateway that describes the model was overwritten",
		},
		{
			label: "a model no bundled catalogue declares",
			model: "some-vendor/never-heard-of-it",
			why:   "invented a capability we never declared",
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			fetched := pconfig.ModelsConfig{{Name: tc.model, Reasoning: tc.reasoning, Thinking: tc.thinking}}

			enriched := EnrichCatalogCapabilities(fetched)

			require.Len(t, enriched, 1)
			assert.Equal(t, tc.model, enriched[0].Name, "the name was rewritten")

			assert.Equal(t, tc.wantReasoning, enriched[0].Reasoning, tc.why)
			assert.Equal(t, tc.wantThinking, enriched[0].Thinking, tc.why)
		})
	}
}

func catalogue(models ...pconfig.ModelConfig) bundledCatalog {
	return bundledCatalog{
		Pkg:  "synthetic",
		Load: func() (pconfig.ModelsConfig, error) { return pconfig.ModelsConfig(models), nil },
	}
}

func efforts(list ...llms.ReasoningEffort) *pconfig.ModelReasoningInfo {
	return &pconfig.ModelReasoningInfo{Efforts: list}
}

func TestCatalog_BuildCapabilityIndex_IndexesOnlyWhatTheCataloguesAgreeOn(t *testing.T) {
	same := func() *pconfig.ModelReasoningInfo { return efforts(llms.ReasoningLow, llms.ReasoningHigh) }
	unreadable := bundledCatalog{
		Pkg:  "unreadable",
		Load: func() (pconfig.ModelsConfig, error) { return nil, errors.New("models.yml does not parse") },
	}

	for _, tc := range []struct {
		label    string
		catalogs []bundledCatalog
		// wantName is the entry indexed under "shared", empty when nothing may be.
		wantName    string
		wantEfforts int
		why         string
	}{
		{
			label: "two catalogues describe the name differently",
			catalogs: []bundledCatalog{
				catalogue(pconfig.ModelConfig{Name: "shared", Reasoning: efforts(llms.ReasoningLow)}),
				catalogue(pconfig.ModelConfig{Name: "vendor/shared", Reasoning: efforts(llms.ReasoningHigh)}),
			},
			why: "a name two catalogues describe differently was resolved by guess instead of dropped",
		},
		{
			label: "two catalogues agree on the name",
			catalogs: []bundledCatalog{
				catalogue(pconfig.ModelConfig{Name: "shared", Reasoning: same()}),
				catalogue(pconfig.ModelConfig{Name: "other/shared", Reasoning: same()}),
			},
			wantName: "shared", wantEfforts: 2,
			why: "two catalogues that agree should still yield the first one's entry",
		},
		{
			label:    "the model declares neither reasoning nor thinking",
			catalogs: []bundledCatalog{catalogue(pconfig.ModelConfig{Name: "shared"})},
			why:      "a model declaring neither reasoning nor thinking was indexed",
		},
		{
			label: "a sibling catalogue does not load",
			catalogs: []bundledCatalog{
				unreadable,
				catalogue(pconfig.ModelConfig{Name: "shared", Reasoning: efforts(llms.ReasoningLow)}),
			},
			wantName: "shared", wantEfforts: 1,
			why: "one catalogue that does not load took the others' entries with it",
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			index := buildCapabilityIndex(tc.catalogs)

			got, ok := index["shared"]
			if tc.wantName == "" {
				assert.False(t, ok, tc.why)
				return
			}
			require.True(t, ok, tc.why)
			assert.Equal(t, tc.wantName, got.Name, tc.why)
			assert.Len(t, got.Reasoning.Efforts, tc.wantEfforts, tc.why)
		})
	}
}

func TestCatalog_BareModelName_KeepsTheSegmentAfterTheLastSlash(t *testing.T) {
	cases := map[string]string{
		"gpt-5.6-luna":                   "gpt-5.6-luna",
		"openai/gpt-5.6-luna":            "gpt-5.6-luna",
		"openrouter/openai/gpt-5.6-luna": "gpt-5.6-luna",
		"Qwen/Qwen3.6-27B-FP8":           "Qwen3.6-27B-FP8",
		"":                               "",
	}
	for input, want := range cases {
		if got := bareModelName(input); got != want {
			t.Errorf("bareModelName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCatalog_BundledCatalogs_ParsesEveryCatalogueOncePerProcess(t *testing.T) {
	first := BundledCatalogs()
	require.Len(t, first, len(bundledCatalogs))
	for prvtype, models := range first {
		require.NotEmpty(t, models, prvtype)
	}

	again := BundledCatalogs()
	assert.Same(t, &first[provider.ProviderOpenAI][0], &again[provider.ProviderOpenAI][0])
	assert.Zero(t, testing.AllocsPerRun(10, func() { _ = BundledCatalogs() }))
}

func TestCatalog_BundledCatalogs_ListsEveryShippedModelsFile(t *testing.T) {
	listed := make(map[string]bool, len(bundledCatalogs))
	for _, catalog := range bundledCatalogs {
		listed[catalog.Pkg] = true
	}

	root := filepath.Join("..", "..", "pkg", "providers")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("cannot read %s: %v", root, err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, e.Name(), "models.yml")); err != nil {
			continue
		}
		if !listed[e.Name()] {
			t.Errorf("%s ships a models.yml but is absent from bundledCatalogs, so nothing indexes it", e.Name())
		}
		delete(listed, e.Name())
	}

	for pkg := range listed {
		t.Errorf("bundledCatalogs lists %q, which ships no models.yml", pkg)
	}
}

// catalogDoor is one bundled catalogue beside the config its door ships.
type catalogDoor struct {
	pkg     string
	ptype   provider.ProviderType
	models  pconfig.ModelsConfig
	shipped *pconfig.ProviderConfig
	// fallback is the model an agent block that names none is sent to.
	fallback string
}

// catalogLoadDoors reads each door's configs through the registry, as the server does; a catalogue with no door fails.
func catalogLoadDoors(t *testing.T) []catalogDoor {
	t.Helper()

	doors := make([]catalogDoor, 0, len(bundledCatalogs))
	for _, catalog := range bundledCatalogs {
		models, err := catalog.Load()
		require.NoError(t, err, "%s catalogue", catalog.Pkg)

		ptype := provider.ProviderType(catalog.Pkg)
		entry, ok := entryForType(ptype)
		require.True(t, ok, "%s ships a catalogue but the registry builds no door for it", catalog.Pkg)

		shipped, err := entry.NewConfig(&config.Config{})
		require.NoError(t, err, "%s: shipped config", catalog.Pkg)
		blank, err := entry.BuildConfig(&config.Config{}, []byte(`{}`))
		require.NoError(t, err, "%s: config with no agent block", catalog.Pkg)

		var options llms.CallOptions
		for _, option := range blank.GetDefaultOptions() {
			option(&options)
		}

		doors = append(doors, catalogDoor{
			pkg: catalog.Pkg, ptype: ptype, models: models, shipped: shipped, fallback: options.GetModel(),
		})
	}

	return doors
}

var limitsPublishedForEveryModel = map[string]bool{
	"anthropic": true,
	"bedrock":   true,
	"gemini":    true,
}

var pricingPage = map[string]string{
	"anthropic": "https://platform.claude.com/docs/en/about-claude/pricing",
	"deepseek":  "https://api-docs.deepseek.com/quick_start/pricing",
	"gemini":    "https://ai.google.dev/gemini-api/docs/pricing",
	"minimax":   "https://platform.minimax.io/docs/guides/pricing-paygo",
	"openai":    "https://developers.openai.com/api/docs/pricing",
	"qwen":      "https://www.alibabacloud.com/help/en/model-studio/model-pricing",
	"xai":       "https://docs.x.ai/developers/pricing",
}

// Door-wide opt-in modes (batch, flex, priority) are left out: the price schema carries the standard rate.
var billedAtMoreThanOneRate = map[string][]string{
	"anthropic": {"claude-opus-5", "claude-opus-4-8"},
	"deepseek":  {"deepseek-flash", "deepseek-v4-pro"},
	"gemini": {
		"gemini-3.8-flash", "gemini-3.7-flash", "gemini-3.6-flash",
		"gemini-3.1-pro-preview", "gemini-3.1-pro-preview-customtools", "gemini-2.5-pro",
	},
	"minimax": {"MiniMax-M3"},
	"openai":  {"gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna", "gpt-5.5", "gpt-5.4"},
	"qwen": {
		"qwen3.7-plus", "qwen3.7-flash", "qwen3.6-max-preview", "qwen3.6-plus", "qwen3.6-flash",
		"qwen3.5-plus", "qwen3-max", "qwen-plus", "qwen-flash",
		"qwen3-coder-plus", "qwen3-coder-flash", "qwen3-coder-next",
		"qwen3-coder-30b-a3b-instruct", "qwen3-coder-480b-a35b-instruct",
		"qwen3-vl-plus", "qwen3-vl-flash", "deepseek-v4-flash-0731",
	},
	"xai": {
		"grok-4.7", "grok-4.6", "grok-4.5", "grok-4.3",
		"grok-4.20-0309-reasoning", "grok-4.20-0309-non-reasoning", "grok-build-0.1",
	},
}

var namesItsTier = regexp.MustCompile(`(?i)\bprices? (below|apply)\b|tier pricing\)`)

var usesTheNewerTokenizer = map[string]bool{
	"claude-fable-5-1": true,
	"claude-fable-5":   true,
	"claude-opus-5":    true,
	"claude-sonnet-5":  true,
	"claude-opus-4-8":  true,
	"claude-opus-4-7":  true,
}

var namesTheTokenizer = regexp.MustCompile(`(?i)tokenizer`)

// Models whose thinking the catalogue and the SDK disagree on, parked until the catalogue declares it.
var awaitingVerdict = map[string]bool{
	"deepseek.v3.2":        true,
	"moonshotai.kimi-k2.5": true,
}

func sdkCallsItReasoning(name string) bool {
	return reasoning.IsReasoningModel(name) || reasoning.ClaudeSupportsThinking(name)
}

func doorHasNoReasoningFor(name string, door provider.ProviderType) bool {
	s := llms.ReasoningSupportFor(name, door.ReasoningProvider())
	return s.Known && !s.Supported
}

// declaresThinking is the catalogue half of the reasoning gate in converter.ConvertModels.
func declaresThinking(m pconfig.ModelConfig) bool {
	return m.Reasoning != nil || (m.Thinking != nil && *m.Thinking)
}

var promotionDate = regexp.MustCompile(`\b(\d{4})-(\d{2})-(\d{2})\b|\b(\d{2})\.(\d{2})\.(\d{4})\b`)

func promotionEnds(description string) (time.Time, bool) {
	earliest := time.Time{}
	for _, sentence := range strings.Split(description, ";") {
		lower := strings.ToLower(sentence)
		if !strings.Contains(lower, "promotion") && !strings.Contains(lower, "promotional") {
			continue
		}
		for _, m := range promotionDate.FindAllStringSubmatch(sentence, -1) {
			layout, value := "2006-01-02", m[0]
			if m[1] == "" {
				layout, value = "02.01.2006", m[0]
			}
			parsed, err := time.Parse(layout, value)
			if err != nil {
				continue
			}
			if earliest.IsZero() || parsed.Before(earliest) {
				earliest = parsed
			}
		}
	}

	return earliest, !earliest.IsZero()
}

var catalogOpenAICompatDoors = map[string]bool{
	"deepseek": true,
	"glm":      true,
	"kimi":     true,
	"minimax":  true,
	"mistral":  true,
	"qwen":     true,
	"xai":      true,
}

func TestCatalog_BundledCatalogs_ShippedDataKeepsItsInvariants(t *testing.T) {
	t.Parallel()

	doors := catalogLoadDoors(t)

	t.Run("a vendor publishing both limits for every model has both in the catalogue", func(t *testing.T) {
		checked := 0
		for _, d := range doors {
			if !limitsPublishedForEveryModel[d.pkg] {
				continue
			}
			for _, m := range d.models {
				checked++

				if m.ContextWindow == nil {
					t.Errorf("%s catalogue leaves %q without a context window, though the vendor documents one for every model",
						d.pkg, m.Name)
				}
				if m.MaxOutputTokens == nil {
					t.Errorf("%s catalogue leaves %q without an output ceiling, though the vendor documents one for every model",
						d.pkg, m.Name)
				}
			}
		}

		if checked == 0 {
			t.Fatal("no catalogue matched limitsPublishedForEveryModel: the guard checks nothing")
		}
	})

	t.Run("an output ceiling stays inside the context window", func(t *testing.T) {
		for _, d := range doors {
			for _, m := range d.models {
				if m.ContextWindow == nil || m.MaxOutputTokens == nil {
					continue
				}
				if *m.MaxOutputTokens > *m.ContextWindow {
					t.Errorf("%s catalogue gives %q an output ceiling of %d above its %d context window",
						d.pkg, m.Name, *m.MaxOutputTokens, *m.ContextWindow)
				}
			}
		}
	})

	t.Run("every entry names its context window", func(t *testing.T) {
		// The vendor catalogue states no window for these two.
		unfilled := map[string]bool{"qwen/qwen3.6-max-preview": true, "qwen/qwen3-max": true}

		for _, d := range doors {
			for _, m := range d.models {
				key := d.pkg + "/" + m.Name
				if m.ContextWindow == nil || *m.ContextWindow <= 0 {
					if !unfilled[key] {
						t.Errorf("%s has no context_window, so the window guard and chain compaction skip every call to it", key)
					}
					continue
				}
				if unfilled[key] {
					t.Errorf("%s now names its window; drop it from the unfilled list", key)
				}
			}
		}
	})

	t.Run("an entry billed at several rates says which one its price is", func(t *testing.T) {
		loaded := map[string]bool{}

		for _, d := range doors {
			loaded[d.pkg] = true

			tiered := map[string]bool{}
			for _, name := range billedAtMoreThanOneRate[d.pkg] {
				tiered[name] = true
			}

			for _, m := range d.models {
				if !tiered[m.Name] {
					continue
				}
				delete(tiered, m.Name)

				if m.Description == nil || !namesItsTier.MatchString(*m.Description) {
					t.Errorf("%s/%s is billed at more than one rate (%s), but its description does not say which one its price is",
						d.pkg, m.Name, pricingPage[d.pkg])
				}
			}

			for name := range tiered {
				t.Errorf("%s/%s is not in the catalogue; drop it from billedAtMoreThanOneRate", d.pkg, name)
			}
		}

		for pkg := range billedAtMoreThanOneRate {
			if !loaded[pkg] {
				t.Errorf("billedAtMoreThanOneRate lists %q, which is not a bundled catalogue", pkg)
			}
		}
	})

	t.Run("only a claude 4.7 or later entry says the newer tokenizer inflates its count", func(t *testing.T) {
		seen := map[string]bool{}

		for _, d := range doors {
			if d.pkg != "anthropic" {
				continue
			}
			for _, m := range d.models {
				desc := ""
				if m.Description != nil {
					desc = *m.Description
				}
				mentions := namesTheTokenizer.MatchString(desc)

				if usesTheNewerTokenizer[m.Name] {
					seen[m.Name] = true
					if !mentions {
						t.Errorf("anthropic/%s runs the Claude 4.7+ tokenizer (~30%% more tokens for the same text), but its description does not say its per-token price is not comparable to 4.6 and earlier",
							m.Name)
					}
					continue
				}

				if mentions {
					t.Errorf("anthropic/%s is not a Claude 4.7+ model, so it must not claim the newer tokenizer", m.Name)
				}
			}
		}

		for name := range usesTheNewerTokenizer {
			if !seen[name] {
				t.Errorf("%s is not in the anthropic catalogue; drop it from usesTheNewerTokenizer", name)
			}
		}
	})

	t.Run("an entry declares the thinking the sdk knows about", func(t *testing.T) {
		settled := map[string]string{}
		shipped := map[string]bool{}

		for _, d := range doors {
			for _, m := range d.models {
				if awaitingVerdict[m.Name] {
					shipped[m.Name] = true
					if declaresThinking(m) {
						settled[m.Name] = d.pkg
					}
					continue
				}

				noReasoningHere := doorHasNoReasoningFor(m.Name, d.ptype)
				switch {
				case declaresThinking(m) && noReasoningHere:
					t.Errorf("%s catalogue declares that %q thinks, but the SDK classifies it as a model that does not reason on this door: the interface would offer a reasoning control the door cannot fill",
						d.pkg, m.Name)
				case !declaresThinking(m) && sdkCallsItReasoning(m.Name) && !noReasoningHere:
					t.Errorf("%s catalogue calls %q a model that does not think, but the SDK classifies it as one: the interface would offer no reasoning control for it",
						d.pkg, m.Name)
				}
			}
		}

		for name, pkg := range settled {
			t.Errorf("%s catalogue now declares that %q thinks: the disagreement is over, drop the entry from awaitingVerdict",
				pkg, name)
		}
		for name := range awaitingVerdict {
			if !shipped[name] {
				t.Errorf("awaitingVerdict parks %q, which no bundled catalogue ships: the entry parks nothing, drop it", name)
			}
		}
	})

	t.Run("the expired-promotion guard reads the end date out of the prose", func(t *testing.T) {
		for _, tc := range []struct {
			label       string
			description string
			want        string
		}{
			{"a day-first date", "Prices below are the 50% promotional rate that ends 09.09.2026; list prices are 0.15 input.", "2026-09-09"},
			{"an iso date", "Prices below are the promotional rate through 2026-12-31; from 2027-01-01 they return to 1.5 input.", "2026-12-31"},
			{"no promotion named", "Retired on 2026-09-14; requests are served by another model.", ""},
			{"a promotion without a date", "Prices below are the vendor's permanent 50% promotional rate.", ""},
			{"a clock range is not a date", "Prices below are the peak rate; off-peak (outside 01:00-04:00 UTC) is half.", ""},
			{"a model name is not a date", "Snapshot claude-sonnet-4-5-20250929 kept for reproducibility.", ""},
		} {
			t.Run(tc.label, func(t *testing.T) {
				ends, ok := promotionEnds(tc.description)
				if tc.want == "" {
					if ok {
						t.Errorf("read %s out of a description that names no promotion end", ends.Format("2006-01-02"))
					}
					return
				}
				if !ok {
					t.Fatalf("read no date out of %q", tc.description)
				}
				if got := ends.Format("2006-01-02"); got != tc.want {
					t.Errorf("read %s, want %s", got, tc.want)
				}
			})
		}
	})

	t.Run("no entry charges an expired promotion", func(t *testing.T) {
		for _, d := range doors {
			for _, m := range d.models {
				if m.Description == nil {
					continue
				}
				ends, ok := promotionEnds(*m.Description)
				if !ok {
					continue
				}
				if !ends.AddDate(0, 0, 1).After(time.Now().UTC()) {
					t.Errorf("%s/%s prices a promotion that ran out on %s, so every flow on it records less than the vendor charges",
						d.pkg, m.Name, ends.Format("2006-01-02"))
				}
			}
		}
	})

	t.Run("an agent block that names no model falls back to a catalogued model", func(t *testing.T) {
		for _, d := range doors {
			if d.fallback == "" {
				t.Errorf("%s: an agent block that names no model calls no model at all", d.pkg)
				continue
			}
			if !slices.ContainsFunc(d.models, func(m pconfig.ModelConfig) bool { return m.Name == d.fallback }) {
				t.Errorf("an agent block that names no model calls %q, which the %s catalogue does not list", d.fallback, d.pkg)
			}
		}
	})

	t.Run("every shipped agent binding names a catalogued model", func(t *testing.T) {
		for _, d := range doors {
			checked := 0
			for _, opt := range pconfig.AllAgentTypes {
				ac := d.shipped.AgentConfigForType(opt)
				if ac == nil || ac.Model == "" {
					continue
				}
				checked++
				if !slices.ContainsFunc(d.models, func(m pconfig.ModelConfig) bool { return m.Name == ac.Model }) {
					t.Errorf("%s/%s binds model %q, which is absent from models.yml", d.pkg, opt, ac.Model)
				}
			}
			if checked == 0 {
				t.Errorf("%s: no agent binding was checked, the guard verifies nothing", d.pkg)
			}
		}
	})

	// GetPriceInfoForType has no catalogue fallback, so a drift mis-prices cost telemetry.
	t.Run("a shipped agent binding carries its model's catalogue price", func(t *testing.T) {
		for _, d := range doors {
			catalog := make(map[string]*pconfig.PriceInfo, len(d.models))
			for i := range d.models {
				catalog[d.models[i].Name] = d.models[i].Price
			}

			for _, opt := range pconfig.AllAgentTypes {
				ac := d.shipped.AgentConfigForType(opt)
				if ac == nil || ac.Model == "" || ac.Price == nil {
					continue
				}
				cat, ok := catalog[ac.Model]
				if !ok || cat == nil {
					continue
				}
				if *ac.Price != *cat {
					t.Errorf("%s/%s model %q: config.yml price %+v != catalog price %+v", d.pkg, opt, ac.Model, *ac.Price, *cat)
				}
			}
		}
	})

	t.Run("a shipped reasoning effort is a level the form offers", func(t *testing.T) {
		for _, d := range doors {
			capabilities := make(map[string]*model.ModelReasoningInfo, len(d.models))
			for _, m := range converter.ConvertModels(d.models, d.ptype.ReasoningProvider()) {
				capabilities[m.Name] = m.Reasoning
			}

			for _, opt := range pconfig.AllAgentTypes {
				ac := d.shipped.AgentConfigForType(opt)
				if ac == nil || ac.Reasoning.Effort == "" {
					continue
				}
				capability, listed := capabilities[ac.Model]
				if !listed || capability == nil {
					continue
				}
				if !slices.Contains(capability.Efforts, model.ReasoningEffort(ac.Reasoning.Effort)) {
					t.Errorf("%s: %s on %s ships reasoning effort %q, which the form does not offer for it (efforts %v)",
						d.pkg, opt, ac.Model, ac.Reasoning.Effort, capability.Efforts)
				}
			}
		}
	})

	// The converter drops such an effort silently; none reaches the form through cannotDisable instead.
	t.Run("the only catalogue effort the graphql enum lacks is none", func(t *testing.T) {
		dropped := map[string][]string{}
		for _, d := range doors {
			for _, m := range d.models {
				if m.Reasoning == nil {
					continue
				}
				for _, effort := range m.Reasoning.Efforts {
					if !model.ReasoningEffort(effort).IsValid() {
						dropped[string(effort)] = append(dropped[string(effort)], d.pkg+"/"+m.Name)
					}
				}
			}
		}

		efforts := make([]string, 0, len(dropped))
		for effort := range dropped {
			efforts = append(efforts, effort)
		}
		slices.Sort(efforts)

		assert.Equal(t, []string{"none"}, efforts,
			"anything but none is a typo in a models.yml or a level the enum owes the catalogue: %v", dropped)
	})

	t.Run("an openai-compatible door declares no adaptive thinking", func(t *testing.T) {
		seen := map[string]bool{}
		for _, d := range doors {
			if !catalogOpenAICompatDoors[d.pkg] {
				continue
			}
			seen[d.pkg] = true

			for _, opt := range pconfig.AllAgentTypes {
				if d.shipped.UsesAdaptiveThinking(d.models, opt) {
					t.Errorf("%s/%s resolves to adaptive thinking, which the openai-compatible transport cannot "+
						"express: stop declaring it", d.pkg, opt)
				}
			}

			for _, m := range d.models {
				if m.Reasoning == nil {
					continue
				}
				if mode := m.Reasoning.Mode; mode == pconfig.ModelReasoningAdaptive || mode == pconfig.ModelReasoningAdaptiveOnly {
					t.Errorf("%s catalog model %q declares reasoning mode %q, which the openai-compatible "+
						"transport cannot express: stop declaring it", d.pkg, m.Name, mode)
				}
			}
		}

		for pkg := range catalogOpenAICompatDoors {
			if !seen[pkg] {
				t.Errorf("catalogOpenAICompatDoors lists %q, which is not a bundled catalogue", pkg)
			}
		}
	})
}

// catalogCustomEntry is the custom door handed enrich, or no enrichment when it is nil.
func catalogCustomEntry(enrich func(pconfig.ModelsConfig) pconfig.ModelsConfig) registryEntry {
	return registryEntry{
		BuildConfig: custom.BuildProviderConfig,
		New: func(cfg *config.Config, name provider.ProviderName, pc *pconfig.ProviderConfig) (provider.Provider, error) {
			return custom.New(cfg, name, pc, enrich)
		},
	}
}

func checkAdaptiveLeftToTheModel(
	t *testing.T, prov provider.Provider, resp *llms.ContentResponse, body map[string]any, wantAdaptive bool,
) {
	t.Helper()

	model := prov.Model(pconfig.OptionsTypeSimple)
	if got := prov.GetProviderConfig().UsesAdaptiveThinking(prov.GetModels(), pconfig.OptionsTypeSimple); got != wantAdaptive {
		t.Errorf("adaptive thinking for %q = %v, want %v", model, got, wantAdaptive)
	}

	_, gotModern := body["reasoning"]
	_, gotLegacy := body["reasoning_effort"]
	if gotModern || gotLegacy {
		t.Errorf("a reasoning parameter reached the wire, but %q thinks adaptively at its default effort "+
			"when the request names none; body=%v", model, body)
	}

	for _, warning := range resp.Warnings {
		if warning.Option == "WithAdaptiveReasoning" {
			t.Errorf("the adaptive request was reported lost on a model that thinks without one: %s", warning)
		}
	}
}

// The model is adaptive-only in the bundled catalogue; the gateway does not say so.
func TestCatalog_EnrichCatalogCapabilities_GivesTheCustomDoorTheAdaptiveModeItsModelRequires(t *testing.T) {
	const model = "claude-sonnet-5"

	for _, tc := range []struct {
		label        string
		block        map[string]any
		enrich       func(pconfig.ModelsConfig) pconfig.ModelsConfig
		wantAdaptive bool
	}{
		{"raw catalogue, the agent naming the model", map[string]any{"model": model}, nil, false},
		{"enriched catalogue, the agent naming the model", map[string]any{"model": model}, EnrichCatalogCapabilities, true},
		{"enriched catalogue, the shipped template's empty block", map[string]any{}, EnrichCatalogCapabilities, true},
		{"enriched catalogue, a block tuning temperature only", map[string]any{"temperature": 0.5}, EnrichCatalogCapabilities, true},
		{"enriched catalogue, a block tuning max_tokens only", map[string]any{"max_tokens": 1024}, EnrichCatalogCapabilities, true},
	} {
		t.Run(tc.label, func(t *testing.T) {
			prov, body := customDoor(t, model, tc.block, catalogCustomEntry(tc.enrich))

			resp, err := prov.CallEx(context.Background(), pconfig.OptionsTypeSimple,
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil)
			if err != nil {
				t.Fatalf("call: %v", err)
			}

			checkAdaptiveLeftToTheModel(t, prov, resp, *body, tc.wantAdaptive)
		})
	}
}

var (
	readmeModelRow    = regexp.MustCompile("^\\|\\s*`([^`]+)`(\\*?)\\s*\\|")
	readmeBacktick    = regexp.MustCompile("`([^`]+)`")
	readmeAssignment  = regexp.MustCompile("^- \\*\\*`([^`]+)`\\*\\* - (.+)$")
	readmeEffortLevel = regexp.MustCompile(`^- \*\*([A-Za-z ]+)\*\*: (.+)$`)
	readmeRoleRow     = regexp.MustCompile("^\\|(.+?)\\|\\s*`([^`]+)`\\s*\\|")
)

func readmeHeading(t *testing.T, lines []string, level, heading string) []string {
	t.Helper()

	start := slices.Index(lines, level+" "+heading)
	require.NotEqual(t, -1, start, "README has no heading %q", heading)

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], level+" ") {
			end = i
			break
		}
	}
	return lines[start+1 : end]
}

func readmeSection(t *testing.T, heading string) []string {
	t.Helper()

	raw, err := os.ReadFile("../../../README.md")
	require.NoError(t, err)

	return readmeHeading(t, strings.Split(string(raw), "\n"), "###", heading)
}

func readmeBlock(t *testing.T, section []string, marker string) []string {
	t.Helper()

	start := slices.Index(section, marker)
	require.NotEqual(t, -1, start, "README section has no %q block", marker)

	var block []string
	for _, line := range section[start+1:] {
		if strings.TrimSpace(line) == "" {
			break
		}
		block = append(block, line)
	}
	require.NotEmpty(t, block, "README block %q is empty", marker)
	return block
}

func readmeModelRows(t *testing.T, section, exclude []string) (listed []string, starred []string) {
	t.Helper()

	for _, line := range readmeHeading(t, section, "####", "Supported Models") {
		if slices.Contains(exclude, line) {
			continue
		}
		if m := readmeModelRow.FindStringSubmatch(line); m != nil {
			listed = append(listed, m[1])
			if m[2] == "*" {
				starred = append(starred, m[1])
			}
		}
	}
	slices.Sort(listed)
	slices.Sort(starred)
	return listed, starred
}

func backtickedRoles(text string) []string {
	var roles []string
	for _, m := range readmeBacktick.FindAllStringSubmatch(text, -1) {
		roles = append(roles, m[1])
	}
	return roles
}

func shippedBindings(t *testing.T, cfg *pconfig.ProviderConfig) (models map[string]string, bound []string) {
	t.Helper()

	models = make(map[string]string)
	for _, opt := range pconfig.AllAgentTypes {
		ac := cfg.AgentConfigForType(opt)
		require.NotNil(t, ac, "%s has no shipped binding", opt)
		models[string(opt)] = ac.Model
		if !slices.Contains(bound, ac.Model) {
			bound = append(bound, ac.Model)
		}
	}
	slices.Sort(bound)
	return models, bound
}

// Each row brings its own reader: the sections lay out the role-to-model list differently.
func TestCatalog_ReadmeProviderSectionsMatchTheShippedConfig(t *testing.T) {
	for _, tc := range []struct {
		name    string
		heading string
		config  func() (*pconfig.ProviderConfig, error)
		models  func() (pconfig.ModelsConfig, error)
		// roles reads the role-to-model list, and the lines of it that are not model rows.
		roles func(t *testing.T, section []string) (assigned map[string]string, notModels []string)
		// efforts, when set, checks the section's effort levels against cfg.
		efforts func(t *testing.T, section []string, cfg *pconfig.ProviderConfig)
	}{
		{
			name:    "the gemini section",
			heading: "Google AI (Gemini) Provider Configuration",
			config:  gemini.DefaultProviderConfig,
			models:  gemini.DefaultModels,
			roles: func(t *testing.T, section []string) (map[string]string, []string) {
				assigned := make(map[string]string)
				for _, line := range readmeBlock(t, section, "**Default Model Assignments (config.yml)**:") {
					m := readmeAssignment.FindStringSubmatch(line)
					require.NotNil(t, m, "unparsable assignment line %q", line)
					for _, role := range backtickedRoles(m[2]) {
						assert.NotContains(t, assigned, role, "role %s is assigned twice", role)
						assigned[role] = m[1]
					}
				}
				return assigned, nil
			},
			efforts: func(t *testing.T, section []string, cfg *pconfig.ProviderConfig) {
				wantEfforts := make(map[string]string)
				for _, opt := range pconfig.AllAgentTypes {
					rc := cfg.AgentConfigForType(opt).Reasoning
					wantEfforts[string(opt)] = string(rc.Effort)
					if rc.EffectiveMode() == pconfig.ReasoningModeBudget {
						wantEfforts[string(opt)] = "budget"
					}
				}
				gotEfforts := make(map[string]string)
				for _, line := range readmeBlock(t, section, "**Reasoning Effort Levels**:") {
					m := readmeEffortLevel.FindStringSubmatch(line)
					require.NotNil(t, m, "unparsable effort line %q", line)
					level := strings.ToLower(m[1])
					if level == "not set" {
						level = ""
					}
					for _, role := range backtickedRoles(m[2]) {
						assert.NotContains(t, gotEfforts, role, "role %s has two effort levels", role)
						gotEfforts[role] = level
					}
				}
				assert.Equal(t, wantEfforts, gotEfforts)
			},
		},
		{
			name:    "the qwen section",
			heading: "Qwen Provider Configuration",
			config:  qwen.DefaultProviderConfig,
			models:  qwen.DefaultModels,
			roles: func(t *testing.T, section []string) (map[string]string, []string) {
				defaults := readmeBlock(t, section, "**Default Agent Configuration**:")
				assigned := make(map[string]string)
				for _, line := range defaults[2:] {
					m := readmeRoleRow.FindStringSubmatch(line)
					require.NotNil(t, m, "unparsable default agent row %q", line)
					for _, role := range backtickedRoles(m[1]) {
						assert.NotContains(t, assigned, role, "role %s is assigned twice", role)
						assigned[role] = m[2]
					}
				}
				return assigned, defaults
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			section := readmeSection(t, tc.heading)

			cfg, err := tc.config()
			require.NoError(t, err)
			catalog, err := tc.models()
			require.NoError(t, err)

			wantModels, wantBound := shippedBindings(t, cfg)

			var catalogued []string
			for _, m := range catalog {
				catalogued = append(catalogued, m.Name)
			}
			slices.Sort(catalogued)

			gotModels, notModels := tc.roles(t, section)

			listed, starred := readmeModelRows(t, section, notModels)
			assert.Equal(t, catalogued, listed, "README model tables must list exactly the models.yml catalogue")
			assert.Equal(t, wantBound, starred, "README must star exactly the models config.yml binds")
			assert.Equal(t, wantModels, gotModels)

			if tc.efforts != nil {
				tc.efforts(t, section, cfg)
			}
		})
	}
}
