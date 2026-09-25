package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/bedrock"
	"pentagi/pkg/providers/custom"
	"pentagi/pkg/providers/deepseek"
	"pentagi/pkg/providers/gemini"
	"pentagi/pkg/providers/glm"
	"pentagi/pkg/providers/kimi"
	"pentagi/pkg/providers/minimax"
	"pentagi/pkg/providers/mistral"
	"pentagi/pkg/providers/ollama"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/qwen"
	"pentagi/pkg/providers/tester"
	"pentagi/pkg/providers/tester/cases"
	"pentagi/pkg/providers/xai"
	"pentagi/pkg/version"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	envFile := flag.String("env", ".env", "Path to environment file")
	providerType := flag.String("type", "custom", "Provider type [custom, openai, anthropic, gemini, bedrock, ollama, deepseek, glm, kimi, qwen, minimax, mistral, xai]")
	providerName := flag.String("name", "", "Provider name using as PROVDER_NAME/MODEL_NAME while building provider config")
	configPath := flag.String("config", "", "Path to provider config file")
	testsPath := flag.String("tests", "", "Path to custom tests YAML file")
	reportPath := flag.String("report", "", "Path to write report file")
	agentTypes := flag.String("agents", "all", "Comma-separated agent types to test")
	testGroups := flag.String("groups", "all", "Comma-separated test groups to run")
	workers := flag.Int("workers", 4, "Number of workers to use")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	checkRouting := flag.Bool("check-routing", false,
		"Ask each door which models it serves and report catalogue entries it did not name")
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

	if *checkRouting {
		reportRouting(providers.CheckRouting(cfg, &http.Client{Timeout: 15 * time.Second}))
		return
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
		allGroups := []cases.TestGroup{
			cases.TestGroupBasic,
			cases.TestGroupAdvanced,
			cases.TestGroupJSON,
			cases.TestGroupKnowledge,
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
			log.Fatalf("Error writing report: %v", err)
		}

		fmt.Printf("Report written to %s\n", *reportPath)
	}
}

func createProvider(providerType string, cfg *config.Config) (provider.Provider, error) {
	switch providerType {
	case "custom":
		providerConfig, err := custom.DefaultProviderConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating custom provider config: %w", err)
		}
		return custom.New(cfg, provider.DefaultProviderNameCustom, providerConfig, providers.EnrichCatalogCapabilities)

	case "openai":
		if cfg.OpenAIKey == "" {
			return nil, fmt.Errorf("OpenAI key is not set")
		}
		providerConfig, err := openai.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating openai provider config: %w", err)
		}
		return openai.New(cfg, provider.DefaultProviderNameOpenAI, providerConfig)

	case "anthropic":
		providerConfig, err := anthropic.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating anthropic provider config: %w", err)
		}
		return anthropic.New(cfg, provider.DefaultProviderNameAnthropic, providerConfig)

	case "gemini":
		if cfg.GeminiAPIKey == "" {
			return nil, fmt.Errorf("API key for Gemini is not set")
		}
		providerConfig, err := gemini.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating gemini provider config: %w", err)
		}
		return gemini.New(cfg, provider.DefaultProviderNameGemini, providerConfig)

	case "bedrock":
		if !cfg.BedrockDefaultAuth && cfg.BedrockBearerToken == "" &&
			(cfg.BedrockAccessKey == "" || cfg.BedrockSecretKey == "") {
			return nil, fmt.Errorf("authentication for Bedrock is not set: provide " +
				"BEDROCK_DEFAULT_AUTH=true, BEDROCK_BEARER_TOKEN, or " +
				"BEDROCK_ACCESS_KEY_ID+BEDROCK_SECRET_ACCESS_KEY")
		}
		providerConfig, err := bedrock.DefaultProviderConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating bedrock provider config: %w", err)
		}
		return bedrock.New(cfg, provider.DefaultProviderNameBedrock, providerConfig)

	case "ollama":
		if cfg.OllamaServerURL == "" {
			return nil, fmt.Errorf("server URL for Ollama is not set")
		}
		providerConfig, err := ollama.DefaultProviderConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating ollama provider config: %w", err)
		}
		return ollama.New(cfg, provider.DefaultProviderNameOllama, providerConfig)

	case "deepseek":
		if cfg.DeepSeekAPIKey == "" {
			return nil, fmt.Errorf("DeepSeek API key is not set")
		}
		providerConfig, err := deepseek.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating deepseek provider config: %w", err)
		}
		return deepseek.New(cfg, provider.DefaultProviderNameDeepSeek, providerConfig)

	case "glm":
		if cfg.GLMAPIKey == "" {
			return nil, fmt.Errorf("GLM Zhipu AI API key is not set")
		}
		providerConfig, err := glm.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating glm provider config: %w", err)
		}
		return glm.New(cfg, provider.DefaultProviderNameGLM, providerConfig)

	case "kimi":
		if cfg.KimiAPIKey == "" {
			return nil, fmt.Errorf("API key for Kimi Moonshot AI is not set")
		}
		providerConfig, err := kimi.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating kimi provider config: %w", err)
		}
		return kimi.New(cfg, provider.DefaultProviderNameKimi, providerConfig)

	case "qwen":
		if cfg.QwenAPIKey == "" {
			return nil, fmt.Errorf("API key for Qwen Alibaba Cloud is not set")
		}
		providerConfig, err := qwen.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating qwen provider config: %w", err)
		}
		return qwen.New(cfg, provider.DefaultProviderNameQwen, providerConfig)

	case "minimax":
		if cfg.MiniMaxAPIKey == "" {
			return nil, fmt.Errorf("MiniMax API key is not set")
		}
		providerConfig, err := minimax.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating minimax provider config: %w", err)
		}
		return minimax.New(cfg, provider.DefaultProviderNameMiniMax, providerConfig)

	case "mistral":
		if cfg.MistralAPIKey == "" {
			return nil, fmt.Errorf("mistral API key is not set")
		}
		providerConfig, err := mistral.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating mistral provider config: %w", err)
		}
		return mistral.New(cfg, provider.DefaultProviderNameMistral, providerConfig)

	case "xai":
		if cfg.XAIAPIKey == "" {
			return nil, fmt.Errorf("xAI API key is not set")
		}
		providerConfig, err := xai.DefaultProviderConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating xai provider config: %w", err)
		}
		return xai.New(cfg, provider.DefaultProviderNameXAI, providerConfig)

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

func parseTestGroups(groupStrings []string) []cases.TestGroup {
	var groups []cases.TestGroup
	validGroups := map[string]cases.TestGroup{
		"basic":     cases.TestGroupBasic,
		"advanced":  cases.TestGroupAdvanced,
		"json":      cases.TestGroupJSON,
		"knowledge": cases.TestGroupKnowledge,
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
				Name:            testResult.Name,
				Type:            string(testResult.Type),
				Capability:      string(testResult.Capability),
				Success:         testResult.Success,
				Unsupported:     testResult.Unsupported,
				ContentFiltered: testResult.ContentFiltered,
				Error:           testResult.Error,
				StopReason:      testResult.StopReason,
				Streaming:       testResult.Streaming,
				Reasoning:       testResult.Reasoning,
				LatencyMs:       testResult.Latency.Milliseconds(),
			}

			switch {
			case testResult.Capability != "":
				result.CapabilityTests = append(result.CapabilityTests, oldResult)
			case testResult.Group == cases.TestGroupBasic:
				result.BasicTests = append(result.BasicTests, oldResult)
			default:
				result.AdvancedTests = append(result.AdvancedTests, oldResult)
			}

			if testResult.Unsupported {
				continue
			}

			result.TotalTests++
			switch {
			case testResult.Success:
				result.TotalSuccess++
			case testResult.ContentFiltered:
				result.TotalFiltered++
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

func loadCustomTests(path string) (*cases.TestRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tests file: %w", err)
	}
	return cases.LoadRegistryFromYAML(data)
}

func reportRouting(checks []providers.RoutingCheck) {
	fmt.Println("Catalogue routing")
	fmt.Println("=================================================")

	var checked, missing, prefixed int
	for _, check := range checks {
		if !check.Checked {
			fmt.Printf("  %-10s %3d shipped   NOT CHECKED: %s\n", check.Door, check.Shipped, check.Skipped)
			continue
		}

		checked++
		missing += len(check.Missing)
		prefixed += len(check.PrefixedOnly)
		endpoint := check.Endpoint
		if check.Prefix != "" {
			endpoint += " as " + check.Prefix + "/<model>"
		}
		fmt.Printf("  %-10s %3d shipped   %d not callable by %s\n",
			check.Door, check.Shipped, len(check.Missing)+len(check.PrefixedOnly), endpoint)
		for _, name := range check.Missing {
			fmt.Printf("       %s\n", name)
		}
		for _, name := range check.PrefixedOnly {
			fmt.Printf("       %s   listed only behind a prefix\n", name)
		}
	}

	fmt.Println("=================================================")
	fmt.Printf("%d of %d doors answered; %d catalogue entries absent, %d listed only behind a prefix\n",
		checked, len(checks), missing, prefixed)
	if checked < len(checks) {
		fmt.Println("a door that did not answer proves nothing about its catalogue")
	}
}
