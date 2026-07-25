package tester

import (
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/testdata"
)

// capabilitySupported reports whether opt's agent config, exactly as loaded
// (models.yml + the operator's provider YAML), would actually exercise
// capability during a real PentAGI flow. ctester exists to answer "will this
// configuration work inside PentAGI", so a capability test must only run when
// PentAGI's own runtime — pconfig.AgentConfig.BuildOptions /
// ProviderConfig.UsesAdaptiveThinking, the same code CallWithTools/Call use on
// every real request — would make this exact call for this exact agent.
// Forcing the wire behavior through some other path the app never takes
// (e.g. via CallWithExtraOptions) would test a call that can never happen in
// production, so it is deliberately not done here: pass or fail, the result
// wouldn't say anything about a flow that will ever actually run.
func capabilitySupported(prv provider.Provider, opt pconfig.ProviderOptionsType, capability testdata.TestCapability) bool {
	switch capability {
	case testdata.CapabilityNone:
		return true
	case testdata.CapabilityStructuredOutput:
		// Deliberate exception to the "only test a call PentAGI already
		// makes" rule above: AgentConfig.BuildOptions doesn't wire
		// llms.WithStructuredOutput yet, but that wiring is landing soon for
		// the simple_json agent specifically. This runs now, ahead of it, as
		// a readiness check — is this LLM/backend even capable of
		// schema-constrained output at all — so there's already 100%
		// confidence in the backend once the real call path ships. Only ever
		// reachable for OptionsTypeSimpleJSON in practice: isTestCompatibleWithAgent
		// restricts TestTypeJSON (what structured_output tests are) to that
		// agent alone.
		return true
	}

	agentConfig := prv.GetProviderConfig().AgentConfigForType(opt)
	if agentConfig == nil {
		return false
	}

	switch capability {
	case testdata.CapabilityAdaptiveThinking:
		// Same gate PrepareAdaptiveCallOptions uses to decide whether to
		// append llms.WithAdaptiveReasoning for this opt: either the agent
		// explicitly chose adaptive mode, or its assigned model only supports
		// adaptive thinking (e.g. Claude Opus 4.7/4.8), in which case PentAGI
		// forces it on regardless of what the agent's reasoning config says.
		return prv.GetProviderConfig().UsesAdaptiveThinking(prv.GetModels(), opt)
	case testdata.CapabilityReasoningOff:
		// Same gate AgentConfig.BuildOptions uses to decide whether to append
		// llms.WithReasoningDisabled: only when this agent's own config sets
		// reasoning mode to off. If it doesn't, PentAGI never sends a disable
		// signal for this agent at all, so there is nothing to verify.
		return agentConfig.Reasoning.EffectiveMode() == pconfig.ReasoningModeOff
	default:
		return false
	}
}
