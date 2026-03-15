package config

import (
	"testing"

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
		OpenAIKey: "  sk-1234  ",
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
