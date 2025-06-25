package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

const trimCharset = "\n\r\t "

func trimString(s string) string {
	return strings.Trim(s, trimCharset)
}

const toolCallingAgentSystemPrompt = `
You are a helpful assistant that follows instructions precisely.
You must use tools instead of generating text.
You must use only provided tools to figure out a question.
`

// Test definitions and execution functions

// GetOutputJSONTestSuites returns the set of test suites specifically for simple_json agent
func GetOutputJSONTestSuites() []TestSuite {
	// JSON response tests specific for simple_json agent
	jsonTests := []SimpleJSONCompletionTest{
		{
			Prompt:         "Return a JSON with a person's information: name='John Doe', age=30, city='New York'",
			RequiredFields: []string{"name", "age", "city"},
		},
		{
			Prompt:         "Create a JSON object with fields: title='Test Project', completed=false, priority=1",
			RequiredFields: []string{"title", "completed", "priority"},
		},
		{
			Prompt:         "Generate a JSON response for a user profile with username='user123', email='user@example.com', active=true",
			RequiredFields: []string{"username", "email", "active"},
		},
	}

	// JSON array test
	jsonArrayTests := []SimpleJSONCompletionTest{
		{
			Prompt:         "Create a JSON array of 3 colors with properties name and hex code",
			RequiredFields: []string{},
			IsArray:        true,
		},
	}

	// Negative test - expects the model to respond with JSON even for non-JSON prompts
	nonJsonPrompts := []SimpleJSONCompletionTest{
		{
			Prompt:         "What is 2+2? Explain your answer.",
			RequiredFields: []string{},
			ExpectFailure:  false, // We actually expect the model to respond with JSON here too
		},
	}

	return []TestSuite{
		{
			Name:  "Simple JSON Objects",
			Basic: true,
			Tests: []interface{}{jsonTests},
		},
		{
			Name:  "JSON Arrays",
			Basic: true,
			Tests: []interface{}{jsonArrayTests},
		},
		{
			Name:  "Non-JSON Prompts",
			Basic: false,
			Tests: []interface{}{nonJsonPrompts},
		},
	}
}

// GetBasicTestSuites returns the set of basic test suites that all general agents must pass
func GetBasicTestSuites() []TestSuite {
	// Simple text completion tests
	simpleTests := []SimpleCompletionTest{
		{
			Prompt:   "What is 2+2? Write only the number without any other text.",
			Expected: "4",
		},
		{
			Prompt:   "Write 'Hello World' in uppercase without any other text.",
			Expected: "HELLO WORLD",
		},
	}

	// Tests with separate system and user prompts
	systemUserTests := []SystemUserPromptTest{
		{
			SystemPrompt: "You are a helpful assistant that follows instructions precisely. Always keep your responses concise and exact.",
			UserPrompt:   "Count from 1 to 5, separated by commas, without any additional text.",
			Expected:     "1, 2, 3, 4, 5",
		},
		{
			SystemPrompt: "You are a math assistant that provides concise answers.",
			UserPrompt:   "Calculate 5 * 10 and provide only the result.",
			Expected:     "50",
		},
	}

	// Simple JSON response test (without function calling)
	simpleJSONTests := []SimpleJSONCompletionTest{
		{
			Prompt:         "Return a JSON with a person's information: name='John Doe', age=30, city='New York'",
			RequiredFields: []string{"name", "age", "city"},
		},
	}

	// Basic function call test with simplified structure
	// Use a simple echo function that just returns the input
	echoTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "echo",
			Description: "Echoes back the input message",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Message to echo back",
					},
				},
				"required": []string{"message"},
			},
		},
	}

	functionTests := []FunctionCallTest{
		{
			Name:         "Basic echo function",
			Prompt:       "JUST choose the echo function and call it with this message: Hello from function test",
			Functions:    []llms.Tool{echoTool},
			ExpectedTool: "echo",
			ValidateFunc: func(args map[string]interface{}) bool {
				message, ok := args["message"].(string)
				return ok && message != ""
			},
		},
	}

	return []TestSuite{
		{
			Name:  "Simple Completion",
			Basic: true,
			Tests: []interface{}{simpleTests},
		},
		{
			Name:  "System User Prompts",
			Basic: true,
			Tests: []interface{}{systemUserTests},
		},
		{
			Name:  "Simple JSON Response",
			Basic: true,
			Tests: []interface{}{simpleJSONTests},
		},
		{
			Name:  "Basic Function Call",
			Basic: true,
			Tests: []interface{}{functionTests},
		},
	}
}

// GetAdvancedTestSuites returns the set of advanced test suites
func GetAdvancedTestSuites() []TestSuite {
	// Function-based JSON structure tests (only for models supporting function calling)
	jsonTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "respond_with_json",
			Description: "Response with JSON structure",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name value",
					},
					"value": map[string]interface{}{
						"type":        "integer",
						"description": "Numeric value",
					},
				},
				"required": []string{"name", "value"},
			},
		},
	}

	jsonTests := []FunctionCallTest{
		{
			Name:         "JSON response function",
			Prompt:       "Call function respond_with_json to create a JSON with name=test and value=123",
			Functions:    []llms.Tool{jsonTool},
			ExpectedTool: "respond_with_json",
			ValidateFunc: func(args map[string]interface{}) bool {
				name, nameOk := args["name"]
				value, valueOk := args["value"]
				return nameOk && valueOk && name != "" && value != nil
			},
		},
	}

	// Advanced function call tests
	// Simplified search tool
	searchTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "search",
			Description: "Performs search for a given query",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []string{"query"},
			},
		},
	}

	// Simplified advice tool
	adviceTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "provide_advice",
			Description: "Provides advice to the user about the problem",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"problem": map[string]interface{}{
						"type":        "string",
						"description": "Problem to solve",
					},
					"solution": map[string]interface{}{
						"type":        "string",
						"description": "Solution to the problem",
					},
				},
				"required": []string{"problem", "solution"},
			},
		},
	}

	functionTests := []FunctionCallTest{
		{
			Name:         "Search query",
			Prompt:       "Find information about Golang programming language by calling the search function",
			Functions:    []llms.Tool{searchTool},
			ExpectedTool: "search",
			ValidateFunc: func(args map[string]interface{}) bool {
				query, ok := args["query"].(string)
				return ok && query != ""
			},
		},
		{
			Name: "Ask advice",
			Prompt: "Please give me advice on how to understand the error: 'cannot find package' " +
				"in Go development by calling the provide_advice function. Before calling the function, " +
				"think about the problem and the solution.",
			Functions:    []llms.Tool{adviceTool},
			ExpectedTool: "provide_advice",
			ValidateFunc: func(args map[string]interface{}) bool {
				problem, isProblemExists := args["problem"].(string)
				solution, isSolutionExists := args["solution"].(string)
				return isProblemExists && isSolutionExists && problem != "" && solution != ""
			},
		},
	}

	return []TestSuite{
		{
			Name:  "Function-based JSON",
			Basic: false,
			Tests: []interface{}{jsonTests},
		},
		{
			Name:  "Advanced Function Calls",
			Basic: false,
			Tests: []interface{}{functionTests},
		},
	}
}

// RunCompletionTest executes a simple completion test
func RunCompletionTest(
	ctx context.Context,
	p provider.Provider,
	agentType provider.ProviderOptionsType,
	test SimpleCompletionTest,
) TestResult {
	result := TestResult{
		Name:     FormatTestName("Completion", test.Prompt, 30, false),
		Type:     "completion",
		Expected: test.Expected,
	}

	startTime := time.Now()

	response, err := p.Call(ctx, agentType, test.Prompt)
	elapsedTime := time.Since(startTime)
	result.LatencyMs = elapsedTime.Milliseconds()
	result.Response = response

	if err != nil {
		result.Success = false
		result.Error = err
		return result
	}

	// Check if the response contains the expected text
	result.Success = ContainsString(response, test.Expected)
	return result
}

// RunSystemUserPromptTest executes a test with system and user prompts
func RunSystemUserPromptTest(
	ctx context.Context,
	p provider.Provider,
	agentType provider.ProviderOptionsType,
	test SystemUserPromptTest,
	useStream bool,
) TestResult {
	result := TestResult{
		Name:      FormatTestName("System-User", test.UserPrompt, 30, false),
		Type:      "system_user",
		Expected:  test.Expected,
		Streaming: useStream,
	}

	respBuf := bytes.NewBuffer(nil)
	reasoningBuf := bytes.NewBuffer(nil)
	var streamCb streaming.Callback
	if useStream {
		streamCb = func(ctx context.Context, chunk streaming.Chunk) error {
			respBuf.WriteString(chunk.Content)
			reasoningBuf.WriteString(chunk.ReasoningContent)
			return nil
		}
	}

	startTime := time.Now()

	response, err := p.CallEx(ctx, agentType, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, test.SystemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, test.UserPrompt),
	}, streamCb)

	elapsedTime := time.Since(startTime)
	result.LatencyMs = elapsedTime.Milliseconds()
	respContent := trimString(respBuf.String())
	reasoningContent := trimString(reasoningBuf.String())

	if err != nil {
		result.Success = false
		result.Error = err
		return result
	}

	if len(response.Choices) == 0 {
		result.Success = false
		result.Error = fmt.Errorf("empty response from model")
		return result
	}

	choice := response.Choices[0]
	if len(reasoningContent) > 0 || len(choice.ReasoningContent) > 0 {
		result.Reasoning = true
	}
	reasoningTokens, ok := choice.GenerationInfo["ReasoningTokens"]
	if ok && reasoningTokens.(int) > 0 {
		result.Reasoning = true
	}

	if useStream && respContent != trimString(choice.Content) {
		result.Success = false
		result.Error = fmt.Errorf("streaming response content mismatch: '%s' != '%s'",
			respContent, choice.Content)
		return result
	}

	if useStream && reasoningContent != trimString(choice.ReasoningContent) {
		result.Success = false
		result.Error = fmt.Errorf("streaming reasoning content mismatch: '%s' != '%s'",
			reasoningContent, choice.ReasoningContent)
		return result
	}

	responseText := choice.Content
	result.Response = responseText

	// Clean up response for better matching
	cleanResponse := strings.TrimSpace(responseText)
	expected := strings.TrimSpace(test.Expected)

	// Special case for number sequences - normalize them by removing spaces
	if test.Expected == "1, 2, 3, 4, 5" || test.Expected == "1,2,3,4,5" {
		// Accept both formats: with or without spaces
		normalizedResponse := strings.ReplaceAll(cleanResponse, " ", "")
		normalizedExpected := "1,2,3,4,5"

		if strings.Contains(normalizedResponse, normalizedExpected) {
			result.Success = true
			return result
		}

		// Also try with spaces
		if strings.Contains(cleanResponse, "1, 2, 3, 4, 5") {
			result.Success = true
			return result
		}
	}

	// Check if the response exactly matches the expected text
	if cleanResponse == expected {
		result.Success = true
		return result
	}

	// Check if the response contains the expected text
	if strings.Contains(cleanResponse, expected) {
		result.Success = true
		return result
	}

	// Convert both to lowercase and try again (less strict matching)
	if strings.Contains(strings.ToLower(cleanResponse), strings.ToLower(expected)) {
		result.Success = true
		return result
	}

	result.Success = false
	result.Error = fmt.Errorf("expected: %s, got: %s", expected, cleanResponse)
	return result
}

// RunSimpleJSONTest executes a test for simple JSON response (no function calling)
func RunSimpleJSONTest(
	ctx context.Context,
	p provider.Provider,
	agentType provider.ProviderOptionsType,
	test SimpleJSONCompletionTest,
) TestResult {
	result := TestResult{
		Name:     FormatTestName("SimpleJSON", test.Prompt, 30, false),
		Type:     "simple_json",
		Expected: strings.Join(test.RequiredFields, ", "),
	}

	startTime := time.Now()

	// Enhance the prompt to encourage JSON response
	enhancedPrompt := fmt.Sprintf("%s\nRespond only with valid JSON. No explanations or additional text.", test.Prompt)

	response, err := p.Call(ctx, agentType, enhancedPrompt)
	elapsedTime := time.Since(startTime)
	result.LatencyMs = elapsedTime.Milliseconds()
	result.Response = response

	if err != nil {
		result.Success = false
		result.Error = err
		return result
	}

	// For simple_json agents, validate that the response is actually JSON
	if agentType == provider.OptionsTypeSimpleJSON {
		// Try to extract JSON from response (in case there's surrounding text)
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		arrayStart := strings.Index(response, "[")
		arrayEnd := strings.LastIndex(response, "]")

		// Check if this is a JSON object
		isValidObject := jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart

		// Check if this is a JSON array (for array tests)
		isValidArray := test.IsArray && arrayStart != -1 && arrayEnd != -1 && arrayEnd > arrayStart

		if !isValidObject && !isValidArray {
			if test.ExpectFailure {
				// For negative tests, this is actually expected
				result.Success = true
				return result
			}

			result.Success = false
			result.Error = fmt.Errorf("response is not valid JSON")
			return result
		}

		// Extract and validate JSON
		var jsonStr string
		var parsed interface{}

		if isValidArray {
			jsonStr = response[arrayStart : arrayEnd+1]
			err = json.Unmarshal([]byte(jsonStr), &parsed)

			if err != nil {
				result.Success = false
				result.Error = fmt.Errorf("JSON array parsing error: %v", err)
				return result
			}

			// Array validation successful
			result.Success = true
			return result
		} else {
			jsonStr = response[jsonStart : jsonEnd+1]
			var parsedObj map[string]interface{}
			err = json.Unmarshal([]byte(jsonStr), &parsedObj)

			if err != nil {
				result.Success = false
				result.Error = fmt.Errorf("JSON parsing error: %v", err)
				return result
			}

			// For regular tests, check required fields
			if len(test.RequiredFields) > 0 {
				for _, field := range test.RequiredFields {
					if _, ok := parsedObj[field]; !ok {
						result.Success = false
						result.Error = fmt.Errorf("missing required field: %s", field)
						return result
					}
				}
			}

			result.Success = true
			return result
		}
	}

	// For other agents, use the original logic
	// Try to extract JSON from response (in case there's surrounding text)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd < jsonStart {
		result.Success = false
		result.Error = fmt.Errorf("no valid JSON found in response")
		return result
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	// Parse JSON
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("JSON parsing error: %v", err)
		return result
	}

	// Check required fields
	for _, field := range test.RequiredFields {
		if _, ok := parsed[field]; !ok {
			result.Success = false
			result.Error = fmt.Errorf("missing required field: %s", field)
			return result
		}
	}

	result.Success = true
	return result
}

// RunFunctionTest executes a function call test
func RunFunctionTest(
	ctx context.Context,
	p provider.Provider,
	agentType provider.ProviderOptionsType,
	test FunctionCallTest,
	useStream bool,
) TestResult {
	testName := test.Name
	if testName == "" {
		testName = FormatTestName("Function", test.Prompt, 30, useStream)
	}

	result := TestResult{
		Name:      testName,
		Type:      "function",
		Expected:  test.ExpectedTool,
		Streaming: useStream,
	}

	respBuf := bytes.NewBuffer(nil)
	reasoningBuf := bytes.NewBuffer(nil)
	var streamCb streaming.Callback
	if useStream {
		streamCb = func(ctx context.Context, chunk streaming.Chunk) error {
			respBuf.WriteString(chunk.Content)
			reasoningBuf.WriteString(chunk.ReasoningContent)
			return nil
		}
	}

	startTime := time.Now()

	response, err := p.CallWithTools(ctx, agentType, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, toolCallingAgentSystemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, test.Prompt),
	}, test.Functions, streamCb)

	elapsedTime := time.Since(startTime)
	result.LatencyMs = elapsedTime.Milliseconds()
	respContent := trimString(respBuf.String())
	reasoningContent := trimString(reasoningBuf.String())

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("API error: %v", err)
		return result
	}

	if len(response.Choices) == 0 {
		result.Success = false
		result.Error = fmt.Errorf("empty response from model")
		return result
	}

	choice := response.Choices[0]
	if len(reasoningContent) > 0 || len(choice.ReasoningContent) > 0 {
		result.Reasoning = true
	}
	reasoningTokens, ok := choice.GenerationInfo["ReasoningTokens"]
	if ok && reasoningTokens.(int) > 0 {
		result.Reasoning = true
	}

	if useStream && respContent != trimString(choice.Content) {
		result.Success = false
		result.Error = fmt.Errorf("streaming response content mismatch: '%s' != '%s'",
			TruncateString(respContent, 20), TruncateString(choice.Content, 20))
		return result
	}

	if useStream && reasoningContent != trimString(choice.ReasoningContent) {
		result.Success = false
		result.Error = fmt.Errorf("streaming reasoning content mismatch: '%s' != '%s'",
			TruncateString(reasoningContent, 20), TruncateString(choice.ReasoningContent, 20))
		return result
	}

	var (
		chunks    []string
		toolCalls []llms.ToolCall
	)
	for _, choice := range response.Choices {
		if len(choice.ToolCalls) > 0 {
			toolCalls = append(toolCalls, choice.ToolCalls...)
		}
		if len(choice.Content) > 0 {
			chunks = append(chunks, choice.Content)
		}
	}

	if len(toolCalls) == 0 {
		result.Response = choice.Content
		result.Success = false
		content := TruncateString(strings.Join(chunks, " | "), 50)
		result.Error = fmt.Errorf("model did not call a function, responded with text: %s", content)
		return result
	}

	// Find tool call
	var toolCallName string
	var arguments string

	for _, toolCall := range toolCalls {
		if functionCall := toolCall.FunctionCall; functionCall != nil {
			toolCallName = functionCall.Name
			arguments = functionCall.Arguments
			break
		}
	}

	if toolCallName == "" {
		result.Success = false
		result.Error = fmt.Errorf("function name not found in tool calls")
		return result
	}

	result.Response = fmt.Sprintf("%s: %s", toolCallName, arguments)

	// Check that the expected function was called
	if toolCallName != test.ExpectedTool {
		result.Success = false
		result.Error = fmt.Errorf("unexpected function called: %s, expected: %s",
			toolCallName, test.ExpectedTool)
		return result
	}

	// Validate function arguments
	var args map[string]interface{}
	err = json.Unmarshal([]byte(arguments), &args)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("error parsing function arguments: %v", err)
		return result
	}

	// Check argument content through specialized function
	if test.ValidateFunc != nil && !test.ValidateFunc(args) {
		result.Success = false
		result.Error = fmt.Errorf("invalid function arguments: %s", arguments)
		return result
	}

	result.Success = true
	return result
}

// RunTestSuite executes all tests in a test suite
func RunTestSuite(ctx context.Context, p provider.Provider, agentType provider.ProviderOptionsType, suite TestSuite) []TestResult {
	results := make([]TestResult, 0)

	for _, testGroup := range suite.Tests {
		switch tests := testGroup.(type) {
		case []SimpleCompletionTest:
			for _, test := range tests {
				result := RunCompletionTest(ctx, p, agentType, test)
				results = append(results, result)
			}
		case []SystemUserPromptTest:
			for _, test := range tests {
				result := RunSystemUserPromptTest(ctx, p, agentType, test, false)
				results = append(results, result)
				result = RunSystemUserPromptTest(ctx, p, agentType, test, true)
				results = append(results, result)
			}
		case []SimpleJSONCompletionTest:
			for _, test := range tests {
				result := RunSimpleJSONTest(ctx, p, agentType, test)
				results = append(results, result)
			}
		case []FunctionCallTest:
			for _, test := range tests {
				result := RunFunctionTest(ctx, p, agentType, test, false)
				results = append(results, result)
				result = RunFunctionTest(ctx, p, agentType, test, true)
				results = append(results, result)
			}
		}
	}

	return results
}

// TestAgent performs a test of a single agent type
func TestAgent(ctx context.Context, p provider.Provider, agentType provider.ProviderOptionsType, basicSuites []TestSuite, advancedSuites []TestSuite, verbose bool) (AgentTestResult, error) {
	result := AgentTestResult{
		AgentType: string(agentType),
		ModelName: p.Model(agentType),
	}

	// Run basic tests first
	for _, suite := range basicSuites {
		basicResults := RunTestSuite(ctx, p, agentType, suite)
		result.BasicTests = append(result.BasicTests, basicResults...)

		// Update success counts
		for _, tr := range basicResults {
			result.TotalTests++
			if tr.Success {
				result.TotalSuccess++
			}
			result.AverageLatency += time.Duration(tr.LatencyMs) * time.Millisecond
		}
	}

	// Calculate basic success rate
	basicSuccessRate := 0.0
	if len(result.BasicTests) > 0 {
		basicSuccessCount := 0
		for _, tr := range result.BasicTests {
			if tr.Success {
				basicSuccessCount++
			}
		}
		basicSuccessRate = float64(basicSuccessCount) / float64(len(result.BasicTests)) * 100
	}

	// Run advanced tests only if basic tests passed (at least 50% success rate)
	if basicSuccessRate >= 50.0 || len(basicSuites) == 0 {
		for _, suite := range advancedSuites {
			advancedResults := RunTestSuite(ctx, p, agentType, suite)
			result.AdvancedTests = append(result.AdvancedTests, advancedResults...)

			// Update success counts
			for _, tr := range advancedResults {
				result.TotalTests++
				if tr.Success {
					result.TotalSuccess++
				}
				result.AverageLatency += time.Duration(tr.LatencyMs) * time.Millisecond
			}
		}
	} else {
		// Store the skipping information in the result instead of printing it
		result.SkippedAdvanced = true
		result.SkippedReason = fmt.Sprintf("Skipped advanced tests due to low basic test success rate: %.1f%%", basicSuccessRate)
	}

	// Calculate average latency
	if result.TotalTests > 0 {
		result.AverageLatency = result.AverageLatency / time.Duration(result.TotalTests)
	}

	return result, nil
}

// TestJSONAgent performs a test of the simple_json agent type and other which are supported JSON output only
func TestJSONAgent(ctx context.Context, p provider.Provider, agentType provider.ProviderOptionsType, simpleJSONSuites []TestSuite, verbose bool) (AgentTestResult, error) {
	result := AgentTestResult{
		AgentType: string(agentType),
		ModelName: p.Model(agentType),
	}

	// Run all simple_json test suites
	for _, suite := range simpleJSONSuites {
		testResults := RunTestSuite(ctx, p, agentType, suite)

		// Determine if basic or advanced test
		if suite.Basic {
			result.BasicTests = append(result.BasicTests, testResults...)
		} else {
			result.AdvancedTests = append(result.AdvancedTests, testResults...)
		}

		// Update success counts
		for _, tr := range testResults {
			result.TotalTests++
			if tr.Success {
				result.TotalSuccess++
			}
			result.AverageLatency += time.Duration(tr.LatencyMs) * time.Millisecond
		}
	}

	// Calculate average latency
	if result.TotalTests > 0 {
		result.AverageLatency = result.AverageLatency / time.Duration(result.TotalTests)
	}

	return result, nil
}
