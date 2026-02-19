package templates

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

//go:embed prompts/*.tmpl
var promptTemplates embed.FS

//go:embed graphiti/*.tmpl
var graphitiTemplates embed.FS

var ErrTemplateNotFound = errors.New("template not found")

type PromptType string

const (
	PromptTypePrimaryAgent          PromptType = "primary_agent"           // orchestrates subtask execution using AI agents
	PromptTypeAssistant             PromptType = "assistant"               // interactive AI assistant for user conversations
	PromptTypePentester             PromptType = "pentester"               // executes security tests and vulnerability scanning
	PromptTypeQuestionPentester     PromptType = "question_pentester"      // human input requesting penetration testing
	PromptTypeCoder                 PromptType = "coder"                   // develops exploits and custom security tools
	PromptTypeQuestionCoder         PromptType = "question_coder"          // human input requesting code development
	PromptTypeInstaller             PromptType = "installer"               // sets up testing environment and tools
	PromptTypeQuestionInstaller     PromptType = "question_installer"      // human input requesting system installation
	PromptTypeSearcher              PromptType = "searcher"                // gathers intelligence from web sources
	PromptTypeQuestionSearcher      PromptType = "question_searcher"       // human input requesting information search
	PromptTypeMemorist              PromptType = "memorist"                // retrieves knowledge from vector memory store
	PromptTypeQuestionMemorist      PromptType = "question_memorist"       // human input querying past experiences
	PromptTypeAdviser               PromptType = "adviser"                 // provides security recommendations and guidance
	PromptTypeQuestionAdviser       PromptType = "question_adviser"        // human input seeking expert advice
	PromptTypeGenerator             PromptType = "generator"               // creates structured subtask breakdown
	PromptTypeSubtasksGenerator     PromptType = "subtasks_generator"      // human input for task decomposition
	PromptTypeRefiner               PromptType = "refiner"                 // optimizes and adjusts planned subtasks
	PromptTypeSubtasksRefiner       PromptType = "subtasks_refiner"        // human input for task refinement
	PromptTypeReporter              PromptType = "reporter"                // generates comprehensive security reports
	PromptTypeTaskReporter          PromptType = "task_reporter"           // human input for result documentation
	PromptTypeReflector             PromptType = "reflector"               // analyzes outcomes and suggests improvements
	PromptTypeQuestionReflector     PromptType = "question_reflector"      // human input for self-assessment
	PromptTypeEnricher              PromptType = "enricher"                // adds context and details to requests
	PromptTypeQuestionEnricher      PromptType = "question_enricher"       // human input for context enhancement
	PromptTypeToolCallFixer         PromptType = "toolcall_fixer"          // corrects malformed security tool commands
	PromptTypeInputToolCallFixer    PromptType = "input_toolcall_fixer"    // human input for tool argument fixing
	PromptTypeSummarizer            PromptType = "summarizer"              // condenses long conversations and results
	PromptTypeImageChooser          PromptType = "image_chooser"           // selects appropriate Docker containers
	PromptTypeLanguageChooser       PromptType = "language_chooser"        // determines user's preferred language
	PromptTypeFlowDescriptor        PromptType = "flow_descriptor"         // generates flow titles from user requests
	PromptTypeTaskDescriptor        PromptType = "task_descriptor"         // generates task titles from user requests
	PromptTypeExecutionLogs         PromptType = "execution_logs"          // formats execution history for display
	PromptTypeFullExecutionContext  PromptType = "full_execution_context"  // prepares complete context for summarization
	PromptTypeShortExecutionContext PromptType = "short_execution_context" // prepares minimal context for quick processing
	PromptTypeToolCallIDCollector   PromptType = "tool_call_id_collector"  // requests function call to collect tool call ID sample
	PromptTypeToolCallIDDetector    PromptType = "tool_call_id_detector"   // analyzes tool call ID samples to detect pattern template
)

var PromptVariables = map[PromptType][]string{
	PromptTypePrimaryAgent: {
		"FinalyToolName",
		"SearchToolName",
		"PentesterToolName",
		"CoderToolName",
		"AdviceToolName",
		"MemoristToolName",
		"MaintenanceToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"AskUserToolName",
		"AskUserEnabled",
		"ExecutionContext",
		"Lang",
		"DockerImage",
		"CurrentTime",
		"ToolPlaceholder",
	},
	PromptTypeAssistant: {
		"SearchToolName",
		"PentesterToolName",
		"CoderToolName",
		"AdviceToolName",
		"MemoristToolName",
		"MaintenanceToolName",
		"TerminalToolName",
		"FileToolName",
		"GoogleToolName",
		"DuckDuckGoToolName",
		"TavilyToolName",
		"TraversaalToolName",
		"PerplexityToolName",
		"BrowserToolName",
		"SearchInMemoryToolName",
		"SearchGuideToolName",
		"SearchAnswerToolName",
		"SearchCodeToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"UseAgents",
		"DockerImage",
		"Cwd",
		"ContainerPorts",
		"ExecutionContext",
		"Lang",
		"CurrentTime",
	},
	PromptTypePentester: {
		"HackResultToolName",
		"SearchGuideToolName",
		"StoreGuideToolName",
		"GraphitiEnabled",
		"GraphitiSearchToolName",
		"SearchToolName",
		"CoderToolName",
		"AdviceToolName",
		"MemoristToolName",
		"MaintenanceToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"IsDefaultDockerImage",
		"DockerImage",
		"Cwd",
		"ContainerPorts",
		"ExecutionContext",
		"Lang",
		"CurrentTime",
		"ToolPlaceholder",
	},
	PromptTypeQuestionPentester: {
		"Question",
	},
	PromptTypeCoder: {
		"CodeResultToolName",
		"SearchCodeToolName",
		"StoreCodeToolName",
		"GraphitiEnabled",
		"GraphitiSearchToolName",
		"SearchToolName",
		"AdviceToolName",
		"MemoristToolName",
		"MaintenanceToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"DockerImage",
		"Cwd",
		"ContainerPorts",
		"ExecutionContext",
		"Lang",
		"CurrentTime",
		"ToolPlaceholder",
	},
	PromptTypeQuestionCoder: {
		"Question",
	},
	PromptTypeInstaller: {
		"MaintenanceResultToolName",
		"SearchGuideToolName",
		"StoreGuideToolName",
		"SearchToolName",
		"AdviceToolName",
		"MemoristToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"DockerImage",
		"Cwd",
		"ContainerPorts",
		"ExecutionContext",
		"Lang",
		"CurrentTime",
		"ToolPlaceholder",
	},
	PromptTypeQuestionInstaller: {
		"Question",
	},
	PromptTypeSearcher: {
		"SearchResultToolName",
		"SearchAnswerToolName",
		"StoreAnswerToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"ExecutionContext",
		"Lang",
		"CurrentTime",
		"ToolPlaceholder",
	},
	PromptTypeQuestionSearcher: {
		"Question",
		"Task",
		"Subtask",
	},
	PromptTypeMemorist: {
		"MemoristResultToolName",
		"GraphitiEnabled",
		"GraphitiSearchToolName",
		"TerminalToolName",
		"FileToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"DockerImage",
		"Cwd",
		"ContainerPorts",
		"ExecutionContext",
		"Lang",
		"CurrentTime",
		"ToolPlaceholder",
	},
	PromptTypeQuestionMemorist: {
		"Question",
		"Task",
		"Subtask",
		"ExecutionDetails",
	},
	PromptTypeAdviser: {
		"ExecutionContext",
		"CurrentTime",
	},
	PromptTypeQuestionAdviser: {
		"Question",
		"Code",
		"Output",
		"Enriches",
	},
	PromptTypeGenerator: {
		"SubtaskListToolName",
		"SearchToolName",
		"TerminalToolName",
		"FileToolName",
		"BrowserToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"DockerImage",
		"Lang",
		"CurrentTime",
		"N",
		"ToolPlaceholder",
	},
	PromptTypeSubtasksGenerator: {
		"Task",
		"Tasks",
		"Subtasks",
	},
	PromptTypeRefiner: {
		"SubtaskPatchToolName",
		"SearchToolName",
		"TerminalToolName",
		"FileToolName",
		"BrowserToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"DockerImage",
		"Lang",
		"CurrentTime",
		"N",
		"ToolPlaceholder",
	},
	PromptTypeSubtasksRefiner: {
		"Task",
		"Tasks",
		"PlannedSubtasks",
		"CompletedSubtasks",
		"ExecutionLogs",
		"ExecutionState",
	},
	PromptTypeReporter: {
		"ReportResultToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"Lang",
		"N",
		"ToolPlaceholder",
	},
	PromptTypeTaskReporter: {
		"Task",
		"Tasks",
		"CompletedSubtasks",
		"PlannedSubtasks",
		"ExecutionLogs",
		"ExecutionState",
	},
	PromptTypeReflector: {
		"BarrierTools",
		"CurrentTime",
		"ExecutionContext",
		"Request",
	},
	PromptTypeQuestionReflector: {
		"Message",
		"BarrierToolNames",
	},
	PromptTypeEnricher: {
		"EnricherToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
		"ExecutionContext",
		"Lang",
		"CurrentTime",
		"ToolPlaceholder",
	},
	PromptTypeQuestionEnricher: {
		"Question",
		"Code",
		"Output",
	},
	PromptTypeToolCallFixer: {},
	PromptTypeInputToolCallFixer: {
		"ToolCallName",
		"ToolCallArgs",
		"ToolCallSchema",
		"ToolCallError",
	},
	PromptTypeSummarizer: {
		"TaskID",
		"SubtaskID",
		"CurrentTime",
		"SummarizedContentPrefix",
	},
	PromptTypeFlowDescriptor: {
		"Input",
		"Lang",
		"CurrentTime",
		"N",
	},
	PromptTypeTaskDescriptor: {
		"Input",
		"Lang",
		"CurrentTime",
		"N",
	},
	PromptTypeExecutionLogs: {
		"MsgLogs",
	},
	PromptTypeFullExecutionContext: {
		"Task",
		"Tasks",
		"CompletedSubtasks",
		"Subtask",
		"PlannedSubtasks",
	},
	PromptTypeShortExecutionContext: {
		"Task",
		"Tasks",
		"CompletedSubtasks",
		"Subtask",
		"PlannedSubtasks",
	},
	PromptTypeImageChooser: {
		"DefaultImage",
		"DefaultImageForPentest",
		"Input",
	},
	PromptTypeLanguageChooser: {
		"Input",
	},
	PromptTypeToolCallIDCollector: {
		"FunctionName",
		"RandomContext",
	},
	PromptTypeToolCallIDDetector: {
		"FunctionName",
		"Samples",
		"PreviousAttempts",
	},
}

type Prompt struct {
	Type      PromptType
	Template  string
	Variables []string
}

type AgentPrompt struct {
	System Prompt
}

type AgentPrompts struct {
	System Prompt
	Human  Prompt
}

type AgentsPrompts struct {
	PrimaryAgent  AgentPrompt
	Assistant     AgentPrompt
	Pentester     AgentPrompts
	Coder         AgentPrompts
	Installer     AgentPrompts
	Searcher      AgentPrompts
	Memorist      AgentPrompts
	Adviser       AgentPrompts
	Generator     AgentPrompts
	Refiner       AgentPrompts
	Reporter      AgentPrompts
	Reflector     AgentPrompts
	Enricher      AgentPrompts
	ToolCallFixer AgentPrompts
	Summarizer    AgentPrompt
}

type ToolsPrompts struct {
	GetFlowDescription       Prompt
	GetTaskDescription       Prompt
	GetExecutionLogs         Prompt
	GetFullExecutionContext  Prompt
	GetShortExecutionContext Prompt
	ChooseDockerImage        Prompt
	ChooseUserLanguage       Prompt
	CollectToolCallID        Prompt
	DetectToolCallIDPattern  Prompt
}

type DefaultPrompts struct {
	AgentsPrompts AgentsPrompts
	ToolsPrompts  ToolsPrompts
}

func GetDefaultPrompts() (*DefaultPrompts, error) {
	prompts, err := promptTemplates.ReadDir("prompts")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates: %w", err)
	}

	promptsMap := make(PromptsMap)
	for _, prompt := range prompts {
		promptBytes, err := promptTemplates.ReadFile(path.Join("prompts", prompt.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read template: %w", err)
		}

		promptName := strings.TrimSuffix(prompt.Name(), ".tmpl")
		promptsMap[PromptType(promptName)] = string(promptBytes)
	}

	getPrompt := func(promptType PromptType) Prompt {
		return Prompt{
			Type:      promptType,
			Template:  promptsMap[promptType],
			Variables: PromptVariables[promptType],
		}
	}

	return &DefaultPrompts{
		AgentsPrompts: AgentsPrompts{
			PrimaryAgent: AgentPrompt{
				System: getPrompt(PromptTypePrimaryAgent),
			},
			Assistant: AgentPrompt{
				System: getPrompt(PromptTypeAssistant),
			},
			Pentester: AgentPrompts{
				System: getPrompt(PromptTypePentester),
				Human:  getPrompt(PromptTypeQuestionPentester),
			},
			Coder: AgentPrompts{
				System: getPrompt(PromptTypeCoder),
				Human:  getPrompt(PromptTypeQuestionCoder),
			},
			Installer: AgentPrompts{
				System: getPrompt(PromptTypeInstaller),
				Human:  getPrompt(PromptTypeQuestionInstaller),
			},
			Searcher: AgentPrompts{
				System: getPrompt(PromptTypeSearcher),
				Human:  getPrompt(PromptTypeQuestionSearcher),
			},
			Memorist: AgentPrompts{
				System: getPrompt(PromptTypeMemorist),
				Human:  getPrompt(PromptTypeQuestionMemorist),
			},
			Adviser: AgentPrompts{
				System: getPrompt(PromptTypeAdviser),
				Human:  getPrompt(PromptTypeQuestionAdviser),
			},
			Generator: AgentPrompts{
				System: getPrompt(PromptTypeGenerator),
				Human:  getPrompt(PromptTypeSubtasksGenerator),
			},
			Refiner: AgentPrompts{
				System: getPrompt(PromptTypeRefiner),
				Human:  getPrompt(PromptTypeSubtasksRefiner),
			},
			Reporter: AgentPrompts{
				System: getPrompt(PromptTypeReporter),
				Human:  getPrompt(PromptTypeTaskReporter),
			},
			Reflector: AgentPrompts{
				System: getPrompt(PromptTypeReflector),
				Human:  getPrompt(PromptTypeQuestionReflector),
			},
			Enricher: AgentPrompts{
				System: getPrompt(PromptTypeEnricher),
				Human:  getPrompt(PromptTypeQuestionEnricher),
			},
			ToolCallFixer: AgentPrompts{
				System: getPrompt(PromptTypeToolCallFixer),
				Human:  getPrompt(PromptTypeInputToolCallFixer),
			},
			Summarizer: AgentPrompt{
				System: getPrompt(PromptTypeSummarizer),
			},
		},
		ToolsPrompts: ToolsPrompts{
			GetFlowDescription:       getPrompt(PromptTypeFlowDescriptor),
			GetTaskDescription:       getPrompt(PromptTypeTaskDescriptor),
			GetExecutionLogs:         getPrompt(PromptTypeExecutionLogs),
			GetFullExecutionContext:  getPrompt(PromptTypeFullExecutionContext),
			GetShortExecutionContext: getPrompt(PromptTypeShortExecutionContext),
			ChooseDockerImage:        getPrompt(PromptTypeImageChooser),
			ChooseUserLanguage:       getPrompt(PromptTypeLanguageChooser),
			CollectToolCallID:        getPrompt(PromptTypeToolCallIDCollector),
			DetectToolCallIDPattern:  getPrompt(PromptTypeToolCallIDDetector),
		},
	}, nil
}

type PromptsMap map[PromptType]string

type Prompter interface {
	GetTemplate(promptType PromptType) (string, error)
	RenderTemplate(promptType PromptType, params any) (string, error)
	DumpTemplates() ([]byte, error)
}

type flowPrompter struct {
	prompts PromptsMap
}

func NewFlowPrompter(prompts PromptsMap) Prompter {
	return &flowPrompter{prompts: prompts}
}

func (fp *flowPrompter) GetTemplate(promptType PromptType) (string, error) {
	if prompt, ok := fp.prompts[promptType]; ok {
		return prompt, nil
	}

	return "", ErrTemplateNotFound
}

func (fp *flowPrompter) RenderTemplate(promptType PromptType, params any) (string, error) {
	prompt, err := fp.GetTemplate(promptType)
	if err != nil {
		return "", err
	}

	return RenderPrompt(string(promptType), prompt, params)
}

func (fp *flowPrompter) DumpTemplates() ([]byte, error) {
	blob, err := json.Marshal(fp.prompts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal templates: %w", err)
	}

	return blob, nil
}

type defaultPrompter struct {
}

func NewDefaultPrompter() Prompter {
	return &defaultPrompter{}
}

func (dp *defaultPrompter) GetTemplate(promptType PromptType) (string, error) {
	promptPath := path.Join("prompts", fmt.Sprintf("%s.tmpl", promptType))
	promptBytes, err := promptTemplates.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %v: %w", err, ErrTemplateNotFound)
	}

	return string(promptBytes), nil
}

func (dp *defaultPrompter) RenderTemplate(promptType PromptType, params any) (string, error) {
	prompt, err := dp.GetTemplate(promptType)
	if err != nil {
		return "", err
	}

	return RenderPrompt(string(promptType), prompt, params)
}

func (dp *defaultPrompter) DumpTemplates() ([]byte, error) {
	prompts, err := promptTemplates.ReadDir("prompts")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates: %w", err)
	}

	promptsMap := make(PromptsMap)
	for _, prompt := range prompts {
		promptBytes, err := promptTemplates.ReadFile(path.Join("prompts", prompt.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read template: %w", err)
		}

		promptName := strings.TrimSuffix(prompt.Name(), ".tmpl")
		promptsMap[PromptType(promptName)] = string(promptBytes)
	}

	blob, err := json.Marshal(promptsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal templates: %w", err)
	}

	return blob, nil
}

func RenderPrompt(name, prompt string, params any) (string, error) {
	t, err := template.New(string(name)).Parse(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, params); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ReadGraphitiTemplate reads a Graphiti template by name
func ReadGraphitiTemplate(name string) (string, error) {
	templateBytes, err := graphitiTemplates.ReadFile(path.Join("graphiti", name))
	if err != nil {
		return "", fmt.Errorf("failed to read graphiti template %s: %w", name, err)
	}
	return string(templateBytes), nil
}

// String pattern template format:
// - Literal parts: any text outside curly braces
// - Random parts: {r:LENGTH:CHARSET}
//   - LENGTH: number of characters to generate
//   - CHARSET: character set type
//     - d, digit: [0-9]
//     - l, lower: [a-z]
//     - u, upper: [A-Z]
//     - a, alpha: [a-zA-Z]
//     - x, alnum: [a-zA-Z0-9]
//     - h, hex: [0-9a-f]
//     - H, HEX: [0-9A-F]
//     - b, base62: [0-9A-Za-z]
// - Function placeholder: {f}
//   - Represents the function/tool name
//   - Used when tool call IDs contain the function name
//
// Examples:
//   - "toolu_{r:24:b}" → "toolu_013wc5CxNCjWGN2rsAR82rJK"
//   - "call_{r:24:x}" → "call_Z8ofZnYOCeOnpu0h2auwOgeR"
//   - "chatcmpl-tool-{r:32:h}" → "chatcmpl-tool-23c5c0da71854f9bbd8774f7d0113a69"
//   - "{f}:{r:1:d}" with function="get_number" → "get_number:0"

const (
	charsetDigit  = "0123456789"
	charsetLower  = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetAlpha  = charsetLower + charsetUpper
	charsetAlnum  = charsetDigit + charsetAlpha
	charsetHex    = "0123456789abcdef"
	charsetHexUp  = "0123456789ABCDEF"
	charsetBase62 = charsetDigit + charsetUpper + charsetLower
)

var patternRegex = regexp.MustCompile(`\{r:(\d+):(d|digit|l|lower|u|upper|a|alpha|x|alnum|h|hex|H|HEX|b|base62)\}|\{f\}`)

type patternPart struct {
	literal    string
	isRandom   bool
	isFunction bool
	length     int
	charset    string
}

// getCharset returns the character set for a given charset name
func getCharset(name string) string {
	switch name {
	case "d", "digit":
		return charsetDigit
	case "l", "lower":
		return charsetLower
	case "u", "upper":
		return charsetUpper
	case "a", "alpha":
		return charsetAlpha
	case "x", "alnum":
		return charsetAlnum
	case "h", "hex":
		return charsetHex
	case "H", "HEX":
		return charsetHexUp
	case "b", "base62":
		return charsetBase62
	default:
		return charsetAlnum // fallback
	}
}

// parsePattern parses a pattern string into parts
func parsePattern(pattern string) []patternPart {
	var parts []patternPart
	lastIndex := 0

	matches := patternRegex.FindAllStringSubmatchIndex(pattern, -1)
	for _, match := range matches {
		// Add literal part before this match
		if match[0] > lastIndex {
			parts = append(parts, patternPart{
				literal:    pattern[lastIndex:match[0]],
				isRandom:   false,
				isFunction: false,
			})
		}

		matchedText := pattern[match[0]:match[1]]

		// Check if it's a function placeholder
		if matchedText == "{f}" {
			parts = append(parts, patternPart{
				isFunction: true,
			})
		} else {
			// Parse random part
			length, _ := strconv.Atoi(pattern[match[2]:match[3]])
			charsetName := pattern[match[4]:match[5]]
			parts = append(parts, patternPart{
				isRandom: true,
				length:   length,
				charset:  getCharset(charsetName),
			})
		}

		lastIndex = match[1]
	}

	// Add remaining literal part
	if lastIndex < len(pattern) {
		parts = append(parts, patternPart{
			literal:    pattern[lastIndex:],
			isRandom:   false,
			isFunction: false,
		})
	}

	return parts
}

// generateRandomString generates a random string of specified length using the given charset
func generateRandomString(length int, charset string) string {
	if length == 0 || charset == "" {
		return ""
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// Fallback to first character if random fails (should never happen)
			result[i] = charset[0]
		} else {
			result[i] = charset[num.Int64()]
		}
	}

	return string(result)
}

// GenerateFromPattern generates a random string matching the given pattern template.
// This function never returns an error - it uses fallback values for invalid patterns.
//
// Pattern format: literal text with {r:LENGTH:CHARSET} for random parts and {f} for function name
// Example: "toolu_{r:24:base62}" → "toolu_xK9pQw2mN5vR8tY7uI6oP3zA"
// Example: "{f}:{r:1:d}" with functionName="get_number" → "get_number:0"
func GenerateFromPattern(pattern string, functionName string) string {
	parts := parsePattern(pattern)
	var result strings.Builder

	for _, part := range parts {
		if part.isRandom {
			result.WriteString(generateRandomString(part.length, part.charset))
		} else if part.isFunction {
			if functionName != "" {
				result.WriteString(functionName)
			} else {
				result.WriteString("function")
			}
		} else {
			result.WriteString(part.literal)
		}
	}

	return result.String()
}

// PatternSample represents a sample value with optional function name for pattern validation
type PatternSample struct {
	Value        string
	FunctionName string
}

// PatternValidationError represents a validation error for a specific value
type PatternValidationError struct {
	Value    string
	Position int
	Expected string
	Got      string
	Message  string
}

func (e *PatternValidationError) Error() string {
	if e.Position >= 0 {
		return fmt.Sprintf("validation failed for '%s' at position %d: expected %s, got '%s': %s",
			e.Value, e.Position, e.Expected, e.Got, e.Message)
	}
	return fmt.Sprintf("validation failed for '%s': %s", e.Value, e.Message)
}

// ValidatePattern validates that all provided samples match the given pattern template.
// Returns a detailed error if any sample doesn't match, nil if all samples are valid.
//
// Pattern format: literal text with {r:LENGTH:CHARSET} for random parts and {f} for function name
// Example: ValidatePattern("call_{r:24:alnum}", []PatternSample{{Value: "call_abc123..."}})
func ValidatePattern(pattern string, samples []PatternSample) error {
	if len(samples) == 0 {
		return nil
	}

	parts := parsePattern(pattern)

	// Validate each sample
	for _, sample := range samples {
		value := sample.Value
		functionName := sample.FunctionName

		// Build expected length and regex pattern for this specific sample
		var expectedLen int
		var regexParts []string

		for _, part := range parts {
			if part.isRandom {
				expectedLen += part.length
				// Build character class from charset
				charClass := buildCharClass(part.charset)
				regexParts = append(regexParts, fmt.Sprintf("%s{%d}", charClass, part.length))
			} else if part.isFunction {
				if functionName != "" {
					expectedLen += len(functionName)
					regexParts = append(regexParts, regexp.QuoteMeta(functionName))
				} else {
					// Fallback if no function name provided
					expectedLen += len("function")
					regexParts = append(regexParts, regexp.QuoteMeta("function"))
				}
			} else {
				expectedLen += len(part.literal)
				regexParts = append(regexParts, regexp.QuoteMeta(part.literal))
			}
		}

		regexPattern := "^" + strings.Join(regexParts, "") + "$"
		re := regexp.MustCompile(regexPattern)

		// Check length
		if len(value) != expectedLen {
			return &PatternValidationError{
				Value:    value,
				Position: -1,
				Expected: fmt.Sprintf("length %d", expectedLen),
				Got:      fmt.Sprintf("length %d", len(value)),
				Message:  fmt.Sprintf("incorrect length: expected %d, got %d", expectedLen, len(value)),
			}
		}

		// Check pattern match
		if !re.MatchString(value) {
			// Find the exact position where it fails
			pos := findMismatchPosition(value, parts, functionName)
			part := getPartAtPosition(parts, pos, functionName)

			var expected string
			if part.isRandom {
				expected = fmt.Sprintf("character from charset [%s]", describeCharset(part.charset))
			} else if part.isFunction {
				if functionName != "" {
					expected = fmt.Sprintf("function name '%s'", functionName)
				} else {
					expected = "function name"
				}
			} else {
				expected = fmt.Sprintf("'%s'", part.literal)
			}

			got := ""
			if pos < len(value) {
				got = string(value[pos])
			}

			return &PatternValidationError{
				Value:    value,
				Position: pos,
				Expected: expected,
				Got:      got,
				Message:  "pattern mismatch",
			}
		}
	}

	return nil
}

// buildCharClass builds a regex character class from a charset string
func buildCharClass(charset string) string {
	// Optimize for common charsets
	switch charset {
	case charsetDigit:
		return `\d`
	case charsetLower:
		return `[a-z]`
	case charsetUpper:
		return `[A-Z]`
	case charsetAlpha:
		return `[a-zA-Z]`
	case charsetAlnum:
		return `[a-zA-Z0-9]`
	case charsetHex:
		return `[0-9a-f]`
	case charsetHexUp:
		return `[0-9A-F]`
	case charsetBase62:
		return `[0-9A-Za-z]`
	default:
		// Build custom character class
		return `[` + regexp.QuoteMeta(charset) + `]`
	}
}

// describeCharset returns a human-readable description of a charset
func describeCharset(charset string) string {
	switch charset {
	case charsetDigit:
		return "0-9"
	case charsetLower:
		return "a-z"
	case charsetUpper:
		return "A-Z"
	case charsetAlpha:
		return "a-zA-Z"
	case charsetAlnum:
		return "a-zA-Z0-9"
	case charsetHex:
		return "0-9a-f"
	case charsetHexUp:
		return "0-9A-F"
	case charsetBase62:
		return "0-9A-Za-z"
	default:
		return charset
	}
}

// findMismatchPosition finds the first position where value doesn't match the pattern
func findMismatchPosition(value string, parts []patternPart, functionName string) int {
	pos := 0

	for _, part := range parts {
		if part.isRandom {
			// Check each character against charset
			for i := 0; i < part.length && pos < len(value); i++ {
				if !strings.ContainsRune(part.charset, rune(value[pos])) {
					return pos
				}
				pos++
			}
		} else if part.isFunction {
			// Check function name match
			fn := functionName
			if fn == "" {
				fn = "function"
			}
			for i := 0; i < len(fn) && pos < len(value); i++ {
				if value[pos] != fn[i] {
					return pos
				}
				pos++
			}
		} else {
			// Check literal match
			for i := 0; i < len(part.literal) && pos < len(value); i++ {
				if value[pos] != part.literal[i] {
					return pos
				}
				pos++
			}
		}
	}

	return pos
}

// getPartAtPosition returns the pattern part at the given position in the generated string
func getPartAtPosition(parts []patternPart, position int, functionName string) patternPart {
	pos := 0

	for _, part := range parts {
		var length int
		if part.isRandom {
			length = part.length
		} else if part.isFunction {
			if functionName != "" {
				length = len(functionName)
			} else {
				length = len("function")
			}
		} else {
			length = len(part.literal)
		}

		if position < pos+length {
			return part
		}
		pos += length
	}

	// Return last part if position is beyond
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return patternPart{}
}
