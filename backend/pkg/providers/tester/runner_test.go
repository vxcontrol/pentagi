package tester

import (
	"fmt"
	"testing"
	"time"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"
	"pentagi/pkg/providers/tester/testdata"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestTestProvider(t *testing.T) {
	// create mock provider
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: "What is 2+2?", Response: "4"},
		{Key: "Hello World", Response: "HELLO WORLD"},
		{Key: "Count from 1 to 5", Response: "1, 2, 3, 4, 5"},
	})
	mockProvider.SetDefaultResponse("Mock response")

	// test basic functionality
	results, err := TestProvider(t.Context(), mockProvider)
	if err != nil {
		t.Fatalf("TestProvider failed: %v", err)
	}

	// verify we got results for all agent types
	agentTypeFields := []struct {
		name    string
		results AgentTestResults
	}{
		{"simple", results.Simple},
		{"simple_json", results.SimpleJSON},
		{"primary_agent", results.PrimaryAgent},
		{"assistant", results.Assistant},
		{"generator", results.Generator},
		{"refiner", results.Refiner},
		{"adviser", results.Adviser},
		{"reflector", results.Reflector},
		{"searcher", results.Searcher},
		{"enricher", results.Enricher},
		{"coder", results.Coder},
		{"installer", results.Installer},
		{"pentester", results.Pentester},
	}

	totalTests := 0
	for _, field := range agentTypeFields {
		if len(field.results) > 0 {
			t.Logf("Agent %s has %d test results", field.name, len(field.results))
			totalTests += len(field.results)
		}
	}

	if totalTests == 0 {
		t.Errorf("Expected some test results, got 0")
	}
}

func TestTestProviderWithOptions(t *testing.T) {
	// create mock provider
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: "What is 2+2?", Response: "4"},
		{Key: "Hello World", Response: "HELLO WORLD"},
	})

	// test with specific agent types
	results, err := TestProvider(
		t.Context(),
		mockProvider,
		WithAgentTypes(pconfig.OptionsTypeSimple, pconfig.OptionsTypePrimaryAgent),
		WithGroups(testdata.TestGroupBasic),
		WithVerbose(false),
		WithParallelWorkers(2),
	)
	if err != nil {
		t.Fatalf("TestProvider with options failed: %v", err)
	}

	// should only have results for Simple and Agent
	if len(results.Simple) == 0 {
		t.Errorf("Expected Simple agent results")
	}
	if len(results.PrimaryAgent) == 0 {
		t.Errorf("Expected Agent results")
	}

	// other agents should have no results since they weren't requested
	if len(results.Generator) > 0 {
		t.Errorf("Expected no Generator results, got %d", len(results.Generator))
	}
}

func TestTestProviderStreamingMode(t *testing.T) {
	// create mock provider with streaming delay
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetStreamingDelay(time.Millisecond * 5)
	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: "What is 2+2?", Response: "The answer is 4"},
	})

	// test with streaming enabled
	results, err := TestProvider(
		t.Context(),
		mockProvider,
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(testdata.TestGroupBasic),
		WithStreamingMode(true),
	)
	if err != nil {
		t.Fatalf("TestProvider streaming failed: %v", err)
	}

	// verify we got results
	if len(results.Simple) == 0 {
		t.Errorf("Expected some Simple agent results")
	}

	// check for streaming tests
	foundStreaming := false
	for _, result := range results.Simple {
		if result.Streaming {
			foundStreaming = true
			if result.Latency == 0 {
				t.Errorf("Expected non-zero latency for streaming test")
			}
		}
	}

	if !foundStreaming {
		t.Logf("No streaming tests found (this may be expected if no streaming tests in testdata)")
	}
}

func TestTestProviderJSONTests(t *testing.T) {
	// create mock provider with JSON responses
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: "Return JSON", Response: `{"name": "John Doe", "age": 30, "city": "New York"}`},
		{Key: "Create JSON array", Response: `[{"name": "red", "hex": "#FF0000"}]`},
	})

	// test with JSON group only - must use SimpleJSON agent
	results, err := TestProvider(
		t.Context(),
		mockProvider,
		WithAgentTypes(pconfig.OptionsTypeSimpleJSON),
		WithGroups(testdata.TestGroupJSON),
	)
	if err != nil {
		t.Fatalf("TestProvider JSON tests failed: %v", err)
	}

	// verify we got JSON test results for SimpleJSON agent
	if len(results.SimpleJSON) == 0 {
		t.Logf("No JSON test results (this may be expected if no JSON tests in testdata)")
		return
	}

	// check for JSON test types
	foundJSON := false
	for _, result := range results.SimpleJSON {
		if result.Type == testdata.TestTypeJSON {
			foundJSON = true
		}
	}

	if !foundJSON {
		t.Logf("No JSON tests found (this may be expected if no JSON tests in testdata)")
	}
}

func TestTestProviderToolTests(t *testing.T) {
	// create mock provider with tool call responses
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")

	// set up tool call response
	toolResponse := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "",
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "echo",
							Arguments: `{"message": "hello"}`,
						},
					},
				},
			},
		},
	}

	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: "tools:Call echo", Response: toolResponse},
		{Key: "Use tools only Call echo", Response: toolResponse},
	})

	// test with basic group (may contain tool tests)
	results, err := TestProvider(
		t.Context(),
		mockProvider,
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(testdata.TestGroupBasic, testdata.TestGroupAdvanced),
	)
	if err != nil {
		t.Fatalf("TestProvider tool tests failed: %v", err)
	}

	// verify we got some results
	if len(results.Simple) == 0 {
		t.Errorf("Expected some test results")
	}

	// check for tool test types
	foundTool := false
	for _, result := range results.Simple {
		if result.Type == testdata.TestTypeTool {
			foundTool = true
		}
	}

	if !foundTool {
		t.Logf("No tool tests found (this may be expected if no tool tests in configured groups)")
	}
}

func TestTestProviderErrorHandling(t *testing.T) {
	// create mock provider that returns errors for specific requests
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: "What is 2+2?", Response: fmt.Errorf("mock API error")},
	})

	// test error handling
	results, err := TestProvider(
		t.Context(),
		mockProvider,
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(testdata.TestGroupBasic),
	)
	if err != nil {
		t.Fatalf("TestProvider error handling failed: %v", err)
	}

	// should have some results (some may be errors)
	if len(results.Simple) == 0 {
		t.Errorf("Expected some results even with errors")
	}

	// check for error results
	foundError := false
	for _, result := range results.Simple {
		if !result.Success && result.Error != nil {
			foundError = true
			t.Logf("Found expected error result: %v", result.Error)
		}
	}

	if !foundError {
		t.Logf("No error results found (this may be expected)")
	}
}

func TestTestProviderGroups(t *testing.T) {
	// create mock provider
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetDefaultResponse("Group test response")

	tests := []struct {
		name      string
		agentType pconfig.ProviderOptionsType
		groups    []testdata.TestGroup
	}{
		{"Basic only", pconfig.OptionsTypeSimple, []testdata.TestGroup{testdata.TestGroupBasic}},
		{"Advanced only", pconfig.OptionsTypeSimple, []testdata.TestGroup{testdata.TestGroupAdvanced}},
		{"JSON only", pconfig.OptionsTypeSimpleJSON, []testdata.TestGroup{testdata.TestGroupJSON}},
		{"Knowledge only", pconfig.OptionsTypeSimple, []testdata.TestGroup{testdata.TestGroupKnowledge}},
		{"Multiple groups", pconfig.OptionsTypeSimple, []testdata.TestGroup{testdata.TestGroupBasic, testdata.TestGroupAdvanced}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := TestProvider(
				t.Context(),
				mockProvider,
				WithAgentTypes(tt.agentType),
				WithGroups(tt.groups...),
			)
			if err != nil {
				t.Fatalf("TestProvider groups failed: %v", err)
			}

			// get results for the correct agent type
			var agentResults AgentTestResults
			switch tt.agentType {
			case pconfig.OptionsTypeSimple:
				agentResults = results.Simple
			case pconfig.OptionsTypeSimpleJSON:
				agentResults = results.SimpleJSON
			default:
				t.Fatalf("Unexpected agent type: %v", tt.agentType)
			}

			// verify all results belong to specified groups
			for _, result := range agentResults {
				groupFound := false
				for _, group := range tt.groups {
					if result.Group == group {
						groupFound = true
						break
					}
				}
				if !groupFound {
					t.Errorf("Result belongs to unexpected group: %s", result.Group)
				}
			}
		})
	}
}

// TestTestProviderCapabilityGating_AdaptiveThinking verifies that
// adaptive_thinking_reasoning_present runs — via the plain CallWithTools/
// CallEx path, not a forced extra option — only when the agent's assigned
// model is adaptive-only, exactly the condition
// ProviderConfig.UsesAdaptiveThinking uses in production to decide whether to
// append llms.WithAdaptiveReasoning. reasoning_off_suppresses_reasoning must
// NOT run: this config never sets reasoning mode to off, so PentAGI's runtime
// would never send that disable signal for this agent either.
func TestTestProviderCapabilityGating_AdaptiveThinking(t *testing.T) {
	const model = "claude-opus-4-8"
	mockProvider := mock.NewProvider(provider.ProviderAnthropic, provider.DefaultProviderNameAnthropic, model)
	mockProvider.SetProviderConfig(&pconfig.ProviderConfig{
		Simple: &pconfig.AgentConfig{Model: model},
	})
	mockProvider.SetModels(pconfig.ModelsConfig{
		{Name: model, Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningAdaptiveOnly}},
	})
	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: "farmer has 17 sheep", Response: &llms.ContentResponse{
			Choices: []*llms.ContentChoice{{
				Content:   "The answer is 13",
				Reasoning: &reasoning.ContentReasoning{Content: "mock reasoning trace"},
			}},
		}},
	})

	results, err := TestProvider(
		t.Context(),
		mockProvider,
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(testdata.TestGroupAdvanced),
		WithStreamingMode(false),
	)
	if err != nil {
		t.Fatalf("TestProvider capability gating failed: %v", err)
	}

	var foundAdaptive, foundOff bool
	for _, result := range results.Simple {
		switch result.ID {
		case "adaptive_thinking_reasoning_present":
			foundAdaptive = true
			if !result.Success {
				t.Errorf("adaptive_thinking test failed: %v", result.Error)
			}
			if result.Capability != testdata.CapabilityAdaptiveThinking {
				t.Errorf("expected capability %q, got %q", testdata.CapabilityAdaptiveThinking, result.Capability)
			}
		case "reasoning_off_suppresses_reasoning":
			foundOff = true
		}
	}

	if !foundAdaptive {
		t.Errorf("expected adaptive_thinking_reasoning_present to run for an adaptive-only model")
	}
	if foundOff {
		t.Errorf("reasoning_off_suppresses_reasoning must not run: this config never sets reasoning mode to off")
	}
}

// TestTestProviderCapabilityGating_ReasoningOff verifies that
// reasoning_off_suppresses_reasoning runs — via the plain CallWithTools/
// CallEx path — only when the agent's own config explicitly sets
// `reasoning: {mode: off}`, exactly the condition AgentConfig.BuildOptions
// uses in production to decide whether to append llms.WithReasoningDisabled.
// adaptive_thinking_reasoning_present must NOT run for this config.
func TestTestProviderCapabilityGating_ReasoningOff(t *testing.T) {
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "some-model")
	mockProvider.SetProviderConfig(&pconfig.ProviderConfig{
		Simple: &pconfig.AgentConfig{
			Model:     "some-model",
			Reasoning: pconfig.ReasoningConfig{Mode: pconfig.ReasoningModeOff},
		},
	})
	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: "capital of France", Response: &llms.ContentResponse{
			Choices: []*llms.ContentChoice{{Content: "Paris"}},
		}},
	})

	results, err := TestProvider(
		t.Context(),
		mockProvider,
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(testdata.TestGroupAdvanced),
		WithStreamingMode(false),
	)
	if err != nil {
		t.Fatalf("TestProvider capability gating failed: %v", err)
	}

	var foundAdaptive, foundOff bool
	for _, result := range results.Simple {
		switch result.ID {
		case "reasoning_off_suppresses_reasoning":
			foundOff = true
			if !result.Success {
				t.Errorf("reasoning_off test failed: %v", result.Error)
			}
			if result.Capability != testdata.CapabilityReasoningOff {
				t.Errorf("expected capability %q, got %q", testdata.CapabilityReasoningOff, result.Capability)
			}
		case "adaptive_thinking_reasoning_present":
			foundAdaptive = true
		}
	}

	if !foundOff {
		t.Errorf("expected reasoning_off_suppresses_reasoning to run when the agent's config sets reasoning mode to off")
	}
	if foundAdaptive {
		t.Errorf("adaptive_thinking_reasoning_present must not run: this config doesn't request adaptive thinking")
	}
}

// TestTestProviderCapabilityGating_SkipsDefaultConfig verifies that capability
// tests are entirely absent (not failed) for a provider whose config sets
// neither an adaptive mode nor an explicit reasoning-off — i.e. a flow that
// would never trigger either CallOption in production.
func TestTestProviderCapabilityGating_SkipsDefaultConfig(t *testing.T) {
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetDefaultResponse("Mock response")

	results, err := TestProvider(
		t.Context(),
		mockProvider,
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(testdata.TestGroupAdvanced),
		WithStreamingMode(false),
	)
	if err != nil {
		t.Fatalf("TestProvider capability gating failed: %v", err)
	}

	for _, result := range results.Simple {
		if result.Capability != testdata.CapabilityNone {
			t.Errorf("default config must not run capability test %s (capability=%s)", result.ID, result.Capability)
		}
	}
}

func TestApplyOptions(t *testing.T) {
	// test default configuration
	config := applyOptions(nil)
	if config == nil {
		t.Fatalf("Expected non-nil config")
	}

	if len(config.agentTypes) == 0 {
		t.Errorf("Expected default agent types")
	}
	if len(config.groups) == 0 {
		t.Errorf("Expected default groups")
	}
	if !config.streamingMode {
		t.Errorf("Expected streaming mode enabled by default")
	}
	if config.verbose {
		t.Errorf("Expected verbose mode disabled by default")
	}
	if config.parallelWorkers != 4 {
		t.Errorf("Expected 4 parallel workers by default, got %d", config.parallelWorkers)
	}

	// test with options
	config = applyOptions([]TestOption{
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(testdata.TestGroupBasic),
		WithStreamingMode(false),
		WithVerbose(true),
		WithParallelWorkers(8),
	})

	if len(config.agentTypes) != 1 || config.agentTypes[0] != pconfig.OptionsTypeSimple {
		t.Errorf("Agent types not applied correctly")
	}
	if len(config.groups) != 1 || config.groups[0] != testdata.TestGroupBasic {
		t.Errorf("Groups not applied correctly")
	}
	if config.streamingMode {
		t.Errorf("Streaming mode not disabled")
	}
	if !config.verbose {
		t.Errorf("Verbose mode not enabled")
	}
	if config.parallelWorkers != 8 {
		t.Errorf("Parallel workers not set correctly")
	}
}
