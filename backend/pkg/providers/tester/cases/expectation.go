package cases

import (
	"errors"
	"slices"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// RefusalKind names a typed error the SDK returns instead of calling the
// provider, because a rule it can evaluate locally says the request cannot
// succeed. Its zero value covers every other outcome, including a wire error.
type RefusalKind string

const (
	// RefusalNone is any outcome the SDK did not decide by itself: a successful
	// call, a provider rejection, a transport failure.
	RefusalNone RefusalKind = ""
	// RefusalStructuredOutputConfig is llms.ErrStructuredOutputConfig: the
	// structured-output request contradicts itself (no JSON mode, an empty or
	// non-object schema, an unusable name, a schema that will not compile).
	RefusalStructuredOutputConfig RefusalKind = "structured_output_config"
	// RefusalStructuredOutputUnsupported is llms.ErrStructuredOutputUnsupported.
	RefusalStructuredOutputUnsupported RefusalKind = "structured_output_unsupported"
	// RefusalStructuredOutputConflict is llms.ErrStructuredOutputConflict.
	RefusalStructuredOutputConflict RefusalKind = "structured_output_conflict"
	// RefusalReasoningOffUnsupported is reasoning.ErrReasoningOffUnsupported.
	RefusalReasoningOffUnsupported RefusalKind = "reasoning_off_unsupported"
)

// ClassifyRefusal reports which local refusal err carries, or RefusalNone when
// err is nil or came from anywhere other than the SDK's own pre-flight rules.
func ClassifyRefusal(err error) RefusalKind {
	if err == nil {
		return RefusalNone
	}

	var (
		unsupported  *llms.ErrStructuredOutputUnsupported
		conflict     *llms.ErrStructuredOutputConflict
		reasoningOff *reasoning.ErrReasoningOffUnsupported
	)

	switch {
	case errors.As(err, &unsupported):
		return RefusalStructuredOutputUnsupported
	case errors.As(err, &conflict):
		return RefusalStructuredOutputConflict
	case errors.As(err, &reasoningOff):
		return RefusalReasoningOffUnsupported
	case errors.Is(err, llms.ErrStructuredOutputConfig):
		return RefusalStructuredOutputConfig
	default:
		return RefusalNone
	}
}

// CountsAsUnsupported reports whether a refusal means this model or provider
// cannot do what was asked — the case TestResult.Unsupported excuses from the
// score. A configuration error is not one: the request itself was malformed, so
// it stays a failure of the caller rather than a limitation of the model.
func (k RefusalKind) CountsAsUnsupported() bool {
	switch k {
	case RefusalStructuredOutputUnsupported,
		RefusalStructuredOutputConflict,
		RefusalReasoningOffUnsupported:
		return true
	default:
		return false
	}
}

func countsReasoning(generationInfo map[string]any) bool {
	tokens, _ := generationInfo["ReasoningTokens"].(int)
	return tokens > 1
}

// The spellings live here only while the pin lacks a machine-readable marker:
// llms.IsTruncated replaces this function once the pin moves.
func IsTruncationStopReason(reason string) bool {
	switch strings.ToLower(strings.TrimSpace(reason)) {
	case "length", "max_tokens", "model_length":
		return true
	default:
		return false
	}
}

func IsContentFilterStopReason(reason string) bool {
	switch strings.ToLower(strings.TrimSpace(reason)) {
	case "content_filter", "content_filtered", "refusal", "safety", "prohibited_content", "blocklist", "spii":
		return true
	default:
		return false
	}
}

func IsContentFilterError(err error) bool {
	if err == nil {
		return false
	}

	var refusal *llms.ErrModelRefusal
	if errors.As(err, &refusal) || llms.IsContentFilterError(err) {
		return true
	}

	message := strings.ToLower(err.Error())
	return slices.ContainsFunc(contentFilterMessages, func(filtered string) bool {
		return strings.Contains(message, filtered)
	})
}

var contentFilterMessages = []string{
	"data may contain inappropriate content",
	"flagged for possible cybersecurity risk",
}
