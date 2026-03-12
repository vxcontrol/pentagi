package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wasilibs/go-re2"
	"github.com/wasilibs/go-re2/experimental"
)

func TestGetSecretPatterns_Empty(t *testing.T) {
	cfg := &Config{}
	patterns := cfg.GetSecretPatterns()

	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns for empty config, got %d", len(patterns))
	}
}

func TestGetSecretPatterns_WithSecrets(t *testing.T) {
	cfg := &Config{
		OpenAIKey:       "sk-proj-1234567890abcdef",
		AnthropicAPIKey: "sk-ant-api03-1234567890",
		GeminiAPIKey:    "AIzaSyC1234567890abcdefghijklmnopqrst",
		DatabaseURL:     "postgres://user:password@localhost:5432/db",
		LicenseKey:      "ABCD-EFGH-IJKL-MNOP",
	}

	patterns := cfg.GetSecretPatterns()

	if len(patterns) != 5 {
		t.Errorf("expected 5 patterns, got %d", len(patterns))
	}

	// check that all patterns have names and regexes
	for i, pattern := range patterns {
		if pattern.Name == "" {
			t.Errorf("pattern at index %d has empty name", i)
		}
		if pattern.Regex == "" {
			t.Errorf("pattern at index %d has empty regex", i)
		}
	}
}

func TestGetSecretPatterns_TrimsWhitespace(t *testing.T) {
	cfg := &Config{
		OpenAIKey:    "  sk-1234  ",
		GeminiAPIKey: "\tAIzaSyC123\n",
	}

	patterns := cfg.GetSecretPatterns()

	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(patterns))
	}
}

func TestGetSecretPatterns_SkipsEmptyStrings(t *testing.T) {
	cfg := &Config{
		OpenAIKey:       "sk-1234",
		AnthropicAPIKey: "",
		GeminiAPIKey:    "   ",
		DatabaseURL:     "\t\n",
		LicenseKey:      "ABCD-EFGH",
	}

	patterns := cfg.GetSecretPatterns()

	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns (only non-empty after trim), got %d", len(patterns))
	}
}

func TestGetSecretPatterns_PatternCompilation(t *testing.T) {
	testCases := []struct {
		name   string
		config *Config
	}{
		{
			name: "OpenAI",
			config: &Config{
				OpenAIKey: "sk-proj-1234567890abcdefghijklmnopqrstuvwxyz",
			},
		},
		{
			name: "Anthropic",
			config: &Config{
				AnthropicAPIKey: "sk-ant-api03-abcdefghijklmnopqrstuvwxyz1234567890",
			},
		},
		{
			name: "Gemini",
			config: &Config{
				GeminiAPIKey: "AIzaSyC1234567890abcdefghijklmnopqrstuvwxyz",
			},
		},
		{
			name: "DeepSeek",
			config: &Config{
				DeepSeekAPIKey: "sk-1234567890abcdefghijklmnopqrstuvwxyz",
			},
		},
		{
			name: "Kimi",
			config: &Config{
				KimiAPIKey: "sk-1234567890abcdefghijklmnopqrstuvwxyz",
			},
		},
		{
			name: "Qwen",
			config: &Config{
				QwenAPIKey: "sk-1234567890abcdefghijklmnopqrstuvwxyz",
			},
		},
		{
			name: "Tavily",
			config: &Config{
				TavilyAPIKey: "tvly-1234567890abcdefghijklmnopqrstuvwxyz",
			},
		},
		{
			name: "Google",
			config: &Config{
				GoogleAPIKey: "AIzaSyC1234567890abcdefghijklmnopqrstuvwxyz",
				GoogleCXKey:  "1234567890abcdef:ghijklmnopqrstuv",
			},
		},
		{
			name: "OAuth",
			config: &Config{
				OAuthGoogleClientID:     "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com",
				OAuthGoogleClientSecret: "GOCSPX-1234567890abcdefghijklmnopqr",
				OAuthGithubClientID:     "Iv1.1234567890abcdef",
				OAuthGithubClientSecret: "1234567890abcdefghijklmnopqrstuvwxyz123456",
			},
		},
		{
			name: "Database",
			config: &Config{
				DatabaseURL: "postgres://user:p@ssw0rd!@localhost:5432/db?sslmode=disable",
			},
		},
		{
			name: "Bedrock",
			config: &Config{
				BedrockAccessKey:    "AKIAIOSFODNN7EXAMPLE",
				BedrockSecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				BedrockBearerToken:  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.example",
				BedrockSessionToken: "FwoGZXIvYXdzEBYaDD1234567890EXAMPLE",
			},
		},
		{
			name: "Langfuse",
			config: &Config{
				LangfusePublicKey: "pk-lf-1234567890abcdefghijklmnopqrstuvwxyz",
				LangfuseSecretKey: "sk-lf-1234567890abcdefghijklmnopqrstuvwxyz",
			},
		},
		{
			name: "Proxy",
			config: &Config{
				ProxyURL: "http://user:password@proxy.example.com:8080",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patterns := tc.config.GetSecretPatterns()

			if len(patterns) == 0 {
				t.Fatal("expected at least one pattern")
			}

			regexes := make([]string, 0, len(patterns))
			for i, pattern := range patterns {
				if pattern.Name == "" {
					t.Errorf("pattern at index %d has empty name", i)
				}
				if pattern.Regex == "" {
					t.Errorf("pattern at index %d has empty regex", i)
				}

				// test individual regex compilation
				if _, err := re2.Compile(pattern.Regex); err != nil {
					t.Errorf("failed to compile regex at index %d with name '%s': %s - error: %v",
						i, pattern.Name, pattern.Regex, err)
				}

				regexes = append(regexes, pattern.Regex)
			}

			// test regex set compilation
			if _, err := experimental.CompileSet(regexes); err != nil {
				t.Errorf("failed to compile regex set: %v", err)
			}

			t.Logf("successfully compiled %d regexes for %s", len(regexes), tc.name)
		})
	}
}

func TestGetSecretPatterns_AllFields(t *testing.T) {
	cfg := &Config{
		DatabaseURL:             "postgres://user:pass@localhost:5432/db",
		LicenseKey:              "ABCD-EFGH-IJKL-MNOP",
		CookieSigningSalt:       "random-salt-string-12345",
		OpenAIKey:               "sk-proj-123",
		AnthropicAPIKey:         "sk-ant-123",
		EmbeddingKey:            "emb-123",
		LLMServerKey:            "llm-123",
		OllamaServerAPIKey:      "ollama-123",
		GeminiAPIKey:            "AIzaSyC123",
		BedrockBearerToken:      "bearer-123",
		BedrockAccessKey:        "AKIA123",
		BedrockSecretKey:        "secret-123",
		BedrockSessionToken:     "session-123",
		DeepSeekAPIKey:          "ds-123",
		GLMAPIKey:               "glm-123",
		KimiAPIKey:              "kimi-123",
		QwenAPIKey:              "qwen-123",
		GoogleAPIKey:            "AIza123",
		GoogleCXKey:             "cx-123",
		OAuthGoogleClientID:     "google-client-id",
		OAuthGoogleClientSecret: "google-client-secret",
		OAuthGithubClientID:     "github-client-id",
		OAuthGithubClientSecret: "github-client-secret",
		TraversaalAPIKey:        "traversaal-123",
		TavilyAPIKey:            "tavily-123",
		PerplexityAPIKey:        "perplexity-123",
		ProxyURL:                "http://proxy:8080",
		LangfusePublicKey:       "lf-public-123",
		LangfuseSecretKey:       "lf-secret-123",
	}

	patterns := cfg.GetSecretPatterns()

	expectedCount := 29
	if len(patterns) != expectedCount {
		t.Errorf("expected %d patterns, got %d", expectedCount, len(patterns))
	}

	// verify all patterns can be compiled
	regexes := make([]string, 0, len(patterns))
	for i, pattern := range patterns {
		if _, err := re2.Compile(pattern.Regex); err != nil {
			t.Errorf("failed to compile regex at index %d with name '%s': error: %v",
				i, pattern.Name, err)
		}
		regexes = append(regexes, pattern.Regex)
	}

	// verify regex set compilation
	if _, err := experimental.CompileSet(regexes); err != nil {
		t.Errorf("failed to compile regex set: %v", err)
	}

	t.Logf("successfully compiled %d total regexes", len(regexes))
}

func TestNewConfig_Defaults(t *testing.T) {
	// Unset env vars that would override defaults
	t.Setenv("DATABASE_URL", "")
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 8080, config.ServerPort)
	assert.Equal(t, "0.0.0.0", config.ServerHost)
	assert.Equal(t, false, config.Debug)
	assert.Equal(t, "./data", config.DataDir)
	assert.Equal(t, false, config.ServerUseSSL)
	assert.Equal(t, "openai", config.EmbeddingProvider)
	assert.Equal(t, 512, config.EmbeddingBatchSize)
	assert.Equal(t, true, config.EmbeddingStripNewLines)
	assert.Equal(t, true, config.DuckDuckGoEnabled)
	assert.Equal(t, "debian:latest", config.DockerDefaultImage)
	assert.Equal(t, "vxcontrol/kali-linux", config.DockerDefaultImageForPentest)
}

func TestNewConfig_EnvOverride(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("DEBUG", "true")
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 9090, config.ServerPort)
	assert.Equal(t, "127.0.0.1", config.ServerHost)
	assert.Equal(t, true, config.Debug)
}

func TestNewConfig_ProviderDefaults(t *testing.T) {
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "https://api.openai.com/v1", config.OpenAIServerURL)
	assert.Equal(t, "https://api.anthropic.com/v1", config.AnthropicServerURL)
	assert.Equal(t, "https://generativelanguage.googleapis.com", config.GeminiServerURL)
	assert.Equal(t, "us-east-1", config.BedrockRegion)
	assert.Equal(t, "https://api.deepseek.com", config.DeepSeekServerURL)
	assert.Equal(t, "https://api.z.ai/api/paas/v4", config.GLMServerURL)
	assert.Equal(t, "https://api.moonshot.ai/v1", config.KimiServerURL)
	assert.Equal(t, "https://dashscope-us.aliyuncs.com/compatible-mode/v1", config.QwenServerURL)
}

func TestNewConfig_StaticURL(t *testing.T) {
	t.Setenv("STATIC_URL", "https://example.com/static")
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, config.StaticURL)

	assert.Equal(t, "https", config.StaticURL.Scheme)
	assert.Equal(t, "example.com", config.StaticURL.Host)
	assert.Equal(t, "/static", config.StaticURL.Path)
}

func TestNewConfig_StaticURL_Empty(t *testing.T) {
	t.Setenv("STATIC_URL", "")
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)
	assert.Nil(t, config.StaticURL)
}

func TestNewConfig_SummarizerDefaults(t *testing.T) {
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, true, config.SummarizerPreserveLast)
	assert.Equal(t, true, config.SummarizerUseQA)
	assert.Equal(t, false, config.SummarizerSumHumanInQA)
	assert.Equal(t, 51200, config.SummarizerLastSecBytes)
	assert.Equal(t, 16384, config.SummarizerMaxBPBytes)
	assert.Equal(t, 10, config.SummarizerMaxQASections)
	assert.Equal(t, 65536, config.SummarizerMaxQABytes)
	assert.Equal(t, 1, config.SummarizerKeepQASections)
}

func TestNewConfig_SearchEngineDefaults(t *testing.T) {
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "sonar", config.PerplexityModel)
	assert.Equal(t, "low", config.PerplexityContextSize)
	assert.Equal(t, "general", config.SearxngCategories)
	assert.Equal(t, "0", config.SearxngSafeSearch)
	assert.Equal(t, "lang_en", config.GoogleLRKey)
}

func TestEnsureInstallationID_GeneratesNewUUID(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		DataDir: tmpDir,
	}

	ensureInstallationID(config)

	assert.NotEmpty(t, config.InstallationID)
	assert.NoError(t, uuid.Validate(config.InstallationID))

	// verify file was written
	data, err := os.ReadFile(filepath.Join(tmpDir, "installation_id"))
	require.NoError(t, err)
	assert.Equal(t, config.InstallationID, string(data))
}

func TestEnsureInstallationID_ReadsExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	existingID := uuid.New().String()
	err := os.WriteFile(filepath.Join(tmpDir, "installation_id"), []byte(existingID), 0644)
	require.NoError(t, err)

	config := &Config{
		DataDir: tmpDir,
	}

	ensureInstallationID(config)

	assert.Equal(t, existingID, config.InstallationID)
}

func TestEnsureInstallationID_KeepsValidEnvValue(t *testing.T) {
	envID := uuid.New().String()
	config := &Config{
		InstallationID: envID,
		DataDir:        t.TempDir(),
	}

	ensureInstallationID(config)

	assert.Equal(t, envID, config.InstallationID)
}

func TestEnsureInstallationID_ReplacesInvalidEnvValue(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		InstallationID: "not-a-valid-uuid",
		DataDir:        tmpDir,
	}

	ensureInstallationID(config)

	assert.NotEqual(t, "not-a-valid-uuid", config.InstallationID)
	assert.NoError(t, uuid.Validate(config.InstallationID))
}

func TestEnsureInstallationID_ReplacesInvalidFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "installation_id"), []byte("garbage"), 0644)
	require.NoError(t, err)

	config := &Config{
		DataDir: tmpDir,
	}

	ensureInstallationID(config)

	assert.NotEqual(t, "garbage", config.InstallationID)
	assert.NoError(t, uuid.Validate(config.InstallationID))
}

func TestNewConfig_CorsOrigins(t *testing.T) {
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, []string{"*"}, config.CorsOrigins)
}

func TestNewConfig_OllamaDefaults(t *testing.T) {
	t.Setenv("LICENSE_KEY", "")

	config, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, 600, config.OllamaServerPullModelsTimeout)
	assert.Equal(t, false, config.OllamaServerPullModelsEnabled)
	assert.Equal(t, false, config.OllamaServerLoadModelsEnabled)
}
