package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vxcontrol/langchaingo/llms"

	"pentagi/pkg/cast"
	"pentagi/pkg/smallmodel"
	"pentagi/pkg/tools"
)

// smallModelEnabled reports whether the small-model profile is active.
func (fp *flowProvider) smallModelEnabled() bool {
	return fp.cfg != nil && fp.cfg.SmallModelMode
}

// smallModelVerifier returns the lazily-built verifier configured from the
// authorized scope and the operator's denied-command list.
func (fp *flowProvider) smallModelVerifier() *smallmodel.Verifier {
	fp.mx.Lock()
	defer fp.mx.Unlock()
	if fp.smVerifier == nil {
		fp.smVerifier = smallmodel.NewVerifier(fp.cfg.SmallModelScope, fp.cfg.SmallModelDeniedCommands)
	}
	return fp.smVerifier
}

// smallModelState returns the per-flow structured-state store.
func (fp *flowProvider) smallModelState() *smallmodel.Store {
	fp.mx.Lock()
	defer fp.mx.Unlock()
	if fp.smState == nil {
		fp.smState = smallmodel.NewStore()
	}
	return fp.smState
}

// smallModelBudget builds the token budget from config.
func (fp *flowProvider) smallModelBudget() smallmodel.Budget {
	return smallmodel.Budget{
		CtxWindow:     fp.cfg.SmallModelCtxWindow,
		Percent:       fp.cfg.SmallModelCtxBudgetPercent,
		BytesPerToken: fp.cfg.SmallModelBytesPerToken,
	}
}

// sanitizeToolArgs repairs malformed JSON arguments from a small model before
// execution. It only substitutes the sanitized form when that form is valid
// JSON, so a value it cannot fix is left for the normal execution-error repair
// path rather than being blanked out.
func (fp *flowProvider) sanitizeToolArgs(funcArgs json.RawMessage) json.RawMessage {
	if !fp.smallModelEnabled() {
		return funcArgs
	}
	sanitized := cast.SanitizeJSONControlChars(string(funcArgs))
	if sanitized != string(funcArgs) && json.Valid([]byte(sanitized)) {
		return json.RawMessage(sanitized)
	}
	return funcArgs
}

// verifyToolCall classifies a proposed tool call before execution. When the call
// is not allowed it returns a tool response to feed back to the model and blocked
// set, so the caller skips execution and keeps the agent inside the authorized
// scope. It only inspects commands today; other tools pass through.
func (fp *flowProvider) verifyToolCall(funcName string, funcArgs json.RawMessage) (string, bool) {
	if !fp.smallModelEnabled() || !fp.cfg.SmallModelVerifierEnabled {
		return "", false
	}
	if funcName != tools.TerminalToolName {
		return "", false
	}

	var action struct {
		Input string `json:"input"`
	}
	if err := json.Unmarshal(funcArgs, &action); err != nil || action.Input == "" {
		return "", false
	}

	verdict := fp.smallModelVerifier().CheckCommand(action.Input)
	if verdict.Decision == smallmodel.Allow {
		return "", false
	}

	return fmt.Sprintf(
		"BLOCKED by pre-execution verifier: %s. The command was not run. "+
			"If this is in scope and intended, it needs explicit operator confirmation; "+
			"otherwise choose an in-scope, non-destructive action.",
		verdict.Reason,
	), true
}

// compactToolResponse pre-processes a raw tool response for a small model: it
// parses known formats, keeps head and tail of long output, and records the
// extracted facts in the structured state so they survive summarization.
func (fp *flowProvider) compactToolResponse(funcName, response string, taskID *int64) string {
	if !fp.smallModelEnabled() {
		return response
	}

	compacted := smallmodel.CompactOutput(funcName, response, fp.cfg.SmallModelToolOutputMaxBytes)
	if len(compacted.Facts) > 0 && taskID != nil {
		fp.smallModelState().MergeFacts(smallmodel.Key{FlowID: fp.flowID, TaskID: *taskID}, compacted.Facts)
	}
	return compacted.Summary
}

// smallModelStateBlock returns the structured-state block to prepend to a
// prompt, or an empty string when the profile is off or there is nothing yet.
func (fp *flowProvider) smallModelStateBlock(taskID *int64) string {
	if !fp.smallModelEnabled() || taskID == nil {
		return ""
	}
	return fp.smallModelState().Block(smallmodel.Key{FlowID: fp.flowID, TaskID: *taskID})
}

// smallModelCompactIfOverBudget compacts the chain when it exceeds the token
// budget, independently of the summarizer's own byte thresholds. It returns the
// chain unchanged on any error so it can never make the chain worse.
func (fp *flowProvider) smallModelCompactIfOverBudget(
	ctx context.Context,
	handler tools.SummarizeHandler,
	chain []llms.MessageContent,
) ([]llms.MessageContent, bool) {
	if !fp.smallModelEnabled() || fp.summarizer == nil {
		return chain, false
	}
	if !fp.smallModelBudget().Exceeded(chain) {
		return chain, false
	}
	compacted, err := fp.summarizer.SummarizeChain(ctx, handler, chain, fp.tcIDTemplate)
	if err != nil {
		return chain, false
	}
	return compacted, true
}
