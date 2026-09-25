package cases

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestExpectation_ClassifyRefusal_NamesTheLocalRefusalAnErrorCarries(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want RefusalKind
	}{
		{"no error", nil, RefusalNone},
		{"a wire failure", errors.New("connection reset by peer"), RefusalNone},
		{"a structured output config error", fmt.Errorf("%w: schema is empty", llms.ErrStructuredOutputConfig), RefusalStructuredOutputConfig},
		{
			"structured output the model does not support",
			&llms.ErrStructuredOutputUnsupported{Provider: "openai", Model: "gpt-4"},
			RefusalStructuredOutputUnsupported,
		},
		{"conflicting structured output", &llms.ErrStructuredOutputConflict{Provider: "anthropic"}, RefusalStructuredOutputConflict},
		{"reasoning that cannot be turned off", &reasoning.ErrReasoningOffUnsupported{Model: "o3"}, RefusalReasoningOffUnsupported},
		{
			"a refusal wrapped several layers deep",
			fmt.Errorf("call failed: %w", fmt.Errorf("agent simple: %w",
				&reasoning.ErrReasoningOffUnsupported{Model: "o3"})),
			RefusalReasoningOffUnsupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyRefusal(tt.err); got != tt.want {
				t.Errorf("ClassifyRefusal(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

func TestExpectation_CountsAsUnsupported_ExcusesOnlyAModelLimitation(t *testing.T) {
	excused := []RefusalKind{
		RefusalStructuredOutputUnsupported,
		RefusalStructuredOutputConflict,
		RefusalReasoningOffUnsupported,
	}
	for _, kind := range excused {
		if !kind.CountsAsUnsupported() {
			t.Errorf("%q describes a model that cannot do what was asked, so it must be excused from the score", kind)
		}
	}

	charged := []RefusalKind{RefusalNone, RefusalStructuredOutputConfig}
	for _, kind := range charged {
		if kind.CountsAsUnsupported() {
			t.Errorf("%q is a malformed request, not a model limitation, so it must count against the score", kind)
		}
	}
}

func TestExpectation_IsContentFilterError_RecognisesEachVendorsFilter(t *testing.T) {
	for _, tc := range []struct {
		label string
		err   error
		want  bool
	}{
		{
			"the DashScope moderation behind the gateway",
			errors.New("API returned unexpected status code: 400: litellm.BadRequestError: DashscopeException - " +
				"Input data may contain inappropriate content. For details, see: " +
				"https://www.alibabacloud.com/help/en/model-studio/error-code#inappropriate-content."),
			true,
		},
		{
			"the OpenAI cyber policy on the vendor's own endpoint",
			errors.New("API returned unexpected status code: 400: This content was flagged for possible cybersecurity risk. " +
				"If this seems wrong, try rephrasing your request. To get authorized for security work, " +
				"join the Trusted Access for Cyber program: https://chatgpt.com/cyber"),
			true,
		},
		{
			"the OpenAI cyber policy behind the gateway",
			errors.New("API returned unexpected status code: 400: litellm.BadRequestError: OpenAIException - {\n" +
				"  \"error\": {\n" +
				"    \"message\": \"This content was flagged for possible cybersecurity risk. " +
				"If this seems wrong, try rephrasing your request. To get authorized for security work, " +
				"join the Trusted Access for Cyber program: https://chatgpt.com/cyber\",\n" +
				"    \"type\": \"invalid_request\",\n    \"param\": null,\n    \"code\": \"cyber_policy\"\n  }\n" +
				"}. Received Model Group=openai/gpt-5.6-sol\nAvailable Model Group Fallbacks=None"),
			true,
		},
		{"an Anthropic refusal", &llms.ErrModelRefusal{Provider: "anthropic", Category: "cyber"}, true},
		{"a transport failure", errors.New("connection reset by peer"), false},
		{"a gateway rejecting a parameter", errors.New("API returned unexpected status code: 400: litellm.UnsupportedParamsError"), false},
	} {
		t.Run(tc.label, func(t *testing.T) {
			if got := IsContentFilterError(tc.err); got != tc.want {
				t.Errorf("IsContentFilterError = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExpectation_IsContentFilterStopReason_RecognisesEachVendorsSpelling(t *testing.T) {
	for _, tc := range []struct {
		label  string
		reason string
		want   bool
	}{
		{"the Bedrock Nova filtered stop", "content_filtered", true},
		{"the OpenAI filtered stop", "content_filter", true},
		{"the Gemini safety stop", "SAFETY", true},
		{"the Gemini prohibited content stop", "PROHIBITED_CONTENT", true},
		{"a natural stop", "STOP", false},
	} {
		t.Run(tc.label, func(t *testing.T) {
			if got := IsContentFilterStopReason(tc.reason); got != tc.want {
				t.Errorf("IsContentFilterStopReason(%q) = %v, want %v", tc.reason, got, tc.want)
			}
		})
	}
}

// The rule is shared by every case type's Execute, so each builder is driven through it.
func TestExpectation_CountsReasoning_FlagsTextOrACountAboveOneTokenInEveryCaseType(t *testing.T) {
	builders := map[string]func(TestDefinition) (TestCase, error){
		"reasoning_off_suppresses_reasoning": newCompletionTestCase,
		"person_info_json":                   newJSONTestCase,
		"echo_function_basic":                newToolTestCase,
	}
	for id, build := range builders {
		for _, tc := range []struct {
			tokens int
			text   string
			want   bool
		}{
			{tokens: 0, want: false},
			{tokens: 1, want: false},
			{tokens: 2, want: true},
			{tokens: 1, text: "thinking it over", want: true},
		} {
			testCase, err := build(builtinDefinition(t, id))
			if err != nil {
				t.Fatalf("%s: %v", id, err)
			}
			choice := llms.ContentChoice{Content: "Paris", GenerationInfo: map[string]any{"ReasoningTokens": tc.tokens}}
			if tc.text != "" {
				choice.Reasoning = &reasoning.ContentReasoning{Content: tc.text}
			}
			result := testCase.Execute(&llms.ContentResponse{Choices: []*llms.ContentChoice{&choice}}, time.Second)
			if result.Reasoning != tc.want {
				t.Errorf("%s with %d reasoning tokens and text %q: reasoning = %v, want %v",
					id, tc.tokens, tc.text, result.Reasoning, tc.want)
			}
		}
	}
}
