package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/bedrock"
	"pentagi/pkg/providers/custom"
	"pentagi/pkg/providers/gemini"
	"pentagi/pkg/providers/ollama"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester"
	"pentagi/pkg/providers/tester/testdata"
	"pentagi/pkg/version"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	envFile := flag.String("env", ".env", "Path to environment file")
	providerType := flag.String("type", "custom", "Provider type [custom, openai, anthropic, gemini, bedrock, ollama]")
	providerName := flag.String("name", "", "Provider name using as PROVDER_NAME/MODEL_NAME while building provider config")
	configPath := flag.String("config", "", "Path to provider config file")
	testsPath := flag.String("tests", "", "Path to custom tests YAML file")
	reportPath := flag.String("report", "", "Path to write report file")
	agentTypes := flag.String("agents", "all", "Comma-separated agent types to test")
	testGroups := flag.String("groups", "all", "Comma-separated test groups to run")
	workers := flag.Int("workers", 4, "Number of workers to use")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	logrus.Infof("Starting PentAGI Provider Configuration Tester %s", version.GetBinaryVersion())

	if err := godotenv.Load(*envFile); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if *configPath != "" {
		cfg.LLMServerConfig = *configPath
		cfg.OllamaServerConfig = *configPath
	}
	if *providerName != "" {
		cfg.LLMServerProvider = *providerName
	}

	prv, err := createProvider(*providerType, cfg)
	if err != nil {
		log.Fatalf("Error creating provider: %v", err)
	}

	fmt.Printf("Testing %s Provider\n", *providerType)
	fmt.Println("=================================================")

	var testOptions []tester.TestOption

	if *agentTypes != "all" {
		selectedTypes := parseAgentTypes(strings.Split(*agentTypes, ","))
		testOptions = append(testOptions, tester.WithAgentTypes(selectedTypes...))
	}

	if *testGroups != "all" {
		selectedGroups := parseTestGroups(strings.Split(*testGroups, ","))
		testOptions = append(testOptions, tester.WithGroups(selectedGroups...))
	} else {
		// Include all available groups when "all" is specified
		allGroups := []testdata.TestGroup{
			testdata.TestGroupBasic,
			testdata.TestGroupAdvanced,
			testdata.TestGroupJSON,
			testdata.TestGroupKnowledge,
		}
		testOptions = append(testOptions, tester.WithGroups(allGroups...))
	}

	if *testsPath != "" {
		registry, err := loadCustomTests(*testsPath)
		if err != nil {
			log.Fatalf("Error loading custom tests: %v", err)
		}
		testOptions = append(testOptions, tester.WithCustomRegistry(registry))
	}

	testOptions = append(
		testOptions,
		tester.WithVerbose(*verbose),
		tester.WithParallelWorkers(*workers),
	)

	results, err := tester.TestProvider(context.Background(), prv, testOptions...)
	if err != nil {
		log.Fatalf("Error running tests: %v", err)
	}

	agentResults := convertToAgentResults(results, prv)
	PrintSummaryReport(agentResults)

	if *reportPath != "" {
		if err := WriteReportToFile(agentResults, *reportPath); err != nil {
			log.Printf("Error writing report: %v", err)
		} else {
			fmt.Printf("Report written to %s\n", *reportPath)
		}
	}
}

func createProvider(providerType string, cfg *config.Config) (provider.Provider, error) {
	switch providerType {
	case "custom":
		providerConfig, err := custom.DefaultProviderConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating custom provider config: %w", err)
		}
		return custom.New(cfg, providerConfig)

	case "openai":
		if cfg.OpenAIKey == "" {
			return nil, fmt.Errorf("OpenAI key is not set")
		}
		providerConfig, err := openai.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating openai provider config: %w", err)
		}
		return openai.New(cfg, providerConfig)

	case "anthropic":
		if cfg.AnthropicAPIKey == "" {
			return nil, fmt.Errorf("Anthropic API key is not set")
		}
		providerConfig, err := anthropic.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating anthropic provider config: %w", err)
		}
		return anthropic.New(cfg, providerConfig)

	case "gemini":
		if cfg.GeminiAPIKey == "" {
			return nil, fmt.Errorf("Gemini API key is not set")
		}
		providerConfig, err := gemini.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating gemini provider config: %w", err)
		}
		return gemini.New(cfg, providerConfig)

	case "bedrock":
		if cfg.BedrockAccessKey == "" {
			return nil, fmt.Errorf("Bedrock access key is not set")
		}
		providerConfig, err := bedrock.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating bedrock provider config: %w", err)
		}
		return bedrock.New(cfg, providerConfig)

	case "ollama":
		if cfg.OllamaServerURL == "" {
			return nil, fmt.Errorf("Ollama server URL is not set")
		}
		providerConfig, err := ollama.DefaultProviderConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating ollama provider config: %w", err)
		}
		return ollama.New(cfg, providerConfig)

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

func parseAgentTypes(agentStrings []string) []pconfig.ProviderOptionsType {
	var agentTypes []pconfig.ProviderOptionsType
	validTypes := map[string]pconfig.ProviderOptionsType{
		"simple":        pconfig.OptionsTypeSimple,
		"simple_json":   pconfig.OptionsTypeSimpleJSON,
		"primary_agent": pconfig.OptionsTypePrimaryAgent,
		"assistant":     pconfig.OptionsTypeAssistant,
		"generator":     pconfig.OptionsTypeGenerator,
		"refiner":       pconfig.OptionsTypeRefiner,
		"adviser":       pconfig.OptionsTypeAdviser,
		"reflector":     pconfig.OptionsTypeReflector,
		"searcher":      pconfig.OptionsTypeSearcher,
		"enricher":      pconfig.OptionsTypeEnricher,
		"coder":         pconfig.OptionsTypeCoder,
		"installer":     pconfig.OptionsTypeInstaller,
		"pentester":     pconfig.OptionsTypePentester,
	}

	for _, agentStr := range agentStrings {
		agentStr = strings.TrimSpace(agentStr)
		if agentType, ok := validTypes[agentStr]; ok {
			agentTypes = append(agentTypes, agentType)
		} else {
			log.Printf("Warning: Unknown agent type '%s', skipping", agentStr)
		}
	}

	return agentTypes
}

func parseTestGroups(groupStrings []string) []testdata.TestGroup {
	var groups []testdata.TestGroup
	validGroups := map[string]testdata.TestGroup{
		"basic":     testdata.TestGroupBasic,
		"advanced":  testdata.TestGroupAdvanced,
		"json":      testdata.TestGroupJSON,
		"knowledge": testdata.TestGroupKnowledge,
	}

	for _, groupStr := range groupStrings {
		groupStr = strings.TrimSpace(groupStr)
		if group, ok := validGroups[groupStr]; ok {
			groups = append(groups, group)
		} else {
			log.Printf("Warning: Unknown test group '%s', skipping", groupStr)
		}
	}

	return groups
}

func convertToAgentResults(results tester.ProviderTestResults, prv provider.Provider) []AgentTestResult {
	var agentResults []AgentTestResult

	// Create mapping of agent types to their data
	agentTypeMap := map[pconfig.ProviderOptionsType]struct {
		name    string
		results tester.AgentTestResults
	}{
		pconfig.OptionsTypeSimple:       {"simple", results.Simple},
		pconfig.OptionsTypeSimpleJSON:   {"simple_json", results.SimpleJSON},
		pconfig.OptionsTypePrimaryAgent: {"primary_agent", results.PrimaryAgent},
		pconfig.OptionsTypeAssistant:    {"assistant", results.Assistant},
		pconfig.OptionsTypeGenerator:    {"generator", results.Generator},
		pconfig.OptionsTypeRefiner:      {"refiner", results.Refiner},
		pconfig.OptionsTypeAdviser:      {"adviser", results.Adviser},
		pconfig.OptionsTypeReflector:    {"reflector", results.Reflector},
		pconfig.OptionsTypeSearcher:     {"searcher", results.Searcher},
		pconfig.OptionsTypeEnricher:     {"enricher", results.Enricher},
		pconfig.OptionsTypeCoder:        {"coder", results.Coder},
		pconfig.OptionsTypeInstaller:    {"installer", results.Installer},
		pconfig.OptionsTypePentester:    {"pentester", results.Pentester},
	}

	// Use deterministic order from AllAgentTypes
	for _, agentType := range pconfig.AllAgentTypes {
		agentData, exists := agentTypeMap[agentType]
		if !exists {
			continue
		}

		agentTypeName := agentData.name
		agentTestResults := agentData.results
		if len(agentTestResults) == 0 {
			continue
		}

		result := AgentTestResult{
			AgentType: agentTypeName,
			ModelName: prv.Model(agentType),
		}

		var totalLatency time.Duration
		for _, testResult := range agentTestResults {
			oldResult := TestResult{
				Name:      testResult.Name,
				Type:      string(testResult.Type),
				Success:   testResult.Success,
				Error:     testResult.Error,
				Streaming: testResult.Streaming,
				Reasoning: testResult.Reasoning,
				LatencyMs: testResult.Latency.Milliseconds(),
			}

			if testResult.Group == testdata.TestGroupBasic {
				result.BasicTests = append(result.BasicTests, oldResult)
			} else {
				result.AdvancedTests = append(result.AdvancedTests, oldResult)
			}

			result.TotalTests++
			if testResult.Success {
				result.TotalSuccess++
			}
			if testResult.Reasoning {
				result.Reasoning = true
			}
			totalLatency += testResult.Latency
		}

		if result.TotalTests > 0 {
			result.AverageLatency = totalLatency / time.Duration(result.TotalTests)
		}

		agentResults = append(agentResults, result)
	}

	return agentResults
}

func loadCustomTests(path string) (*testdata.TestRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tests file: %w", err)
	}
	return testdata.LoadRegistryFromYAML(data)
}
