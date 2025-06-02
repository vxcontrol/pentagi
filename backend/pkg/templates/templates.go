package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"text/template"
)

//go:embed prompts/*.tmpl
var promptTemplates embed.FS

var ErrTemplateNotFound = errors.New("template not found")

type PromptType string

const (
	PromptTypeExecutionLogs         PromptType = "execution_logs"          // user prompt to get execution logs for the task
	PromptTypeFullExecutionContext  PromptType = "full_execution_context"  // user prompt to get context to be summarized
	PromptTypeShortExecutionContext PromptType = "short_execution_context" // user prompt to get context to be passed as is
	PromptTypeQuestionEnricher      PromptType = "question_enricher"       // user prompt to enrich the question with more details
	PromptTypeQuestionAdviser       PromptType = "question_adviser"        // user prompt to give advice to the question
	PromptTypeQuestionCoder         PromptType = "question_coder"          // user prompt to write a code to solve the subtask issue
	PromptTypeQuestionInstaller     PromptType = "question_installer"      // user prompt to install and configure the infrastructure
	PromptTypeQuestionMemorist      PromptType = "question_memorist"       // user prompt to get information from long-term memory
	PromptTypeQuestionPentester     PromptType = "question_pentester"      // user prompt to perform penetration testing
	PromptTypeQuestionSearcher      PromptType = "question_searcher"       // user prompt to search the internet for more information
	PromptTypeQuestionReflector     PromptType = "question_reflector"      // user prompt to reflect on the invalid tool call message
	PromptTypeInputToolCallFixer    PromptType = "input_toolcall_fixer"    // user prompt to fix the tool call arguments
	PromptTypeAssistant             PromptType = "assistant"               // perform assistant prompt as an agent for the flow
	PromptTypePrimaryAgent          PromptType = "primary_agent"           // perform primary agent chain for subtask
	PromptTypeFlowDescriptor        PromptType = "flow_descriptor"         // write the flow title by user input
	PromptTypeTaskDescriptor        PromptType = "task_descriptor"         // write the task title by user input
	PromptTypeImageChooser          PromptType = "image_chooser"           // choose the docker image for the flow
	PromptTypeLanguageChooser       PromptType = "language_chooser"        // choose the language for the flow
	PromptTypeTaskReporter          PromptType = "task_reporter"           // validate and write result of the task
	PromptTypeToolCallFixer         PromptType = "toolcall_fixer"          // fix the tool call arguments
	PromptTypeReporter              PromptType = "reporter"                // common system prompt for result writer
	PromptTypeSubtasksGenerator     PromptType = "subtasks_generator"      // user prompt to generate subtasks for the task
	PromptTypeGenerator             PromptType = "generator"               // common system prompt for subtask generator
	PromptTypeSubtasksRefiner       PromptType = "subtasks_refiner"        // user prompt to refine subtasks for the task
	PromptTypeRefiner               PromptType = "refiner"                 // common system prompt for subtask refiner
	PromptTypeEnricher              PromptType = "enricher"                // enrich the question with more details
	PromptTypeReflector             PromptType = "reflector"               // self-reflection analog for completion result
	PromptTypeAdviser               PromptType = "adviser"                 // give advice to the question and context
	PromptTypeCoder                 PromptType = "coder"                   // write a code to solve the subtask issue
	PromptTypeInstaller             PromptType = "installer"               // install and configure the infrastructure
	PromptTypePentester             PromptType = "pentester"               // perform penetration testing
	PromptTypeMemorist              PromptType = "memorist"                // get information from long-term memory
	PromptTypeSearcher              PromptType = "searcher"                // search the internet for more information
	PromptTypeSummarizer            PromptType = "summarizer"              // summarize the result of the tool call
)

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

	return renderPrompt(string(promptType), prompt, params)
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

	return renderPrompt(string(promptType), prompt, params)
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

		promptsMap[PromptType(prompt.Name())] = string(promptBytes)
	}

	blob, err := json.Marshal(promptsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal templates: %w", err)
	}

	return blob, nil
}

func renderPrompt(name, prompt string, params any) (string, error) {
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
