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
	DockerInside   bool   `env:"DOCKER_INSIDE" envDefault:"false"`
	DockerSocket   string `env:"DOCKER_SOCKET"`
	DockerNetwork  string `env:"DOCKER_NETWORK"`
	DockerPublicIP string `env:"DOCKER_PUBLIC_IP" envDefault:"0.0.0.0"`
	DockerWorkDir  string `env:"DOCKER_WORK_DIR"`

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

	// Custom LLM provider
	LLMServerURL   string `env:"LLM_SERVER_URL"`
	LLMServerKey   string `env:"LLM_SERVER_KEY"`
	LLMServerModel string `env:"LLM_SERVER_MODEL"`

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
