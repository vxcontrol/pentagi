package security

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	MaxResponseLength     = 100000
	MaxThinkingLength     = 50000
	SanitizeRedacted      = "[REDACTED]"
	SanitizeTruncated     = "... [truncated]"
)

type SanitizerConfig struct {
	MaxResponseLength     int
	MaxThinkingLength     int
	RedactSecrets         bool
	RedactIPs             bool
	RedactPaths           bool
	RedactCredentials     bool
	StripANSI             bool
	NormalizeWhitespace   bool
	RemoveControlChars    bool
	ValidateUTF8          bool
}

type SanitizationResult struct {
	Original      string            `json:"original,omitempty"`
	Sanitized     string            `json:"sanitized"`
	WasTruncated  bool              `json:"was_truncated"`
	WasRedacted   bool              `json:"was_redacted"`
	Redactions    []RedactionInfo   `json:"redactions,omitempty"`
	Warnings      []string          `json:"warnings,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type RedactionInfo struct {
	Type     string `json:"type"`
	Position int    `json:"position"`
	Length   int    `json:"length"`
}

type ResponseSanitizer struct {
	config SanitizerConfig
}

var (
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|passwd|pwd|secret|api_key|apikey|api-key|token|auth|credential|private_key|private-key)\s*[=:]\s*['"]?([^\s'"<>]{8,})['"]?`),
		regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_-]{20,}`),
		regexp.MustCompile(`(?i)basic\s+[a-zA-Z0-9+/=]{20,}`),
		regexp.MustCompile(`(?i)x-api-key:\s*[a-zA-Z0-9_-]{20,}`),
		regexp.MustCompile(`(?i)authorization:\s*[a-zA-Z0-9_-]{20,}`),
		regexp.MustCompile(`-----BEGIN (?:RSA |DSA |EC |OPENSSH )?PRIVATE KEY-----`),
		regexp.MustCompile(`(?i)aws_access_key_id\s*=\s*[A-Z0-9]{20}`),
		regexp.MustCompile(`(?i)aws_secret_access_key\s*=\s*[a-zA-Z0-9+/]{40}`),
		regexp.MustCompile(`(?i)sk-[a-zA-Z0-9]{48}`),
		regexp.MustCompile(`(?i)sk_live_[a-zA-Z0-9]{24,}`),
		regexp.MustCompile(`(?i)rk_live_[a-zA-Z0-9]{24,}`),
		regexp.MustCompile(`(?i)ghp_[a-zA-Z0-9]{36}`),
		regexp.MustCompile(`(?i)gho_[a-zA-Z0-9]{36}`),
		regexp.MustCompile(`(?i)github_pat_[a-zA-Z0-9_]{22,}`),
		regexp.MustCompile(`(?i)glpat-[a-zA-Z0-9_-]{20,}`),
		regexp.MustCompile(`(?i)npm_[a-zA-Z0-9]{36}`),
		regexp.MustCompile(`(?i)eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`),
	}

	ipPattern = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)

	privateIPPattern = regexp.MustCompile(`\b(?:10\.(?:\d{1,3}\.){2}\d{1,3}|172\.(?:1[6-9]|2[0-9]|3[0-1])\.(?:\d{1,3}\.)\d{1,3}|192\.168\.(?:\d{1,3}\.)\d{1,3}|127\.(?:\d{1,3}\.){2}\d{1,3})\b`)

	pathPattern = regexp.MustCompile(`(?i)(?:/home/|/root/|/etc/|/var/|/usr/|/opt/|/tmp/|C:\\Users\\|C:\\Windows\\)[a-zA-Z0-9_/.-]+`)

	ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b[()][AB012]`)

	controlCharPattern = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

	emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

	creditCardPattern = regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`)

	maliciousResponsePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)ignore\s+(?:all\s+)?(?:previous|above)\s+(?:instructions|prompts|rules)`),
		regexp.MustCompile(`(?i)disregard\s+(?:all\s+)?(?:previous|above)\s+(?:instructions|prompts|rules)`),
		regexp.MustCompile(`(?i)forget\s+(?:all\s+)?(?:previous|above)\s+(?:instructions|prompts|rules)`),
		regexp.MustCompile(`(?i)you\s+are\s+now\s+(?:in\s+)?(?:developer|admin|root|sudo)\s+mode`),
		regexp.MustCompile(`(?i)system:\s*override`),
		regexp.MustCompile(`(?i)jailbreak`),
		regexp.MustCompile(`(?i)<\|im_start\|>`),
		regexp.MustCompile(`(?i)\[SYSTEM\]`),
		regexp.MustCompile(`(?i)\[INST\]`),
	}
)

func NewResponseSanitizer(config SanitizerConfig) *ResponseSanitizer {
	if config.MaxResponseLength <= 0 {
		config.MaxResponseLength = MaxResponseLength
	}
	if config.MaxThinkingLength <= 0 {
		config.MaxThinkingLength = MaxThinkingLength
	}
	return &ResponseSanitizer{config: config}
}

func DefaultSanitizerConfig() SanitizerConfig {
	return SanitizerConfig{
		MaxResponseLength:   MaxResponseLength,
		MaxThinkingLength:   MaxThinkingLength,
		RedactSecrets:       true,
		RedactIPs:           false,
		RedactPaths:         false,
		RedactCredentials:   true,
		StripANSI:           true,
		NormalizeWhitespace: true,
		RemoveControlChars:  true,
		ValidateUTF8:        true,
	}
}

func (s *ResponseSanitizer) Sanitize(input string) *SanitizationResult {
	result := &SanitizationResult{
		Sanitized:  input,
		Redactions: []RedactionInfo{},
		Warnings:   []string{},
		Metadata:   make(map[string]string),
	}

	if s.config.RemoveControlChars {
		result.Sanitized = s.removeControlChars(result.Sanitized)
	}

	if s.config.StripANSI {
		result.Sanitized = s.stripANSI(result.Sanitized)
	}

	if s.config.NormalizeWhitespace {
		result.Sanitized = s.normalizeWhitespace(result.Sanitized)
	}

	if s.config.ValidateUTF8 {
		result.Sanitized = s.validateUTF8(result.Sanitized)
	}

	if s.config.RedactCredentials {
		result = s.redactPatterns(result, secretPatterns, "credential")
		result = s.redactPattern(result, creditCardPattern, "credit_card")
	}

	if s.config.RedactIPs {
		result = s.redactPattern(result, ipPattern, "ip_address")
	}

	if s.config.RedactPaths {
		result = s.redactPattern(result, pathPattern, "file_path")
	}

	if len(result.Sanitized) > s.config.MaxResponseLength {
		result.Sanitized = result.Sanitized[:s.config.MaxResponseLength] + SanitizeTruncated
		result.WasTruncated = true
		result.Warnings = append(result.Warnings, "Response was truncated due to length limit")
	}

	result.Metadata["original_length"] = fmt.Sprintf("%d", len(input))
	result.Metadata["sanitized_length"] = fmt.Sprintf("%d", len(result.Sanitized))
	result.Metadata["redaction_count"] = fmt.Sprintf("%d", len(result.Redactions))

	return result
}

func (s *ResponseSanitizer) SanitizeThinking(thinking string) *SanitizationResult {
	config := s.config
	config.MaxResponseLength = s.config.MaxThinkingLength
	tempSanitizer := NewResponseSanitizer(config)
	return tempSanitizer.Sanitize(thinking)
}

func (s *ResponseSanitizer) SanitizeTargetResponse(response string, targetHost string) *SanitizationResult {
	result := s.Sanitize(response)

	result = s.detectInjectionAttempts(result)

	if targetHost != "" && s.config.RedactIPs {
		privateIPPattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(targetHost) + `\b`)
		if privateIPPattern.MatchString(result.Sanitized) {
			result.Sanitized = privateIPPattern.ReplaceAllString(result.Sanitized, "[TARGET_HOST]")
			result.WasRedacted = true
			result.Redactions = append(result.Redactions, RedactionInfo{
				Type: "target_host",
			})
		}
	}

	return result
}

func (s *ResponseSanitizer) redactPatterns(result *SanitizationResult, patterns []*regexp.Regexp, redactionType string) *SanitizationResult {
	for _, pattern := range patterns {
		result = s.redactPattern(result, pattern, redactionType)
	}
	return result
}

func (s *ResponseSanitizer) redactPattern(result *SanitizationResult, pattern *regexp.Regexp, redactionType string) *SanitizationResult {
	matches := pattern.FindAllStringIndex(result.Sanitized, -1)
	if len(matches) == 0 {
		return result
	}

	result.WasRedacted = true
	for _, match := range matches {
		result.Redactions = append(result.Redactions, RedactionInfo{
			Type:     redactionType,
			Position: match[0],
			Length:   match[1] - match[0],
		})
	}

	result.Sanitized = pattern.ReplaceAllStringFunc(result.Sanitized, func(match string) string {
		if redactionType == "credential" {
			if len(match) > 20 {
				return match[:10] + SanitizeRedacted + match[len(match)-5:]
			}
		}
		return SanitizeRedacted
	})

	return result
}

func (s *ResponseSanitizer) detectInjectionAttempts(result *SanitizationResult) *SanitizationResult {
	for _, pattern := range maliciousResponsePatterns {
		if pattern.MatchString(result.Sanitized) {
			result.Warnings = append(result.Warnings, "Potential prompt injection attempt detected")
			result.Metadata["injection_detected"] = "true"
			break
		}
	}
	return result
}

func (s *ResponseSanitizer) stripANSI(input string) string {
	return ansiPattern.ReplaceAllString(input, "")
}

func (s *ResponseSanitizer) normalizeWhitespace(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")

	lines := strings.Split(input, "\n")
	var result []string
	prevEmpty := false

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed == "" {
			if !prevEmpty {
				result = append(result, "")
				prevEmpty = true
			}
		} else {
			result = append(result, trimmed)
			prevEmpty = false
		}
	}

	return strings.Join(result, "\n")
}

func (s *ResponseSanitizer) removeControlChars(input string) string {
	return controlCharPattern.ReplaceAllString(input, "")
}

func (s *ResponseSanitizer) validateUTF8(input string) string {
	runes := []rune(input)
	validRunes := make([]rune, 0, len(runes))

	for _, r := range runes {
		if r == unicode.ReplacementChar {
			continue
		}
		if !unicode.IsPrint(r) && r != '\n' && r != '\t' && r != '\r' {
			continue
		}
		validRunes = append(validRunes, r)
	}

	return string(validRunes)
}

func (r *SanitizationResult) ToJSON() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal sanitization result: %s"}`, err)
	}
	return string(data)
}

func (r *SanitizationResult) IsClean() bool {
	return !r.WasRedacted && !r.WasTruncated && len(r.Warnings) == 0
}

func SanitizeCommand(command string) string {
	dangerousPatterns := []string{
		";", "&&", "||", "|", "`", "$(", "${",
		">", ">>", "<", "<<",
		"rm -rf", "mkfs", "dd if=",
		":(){ :|:& };:",
		"chmod 777", "chown root",
	}

	result := command
	for _, pattern := range dangerousPatterns {
		if strings.Contains(result, pattern) {
			result = strings.ReplaceAll(result, pattern, "")
		}
	}

	return strings.TrimSpace(result)
}

func ValidateTargetHost(host string) error {
	dangerousHosts := []string{
		"localhost", "127.0.0.1", "0.0.0.0",
		"169.254.169.254",
		"internal", "private",
	}

	hostLower := strings.ToLower(host)
	for _, dangerous := range dangerousHosts {
		if strings.Contains(hostLower, dangerous) {
			return fmt.Errorf("potentially dangerous target host: %s", host)
		}
	}

	return nil
}
