package tester

import (
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/cases"
	"pentagi/pkg/providers/tester/mock"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestCapability_CapabilitySupported_RunsACapabilityCaseOnlyWhereTheAgentConfigSendsIt(t *testing.T) {
	const (
		adaptiveCase = "adaptive_thinking_reasoning_present"
		offCase      = "reasoning_off_suppresses_reasoning"
	)

	for _, tt := range []struct {
		name     string
		provider func() *mock.Provider
		want     map[string]cases.TestCapability
	}{
		{
			name: "an adaptive-only model runs the adaptive thinking case",
			provider: func() *mock.Provider {
				const model = "claude-opus-4-8"
				prv := mock.NewProvider(provider.ProviderAnthropic, provider.DefaultProviderNameAnthropic, model)
				prv.SetProviderConfig(&pconfig.ProviderConfig{Simple: &pconfig.AgentConfig{Model: model}})
				prv.SetModels(pconfig.ModelsConfig{
					{Name: model, Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningAdaptiveOnly}},
				})
				prv.SetResponses([]mock.ResponseConfig{{Key: "farmer has 17 sheep", Response: &llms.ContentResponse{
					Choices: []*llms.ContentChoice{{
						Content:   "The answer is 13",
						Reasoning: &reasoning.ContentReasoning{Content: "mock reasoning trace"},
					}},
				}}})
				return prv
			},
			want: map[string]cases.TestCapability{adaptiveCase: cases.CapabilityAdaptiveThinking},
		},
		{
			name: "an agent config turning reasoning off runs the reasoning off case",
			provider: func() *mock.Provider {
				prv := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "some-model")
				prv.SetProviderConfig(&pconfig.ProviderConfig{Simple: &pconfig.AgentConfig{
					Model:     "some-model",
					Reasoning: pconfig.ReasoningConfig{Mode: pconfig.ReasoningModeOff},
				}})
				prv.SetResponses([]mock.ResponseConfig{{Key: "capital of France", Response: &llms.ContentResponse{
					Choices: []*llms.ContentChoice{{Content: "Paris"}},
				}}})
				return prv
			},
			want: map[string]cases.TestCapability{offCase: cases.CapabilityReasoningOff},
		},
		{
			name: "a default config runs no capability case",
			provider: func() *mock.Provider {
				prv := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
				prv.SetDefaultResponse("Mock response")
				return prv
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			results, err := TestProvider(
				t.Context(),
				tt.provider(),
				WithAgentTypes(pconfig.OptionsTypeSimple),
				WithGroups(cases.TestGroupAdvanced),
				WithStreamingMode(false),
			)
			if err != nil {
				t.Fatalf("TestProvider capability gating failed: %v", err)
			}

			ran := make(map[string]bool)
			for _, result := range results.Simple {
				ran[result.ID] = true
				capability, wanted := tt.want[result.ID]
				if !wanted {
					if result.Capability != cases.CapabilityNone {
						t.Errorf("%s ran with capability %q, which this config never sends", result.ID, result.Capability)
					}
					continue
				}
				if !result.Success {
					t.Errorf("%s failed: %v", result.ID, result.Error)
				}
				if result.Capability != capability {
					t.Errorf("%s reports capability %q, want %q", result.ID, result.Capability, capability)
				}
			}

			for _, id := range []string{adaptiveCase, offCase} {
				if _, wanted := tt.want[id]; ran[id] != wanted {
					t.Errorf("%s ran = %v, want %v", id, ran[id], wanted)
				}
			}
		})
	}
}
