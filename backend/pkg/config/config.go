package config

import (
	"net/url"
	"reflect"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	// General
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://pentagiuser:pentagipass@pgvector:5432/pentagidb?sslmode=disable"`
	Debug       bool   `env:"DEBUG" envDefault:"false"`
	DataDir     string `env:"DATA_DIR" envDefault:"./data"`
	AskUser     bool   `env:"ASK_USER" envDefault:"false"`

	// Docker (terminal) settings
	DockerInside                 bool   `env:"DOCKER_INSIDE" envDefault:"false"`
	DockerNetAdmin               bool   `env:"DOCKER_NET_ADMIN" envDefault:"false"`
	DockerSocket                 string `env:"DOCKER_SOCKET"`
	DockerNetwork                string `env:"DOCKER_NETWORK"`
	DockerPublicIP               string `env:"DOCKER_PUBLIC_IP" envDefault:"0.0.0.0"`
	DockerWorkDir                string `env:"DOCKER_WORK_DIR"`
	DockerDefaultImage           string `env:"DOCKER_DEFAULT_IMAGE" envDefault:"debian:latest"`
	DockerDefaultImageForPentest string `env:"DOCKER_DEFAULT_IMAGE_FOR_PENTEST" envDefault:"vxcontrol/kali-linux"`

	// HTTP and GraphQL server settings
	ServerPort   int    `env:"SERVER_PORT" envDefault:"8080"`
	ServerHost   string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	ServerUseSSL bool   `env:"SERVER_USE_SSL" envDefault:"false"`
	ServerSSLKey string `env:"SERVER_SSL_KEY"`
	ServerSSLCrt string `env:"SERVER_SSL_CRT"`

	// Frontend static URL
	StaticURL   *url.URL `env:"STATIC_URL"`
	StaticDir   string   `env:"STATIC_DIR" envDefault:"./fe"`
	CorsOrigins []string `env:"CORS_ORIGINS" envDefault:"*"`

	// Cookie signing salt
	CookieSigningSalt string `env:"COOKIE_SIGNING_SALT"`

	// Scraper (browser)
	ScraperPublicURL  string `env:"SCRAPER_PUBLIC_URL" envDefault:"https://someuser:somepass@scraper"`
	ScraperPrivateURL string `env:"SCRAPER_PRIVATE_URL" envDefault:"https://someuser:somepass@scraper"`

	// OpenAI
	OpenAIKey       string `env:"OPEN_AI_KEY"`
	OpenAIServerURL string `env:"OPEN_AI_SERVER_URL" envDefault:"https://api.openai.com/v1"`

	// Anthropic
	AnthropicAPIKey    string `env:"ANTHROPIC_API_KEY"`
	AnthropicServerURL string `env:"ANTHROPIC_SERVER_URL" envDefault:"https://api.anthropic.com/v1"`

	// Embedding provider
	EmbeddingURL           string `env:"EMBEDDING_URL"`
	EmbeddingKey           string `env:"EMBEDDING_KEY"`
	EmbeddingModel         string `env:"EMBEDDING_MODEL"`
	EmbeddingStripNewLines bool   `env:"EMBEDDING_STRIP_NEW_LINES" envDefault:"true"`
	EmbeddingBatchSize     int    `env:"EMBEDDING_BATCH_SIZE" envDefault:"512"`
	EmbeddingProvider      string `env:"EMBEDDING_PROVIDER" envDefault:"openai"`

	// Summarizer
	SummarizerPreserveLast   bool `env:"SUMMARIZER_PRESERVE_LAST" envDefault:"true"`
	SummarizerUseQA          bool `env:"SUMMARIZER_USE_QA" envDefault:"true"`
	SummarizerSumHumanInQA   bool `env:"SUMMARIZER_SUM_MSG_HUMAN_IN_QA" envDefault:"false"`
	SummarizerLastSecBytes   int  `env:"SUMMARIZER_LAST_SEC_BYTES" envDefault:"51200"`
	SummarizerMaxBPBytes     int  `env:"SUMMARIZER_MAX_BP_BYTES" envDefault:"16384"`
	SummarizerMaxQASections  int  `env:"SUMMARIZER_MAX_QA_SECTIONS" envDefault:"10"`
	SummarizerMaxQABytes     int  `env:"SUMMARIZER_MAX_QA_BYTES" envDefault:"65536"`
	SummarizerKeepQASections int  `env:"SUMMARIZER_KEEP_QA_SECTIONS" envDefault:"1"`

	// Custom LLM provider
	LLMServerURL             string `env:"LLM_SERVER_URL"`
	LLMServerKey             string `env:"LLM_SERVER_KEY"`
	LLMServerModel           string `env:"LLM_SERVER_MODEL"`
	LLMServerConfig          string `env:"LLM_SERVER_CONFIG_PATH"`
	LLMServerLegacyReasoning bool   `env:"LLM_SERVER_LEGACY_REASONING" envDefault:"false"`

	// DuckDuckGo search engine
	DuckDuckGoEnabled bool `env:"DUCKDUCKGO_ENABLED" envDefault:"true"`

	// Google search engine
	GoogleAPIKey string `env:"GOOGLE_API_KEY"`
	GoogleCXKey  string `env:"GOOGLE_CX_KEY"`
	GoogleLRKey  string `env:"GOOGLE_LR_KEY" envDefault:"lang_en"`

	// OAuth google
	OAuthGoogleClientID     string `env:"OAUTH_GOOGLE_CLIENT_ID"`
	OAuthGoogleClientSecret string `env:"OAUTH_GOOGLE_CLIENT_SECRET"`

	// OAuth github
	OAuthGithubClientID     string `env:"OAUTH_GITHUB_CLIENT_ID"`
	OAuthGithubClientSecret string `env:"OAUTH_GITHUB_CLIENT_SECRET"`

	// Public URL for auth callback
	PublicURL string `env:"PUBLIC_URL" envDefault:""`

	// Traversaal search engine
	TraversaalAPIKey string `env:"TRAVERSAAL_API_KEY"`

	// Tavily search engine
	TavilyAPIKey string `env:"TAVILY_API_KEY"`

	// Perplexity search engine
	PerplexityAPIKey      string `env:"PERPLEXITY_API_KEY"`
	PerplexityModel       string `env:"PERPLEXITY_MODEL" envDefault:"sonar"`
	PerplexityContextSize string `env:"PERPLEXITY_CONTEXT_SIZE" envDefault:"low"`

	// Searxng search engine
	SearxngURL        string `env:"SEARXNG_URL"`
	SearxngCategories string `env:"SEARXNG_CATEGORIES" envDefault:"general"`
	SearxngLanguage   string `env:"SEARXNG_LANGUAGE" envDefault:"en"`
	SearxngSafeSearch string `env:"SEARXNG_SAFESearch" envDefault:"0"`
	SearxngTimeRange  string `env:"SEARXNG_TIME_RANGE"`
	SearxngProxyURL   string `env:"SEARXNG_PROXY_URL"`

	// Assistant
	AssistantUseAgents                bool `env:"ASSISTANT_USE_AGENTS" envDefault:"false"`
	AssistantSummarizerPreserveLast   bool `env:"ASSISTANT_SUMMARIZER_PRESERVE_LAST" envDefault:"true"`
	AssistantSummarizerLastSecBytes   int  `env:"ASSISTANT_SUMMARIZER_LAST_SEC_BYTES" envDefault:"76800"`
	AssistantSummarizerMaxBPBytes     int  `env:"ASSISTANT_SUMMARIZER_MAX_BP_BYTES" envDefault:"16384"`
	AssistantSummarizerMaxQASections  int  `env:"ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS" envDefault:"7"`
	AssistantSummarizerMaxQABytes     int  `env:"ASSISTANT_SUMMARIZER_MAX_QA_BYTES" envDefault:"76800"`
	AssistantSummarizerKeepQASections int  `env:"ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS" envDefault:"3"`

	// Proxy
	ProxyURL string `env:"PROXY_URL"`

	// Telemetry
	TelemetryEndpoint string `env:"OTEL_HOST"`

	// Langfuse
	LangfuseBaseURL   string `env:"LANGFUSE_BASE_URL"`
	LangfuseProjectID string `env:"LANGFUSE_PROJECT_ID"`
	LangfusePublicKey string `env:"LANGFUSE_PUBLIC_KEY"`
	LangfuseSecretKey string `env:"LANGFUSE_SECRET_KEY"`
}

func NewConfig() (*Config, error) {
	godotenv.Load()

	var config Config
	if err := env.ParseWithOptions(&config, env.Options{
		RequiredIfNoDef: false,
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(&url.URL{}): func(s string) (interface{}, error) {
				if s == "" {
					return nil, nil
				}
				return url.Parse(s)
			},
		},
	}); err != nil {
		return nil, err
	}

	return &config, nil
}
