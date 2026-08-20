// Package deafguard implements pre-execution command classification for PentAGI.
//
// It intercepts terminal commands before they execute and classifies them through
// a 9-tier risk taxonomy using regex-based pattern matching. The classification
// result determines whether a command is allowed, warned, or blocked based on
// the current enforcement mode.
//
// Known limitations (Phase 1):
//   - Regex evasion: Commands using base64 encoding, $() substitution, backtick
//     expansion, escaped characters, or PATH-qualified binaries (/usr/bin/nsenter)
//     can bypass pattern matching. Phase 3 adds LLM-assisted classification.
//   - File tool: The "file" tool is not classified. It can write to arbitrary paths
//     (e.g., /etc/cron.d/, .ssh/authorized_keys). Planned for Phase 2.
//   - Shell parsing: splitCommand is simplified and does not handle heredocs,
//     escaped quotes within strings, or process substitution.
//   - Per-flow snapshot: DeafGuard config is captured at flow creation time.
//     Runtime changes via the REST API only affect subsequent flows.
package deafguard

import (
	"encoding/json"
	"strings"
	"time"

	"pentagi/pkg/config"

	"github.com/sirupsen/logrus"
)

// terminalToolName must match tools.TerminalToolName in registry.go.
// Defined locally to avoid circular import (deafguard is a subpackage of tools).
const terminalToolName = "terminal"

// Mode controls how the Deaf Guard handles classified commands.
type Mode string

const (
	ModeLog     Mode = "log"     // Classify and log only — no blocking.
	ModeWarn    Mode = "warn"    // Return warning to agent for BLOCK-tier commands.
	ModeEnforce Mode = "enforce" // Hard block — agent must request operator approval.
)

// ClassificationResult holds the outcome of a command classification.
type ClassificationResult struct {
	Allowed   bool      `json:"allowed"`
	Command   string    `json:"command"`
	Category  Category  `json:"category"`
	Risk      RiskLevel `json:"risk"`
	Action    Action    `json:"action"`
	Reason    string    `json:"reason"`
	Mode      Mode      `json:"mode"`
	Tier      int       `json:"tier"`
	Timestamp int64     `json:"timestamp"`
}

// DeafGuard is the command interception engine.
type DeafGuard struct {
	enabled       bool
	mode          Mode
	disabledTiers map[int]bool
}

// New creates a new DeafGuard from configuration.
func New(cfg *config.Config) *DeafGuard {
	mode := Mode(cfg.DeafGuardMode)
	if mode != ModeLog && mode != ModeWarn && mode != ModeEnforce {
		mode = ModeLog
	}

	return &DeafGuard{
		enabled:       cfg.DeafGuardEnabled,
		mode:          mode,
		disabledTiers: make(map[int]bool),
	}
}

// IsEnabled returns whether the Deaf Guard is active.
func (dg *DeafGuard) IsEnabled() bool {
	return dg.enabled
}

// GetMode returns the current enforcement mode.
func (dg *DeafGuard) GetMode() Mode {
	return dg.mode
}

// SetMode updates the enforcement mode at runtime.
func (dg *DeafGuard) SetMode(mode Mode) {
	if mode == ModeLog || mode == ModeWarn || mode == ModeEnforce {
		dg.mode = mode
	}
}

// SetTierEnabled enables or disables a specific classification tier.
func (dg *DeafGuard) SetTierEnabled(tier int, enabled bool) {
	if enabled {
		delete(dg.disabledTiers, tier)
	} else {
		dg.disabledTiers[tier] = true
	}
}

// IsTierEnabled checks if a specific tier is active.
func (dg *DeafGuard) IsTierEnabled(tier int) bool {
	return !dg.disabledTiers[tier]
}

// terminalAction mirrors the terminal tool's action struct for JSON parsing.
type terminalAction struct {
	Input string `json:"input"`
}

// Classify evaluates a tool call and returns the classification result.
// For non-terminal tools, it returns an allow-all result.
// For terminal tools, it runs the command through the classification pipeline.
func (dg *DeafGuard) Classify(toolName string, args json.RawMessage) *ClassificationResult {
	if !dg.enabled {
		return &ClassificationResult{Allowed: true, Category: CategoryLocalUtility, Risk: RiskNone, Action: ActionLog}
	}

	// Only classify terminal commands
	if toolName != terminalToolName {
		return &ClassificationResult{Allowed: true, Category: CategoryLocalUtility, Risk: RiskNone, Action: ActionLog}
	}

	// Extract the command string from args
	var action terminalAction
	if err := json.Unmarshal(args, &action); err != nil || action.Input == "" {
		return &ClassificationResult{Allowed: true, Category: CategoryLocalUtility, Risk: RiskNone, Action: ActionLog}
	}

	return dg.classifyCommand(action.Input)
}

// classifyCommand runs the command through the three-stage classification pipeline.
func (dg *DeafGuard) classifyCommand(command string) *ClassificationResult {
	// Stage 1: Classify the raw command and each split segment.
	// The raw command is included so rules that span separators
	// (for example a fork bomb that contains `|`) still match.
	// The worst classification wins.
	segments := append([]string{command}, splitCommand(command)...)

	var worstResult *ClassificationResult
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		result := dg.classifySegment(seg)
		if worstResult == nil || actionSeverity(result.Action) > actionSeverity(worstResult.Action) {
			worstResult = result
		}
	}

	if worstResult == nil {
		worstResult = &ClassificationResult{
			Allowed:   true,
			Command:   command,
			Category:  CategoryLocalUtility,
			Risk:      RiskNone,
			Action:    ActionLog,
			Reason:    "No matching rules",
			Mode:      dg.mode,
			Tier:      9,
			Timestamp: time.Now().Unix(),
		}
	}

	worstResult.Command = command
	worstResult.Mode = dg.mode
	worstResult.Timestamp = time.Now().Unix()

	// Apply mode: in log mode, everything is allowed regardless of classification.
	switch dg.mode {
	case ModeLog:
		worstResult.Allowed = true
	case ModeWarn:
		worstResult.Allowed = worstResult.Action != ActionBlock
	case ModeEnforce:
		worstResult.Allowed = worstResult.Action == ActionLog
	}

	// Log the classification
	dg.logClassification(worstResult)

	return worstResult
}

// classifySegment evaluates a single command segment against all rules.
func (dg *DeafGuard) classifySegment(segment string) *ClassificationResult {
	for _, rule := range rules {
		tier := CategoryTier[rule.Category]
		// Skip disabled tiers
		if dg.disabledTiers[tier] {
			continue
		}
		if rule.Pattern.MatchString(segment) {
			return &ClassificationResult{
				Allowed:  false, // Explicitly blocked pending mode evaluation in classifyCommand
				Category: rule.Category,
				Risk:     rule.Risk,
				Action:   rule.Action,
				Reason:   rule.Reason,
				Tier:     tier,
			}
		}
	}

	// No rule matched — safe command
	return &ClassificationResult{
		Allowed:  true,
		Category: CategoryLocalUtility,
		Risk:     RiskNone,
		Action:   ActionLog,
		Reason:   "No matching rules — allowed",
		Tier:     9,
	}
}

// splitCommand splits a shell command on ; && || and | boundaries.
// This is intentionally simple — not a full shell parser.
func splitCommand(cmd string) []string {
	var segments []string
	var current strings.Builder
	runes := []rune(cmd)
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		// Handle quotes
		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			current.WriteRune(ch)
			continue
		}
		if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			current.WriteRune(ch)
			continue
		}

		// Only split outside of quotes
		if !inSingleQuote && !inDoubleQuote {
			if ch == ';' {
				segments = append(segments, current.String())
				current.Reset()
				continue
			}
			if ch == '|' {
				if i+1 < len(runes) && runes[i+1] == '|' {
					// ||
					segments = append(segments, current.String())
					current.Reset()
					i++ // skip second |
					continue
				}
				// Single pipe — still a boundary (the piped-to command matters)
				segments = append(segments, current.String())
				current.Reset()
				continue
			}
			if ch == '&' && i+1 < len(runes) && runes[i+1] == '&' {
				segments = append(segments, current.String())
				current.Reset()
				i++ // skip second &
				continue
			}
		}

		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	return segments
}

// actionSeverity returns a numeric severity for action comparison.
func actionSeverity(a Action) int {
	switch a {
	case ActionBlock:
		return 3
	case ActionWarn:
		return 2
	case ActionLog:
		return 1
	default:
		return 0
	}
}

// logClassification writes a structured log entry for the classification.
func (dg *DeafGuard) logClassification(result *ClassificationResult) {
	entry := logrus.WithFields(logrus.Fields{
		"component": "deaf_guard",
		"command":   truncateForLog(result.Command, 200),
		"category":  result.Category,
		"tier":      result.Tier,
		"risk":      result.Risk,
		"action":    result.Action,
		"allowed":   result.Allowed,
		"mode":      result.Mode,
		"reason":    result.Reason,
	})

	switch {
	case result.Action == ActionBlock:
		entry.Warn("deaf guard: command classified as BLOCK")
	case result.Action == ActionWarn:
		entry.Info("deaf guard: command classified as WARN")
	default:
		entry.Debug("deaf guard: command classified as LOG")
	}
}

// truncateForLog limits string length for log output (rune-safe).
func truncateForLog(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// GetConfig returns the current Deaf Guard configuration as a serializable struct.
func (dg *DeafGuard) GetConfig() map[string]interface{} {
	tiers := make(map[int]map[string]interface{})
	for cat, tier := range CategoryTier {
		info := CategoryInfo[cat]
		tiers[tier] = map[string]interface{}{
			"category":       cat,
			"name":           info.Name,
			"description":    info.Description,
			"default_action": info.DefaultAction,
			"enabled":        dg.IsTierEnabled(tier),
		}
	}

	return map[string]interface{}{
		"enabled": dg.enabled,
		"mode":    dg.mode,
		"tiers":   tiers,
	}
}
