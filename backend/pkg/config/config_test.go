package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wasilibs/go-re2"
	"github.com/wasilibs/go-re2/experimental"
)

func TestConfig_GetSecretPatterns_SkipsBlankAndTrimsPaddedValues(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want []string
	}{
		{name: "an empty config", cfg: &Config{}},
		{
			name: "blank values are skipped",
			cfg:  &Config{OpenAIKey: "sk-1234", AnthropicAPIKey: "", GeminiAPIKey: "   ", DatabaseURL: "\t\n", LicenseKey: "ABCD-EFGH"},
			want: []string{"(?P<replace>ABCD-EFGH)", "(?P<replace>sk-1234)"},
		},
		{
			name: "padded values are trimmed",
			cfg:  &Config{OpenAIKey: "  sk-1234  ", GeminiAPIKey: "\tAIzaSyC123\n"},
			want: []string{"(?P<replace>sk-1234)", "(?P<replace>AIzaSyC123)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, pattern := range tt.cfg.GetSecretPatterns() {
				got = append(got, pattern.Regex)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// The values carry regex metacharacters, so a pattern that is not quoted fails to find its own secret.
func TestConfig_GetSecretPatterns_FindsEachSecretVerbatim(t *testing.T) {
	cfg := &Config{
		DatabaseURL:             "postgres://user:p@ssw0rd!@localhost:5432/db?sslmode=disable",
		LicenseKey:              "ABCD-EFGH-IJKL-MNOP",
		CookieSigningSalt:       "random-salt-string-12345",
		OpenAIKey:               "sk-proj-1234567890abcdefghijklmnopqrstuvwxyz",
		AnthropicAPIKey:         "sk-ant-api03-abcdefghijklmnopqrstuvwxyz1234567890",
		AnthropicIdentityToken:  "eyJhbGciOiJSUzI1NiJ9.e30.sig",
		EmbeddingKey:            "emb-123",
		LLMServerKey:            "llm-123",
		OllamaServerAPIKey:      "ollama-123",
		GeminiAPIKey:            "AIzaSyC1234567890abcdefghijklmnopqrstuvwxyz",
		BedrockBearerToken:      "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.example",
		BedrockAccessKey:        "AKIAIOSFODNN7EXAMPLE",
		BedrockSecretKey:        "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		BedrockSessionToken:     "FwoGZXIvYXdzEBYaDD1234567890EXAMPLE",
		DeepSeekAPIKey:          "ds-123",
		GLMAPIKey:               "glm-123",
		KimiAPIKey:              "kimi-123",
		QwenAPIKey:              "qwen-123",
		MiniMaxAPIKey:           "minimax-123",
		MistralAPIKey:           "mistral-123",
		XAIAPIKey:               "xai-123",
		GoogleAPIKey:            "AIza123",
		GoogleCXKey:             "1234567890abcdef:ghijklmnopqrstuv",
		OAuthGoogleClientID:     "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com",
		OAuthGoogleClientSecret: "GOCSPX-1234567890abcdefghijklmnopqr",
		OAuthGithubClientID:     "Iv1.1234567890abcdef",
		OAuthGithubClientSecret: "1234567890abcdefghijklmnopqrstuvwxyz123456",
		TraversaalAPIKey:        "traversaal-123",
		TavilyAPIKey:            "tvly-1234567890abcdefghijklmnopqrstuvwxyz",
		FirecrawlAPIKey:         "fc-123",
		PerplexityAPIKey:        "perplexity-123",
		ProxyURL:                "http://user:password@proxy.example.com:8080",
		LangfusePublicKey:       "pk-lf-1234567890abcdefghijklmnopqrstuvwxyz",
		LangfuseSecretKey:       "sk-lf-1234567890abcdefghijklmnopqrstuvwxyz",
	}

	var secrets []string
	fields := reflect.ValueOf(*cfg)
	for i := range fields.NumField() {
		if value, ok := fields.Field(i).Interface().(string); ok && value != "" {
			secrets = append(secrets, value)
		}
	}
	transcript := strings.Join(secrets, "\n")

	patterns := cfg.GetSecretPatterns()
	regexes := make([]string, 0, len(patterns))
	found := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		assert.NotEmpty(t, pattern.Name)
		re, err := re2.Compile(pattern.Regex)
		if !assert.NoError(t, err, pattern.Name) {
			continue
		}
		found = append(found, re.FindString(transcript))
		regexes = append(regexes, pattern.Regex)
	}
	assert.ElementsMatch(t, secrets, found, "each secret must be found verbatim by exactly one pattern")

	_, err := experimental.CompileSet(regexes)
	assert.NoError(t, err, "the patterns must compile as one set")
}

// clearConfigEnv blanks every variable Config reads, so no test depends on the ambient environment.
func clearConfigEnv(t *testing.T) {
	t.Helper()

	envVars := []string{
		"DATABASE_URL", "DEBUG", "DATA_DIR", "ASK_USER", "TENANT_ID", "INSTALLATION_ID", "LICENSE_KEY",
		"DOCKER_INSIDE", "DOCKER_NET_ADMIN", "DOCKER_SOCKET", "DOCKER_NETWORK",
		"DOCKER_INSIDE_HOST", "DOCKER_INSIDE_TLS_VERIFY", "DOCKER_INSIDE_CERT_PATH",
		"DOCKER_INSIDE_POLICY_TESTS",
		"DOCKER_DEFAULT_IMAGE_FOR_TEST",
		"DOCKER_PUBLIC_IP", "DOCKER_WORK_DIR", "DOCKER_DEFAULT_IMAGE", "DOCKER_DEFAULT_IMAGE_FOR_PENTEST", "TERMINAL_TOOL_TIMEOUT",
		"SERVER_PORT", "SERVER_HOST", "SERVER_USE_SSL", "SERVER_SSL_KEY", "SERVER_SSL_CRT",
		"STATIC_URL", "STATIC_DIR", "CORS_ORIGINS", "COOKIE_SIGNING_SALT",
		"SCRAPER_PUBLIC_URL", "SCRAPER_PRIVATE_URL",
		"OPEN_AI_KEY", "OPEN_AI_SERVER_URL",
		"ANTHROPIC_API_KEY", "ANTHROPIC_SERVER_URL",
		"ANTHROPIC_FEDERATION_RULE_ID", "ANTHROPIC_ORGANIZATION_ID", "ANTHROPIC_SERVICE_ACCOUNT_ID",
		"ANTHROPIC_WORKSPACE_ID", "ANTHROPIC_IDENTITY_TOKEN_FILE", "ANTHROPIC_IDENTITY_TOKEN",
		"EMBEDDING_URL", "EMBEDDING_KEY", "EMBEDDING_MODEL",
		"EMBEDDING_STRIP_NEW_LINES", "EMBEDDING_BATCH_SIZE", "EMBEDDING_MAX_TEXT_BYTES", "EMBEDDING_PROVIDER",
		"SUMMARIZER_PRESERVE_LAST", "SUMMARIZER_USE_QA", "SUMMARIZER_SUM_MSG_HUMAN_IN_QA",
		"SUMMARIZER_LAST_SEC_BYTES", "SUMMARIZER_MAX_BP_BYTES",
		"SUMMARIZER_MAX_QA_SECTIONS", "SUMMARIZER_MAX_QA_BYTES", "SUMMARIZER_KEEP_QA_SECTIONS",
		"LLM_SERVER_URL", "LLM_SERVER_KEY", "LLM_SERVER_MODEL", "LLM_SERVER_PROVIDER",
		"LLM_SERVER_CONFIG_PATH", "LLM_SERVER_PRESERVE_REASONING",
		"OLLAMA_SERVER_URL", "OLLAMA_SERVER_API_KEY", "OLLAMA_SERVER_MODEL",
		"OLLAMA_SERVER_CONFIG_PATH", "OLLAMA_SERVER_PULL_MODELS_TIMEOUT",
		"OLLAMA_SERVER_PULL_MODELS_ENABLED", "OLLAMA_SERVER_LOAD_MODELS_ENABLED",
		"GEMINI_API_KEY", "GEMINI_SERVER_URL",
		"BEDROCK_REGION", "BEDROCK_DEFAULT_AUTH", "BEDROCK_BEARER_TOKEN",
		"BEDROCK_ACCESS_KEY_ID", "BEDROCK_SECRET_ACCESS_KEY", "BEDROCK_SESSION_TOKEN", "BEDROCK_SERVER_URL",
		"DEEPSEEK_API_KEY", "DEEPSEEK_SERVER_URL", "DEEPSEEK_PROVIDER",
		"GLM_API_KEY", "GLM_SERVER_URL", "GLM_PROVIDER",
		"KIMI_API_KEY", "KIMI_SERVER_URL", "KIMI_PROVIDER",
		"QWEN_API_KEY", "QWEN_SERVER_URL", "QWEN_PROVIDER",
		"MINIMAX_API_KEY", "MINIMAX_SERVER_URL", "MINIMAX_PROVIDER",
		"MISTRAL_API_KEY", "MISTRAL_SERVER_URL", "MISTRAL_PROVIDER",
		"XAI_API_KEY", "XAI_SERVER_URL", "XAI_PROVIDER",
		"DUCKDUCKGO_ENABLED", "DUCKDUCKGO_REGION", "DUCKDUCKGO_SAFESEARCH", "DUCKDUCKGO_TIME_RANGE",
		"SPLOITUS_ENABLED",
		"GOOGLE_API_KEY", "GOOGLE_CX_KEY", "GOOGLE_LR_KEY",
		"OAUTH_GOOGLE_CLIENT_ID", "OAUTH_GOOGLE_CLIENT_SECRET",
		"OAUTH_GITHUB_CLIENT_ID", "OAUTH_GITHUB_CLIENT_SECRET",
		"PUBLIC_URL", "TRAVERSAAL_API_KEY", "TAVILY_API_KEY",
		"PERPLEXITY_API_KEY", "PERPLEXITY_MODEL", "PERPLEXITY_CONTEXT_SIZE", "PERPLEXITY_TIMEOUT",
		"SEARXNG_URL", "SEARXNG_CATEGORIES", "SEARXNG_LANGUAGE",
		"SEARXNG_SAFESEARCH", "SEARXNG_TIME_RANGE", "SEARXNG_TIMEOUT",
		"ASSISTANT_USE_AGENTS", "ASSISTANT_SUMMARIZER_PRESERVE_LAST",
		"ASSISTANT_SUMMARIZER_LAST_SEC_BYTES", "ASSISTANT_SUMMARIZER_MAX_BP_BYTES",
		"ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS", "ASSISTANT_SUMMARIZER_MAX_QA_BYTES",
		"ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS",
		"PROXY_URL", "EXTERNAL_SSL_CA_PATH", "EXTERNAL_SSL_INSECURE", "HTTP_CLIENT_TIMEOUT",
		"OTEL_HOST", "LANGFUSE_BASE_URL", "LANGFUSE_PROJECT_ID", "LANGFUSE_PUBLIC_KEY", "LANGFUSE_SECRET_KEY",
		"GRAPHITI_ENABLED", "GRAPHITI_TIMEOUT", "GRAPHITI_URL",
		"EXECUTION_MONITOR_ENABLED", "EXECUTION_MONITOR_SAME_TOOL_LIMIT", "EXECUTION_MONITOR_TOTAL_TOOL_LIMIT",
		"MAX_GENERAL_AGENT_TOOL_CALLS", "MAX_LIMITED_AGENT_TOOL_CALLS",
		"AGENT_PLANNING_STEP_ENABLED",
	}
	for _, v := range envVars {
		t.Setenv(v, "")
	}
}

func TestConfig_NewConfig_FillsDefaultsForAnEmptyEnvironment(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())

	config, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 8080, config.ServerPort)
	assert.Equal(t, "0.0.0.0", config.ServerHost)
	assert.Equal(t, false, config.Debug)
	assert.Equal(t, "./data", config.DataDir)
	assert.Equal(t, false, config.ServerUseSSL)
	assert.Nil(t, config.StaticURL)
	assert.Equal(t, []string{"*"}, config.CorsOrigins)
	assert.Equal(t, 600, config.HTTPClientTimeout)
	assert.Equal(t, 1200, config.TerminalToolTimeout)
	assert.Equal(t, "debian:latest", config.DockerDefaultImage)
	assert.Equal(t, "vxcontrol/kali-linux", config.DockerDefaultImageForPentest)

	assert.Equal(t, "openai", config.EmbeddingProvider)
	assert.Equal(t, 512, config.EmbeddingBatchSize)
	assert.Equal(t, true, config.EmbeddingStripNewLines)

	assert.Equal(t, "https://api.openai.com/v1", config.OpenAIServerURL)
	assert.Equal(t, "https://api.anthropic.com/v1", config.AnthropicServerURL)
	assert.Equal(t, "https://generativelanguage.googleapis.com", config.GeminiServerURL)
	assert.Equal(t, "us-east-1", config.BedrockRegion)
	assert.Equal(t, "https://api.deepseek.com", config.DeepSeekServerURL)
	assert.Equal(t, "https://api.z.ai/api/paas/v4", config.GLMServerURL)
	assert.Equal(t, "https://api.moonshot.ai/v1", config.KimiServerURL)
	assert.Equal(t, "https://dashscope-us.aliyuncs.com/compatible-mode/v1", config.QwenServerURL)
	assert.Equal(t, "2024-10-21", config.LLMServerAPIVersion)
	assert.Empty(t, config.LLMServerAPIType, "a plain OpenAI-compatible endpoint is the default")
	assert.Equal(t, 600, config.OllamaServerPullModelsTimeout)
	assert.Equal(t, false, config.OllamaServerPullModelsEnabled)
	assert.Equal(t, false, config.OllamaServerLoadModelsEnabled)

	assert.Equal(t, true, config.SummarizerPreserveLast)
	assert.Equal(t, true, config.SummarizerUseQA)
	assert.Equal(t, false, config.SummarizerSumHumanInQA)
	assert.Equal(t, 51200, config.SummarizerLastSecBytes)
	assert.Equal(t, 16384, config.SummarizerMaxBPBytes)
	assert.Equal(t, 10, config.SummarizerMaxQASections)
	assert.Equal(t, 65536, config.SummarizerMaxQABytes)
	assert.Equal(t, 1, config.SummarizerKeepQASections)

	assert.Equal(t, true, config.DuckDuckGoEnabled)
	assert.Equal(t, "sonar", config.PerplexityModel, "chat/completions names its models bare, so an unset value defaults to the Sonar model")
	assert.Equal(t, "low", config.PerplexityContextSize)
	assert.Equal(t, "general", config.SearxngCategories)
	assert.Equal(t, "0", config.SearxngSafeSearch)
	assert.Equal(t, "lang_en", config.GoogleLRKey)
	assert.False(t, config.WebSearchInternalEnabled)
	assert.Equal(t, 5, config.WebSearchInternalMaxSites)
	assert.Equal(t, 10240, config.WebSearchInternalMaxSiteBytes)

	assert.Equal(t, false, config.ExecutionMonitorEnabled)
	assert.Equal(t, 5, config.ExecutionMonitorSameToolLimit)
	assert.Equal(t, 10, config.ExecutionMonitorTotalToolLimit)
	assert.Equal(t, 100, config.MaxGeneralAgentToolCalls)
	assert.Equal(t, 20, config.MaxLimitedAgentToolCalls)
	assert.Equal(t, false, config.AgentPlanningStepEnabled)
}

func TestConfig_NewConfig_ReadsTheEnvironment(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())

	t.Run("set values replace the defaults", func(t *testing.T) {
		for name, value := range map[string]string{
			"SERVER_PORT":                        "9090",
			"SERVER_HOST":                        "127.0.0.1",
			"DEBUG":                              "true",
			"STATIC_URL":                         "https://example.com/static",
			"HTTP_CLIENT_TIMEOUT":                "300",
			"TERMINAL_TOOL_TIMEOUT":              "900",
			"EXECUTION_MONITOR_ENABLED":          "true",
			"EXECUTION_MONITOR_SAME_TOOL_LIMIT":  "7",
			"EXECUTION_MONITOR_TOTAL_TOOL_LIMIT": "15",
			"MAX_GENERAL_AGENT_TOOL_CALLS":       "150",
			"MAX_LIMITED_AGENT_TOOL_CALLS":       "30",
			"AGENT_PLANNING_STEP_ENABLED":        "true",
		} {
			t.Setenv(name, value)
		}

		config, err := NewConfig()
		require.NoError(t, err)

		assert.Equal(t, 9090, config.ServerPort)
		assert.Equal(t, "127.0.0.1", config.ServerHost)
		assert.Equal(t, true, config.Debug)
		require.NotNil(t, config.StaticURL)
		assert.Equal(t, "https", config.StaticURL.Scheme)
		assert.Equal(t, "example.com", config.StaticURL.Host)
		assert.Equal(t, "/static", config.StaticURL.Path)
		assert.Equal(t, 300, config.HTTPClientTimeout)
		assert.Equal(t, 900, config.TerminalToolTimeout)
		assert.Equal(t, true, config.ExecutionMonitorEnabled)
		assert.Equal(t, 7, config.ExecutionMonitorSameToolLimit)
		assert.Equal(t, 15, config.ExecutionMonitorTotalToolLimit)
		assert.Equal(t, 150, config.MaxGeneralAgentToolCalls)
		assert.Equal(t, 30, config.MaxLimitedAgentToolCalls)
		assert.Equal(t, true, config.AgentPlanningStepEnabled)
	})

	t.Run("zero timeouts are kept rather than defaulted", func(t *testing.T) {
		t.Setenv("HTTP_CLIENT_TIMEOUT", "0")
		t.Setenv("TERMINAL_TOOL_TIMEOUT", "0")

		config, err := NewConfig()
		require.NoError(t, err)

		assert.Equal(t, 0, config.HTTPClientTimeout)
		assert.Equal(t, 0, config.TerminalToolTimeout)
	})
}

func TestConfig_EnsureInstallationID_KeepsAValidIDAndReplacesTheRest(t *testing.T) {
	validEnv, validFile := uuid.New().String(), uuid.New().String()

	tests := []struct {
		name       string
		env        string
		file       string
		missingDir bool
		want       string
		wantStored bool
	}{
		{name: "nothing yet generates and stores an id", wantStored: true},
		{name: "a missing data directory is created", missingDir: true, wantStored: true},
		{name: "a stored id is read back", file: validFile, want: validFile, wantStored: true},
		{name: "invalid stored content is replaced", file: "garbage", wantStored: true},
		{name: "a valid env value is kept", env: validEnv, want: validEnv},
		{name: "an invalid env value is replaced", env: "not-a-valid-uuid", wantStored: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataDir := t.TempDir()
			if tt.missingDir {
				dataDir = filepath.Join(dataDir, "data")
			}
			idPath := filepath.Join(dataDir, "installation_id")
			if tt.file != "" {
				require.NoError(t, os.WriteFile(idPath, []byte(tt.file), 0644))
			}

			config := &Config{DataDir: dataDir, InstallationID: tt.env}
			ensureInstallationID(config)

			require.NoError(t, uuid.Validate(config.InstallationID))
			if tt.want != "" {
				assert.Equal(t, tt.want, config.InstallationID)
			} else {
				assert.NotContains(t, []string{tt.env, tt.file}, config.InstallationID)
			}
			if tt.wantStored {
				stored, err := os.ReadFile(idPath)
				require.NoError(t, err)
				assert.Equal(t, config.InstallationID, string(stored))
			}
		})
	}
}

func TestConfig_WorkerDockerEnv_PassesInsideValuesWithoutTheInsideSegment(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want []string
	}{
		{name: "a nil config", cfg: nil},
		{name: "nothing configured", cfg: &Config{}},
		{
			name: "docker inside disabled despite its values",
			cfg:  &Config{DockerInsideHost: "tcp://dind:2376", DockerInsideTLSVerify: "1", DockerInsideCertPath: "/certs"},
		},
		{name: "docker inside enabled with nothing else", cfg: &Config{DockerInside: true}},
		{
			name: "all three set",
			cfg: &Config{
				DockerInside:          true,
				DockerInsideHost:      "tcp://dind:2376",
				DockerInsideTLSVerify: "1",
				DockerInsideCertPath:  "/certs/client",
			},
			want: []string{"DOCKER_HOST=tcp://dind:2376", "DOCKER_TLS_VERIFY=1", "DOCKER_CERT_PATH=/certs/client"},
		},
		{
			name: "empty values are omitted, not passed as blanks",
			cfg:  &Config{DockerInside: true, DockerInsideHost: "tcp://dind:2375"},
			want: []string{"DOCKER_HOST=tcp://dind:2375"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.WorkerDockerEnv())
		})
	}
}

func TestConfig_WorkerDockerCertPath_MountsOnlyWhenDockerInsideIsOn(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{name: "a nil config", cfg: nil},
		{name: "docker inside disabled despite a cert path", cfg: &Config{DockerInsideCertPath: "/certs"}},
		{name: "docker inside enabled with a cert path", cfg: &Config{DockerInside: true, DockerInsideCertPath: "/certs"}, want: "/certs"},
		{name: "docker inside enabled without a cert path", cfg: &Config{DockerInside: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.WorkerDockerCertPath())
		})
	}
}

// Autodetecting the host socket must stop once a dedicated daemon endpoint is designated for sandboxes.
func TestConfig_WorkerDockerSocket_StopsAutodetectingForAnInsideHost(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *Config
		wantSocket     string
		wantAutodetect bool
	}{
		{name: "a nil config", cfg: nil, wantAutodetect: true},
		{name: "nothing configured", cfg: &Config{}, wantAutodetect: true},
		{name: "docker inside enabled with nothing else", cfg: &Config{DockerInside: true}, wantAutodetect: true},
		{name: "an explicit socket", cfg: &Config{DockerSocket: "/var/run/docker.sock"}, wantSocket: "/var/run/docker.sock"},
		{name: "an inside host", cfg: &Config{DockerInsideHost: "tcp://dind:2376"}},
		{
			name:       "an explicit socket still wins over an inside host",
			cfg:        &Config{DockerSocket: "/run/custom.sock", DockerInsideHost: "tcp://dind:2376"},
			wantSocket: "/run/custom.sock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			socket, autodetect := tt.cfg.WorkerDockerSocket()
			assert.Equal(t, tt.wantSocket, socket)
			assert.Equal(t, tt.wantAutodetect, autodetect)
		})
	}
}

func TestConfig_EveryVariableReachesTheContainer(t *testing.T) {
	compose, err := os.ReadFile(filepath.Join("..", "..", "..", "docker-compose.yml"))
	if err != nil {
		t.Fatalf("reading the compose file: %v", err)
	}

	source, err := os.ReadFile("config.go")
	if err != nil {
		t.Fatalf("reading the config: %v", err)
	}

	tags := re2.MustCompile(`env:"([A-Z0-9_]+)"`).FindAllStringSubmatch(string(source), -1)
	if len(tags) < 100 {
		t.Fatalf("found only %d env tags; this check has lost its subject", len(tags))
	}

	for _, tag := range tags {
		if !strings.Contains(string(compose), tag[1]) {
			t.Errorf(
				"%s is read from the environment but the compose stack never passes it in, so setting it in .env does nothing",
				tag[1],
			)
		}
	}
}
