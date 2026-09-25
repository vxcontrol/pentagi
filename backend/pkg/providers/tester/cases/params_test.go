package cases

import (
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestParams_ValidateDefinition_RefusesADefinitionThatCannotRunAsWritten(t *testing.T) {
	withParams := func(block string) string {
		return "  type: \"completion\"\n  group: \"basic\"\n  params:\n" + block +
			"  messages:\n    - role: \"user\"\n      content: \"hi\"\n  expected: \"ok\"\n"
	}

	tests := []struct {
		name string
		body string
		want string
	}{
		{"params overriding the model", withParams("    model: \"some-other-model\"\n"), "must not set a model"},
		{"params asking for adaptive reasoning", withParams("    reasoning:\n      mode: adaptive\n"), "cannot request adaptive reasoning"},
		{"an empty params block", withParams("    {}\n"), "produces no call options"},
		{"params turning reasoning off", withParams("    reasoning:\n      mode: \"off\"\n"), "cannot disable reasoning"},
		{"a value an operator could not save", withParams("    temperature: 5\n"), "temperature 5 out of range"},
		{
			"a budget past the vendor limit",
			withParams("    reasoning:\n      mode: budget\n      max_tokens: 999999\n"),
			"reasoning.max_tokens 999999 out of range",
		},
		{
			"a tool choice with no tools at all",
			`  type: "completion"
  group: "basic"
  tool_choice: "required"
  messages:
    - role: "user"
      content: "hi"
  expected: "ok"
`,
			"no other type puts tools on the request",
		},
		{
			"a tool choice with tools declared on a type that never sends them",
			`  type: "completion"
  group: "basic"
  tool_choice: "required"
  tools:
    - name: "noop"
      description: "does nothing"
      parameters:
        type: "object"
        properties: {}
  messages:
    - role: "user"
      content: "hi"
  expected: "ok"
`,
			"no other type puts tools on the request",
		},
		{
			"a tool choice that forbids the tool call the case demands",
			`  type: "tool"
  group: "basic"
  tool_choice: "none"
  tools:
    - name: "noop"
      description: "d"
      parameters:
        type: "object"
  messages:
    - role: "user"
      content: "hi"
  expected:
    - function_name: "noop"
`,
			"can never pass",
		},
		{
			"a tool choice on a tool case with the tools left out",
			`  type: "tool"
  group: "basic"
  tool_choice: "required"
  messages:
    - role: "user"
      content: "hi"
  expected:
    - function_name: "noop"
      arguments: {}
`,
			"tool_choice needs tools to choose from",
		},
		{"an unknown group", "  type: \"completion\"\n  group: \"advnaced\"\n  prompt: \"hi\"\n  expected: \"ok\"\n", `unknown group "advnaced"`},
		{
			"an unknown capability",
			"  type: \"completion\"\n  group: \"basic\"\n  capability: \"structured_ouput\"\n  prompt: \"hi\"\n  expected: \"ok\"\n",
			`unknown capability "structured_ouput"`,
		},
		{
			"an unknown refusal",
			"  type: \"completion\"\n  group: \"basic\"\n  expect_refusal: \"structured_output_confg\"\n  prompt: \"hi\"\n  expected: \"ok\"\n",
			`unknown expect_refusal "structured_output_confg"`,
		},
		{"an unknown type", "  type: \"complection\"\n  group: \"basic\"\n  prompt: \"hi\"\n  expected: \"ok\"\n", `unknown test type "complection"`},
		{
			"truncation expected with no output cap",
			`  type: "completion"
  group: "basic"
  messages:
    - role: "user"
      content: "write a long essay"
  expected: ""
  expect_truncated: true
`,
			"needs params.max_tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadRegistryFromYAML([]byte("- id: \"p\"\n  name: \"P\"\n" + tt.body))
			if err == nil {
				t.Fatalf("%s must be refused at load", tt.name)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q does not mention %q", err, tt.want)
			}
		})
	}
}

func TestParams_CaseOptions_CarriesTheToolChoiceAndEveryParamIncludingAZero(t *testing.T) {
	registry, err := LoadRegistryFromYAML([]byte(`
- id: "p"
  name: "P"
  type: "completion"
  group: "basic"
  params:
    max_tokens: 0
    temperature: 0.25
  messages:
    - role: "user"
      content: "hi"
  expected: "ok"

- id: "t"
  name: "T"
  type: "tool"
  group: "basic"
  tool_choice: "required"
  tools:
    - name: "noop"
      description: "does nothing"
      parameters:
        type: "object"
        properties: {}
  messages:
    - role: "user"
      content: "hi"
  expected:
    - function_name: "noop"
      arguments: {}
`))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	suite, err := registry.GetTestSuite(TestGroupBasic)
	if err != nil {
		t.Fatalf("suite: %v", err)
	}
	if len(suite.Tests) != 2 {
		t.Fatalf("expected two cases, got %d", len(suite.Tests))
	}

	var applied, tool llms.CallOptions
	for _, option := range suite.Tests[0].ExtraOptions() {
		option(&applied)
	}
	for _, option := range suite.Tests[1].ExtraOptions() {
		option(&tool)
	}

	switch {
	case applied.MaxTokens == nil:
		t.Error("max_tokens: 0 never reached the call options, so a provider would read it as unset")
	case *applied.MaxTokens != 0:
		t.Errorf("max_tokens = %d, want 0", *applied.MaxTokens)
	}
	if applied.Temperature == nil || *applied.Temperature != 0.25 {
		t.Errorf("temperature = %v, want 0.25", applied.Temperature)
	}
	if tool.ToolChoice != "required" {
		t.Errorf("tool_choice = %v, want \"required\"", tool.ToolChoice)
	}
}
