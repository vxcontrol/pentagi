package cases

import (
	"fmt"

	"pentagi/pkg/providers/pconfig"

	"github.com/vxcontrol/langchaingo/llms"
)

func caseOptions(def TestDefinition) []llms.CallOption {
	var options []llms.CallOption
	if def.Params != nil {
		options = append(options, def.Params.BuildOptions()...)
	}
	if def.ToolChoice != nil {
		options = append(options, llms.WithToolChoice(def.ToolChoice))
	}
	return options
}

func validateDefinition(def TestDefinition) error {
	switch def.Type {
	case TestTypeCompletion, TestTypeJSON, TestTypeTool:
	default:
		return fmt.Errorf("unknown test type %q", def.Type)
	}

	switch def.Group {
	case TestGroupBasic, TestGroupAdvanced, TestGroupJSON, TestGroupKnowledge:
	default:
		return fmt.Errorf("unknown group %q: the case would belong to no run", def.Group)
	}

	switch def.Capability {
	case CapabilityNone, CapabilityAdaptiveThinking, CapabilityReasoningOff, CapabilityStructuredOutput:
	default:
		return fmt.Errorf("unknown capability %q: the gate would refuse the case for every agent", def.Capability)
	}

	switch def.ExpectRefusal {
	case RefusalNone, RefusalStructuredOutputConfig, RefusalStructuredOutputUnsupported,
		RefusalStructuredOutputConflict, RefusalReasoningOffUnsupported:
	default:
		return fmt.Errorf("unknown expect_refusal %q: no SDK refusal classifies as that, so the case can never pass", def.ExpectRefusal)
	}

	if def.ToolChoice != nil {
		if def.Type != TestTypeTool {
			return fmt.Errorf("tool_choice needs a %q test: no other type puts tools on the request", TestTypeTool)
		}
		if len(def.Tools) == 0 {
			return fmt.Errorf("tool_choice needs tools to choose from")
		}
		if choice, isString := def.ToolChoice.(string); isString && choice == "none" {
			return fmt.Errorf(`tool_choice %q can never pass: a %q test fails a response with no tool calls`,
				choice, TestTypeTool)
		}
	}

	if def.Params != nil || def.ToolChoice != nil || def.Schema != nil {
		if len(def.Messages) == 0 {
			return fmt.Errorf("params, tool_choice and schema need messages: the prompt-only path cannot carry call options")
		}
	}

	if def.ExpectTruncated && (def.Params == nil || def.Params.MaxTokens <= 0) {
		return fmt.Errorf("expect_truncated needs params.max_tokens: without a cap the answer has nothing to stop at")
	}

	if def.Params == nil {
		return nil
	}

	if def.Params.Model != "" {
		return fmt.Errorf("params must not set a model: the result is reported against the agent's own model")
	}

	switch def.Params.Reasoning.EffectiveMode() {
	case pconfig.ReasoningModeAdaptive:
		return fmt.Errorf("params cannot request adaptive reasoning: use capability %q", CapabilityAdaptiveThinking)
	case pconfig.ReasoningModeOff:
		return fmt.Errorf("params cannot disable reasoning: use capability %q, which gates the case on an agent PentAGI would send the signal for",
			CapabilityReasoningOff)
	}

	if err := def.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	if len(def.Params.BuildOptions()) == 0 {
		return fmt.Errorf("params block produces no call options")
	}

	return nil
}
