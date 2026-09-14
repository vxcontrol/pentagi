package provider

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/vxcontrol/langchaingo/llms"
)

// Preflight clamping of max_tokens against the model context window.
//
// OpenAI-compatible backends such as vLLM reject a request outright when
//
//	prompt_tokens + tool_schema_tokens + max_tokens > max_model_len
//
// The rejection is an HTTP 400 with an empty body, so the agent chain retries,
// exhausts its attempts and stops with nothing useful logged. It is easy to
// reach on a 32k window: provider settings routinely persist max_tokens near
// the whole window, and the Generator sends a large system prompt plus several
// tool schemas on its very first call, before any history exists to summarize.
//
// Clamping the completion budget to whatever is left of the window keeps the
// request legal. It cannot help when the prompt alone exceeds the window —
// that needs history compression, which is deliberately out of scope here.

const (
	// MinCompletionTokens is the floor for a clamped budget. Trimming below
	// this produces a truncated reply that is worse than a clear failure, so
	// the clamp stops here and lets the request go out as configured.
	MinCompletionTokens = 256

	// WindowSafetyMargin absorbs the gap between the byte-based estimate and
	// the backend's real tokenizer. The estimate is deliberately rough, so the
	// margin is what stops a near-miss from still being rejected.
	WindowSafetyMargin = 512

	// BytesPerTokenEstimate is the conservative bytes-per-token ratio used to
	// size a prompt without a tokenizer. Real ratios run ~3.5-4.5 for English
	// prose and lower for code and CJK; 4 with the margin above is close
	// enough to keep a request inside the window.
	BytesPerTokenEstimate = 4

	// MaxModelLenEnv names the context window of the configured server. Unset
	// or unparseable disables clamping entirely, so this changes nothing for
	// anyone who has not opted in.
	MaxModelLenEnv = "LLM_SERVER_MAX_MODEL_LEN"
)

var (
	maxModelLenOnce  sync.Once
	maxModelLenValue int
)

// modelWindow reports the configured context window, or 0 when clamping is off.
func modelWindow() int {
	maxModelLenOnce.Do(func() {
		raw := os.Getenv(MaxModelLenEnv)
		if raw == "" {
			return
		}
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return // malformed config disables the clamp rather than guessing
		}
		maxModelLenValue = parsed
	})

	return maxModelLenValue
}

// EstimatePromptTokens approximates the tokens a request will occupy, counting
// message parts and tool schemas.
//
// It errs high: a tool schema is measured as its marshalled JSON, which is what
// providers actually send, and unmarshallable parts fall back to a coarse
// per-part cost rather than being skipped. Undercounting would defeat the
// clamp, so every unknown is rounded against us.
func EstimatePromptTokens(messages []llms.MessageContent, tools []llms.Tool) int {
	bytes := 0

	for _, message := range messages {
		// Role and message framing cost tokens on every provider.
		bytes += len(message.Role) + perMessageOverheadBytes

		for _, part := range message.Parts {
			switch typed := part.(type) {
			case llms.TextContent:
				bytes += len(typed.Text)
			case llms.ToolCall:
				bytes += len(typed.ID) + len(typed.Type)
				if typed.FunctionCall != nil {
					bytes += len(typed.FunctionCall.Name) + len(typed.FunctionCall.Arguments)
				}
			case llms.ToolCallResponse:
				bytes += len(typed.ToolCallID) + len(typed.Name) + len(typed.Content)
			default:
				// Binary and provider-specific parts have no cheap size. Charge
				// a flat cost so they are never counted as free.
				bytes += unknownPartBytes
			}
		}
	}

	for _, tool := range tools {
		bytes += len(tool.Type)
		if tool.Function == nil {
			continue
		}
		bytes += len(tool.Function.Name) + len(tool.Function.Description)
		if tool.Function.Parameters != nil {
			// The schema is sent as JSON, so measure it as JSON.
			if encoded, err := json.Marshal(tool.Function.Parameters); err == nil {
				bytes += len(encoded)
			} else {
				bytes += unknownSchemaBytes
			}
		}
	}

	return bytes / BytesPerTokenEstimate
}

const (
	perMessageOverheadBytes = 16
	unknownPartBytes        = 1024
	unknownSchemaBytes      = 2048
)

// ClampMaxTokens returns options whose max_tokens fits the remaining window,
// and reports whether it changed anything.
//
// It is a no-op when clamping is disabled, when the request already fits, or
// when the prompt leaves less than MinCompletionTokens — the last case is a
// prompt too large for the window, which clamping cannot fix.
func ClampMaxTokens(
	window int,
	messages []llms.MessageContent,
	options []llms.CallOption,
) ([]llms.CallOption, int, bool) {
	if window <= 0 {
		return options, 0, false
	}

	resolved := llms.CallOptions{}
	for _, option := range options {
		option(&resolved)
	}

	// An unset max_tokens means the provider's own default applies; there is
	// nothing of ours to clamp.
	if resolved.MaxTokens == nil {
		return options, 0, false
	}
	configured := *resolved.MaxTokens
	if configured <= 0 {
		return options, 0, false
	}

	estimate := EstimatePromptTokens(messages, resolved.Tools)
	if estimate+configured+WindowSafetyMargin <= window {
		return options, 0, false
	}

	budget := window - estimate - WindowSafetyMargin
	if budget < MinCompletionTokens {
		// The prompt itself is at or over the window. Clamping would yield a
		// uselessly short reply and mask the real problem, so leave the request
		// alone and let the backend report it.
		return options, 0, false
	}
	if budget >= configured {
		return options, 0, false
	}

	// Appended last so it overrides the configured value: llms.WithMaxTokens
	// assigns unconditionally and options are applied in order.
	return append(options, llms.WithMaxTokens(budget)), budget, true
}
