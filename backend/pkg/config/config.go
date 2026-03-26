package config

import (
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/vxcontrol/cloud/anonymizer/patterns"
	"github.com/vxcontrol/cloud/sdk"
)

type Config struct {
	// === Core System Configuration ===
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://pentagiuser:pentagipass@pgvector:5432/pentagidb?sslmode=disable"`
	Debug       bool   `env:"DEBUG" envDefault:"false"`
	DataDir     string `env:"DATA_DIR" envDefault:"./data"`
	AskUser     bool   `env:"ASK_USER" envDefault:"false"`

	// === PentAGI Cloud Service Integration ===
	InstallationID string `env:"INSTALLATION_ID"`
	LicenseKey     string `env:"LICENSE_KEY"`

	// === Container Runtime Configuration ===
	DockerInside                 bool   `env:"DOCKER_INSIDE" envDefault:"false"`
	DockerNetAdmin               bool   `env:"DOCKER_NET_ADMIN" envDefault:"false"`
	DockerSocket                 string `env:"DOCKER_SOCKET"`
	DockerNetwork                string `env:"DOCKER_NETWORK"`
	DockerPublicIP               string `env:"DOCKER_PUBLIC_IP" envDefault:"0.0.0.0"`
	DockerWorkDir                string `env:"DOCKER_WORK_DIR"`
	DockerDefaultImage           string `env:"DOCKER_DEFAULT_IMAGE" envDefault:"debian:latest"`
	DockerDefaultImageForPentest string `env:"DOCKER_DEFAULT_IMAGE_FOR_PENTEST" envDefault:"vxcontrol/kali-linux"`

	// === API Server Configuration ===
	ServerPort   int    `env:"SERVER_PORT" envDefault:"8080"`
	ServerHost   string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	ServerUseSSL bool   `env:"SERVER_USE_SSL" envDefault:"false"`
	ServerSSLKey string `env:"SERVER_SSL_KEY"`
	ServerSSLCrt string `env:"SERVER_SSL_CRT"`

	// === Frontend Static Assets ===
	StaticURL   *url.URL `env:"STATIC_URL"`
	StaticDir   string   `env:"STATIC_DIR" envDefault:"./fe"`
	CorsOrigins []string `env:"CORS_ORIGINS" envDefault:"*"`

	// === Authentication & Session Security ===
	CookieSigningSalt string `env:"COOKIE_SIGNING_SALT"`

	// === Web Scraper Service Endpoints ===
	ScraperPublicURL  string `env:"SCRAPER_PUBLIC_URL"`
	ScraperPrivateURL string `env:"SCRAPER_PRIVATE_URL"`

	// === LLM Provider: OpenAI ===
	OpenAIKey       string `env:"OPEN_AI_KEY"`
	OpenAIServerURL string `env:"OPEN_AI_SERVER_URL" envDefault:"https://api.openai.com/v1"`

	// === LLM Provider: Anthropic ===
	AnthropicAPIKey    string `env:"ANTHROPIC_API_KEY"`
	AnthropicServerURL string `env:"ANTHROPIC_SERVER_URL" envDefault:"https://api.anthropic.com/v1"`

	// === Vector Embedding Configuration ===
	EmbeddingURL           string `env:"EMBEDDING_URL"`
	EmbeddingKey           string `env:"EMBEDDING_KEY"`
	EmbeddingModel         string `env:"EMBEDDING_MODEL"`
	EmbeddingStripNewLines bool   `env:"EMBEDDING_STRIP_NEW_LINES" envDefault:"true"`
	EmbeddingBatchSize     int    `env:"EMBEDDING_BATCH_SIZE" envDefault:"512"`
	EmbeddingProvider      string `env:"EMBEDDING_PROVIDER" envDefault:"openai"`

	// === Chain Summarization Engine ===
	SummarizerPreserveLast   bool `env:"SUMMARIZER_PRESERVE_LAST" envDefault:"true"`
	SummarizerUseQA          bool `env:"SUMMARIZER_USE_QA" envDefault:"true"`
	SummarizerSumHumanInQA   bool `env:"SUMMARIZER_SUM_MSG_HUMAN_IN_QA" envDefault:"false"`
	SummarizerLastSecBytes   int  `env:"SUMMARIZER_LAST_SEC_BYTES" envDefault:"51200"`
	SummarizerMaxBPBytes     int  `env:"SUMMARIZER_MAX_BP_BYTES" envDefault:"16384"`
	SummarizerMaxQASections  int  `env:"SUMMARIZER_MAX_QA_SECTIONS" envDefault:"10"`
	SummarizerMaxQABytes     int  `env:"SUMMARIZER_MAX_QA_BYTES" envDefault:"65536"`
	SummarizerKeepQASections int  `env:"SUMMARIZER_KEEP_QA_SECTIONS" envDefault:"1"`

	// === LLM Provider: Custom/Self-Hosted ===
	LLMServerURL               string `env:"LLM_SERVER_URL"`
	LLMServerKey               string `env:"LLM_SERVER_KEY"`
	LLMServerModel             string `env:"LLM_SERVER_MODEL"`
	LLMServerProvider          string `env:"LLM_SERVER_PROVIDER"`
	LLMServerConfig            string `env:"LLM_SERVER_CONFIG_PATH"`
	LLMServerLegacyReasoning   bool   `env:"LLM_SERVER_LEGACY_REASONING" envDefault:"false"`
	LLMServerPreserveReasoning bool   `env:"LLM_SERVER_PRESERVE_REASONING" envDefault:"false"`

	// === LLM Provider: Ollama (Local/Remote) ===
	OllamaServerURL               string `env:"OLLAMA_SERVER_URL"`
	OllamaServerAPIKey            string `env:"OLLAMA_SERVER_API_KEY"`
	OllamaServerModel             string `env:"OLLAMA_SERVER_MODEL"`
	OllamaServerConfig            string `env:"OLLAMA_SERVER_CONFIG_PATH"`
	OllamaServerPullModelsTimeout int    `env:"OLLAMA_SERVER_PULL_MODELS_TIMEOUT" envDefault:"600"`
	OllamaServerPullModelsEnabled bool   `env:"OLLAMA_SERVER_PULL_MODELS_ENABLED" envDefault:"false"`
	OllamaServerLoadModelsEnabled bool   `env:"OLLAMA_SERVER_LOAD_MODELS_ENABLED" envDefault:"false"`

	// === LLM Provider: Google Gemini ===
	GeminiAPIKey    string `env:"GEMINI_API_KEY"`
	GeminiServerURL string `env:"GEMINI_SERVER_URL" envDefault:"https://generativelanguage.googleapis.com"`

	// === LLM Provider: AWS Bedrock ===
	BedrockRegion       string `env:"BEDROCK_REGION" envDefault:"us-east-1"`
	BedrockDefaultAuth  bool   `env:"BEDROCK_DEFAULT_AUTH" envDefault:"false"`
	BedrockBearerToken  string `env:"BEDROCK_BEARER_TOKEN"`
	BedrockAccessKey    string `env:"BEDROCK_ACCESS_KEY_ID"`
	BedrockSecretKey    string `env:"BEDROCK_SECRET_ACCESS_KEY"`
	BedrockSessionToken string `env:"BEDROCK_SESSION_TOKEN"`
	BedrockServerURL    string `env:"BEDROCK_SERVER_URL"`

	// === LLM Provider: DeepSeek ===
	DeepSeekAPIKey    string `env:"DEEPSEEK_API_KEY"`
	DeepSeekServerURL string `env:"DEEPSEEK_SERVER_URL" envDefault:"https://api.deepseek.com"`
	DeepSeekProvider  string `env:"DEEPSEEK_PROVIDER"`

	// === LLM Provider: GLM (Zhipu AI) ===
	GLMAPIKey    string `env:"GLM_API_KEY"`
	GLMServerURL string `env:"GLM_SERVER_URL" envDefault:"https://api.z.ai/api/paas/v4"`
	GLMProvider  string `env:"GLM_PROVIDER"`

	// === LLM Provider: Kimi (Moonshot AI) ===
	KimiAPIKey    string `env:"KIMI_API_KEY"`
	KimiServerURL string `env:"KIMI_SERVER_URL" envDefault:"https://api.moonshot.ai/v1"`
	KimiProvider  string `env:"KIMI_PROVIDER"`

	// === LLM Provider: Qwen (Tongyi Qianwen) ===
	QwenAPIKey    string `env:"QWEN_API_KEY"`
	QwenServerURL string `env:"QWEN_SERVER_URL" envDefault:"https://dashscope-us.aliyuncs.com/compatible-mode/v1"`
	QwenProvider  string `env:"QWEN_PROVIDER"`

	// === Search Engine: DuckDuckGo ===
	DuckDuckGoEnabled    bool   `env:"DUCKDUCKGO_ENABLED" envDefault:"true"`
	DuckDuckGoRegion     string `env:"DUCKDUCKGO_REGION"`
	DuckDuckGoSafeSearch string `env:"DUCKDUCKGO_SAFESEARCH"`
	DuckDuckGoTimeRange  string `env:"DUCKDUCKGO_TIME_RANGE"`

	// Sploitus exploit aggregator (https://sploitus.com)
	// service under cloudflare protection, IP should have good reputation to avoid being blocked
	SploitusEnabled bool `env:"SPLOITUS_ENABLED" envDefault:"false"`

	// === Search Engine: Google Custom Search ===
	GoogleAPIKey string `env:"GOOGLE_API_KEY"`
	GoogleCXKey  string `env:"GOOGLE_CX_KEY"`
	GoogleLRKey  string `env:"GOOGLE_LR_KEY" envDefault:"lang_en"`

	// === OAuth Provider: Google ===
	OAuthGoogleClientID     string `env:"OAUTH_GOOGLE_CLIENT_ID"`
	OAuthGoogleClientSecret string `env:"OAUTH_GOOGLE_CLIENT_SECRET"`

	// === OAuth Provider: GitHub ===
	OAuthGithubClientID     string `env:"OAUTH_GITHUB_CLIENT_ID"`
	OAuthGithubClientSecret string `env:"OAUTH_GITHUB_CLIENT_SECRET"`

	// === OAuth Callback Configuration ===
	PublicURL string `env:"PUBLIC_URL" envDefault:""`

	// === Search Engine: Traversaal AI ===
	TraversaalAPIKey string `env:"TRAVERSAAL_API_KEY"`

	// === Search Engine: Tavily AI ===
	TavilyAPIKey string `env:"TAVILY_API_KEY"`

	// === Search Engine: Perplexity AI ===
	PerplexityAPIKey      string `env:"PERPLEXITY_API_KEY"`
	PerplexityModel       string `env:"PERPLEXITY_MODEL" envDefault:"sonar"`
	PerplexityContextSize string `env:"PERPLEXITY_CONTEXT_SIZE" envDefault:"low"`

	// === Search Engine: SearXNG (Self-Hosted) ===
	SearxngURL        string `env:"SEARXNG_URL"`
	SearxngCategories string `env:"SEARXNG_CATEGORIES" envDefault:"general"`
	SearxngLanguage   string `env:"SEARXNG_LANGUAGE"`
	SearxngSafeSearch string `env:"SEARXNG_SAFESEARCH" envDefault:"0"`
	SearxngTimeRange  string `env:"SEARXNG_TIME_RANGE"`
	SearxngTimeout    int    `env:"SEARXNG_TIMEOUT"`

	// === AI Assistant Mode Configuration ===
	AssistantUseAgents                bool `env:"ASSISTANT_USE_AGENTS" envDefault:"false"`
	AssistantSummarizerPreserveLast   bool `env:"ASSISTANT_SUMMARIZER_PRESERVE_LAST" envDefault:"true"`
	AssistantSummarizerLastSecBytes   int  `env:"ASSISTANT_SUMMARIZER_LAST_SEC_BYTES" envDefault:"76800"`
	AssistantSummarizerMaxBPBytes     int  `env:"ASSISTANT_SUMMARIZER_MAX_BP_BYTES" envDefault:"16384"`
	AssistantSummarizerMaxQASections  int  `env:"ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS" envDefault:"7"`
	AssistantSummarizerMaxQABytes     int  `env:"ASSISTANT_SUMMARIZER_MAX_QA_BYTES" envDefault:"76800"`
	AssistantSummarizerKeepQASections int  `env:"ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS" envDefault:"3"`

	// === Network Proxy Settings ===
	ProxyURL string `env:"PROXY_URL"`

	// SSL Trusted CA Certificate Path (for external communication with LLM backends)
	ExternalSSLCAPath   string `env:"EXTERNAL_SSL_CA_PATH" envDefault:""`
	ExternalSSLInsecure bool   `env:"EXTERNAL_SSL_INSECURE" envDefault:"false"`

	// HTTP client timeout in seconds for external API calls (LLM providers, search tools, etc.)
	// A value of 0 means no timeout (not recommended).
	HTTPClientTimeout int `env:"HTTP_CLIENT_TIMEOUT" envDefault:"600"`

	// === Observability: OpenTelemetry Collector ===
	TelemetryEndpoint string `env:"OTEL_HOST"`

	// === Observability: Langfuse LLM Analytics ===
	LangfuseBaseURL   string `env:"LANGFUSE_BASE_URL"`
	LangfuseProjectID string `env:"LANGFUSE_PROJECT_ID"`
	LangfusePublicKey string `env:"LANGFUSE_PUBLIC_KEY"`
	LangfuseSecretKey string `env:"LANGFUSE_SECRET_KEY"`

	// === Knowledge Graph: Graphiti + Neo4j ===
	GraphitiEnabled bool   `env:"GRAPHITI_ENABLED" envDefault:"false"`
	GraphitiTimeout int    `env:"GRAPHITI_TIMEOUT" envDefault:"30"`
	GraphitiURL     string `env:"GRAPHITI_URL"`

	// === Agent Execution Monitoring ===
	ExecutionMonitorEnabled        bool `env:"EXECUTION_MONITOR_ENABLED" envDefault:"false"`
	ExecutionMonitorSameToolLimit  int  `env:"EXECUTION_MONITOR_SAME_TOOL_LIMIT" envDefault:"5"`
	ExecutionMonitorTotalToolLimit int  `env:"EXECUTION_MONITOR_TOTAL_TOOL_LIMIT" envDefault:"10"`

	// === Agent Tool Execution Limits ===
	MaxGeneralAgentToolCalls int `env:"MAX_GENERAL_AGENT_TOOL_CALLS" envDefault:"100"`
	MaxLimitedAgentToolCalls int `env:"MAX_LIMITED_AGENT_TOOL_CALLS" envDefault:"20"`

	// === Agent Planning Phase Configuration ===
	AgentPlanningStepEnabled bool `env:"AGENT_PLANNING_STEP_ENABLED" envDefault:"false"`
}

func NewConfig() (*Config, error) {
	// Attempt to load .env file (silently ignore if not present)
	_ = godotenv.Load()

	var config Config
	if err := env.ParseWithOptions(&config, env.Options{
		RequiredIfNoDef: false,
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(&url.URL{}): func(s string) (any, error) {
				if s == "" {
					return nil, nil
				}
				return url.Parse(s)
			},
		},
	}); err != nil {
		return nil, err
	}

	ensureInstallationID(&config)
	ensureLicenseKey(&config)

	return &config, nil
}

func ensureInstallationID(config *Config) {
	// validate current installation ID from environment
	if config.InstallationID != "" && uuid.Validate(config.InstallationID) == nil {
		return
	}

	// check local file for installation ID
	installationIDPath := filepath.Join(config.DataDir, "installation_id")
	installationID, err := os.ReadFile(installationIDPath)
	if err != nil {
		config.InstallationID = uuid.New().String()
	} else if uuid.Validate(string(installationID)) == nil {
		config.InstallationID = string(installationID)
	} else {
		config.InstallationID = uuid.New().String()
	}

	// write installation ID to local file
	_ = os.WriteFile(installationIDPath, []byte(config.InstallationID), 0644)
}

func ensureLicenseKey(config *Config) {
	// validate current license key from environment
	if config.LicenseKey == "" {
		return
	}

	// check license key validity, if invalid, set to empty
	info, err := sdk.IntrospectLicenseKey(config.LicenseKey)
	if err != nil {
		config.LicenseKey = ""
	} else if !info.IsValid() {
		config.LicenseKey = ""
	}
}

// GetSecretPatterns returns a list of patterns for all secrets in the config
func (c *Config) GetSecretPatterns() []patterns.Pattern {
	var result []patterns.Pattern

	secrets := []struct {
		value string
		name  string
	}{
		{c.DatabaseURL, "Database URL"},
		{c.LicenseKey, "License Key"},
		{c.CookieSigningSalt, "Cookie Salt"},
		{c.OpenAIKey, "OpenAI Key"},
		{c.AnthropicAPIKey, "Anthropic Key"},
		{c.EmbeddingKey, "Embedding Key"},
		{c.LLMServerKey, "LLM Server Key"},
		{c.OllamaServerAPIKey, "Ollama Key"},
		{c.GeminiAPIKey, "Gemini Key"},
		{c.BedrockBearerToken, "Bedrock Token"},
		{c.BedrockAccessKey, "Bedrock Access Key"},
		{c.BedrockSecretKey, "Bedrock Secret Key"},
		{c.BedrockSessionToken, "Bedrock Session Token"},
		{c.DeepSeekAPIKey, "DeepSeek Key"},
		{c.GLMAPIKey, "GLM Key"},
		{c.KimiAPIKey, "Kimi Key"},
		{c.QwenAPIKey, "Qwen Key"},
		{c.GoogleAPIKey, "Google API Key"},
		{c.GoogleCXKey, "Google CX Key"},
		{c.OAuthGoogleClientID, "Google Client ID"},
		{c.OAuthGoogleClientSecret, "Google Client Secret"},
		{c.OAuthGithubClientID, "Github Client ID"},
		{c.OAuthGithubClientSecret, "Github Client Secret"},
		{c.TraversaalAPIKey, "Traversaal Key"},
		{c.TavilyAPIKey, "Tavily Key"},
		{c.PerplexityAPIKey, "Perplexity Key"},
		{c.ProxyURL, "Proxy URL"},
		{c.LangfusePublicKey, "Langfuse Public Key"},
		{c.LangfuseSecretKey, "Langfuse Secret Key"},
	}

	for _, s := range secrets {
		trimmed := strings.TrimSpace(s.value)
		if trimmed == "" {
			continue
		}

		// escape regex special characters
		escaped := regexp.QuoteMeta(trimmed)
		pattern := patterns.Pattern{
			Name:  s.name,
			Regex: "(?P<replace>" + escaped + ")",
		}
		result = append(result, pattern)
	}

	return result
}
