package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
