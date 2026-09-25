package validator

import (
	"testing"

	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_ExtractTemplateVariables_ReturnsSortedTopLevelNames(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
		wantErr  string
	}{
		{name: "an empty template", template: "", wantErr: "template content is empty"},
		{name: "a syntax error", template: "{{.Name", wantErr: "failed to parse template"},
		{name: "a single field", template: "Hello {{.Name}}!", want: []string{"Name"}},
		{
			name:     "several fields come back sorted",
			template: "User {{.Name}} has {{.Age}} years and {{.Email}} email",
			want:     []string{"Age", "Email", "Name"},
		},
		{name: "a nested field names its root", template: "{{.User.Name}} works at {{.Company.Name}}", want: []string{"Company", "User"}},
		{
			name:     "fields inside a range belong to its items",
			template: "{{range .Items}}Item: {{.Name}} - {{.Value}}{{end}}",
			want:     []string{"Items"},
		},
		{name: "nested ranges", template: "{{range .Categories}}{{range .Items}}{{.Name}}{{end}}{{end}}", want: []string{"Categories", "Items"}},
		{name: "local variables are skipped", template: "{{range .Items}}{{$item := .}}{{$item.Name}}{{end}}", want: []string{"Items"}},
		{
			name:     "if and else branches",
			template: `{{if .UseAgents}}Agent: {{.AgentName}}{{else}}Tool: {{.ToolName}}{{end}}`,
			want:     []string{"AgentName", "ToolName", "UseAgents"},
		},
		{name: "nested conditions", template: "{{if .A}}{{.C}}{{else}}{{if .D}}{{.E}}{{end}}{{end}}", want: []string{"A", "C", "D", "E"}},
		{
			name:     "keywords around empty bodies",
			template: "{{.Items}} {{range .Items}}{{end}} {{if .Condition}}{{end}}",
			want:     []string{"Condition", "Items"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTemplateVariables(tt.template)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidator_ValidatePrompt_ClassifiesEveryFailure(t *testing.T) {
	defaultPrompts, err := templates.GetDefaultPrompts()
	require.NoError(t, err)

	const (
		primary = templates.PromptTypePrimaryAgent
		chooser = templates.PromptTypeImageChooser
	)

	tests := []struct {
		name        string
		promptType  templates.PromptType
		template    string
		wantType    ErrorType
		wantMessage string
		wantDetails string
		wantLine    int
	}{
		{
			name:       "declared variables",
			promptType: primary,
			template:   "You are an AI assistant. Your name is {{.FinalyToolName}} and you can use {{.SearchToolName}} for searches.",
		},
		{
			name:       "declared variables under a condition",
			promptType: templates.PromptTypeAssistant,
			template: `You are an assistant with the following tools:
{{if .UseAgents}}
- {{.SearchToolName}}
- {{.PentesterToolName}}
- {{.CoderToolName}}
{{end}}
Current time: {{.CurrentTime}}
Language: {{.Lang}}`,
		},
		{name: "the shipped primary agent prompt", promptType: primary, template: defaultPrompts.AgentsPrompts.PrimaryAgent.System.Template},
		{
			name:       "the shipped assistant prompt",
			promptType: templates.PromptTypeAssistant,
			template:   defaultPrompts.AgentsPrompts.Assistant.System.Template,
		},
		{
			name:       "the shipped pentester prompt",
			promptType: templates.PromptTypePentester,
			template:   defaultPrompts.AgentsPrompts.Pentester.System.Template,
		},
		{
			name:        "an empty template",
			promptType:  primary,
			template:    "",
			wantType:    "Empty Template",
			wantMessage: "template content cannot be empty",
		},
		{
			name:        "an undeclared variable",
			promptType:  primary,
			template:    "{{.FinalyToolName}} and {{.NonExistentVar}}",
			wantType:    "Unauthorized Variable",
			wantMessage: "[NonExistentVar]",
			wantDetails: "Backend code cannot provide these variables",
		},
		{
			name:        "several undeclared variables",
			promptType:  primary,
			template:    "{{.UnauthorizedVar1}} and {{.UnauthorizedVar2}} with valid {{.FinalyToolName}}",
			wantType:    "Unauthorized Variable",
			wantMessage: "[UnauthorizedVar1 UnauthorizedVar2]",
		},
		{
			name:        "an unknown prompt type",
			promptType:  "unknown_prompt_type",
			template:    "{{.SomeVar}}",
			wantType:    "Unauthorized Variable",
			wantMessage: "unknown prompt type: unknown_prompt_type",
		},
		{
			name:        "an unclosed action",
			promptType:  primary,
			template:    "{{.FinalyToolName",
			wantType:    "Syntax Error",
			wantDetails: "missing closing braces",
			wantLine:    1,
		},
		{
			name:       "an unclosed action on the third line",
			promptType: chooser,
			template:   "line one\nline two\n{{ .Input }\nline four",
			wantType:   "Syntax Error",
			wantLine:   3,
		},
		{name: "an unclosed action on the first line", promptType: chooser, template: "{{ .Input }\nline two", wantType: "Syntax Error", wantLine: 1},
		{
			name:       "an unterminated action at the end",
			promptType: chooser,
			template:   "line one\nline two\nline three\n{{ .Input",
			wantType:   "Syntax Error",
			wantLine:   4,
		},
		{
			name:        "a declared variable used as the wrong type",
			promptType:  primary,
			template:    "{{.FinalyToolName.Foo}}",
			wantType:    "Rendering Failed",
			wantMessage: "template rendering failed",
			wantDetails: "Variable type mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrompt(tt.promptType, tt.template)
			if tt.wantType == "" {
				require.NoError(t, err)
				return
			}

			var vErr *ValidationError
			require.ErrorAs(t, err, &vErr)
			assert.Equal(t, tt.wantType, vErr.Type)
			assert.Contains(t, err.Error(), string(tt.wantType))
			assert.Contains(t, vErr.Message, tt.wantMessage)
			assert.Contains(t, vErr.Details, tt.wantDetails)
			assert.Equal(t, tt.wantLine, vErr.Line, "message: %s", vErr.Message)
		})
	}
}
