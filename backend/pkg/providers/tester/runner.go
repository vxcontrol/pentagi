package tester

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/cases"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/llms"
)

// testRequest represents a test execution request
type testRequest struct {
	agentType pconfig.ProviderOptionsType
	testCase  cases.TestCase
	provider  provider.Provider
}

// testResponse represents a test execution result
type testResponse struct {
	agentType pconfig.ProviderOptionsType
	result    cases.TestResult
	err       error
}

// TestProvider executes tests for a provider with given options
func TestProvider(ctx context.Context, prv provider.Provider, opts ...TestOption) (ProviderTestResults, error) {
	config := applyOptions(opts)

	// load test registry
	var registry *cases.TestRegistry
	var err error

	if config.customRegistry != nil {
		registry = config.customRegistry
	} else {
		registry, err = cases.LoadBuiltinRegistry()
		if err != nil {
			return ProviderTestResults{}, fmt.Errorf("failed to load test registry: %w", err)
		}
	}

	var toolCallIDTemplate string
	if registry.ReplaysToolCalls(config.groups...) {
		toolCallIDTemplate = providerToolCallIDTemplate(ctx, prv, config.verbose)
	}

	// collect all test requests
	requests, err := collectTestRequests(registry, prv, config, toolCallIDTemplate)
	if err != nil {
		return ProviderTestResults{}, err
	}
	if len(requests) == 0 {
		return ProviderTestResults{}, fmt.Errorf("no tests to execute")
	}

	// execute tests in parallel
	responses := executeTestsParallel(ctx, requests, config)

	// group results by agent type
	return groupResults(responses), nil
}

func providerToolCallIDTemplate(ctx context.Context, prv provider.Provider, verbose bool) string {
	template, err := prv.GetToolCallIDTemplate(provider.WithoutToolCallIDTemplateCache(ctx), templates.NewDefaultPrompter())
	if err != nil && verbose {
		log.Printf("Warning: no tool call ID template, replayed tool calls keep the ids of the case: %v", err)
	}
	return template
}

// collectTestRequests gathers all test requests based on configuration
func collectTestRequests(
	registry *cases.TestRegistry,
	prv provider.Provider,
	config *testConfig,
	toolCallIDTemplate string,
) ([]testRequest, error) {
	var requests []testRequest

	// create agent type filter
	agentFilter := make(map[pconfig.ProviderOptionsType]bool)
	for _, agentType := range config.agentTypes {
		agentFilter[agentType] = true
	}

	// collect tests from each group
	for _, group := range config.groups {
		for _, agentType := range config.agentTypes {
			if len(agentFilter) > 0 && !agentFilter[agentType] {
				continue
			}

			suite, err := registry.GetTestSuite(group, cases.WithToolCallIDTemplate(toolCallIDTemplate))
			if err != nil {
				return nil, fmt.Errorf("test group %s: %w", group, err)
			}

			for _, testCase := range suite.Tests {
				if testCase.Streaming() && !config.streamingMode {
					continue
				}

				// filter test types based on agent type
				if !isTestCompatibleWithAgent(testCase.Type(), agentType) {
					continue
				}

				// skip capability-gated tests (adaptive thinking, reasoning
				// off, structured output) whose wire behavior this agent's
				// ACTUAL config wouldn't produce in a real PentAGI flow — see
				// capabilitySupported for why.
				if !capabilitySupported(prv, agentType, testCase.Capability()) {
					continue
				}

				requests = append(requests, testRequest{
					agentType: agentType,
					testCase:  testCase,
					provider:  prv,
				})
			}
		}

		if group != cases.TestGroupAdvanced {
			continue
		}

		// fileEditTestCase is hand-built in Go, not tests.yml (see
		// newFileEditTestCase) - its whole point is a genuinely dynamic
		// exchange that can't be expressed as a fixed message list, so it
		// can't come from the YAML-driven registry like every other test
		// here. It's added for TestGroupAdvanced so it still runs by default
		// without a special opt-in.
		for _, agentType := range config.agentTypes {
			if len(agentFilter) > 0 && !agentFilter[agentType] {
				continue
			}
			if !isTestCompatibleWithAgent(cases.TestTypeFileEdit, agentType) {
				continue
			}
			if !capabilitySupported(prv, agentType, cases.CapabilityNone) {
				continue
			}

			fileEdit, err := newFileEditTestCase()
			if err != nil {
				if config.verbose {
					log.Printf("Warning: failed to build file_edit test case: %v", err)
				}
				break // same failure for every agentType - no point retrying
			}

			requests = append(requests, testRequest{
				agentType: agentType,
				testCase:  fileEdit,
				provider:  prv,
			})
		}
	}

	return requests, nil
}

// executeTestsParallel runs tests concurrently using worker pool
func executeTestsParallel(ctx context.Context, requests []testRequest, config *testConfig) []testResponse {
	requestChan := make(chan testRequest, len(requests))
	responseChan := make(chan testResponse, len(requests))

	// start workers
	var wg sync.WaitGroup
	for i := 0; i < config.parallelWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testWorker(ctx, requestChan, responseChan, config.verbose)
		}()
	}

	// send requests
	for _, req := range requests {
		requestChan <- req
	}
	close(requestChan)

	// collect responses
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	var responses []testResponse
	for resp := range responseChan {
		responses = append(responses, resp)
	}

	return responses
}

// testWorker executes individual tests
func testWorker(ctx context.Context, requests <-chan testRequest, responses chan<- testResponse, verbose bool) {
	for req := range requests {
		resp := testResponse{
			agentType: req.agentType,
		}

		result, err := executeTest(ctx, req)
		if err != nil {
			resp.err = err
			if verbose {
				log.Printf("Test execution failed: %v", err)
			}
		} else {
			resp.result = result
			if verbose {
				status := "PASS"
				if !result.Success {
					status = "FAIL"
				}
				var errorStr string
				if result.Error != nil {
					errorStr = fmt.Sprintf("\n%v", result.Error)
				}
				log.Printf("[%s] %s - %s (%v)%s", status, req.agentType, result.Name, result.Latency, errorStr)
			}
		}

		responses <- resp
	}
}

// executeTest runs a single test case
func executeTest(ctx context.Context, req testRequest) (cases.TestResult, error) {
	startTime := time.Now()

	var response any
	var err error

	extra := req.testCase.ExtraOptions()

	var call func() (any, error)

	switch {
	case len(extra) > 0:
		if len(req.testCase.Messages()) == 0 {
			return cases.TestResult{}, fmt.Errorf(
				"test case carries extra call options but no messages: only the prompt-only path is available, and it cannot carry them")
		}
		call = func() (any, error) {
			return req.provider.CallWithExtraOptions(
				ctx,
				req.agentType,
				req.testCase.Messages(),
				req.testCase.Tools(),
				req.testCase.StreamingCallback(),
				extra...,
			)
		}
	case len(req.testCase.Messages()) > 0 && len(req.testCase.Tools()) > 0:
		call = func() (any, error) {
			return req.provider.CallWithTools(
				ctx,
				req.agentType,
				req.testCase.Messages(),
				req.testCase.Tools(),
				req.testCase.StreamingCallback(),
			)
		}
	case len(req.testCase.Messages()) > 0:
		call = func() (any, error) {
			return req.provider.CallEx(
				ctx,
				req.agentType,
				req.testCase.Messages(),
				req.testCase.StreamingCallback(),
			)
		}
	case req.testCase.Prompt() != "":
		call = func() (any, error) {
			return req.provider.Call(ctx, req.agentType, req.testCase.Prompt())
		}
	default:
		return cases.TestResult{}, fmt.Errorf("test case has no prompt or messages")
	}

	response, err = call()

	// MultiTurnTestCase (e.g. fileEditTestCase's read_file -> edit_file
	// exchange) answers each tool call itself and asks for another round by
	// returning true; every other TestCase leaves this a no-op.
	if multiTurn, ok := req.testCase.(cases.MultiTurnTestCase); ok {
		for err == nil {
			contentResp, isContentResp := response.(*llms.ContentResponse)
			if !isContentResp || !multiTurn.HandleToolResponse(contentResp) {
				break
			}
			response, err = call()
		}
	}

	latency := time.Since(startTime)

	if want := req.testCase.ExpectRefusal(); err != nil || want != cases.RefusalNone {
		return judgeRefusal(req.testCase, want, err, latency), nil
	}

	if req.testCase.ExpectTruncated() {
		return judgeTruncation(req.testCase, response, latency), nil
	}

	// let test case validate and produce result
	result := req.testCase.Execute(response, latency)
	result.Capability = req.testCase.Capability()
	result.StopReason = stopReasonOf(response)
	if !result.Success && cases.IsContentFilterStopReason(result.StopReason) {
		result.ContentFiltered = true
		result.Error = fmt.Errorf("the vendor's content filter stopped the answer (stop reason %q): %w",
			result.StopReason, result.Error)
	}
	return result, nil
}

func stopReasonOf(response any) string {
	contentResp, isContentResp := response.(*llms.ContentResponse)
	if !isContentResp || len(contentResp.Choices) == 0 || contentResp.Choices[0] == nil {
		return ""
	}
	return contentResp.Choices[0].StopReason
}

func judgeTruncation(tc cases.TestCase, response any, latency time.Duration) cases.TestResult {
	result := cases.TestResult{
		ID:         tc.ID(),
		Name:       tc.Name(),
		Type:       tc.Type(),
		Group:      tc.Group(),
		Capability: tc.Capability(),
		Latency:    latency,
		StopReason: stopReasonOf(response),
	}

	limitChange, limitChanged := outputLimitChange(response)

	switch {
	case result.StopReason == "":
		result.Error = fmt.Errorf("the answer carries no stop reason, so the output limit cannot be observed")
	case cases.IsTruncationStopReason(result.StopReason):
		result.Success = true
	case limitChanged:
		result.Success = true
		result.Unsupported = true
		result.Error = fmt.Errorf("the output limit the case asks for did not reach the vendor, "+
			"so the answer had room to finish (stop reason %q): %s", result.StopReason, limitChange)
	default:
		result.Error = fmt.Errorf("expected the answer to stop at the output limit, got stop reason %q", result.StopReason)
	}

	return result
}

func outputLimitChange(response any) (llms.Warning, bool) {
	contentResp, isContentResp := response.(*llms.ContentResponse)
	if !isContentResp {
		return llms.Warning{}, false
	}
	for _, warning := range contentResp.Warnings {
		if warning.Option == "WithMaxTokens" {
			return warning, true
		}
	}
	return llms.Warning{}, false
}

// judgeRefusal builds the result for a call that either failed or was required to fail.
func judgeRefusal(
	tc cases.TestCase,
	want cases.RefusalKind,
	err error,
	latency time.Duration,
) cases.TestResult {
	result := cases.TestResult{
		ID:         tc.ID(),
		Name:       tc.Name(),
		Type:       tc.Type(),
		Group:      tc.Group(),
		Capability: tc.Capability(),
		Latency:    latency,
	}

	switch got := cases.ClassifyRefusal(err); {
	case want == cases.RefusalNone:
		result.Unsupported = got.CountsAsUnsupported()
		result.Success = got == cases.RefusalStructuredOutputUnsupported
		result.Error = err
		if cases.IsContentFilterError(err) {
			result.ContentFiltered = true
			result.Error = fmt.Errorf("the vendor's content filter declined the request: %w", err)
		}
	case err == nil:
		result.Error = fmt.Errorf("expected the SDK to refuse with %q before calling the provider, got an answer", want)
	case got == want:
		result.Success = true
	default:
		result.Error = fmt.Errorf("expected the SDK to refuse with %q, got: %w", want, err)
	}

	return result
}

// groupResults organizes test results by agent type
func groupResults(responses []testResponse) ProviderTestResults {
	resultMap := make(map[pconfig.ProviderOptionsType][]cases.TestResult)

	// group by agent type
	for _, resp := range responses {
		if resp.err != nil {
			// create error result
			errorResult := cases.TestResult{
				ID:      "error",
				Name:    fmt.Sprintf("Execution Error: %v", resp.err),
				Success: false,
				Error:   resp.err,
			}
			resultMap[resp.agentType] = append(resultMap[resp.agentType], errorResult)
		} else {
			resultMap[resp.agentType] = append(resultMap[resp.agentType], resp.result)
		}
	}

	// map to ProviderTestResults structure
	return ProviderTestResults{
		Simple:       AgentTestResults(resultMap[pconfig.OptionsTypeSimple]),
		SimpleJSON:   AgentTestResults(resultMap[pconfig.OptionsTypeSimpleJSON]),
		PrimaryAgent: AgentTestResults(resultMap[pconfig.OptionsTypePrimaryAgent]),
		Assistant:    AgentTestResults(resultMap[pconfig.OptionsTypeAssistant]),
		Generator:    AgentTestResults(resultMap[pconfig.OptionsTypeGenerator]),
		Refiner:      AgentTestResults(resultMap[pconfig.OptionsTypeRefiner]),
		Adviser:      AgentTestResults(resultMap[pconfig.OptionsTypeAdviser]),
		Reflector:    AgentTestResults(resultMap[pconfig.OptionsTypeReflector]),
		Searcher:     AgentTestResults(resultMap[pconfig.OptionsTypeSearcher]),
		Enricher:     AgentTestResults(resultMap[pconfig.OptionsTypeEnricher]),
		Coder:        AgentTestResults(resultMap[pconfig.OptionsTypeCoder]),
		Installer:    AgentTestResults(resultMap[pconfig.OptionsTypeInstaller]),
		Pentester:    AgentTestResults(resultMap[pconfig.OptionsTypePentester]),
	}
}

// isTestCompatibleWithAgent determines if a test type is compatible with an agent type
func isTestCompatibleWithAgent(testType cases.TestType, agentType pconfig.ProviderOptionsType) bool {
	switch agentType {
	case pconfig.OptionsTypeSimpleJSON:
		// simpleJSON agent only handles JSON tests
		return testType == cases.TestTypeJSON
	default:
		switch testType {
		case cases.TestTypeJSON:
			return false
		case cases.TestTypeTool, cases.TestTypeFileEdit:
			return agentType.UsesTools()
		default:
			return true
		}
	}
}
