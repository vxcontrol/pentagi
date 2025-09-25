package tester

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/testdata"
)

// testRequest represents a test execution request
type testRequest struct {
	agentType pconfig.ProviderOptionsType
	testCase  testdata.TestCase
	provider  provider.Provider
}

// testResponse represents a test execution result
type testResponse struct {
	agentType pconfig.ProviderOptionsType
	result    testdata.TestResult
	err       error
}

// TestProvider executes tests for a provider with given options
func TestProvider(ctx context.Context, prv provider.Provider, opts ...TestOption) (ProviderTestResults, error) {
	config := applyOptions(opts)

	// load test registry
	var registry *testdata.TestRegistry
	var err error

	if config.customRegistry != nil {
		registry = config.customRegistry
	} else {
		registry, err = testdata.LoadBuiltinRegistry()
		if err != nil {
			return ProviderTestResults{}, fmt.Errorf("failed to load test registry: %w", err)
		}
	}

	// collect all test requests
	requests := collectTestRequests(registry, prv, config)
	if len(requests) == 0 {
		return ProviderTestResults{}, fmt.Errorf("no tests to execute")
	}

	// execute tests in parallel
	responses := executeTestsParallel(ctx, requests, config)

	// group results by agent type
	return groupResults(responses), nil
}

// collectTestRequests gathers all test requests based on configuration
func collectTestRequests(registry *testdata.TestRegistry, prv provider.Provider, config *testConfig) []testRequest {
	var requests []testRequest

	// create agent type filter
	agentFilter := make(map[pconfig.ProviderOptionsType]bool)
	for _, agentType := range config.agentTypes {
		agentFilter[agentType] = true
	}

	// collect tests from each group
	for _, group := range config.groups {
		suite, err := registry.GetTestSuite(group)
		if err != nil {
			if config.verbose {
				log.Printf("Warning: failed to get test suite for group %s: %v", group, err)
			}
			continue
		}

		for _, testCase := range suite.Tests {
			// skip streaming tests if disabled
			if testCase.Streaming() && !config.streamingMode {
				continue
			}

			// create requests for each agent type
			for _, agentType := range config.agentTypes {
				if len(agentFilter) > 0 && !agentFilter[agentType] {
					continue
				}

				// filter test types based on agent type
				if !isTestCompatibleWithAgent(testCase.Type(), agentType) {
					continue
				}

				requests = append(requests, testRequest{
					agentType: agentType,
					testCase:  testCase,
					provider:  prv,
				})
			}
		}
	}

	return requests
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
func executeTest(ctx context.Context, req testRequest) (testdata.TestResult, error) {
	startTime := time.Now()

	var response interface{}
	var err error

	// execute based on test type and available data
	switch {
	case len(req.testCase.Messages()) > 0 && len(req.testCase.Tools()) > 0:
		// tool calling with messages
		response, err = req.provider.CallWithTools(
			ctx,
			req.agentType,
			req.testCase.Messages(),
			req.testCase.Tools(),
			req.testCase.StreamingCallback(),
		)
	case len(req.testCase.Messages()) > 0:
		// messages without tools
		response, err = req.provider.CallEx(
			ctx,
			req.agentType,
			req.testCase.Messages(),
			req.testCase.StreamingCallback(),
		)
	case req.testCase.Prompt() != "":
		// simple prompt
		response, err = req.provider.Call(ctx, req.agentType, req.testCase.Prompt())
	default:
		return testdata.TestResult{}, fmt.Errorf("test case has no prompt or messages")
	}

	latency := time.Since(startTime)

	if err != nil {
		return testdata.TestResult{
			ID:      req.testCase.ID(),
			Name:    req.testCase.Name(),
			Type:    req.testCase.Type(),
			Group:   req.testCase.Group(),
			Success: false,
			Error:   err,
			Latency: latency,
		}, nil
	}

	// let test case validate and produce result
	return req.testCase.Execute(response, latency), nil
}

// groupResults organizes test results by agent type
func groupResults(responses []testResponse) ProviderTestResults {
	resultMap := make(map[pconfig.ProviderOptionsType][]testdata.TestResult)

	// group by agent type
	for _, resp := range responses {
		if resp.err != nil {
			// create error result
			errorResult := testdata.TestResult{
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
func isTestCompatibleWithAgent(testType testdata.TestType, agentType pconfig.ProviderOptionsType) bool {
	switch agentType {
	case pconfig.OptionsTypeSimpleJSON:
		// simpleJSON agent only handles JSON tests
		return testType == testdata.TestTypeJSON
	default:
		// all other agents handle everything except JSON tests
		return testType != testdata.TestTypeJSON
	}
}
