package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/custom"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/queue"

	"github.com/joho/godotenv"
)

// testRequest encapsulates a request to test a specific agent type
type testRequest struct {
	agentType        provider.ProviderOptionsType
	provider         provider.Provider
	basicSuites      []TestSuite
	advancedSuites   []TestSuite
	outputJSONSuites []TestSuite
	verboseMode      bool
}

// testResult encapsulates a result from testing a specific agent type
type testResult struct {
	agentType provider.ProviderOptionsType
	result    AgentTestResult
	err       error
}

// RunTests runs all tests for the specified provider with the given configuration
func RunTests(ctx context.Context, p provider.Provider, config TestConfig) ([]AgentTestResult, error) {
	// Get all available agent types
	allAgentTypes := []provider.ProviderOptionsType{
		provider.OptionsTypeSimple,
		provider.OptionsTypeSimpleJSON,
		provider.OptionsTypeAgent,
		provider.OptionsTypeAssistant,
		provider.OptionsTypeGenerator,
		provider.OptionsTypeRefiner,
		provider.OptionsTypeAdviser,
		provider.OptionsTypeReflector,
		provider.OptionsTypeSearcher,
		provider.OptionsTypeEnricher,
		provider.OptionsTypeCoder,
		provider.OptionsTypeInstaller,
		provider.OptionsTypePentester,
	}

	// Convert to string slice for filtering
	allAgentTypeStrings := make([]string, len(allAgentTypes))
	for i, at := range allAgentTypes {
		allAgentTypeStrings[i] = string(at)
	}

	// Filter agent types based on config
	selectedAgentTypes := FilterAgentTypes(allAgentTypeStrings, config.AgentTypes)

	// Filter agent types back to provider options
	selectedOptions := make([]provider.ProviderOptionsType, 0, len(selectedAgentTypes))
	for _, at := range selectedAgentTypes {
		selectedOptions = append(selectedOptions, provider.ProviderOptionsType(at))
	}

	// Get test suites
	basicSuites := GetBasicTestSuites()
	if IsTestGroupsFiltered("basic", config.TestGroups) {
		basicSuites = []TestSuite{}
	}
	advancedSuites := GetAdvancedTestSuites()
	if IsTestGroupsFiltered("advanced", config.TestGroups) {
		advancedSuites = []TestSuite{}
	}
	outputJSONSuites := GetOutputJSONTestSuites()
	if IsTestGroupsFiltered("json", config.TestGroups) {
		outputJSONSuites = []TestSuite{}
	}

	// Use queue for parallel execution
	numWorkers := len(selectedOptions)
	if numWorkers < 1 {
		numWorkers = 1
	}

	// Setup input/output channels
	input := make(chan testRequest, numWorkers)
	output := make(chan testResult, numWorkers)

	// Create processing function
	processFn := func(req testRequest) (testResult, error) {
		var res AgentTestResult
		var err error

		if req.agentType == provider.OptionsTypeSimpleJSON {
			res, err = TestJSONAgent(ctx, req.provider, req.agentType, req.outputJSONSuites, req.verboseMode)
		} else {
			res, err = TestAgent(ctx, req.provider, req.agentType, req.basicSuites, req.advancedSuites, req.verboseMode)
		}

		// Is reasoning used while testing?
		for _, test := range append(res.BasicTests, res.AdvancedTests...) {
			if test.Reasoning {
				res.Reasoning = true
				break
			}
		}

		return testResult{
			agentType: req.agentType,
			result:    res,
			err:       err,
		}, nil
	}

	// Create queue
	q := queue.NewQueue(input, output, numWorkers, processFn)
	if err := q.Start(); err != nil {
		return nil, fmt.Errorf("failed to start test queue: %w", err)
	}

	// Submit test requests
	for _, agentType := range selectedOptions {
		input <- testRequest{
			agentType:        agentType,
			provider:         p,
			basicSuites:      basicSuites,
			advancedSuites:   advancedSuites,
			outputJSONSuites: outputJSONSuites,
			verboseMode:      config.VerboseMode,
		}
	}
	close(input)

	// Collect results
	results := make([]AgentTestResult, 0, len(selectedOptions))
	resultMap := make(map[provider.ProviderOptionsType]AgentTestResult)
	var errs []error

	for range selectedOptions {
		result := <-output
		if result.err != nil {
			errs = append(errs, fmt.Errorf("error testing agent %s: %w", result.agentType, result.err))
		} else {
			resultMap[result.agentType] = result.result

			// Print detailed results in an unified format if verbose mode is enabled
			if config.VerboseMode {
				PrintTestingResults(result.result)
			}
		}
	}

	// Check for errors
	if len(errs) > 0 {
		errMsg := "Errors occurred during testing:"
		for _, err := range errs {
			errMsg += "\n  - " + err.Error()
		}
		return nil, errors.New(errMsg)
	}

	// Reorder results to maintain consistent order
	for _, agentType := range selectedOptions {
		if result, ok := resultMap[agentType]; ok {
			results = append(results, result)
		}
	}

	// Stop the queue
	if err := q.Stop(); err != nil && err != queue.ErrAlreadyStopped {
		log.Printf("Warning: Error stopping test queue: %v", err)
	}

	return results, nil
}

// PrintTestingResults prints detailed results in an unified format
func PrintTestingResults(result AgentTestResult) {
	// Print detailed results for each agent
	fmt.Printf("\nTesting agent: %s\n", result.AgentType)
	fmt.Printf("Model: %s\n", result.ModelName)
	fmt.Printf("Reasoning: %t\n", result.Reasoning)
	fmt.Println("-------------------------------------------------")

	PrintAgentResults(result)
}

func main() {
	// Parse command-line arguments
	envFile := flag.String("env", ".env", "Path to environment file")
	configType := flag.String("type", "custom", "Type of provider [custom, openai, anthropic]")
	configPath := flag.String("config", "", "Path to custom provider config")
	reportPath := flag.String("report", "", "Path to write the report file")
	agentTypes := flag.String("agents", "all", "Comma-separated list of agent types to test")
	testGroups := flag.String("tests", "all", "Comma-separated list of test groups to run")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Load environment variables
	err := godotenv.Load(*envFile)
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Initialize configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Set provider configuration path
	if *configPath != "" {
		cfg.LLMServerConfig = *configPath
	}

	// Initialize provider
	var provider provider.Provider
	switch *configType {
	case "custom":
		provider, err = custom.New(cfg)
		if err != nil {
			log.Fatalf("Error initializing custom provider: %v", err)
		}
	case "openai":
		if cfg.OpenAIKey == "" {
			log.Fatalf("OpenAI key is not set")
		}
		provider, err = openai.New(cfg)
		if err != nil {
			log.Fatalf("Error initializing openai provider: %v", err)
		}
	case "anthropic":
		if cfg.AnthropicAPIKey == "" {
			log.Fatalf("Anthropic API key is not set")
		}
		provider, err = anthropic.New(cfg)
		if err != nil {
			log.Fatalf("Error initializing anthropic provider: %v", err)
		}
	default:
		log.Fatalf("Invalid provider type: %s", *configType)
	}

	fmt.Println("Testing Custom Provider with configuration:", *configPath)
	fmt.Println("=================================================")

	// Parse agent types to test
	agentTypeList := []string{"all"}
	if *agentTypes != "all" {
		agentTypeList = strings.Split(*agentTypes, ",")
	}

	// Parse test groups to run
	testGroupList := []string{"all"}
	if *testGroups != "all" {
		testGroupList = strings.Split(*testGroups, ",")
	}

	// Create test configuration
	testConfig := TestConfig{
		AgentTypes:  agentTypeList,
		TestGroups:  testGroupList,
		ConfigPath:  *configPath,
		ReportPath:  *reportPath,
		VerboseMode: *verbose,
	}

	// Run tests for each agent
	ctx := context.Background()
	results, err := RunTests(ctx, provider, testConfig)
	if err != nil {
		log.Fatalf("Error running tests: %v", err)
	}

	// Print summary report
	PrintSummaryReport(results)

	// Write report to file if requested
	if *reportPath != "" {
		if err := WriteReportToFile(results, *reportPath); err != nil {
			log.Printf("Error writing report to file: %v", err)
		} else {
			fmt.Printf("Report written to %s\n", *reportPath)
		}
	}
}
