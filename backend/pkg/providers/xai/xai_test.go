package xai

import (
	"strings"
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestXai_BuildProviderConfig_UsesGrok43AsTheDefaultModel(t *testing.T) {
	cfg, err := BuildProviderConfig([]byte("simple:\n  n: 1\n"))
	if err != nil {
		t.Fatalf("build provider config: %v", err)
	}

	options := llms.CallOptions{}
	for _, option := range cfg.GetOptionsForType(pconfig.OptionsTypeSimple) {
		option(&options)
	}
	if got := options.GetModel(); got != "grok-4.3" {
		t.Errorf("default model = %q, want grok-4.3", got)
	}
}

func TestXai_DefaultToolRolesUseTheReplayedModelAndItsPrice(t *testing.T) {
	cfg, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("load bundled config: %v", err)
	}

	for _, role := range []pconfig.ProviderOptionsType{
		pconfig.OptionsTypePrimaryAgent,
		pconfig.OptionsTypeAssistant,
		pconfig.OptionsTypeGenerator,
		pconfig.OptionsTypeRefiner,
		pconfig.OptionsTypeAdviser,
		pconfig.OptionsTypeSearcher,
		pconfig.OptionsTypeCoder,
		pconfig.OptionsTypeInstaller,
		pconfig.OptionsTypePentester,
	} {
		options := llms.CallOptions{}
		for _, option := range cfg.GetOptionsForType(role) {
			option(&options)
		}
		if got := options.GetModel(); got != "grok-4.3" {
			t.Errorf("%s model = %q, want grok-4.3", role, got)
		}
		if got := cfg.GetPriceInfoForType(role); got == nil ||
			got.Input != 1.25 || got.Output != 2.50 || got.CacheRead != 0.20 {
			t.Errorf("%s price = %+v, want grok-4.3 catalogue price", role, got)
		}
	}
}

func TestXai_DefaultUtilityRolesUseTheNonReasoningModel(t *testing.T) {
	cfg, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("load bundled config: %v", err)
	}

	for _, role := range []pconfig.ProviderOptionsType{
		pconfig.OptionsTypeSimple,
		pconfig.OptionsTypeSimpleJSON,
		pconfig.OptionsTypeReflector,
		pconfig.OptionsTypeEnricher,
	} {
		options := llms.CallOptions{}
		for _, option := range cfg.GetOptionsForType(role) {
			option(&options)
		}
		if got := options.GetModel(); got != "grok-4.20-0309-non-reasoning" {
			t.Errorf("%s model = %q, want grok-4.20-0309-non-reasoning", role, got)
		}
	}
}

func TestXai_DefaultModels_KeepsAPIAvailableModelsEvenWhenAReplayFails(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("cannot read the bundled xai catalogue: %v", err)
	}

	want := map[string]bool{
		"grok-4.7": false, "grok-4.6": false, "grok-4.5": false,
		"grok-4.3": false, "grok-4.20-0309-reasoning": false,
		"grok-4.20-0309-non-reasoning": false, "grok-build-0.1": false,
	}
	for _, model := range models {
		if _, ok := want[model.Name]; !ok {
			t.Errorf("unexpected catalogue model %q", model.Name)
		} else {
			want[model.Name] = true
		}
		if model.Price == nil || model.Price.Input <= 0 || model.Price.Output <= 0 {
			t.Errorf("model %q has no positive input/output price", model.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("current model %q is missing", name)
		}
	}
	if len(models) != len(want) {
		t.Fatalf("catalogue has %d models, want %d API-available models", len(models), len(want))
	}
}

func TestXai_DefaultModels_OffersNoMultiAgentVariant(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("cannot read the bundled xai catalogue: %v", err)
	}

	for _, m := range models {
		if strings.Contains(m.Name, "multi-agent") {
			t.Errorf("the xai catalogue offers %q, but xAI serves its multi-agent variants neither over Chat Completions, which this door speaks, nor with client-side tools, which every agent calls",
				m.Name)
		}
	}
}

func TestXai_DefaultModels_AdvertisesOnlyDocumentedReasoningEffort(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("cannot read the bundled xai catalogue: %v", err)
	}

	checked := 0
	for _, m := range models {
		if m.Description == nil || !strings.Contains(*m.Description, "Reasoning cannot be turned off") {
			if m.Reasoning != nil && len(m.Reasoning.Efforts) > 0 {
				t.Errorf("the xai catalogue advertises undocumented reasoning effort values for %q", m.Name)
			}
			continue
		}
		checked++
		if !llms.ReasoningSupportFor(m.Name, provider.ProviderXAI.ReasoningProvider()).CannotDisable {
			t.Errorf("the xai catalogue says reasoning on %q cannot be turned off, but the library lets the form offer Off", m.Name)
		}
	}
	if checked != 3 {
		t.Fatalf("checked %d forced-reasoning models, want three", checked)
	}
}
