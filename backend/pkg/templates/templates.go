package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"path"
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
		"GraphitiSearchToolName",
		"SummarizationToolName",
		"SummarizedContentPrefix",
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
