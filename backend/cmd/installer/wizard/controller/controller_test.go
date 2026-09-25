package controller

import (
	"os"
	"path/filepath"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func controllerOverExampleEnv(t *testing.T) (*controller, string) {
	t.Helper()

	example, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", ".env.example"))
	require.NoError(t, err)

	envPath := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(envPath, example, 0o600))

	return openController(t, envPath), envPath
}

func openController(t *testing.T, envPath string) *controller {
	t.Helper()

	st, err := state.NewState(envPath)
	require.NoError(t, err)

	return NewController(st, files.NewFiles(), checker.CheckResult{}).(*controller)
}

func TestController_GetEmbedderConfig_CountsAsConfiguredOnlyWithAKeyOrServerItCanUse(t *testing.T) {
	openai := func(url, key string) map[string]string {
		return map[string]string{"EMBEDDING_PROVIDER": "openai", "EMBEDDING_URL": url, "EMBEDDING_KEY": key,
			"OPEN_AI_KEY": "chat-key", "OPEN_AI_SERVER_URL": ""}
	}
	mistral := func(url, key, chatKey string) map[string]string {
		return map[string]string{"EMBEDDING_PROVIDER": "mistral", "EMBEDDING_URL": url, "EMBEDDING_KEY": key,
			"MISTRAL_API_KEY": chatKey, "MISTRAL_SERVER_URL": "https://gateway.example/v1"}
	}
	ollama := func(url, chatURL string) map[string]string {
		return map[string]string{"EMBEDDING_PROVIDER": "ollama", "EMBEDDING_URL": url, "OLLAMA_SERVER_URL": chatURL}
	}

	for _, tc := range []struct {
		label        string
		vars         map[string]string
		isConfigured bool
	}{
		{"openai on the chat door's key and server", openai("", ""), true},
		{"openai on the chat door's server named again", openai("https://api.openai.com/v1/", ""), true},
		{"openai on a third-party endpoint without its own key", openai("https://embedder.example/v1", ""), false},
		{"openai on a third-party endpoint with its own key", openai("https://embedder.example/v1", "embedder-key"), true},
		{"mistral on the chat door's key and server", mistral("", "", "chat-key"), true},
		{"mistral on the chat door's server named again", mistral("https://gateway.example/v1/", "", "chat-key"), true},
		{"mistral on a third-party endpoint without its own key", mistral("https://embedder.example/v1", "", "chat-key"), false},
		{"mistral on a third-party endpoint with its own key", mistral("https://embedder.example/v1", "embedder-key", "chat-key"), true},
		{"mistral with its own key for the vendor", mistral("", "embedder-key", ""), true},
		{"mistral with no key at all", mistral("", "", ""), false},
		{"ollama on its own endpoint", ollama("http://embedder:11434", ""), true},
		{"ollama on the chat door's server", ollama("", "http://ollama:11434"), true},
		{"ollama with no server at all", ollama("", ""), false},
	} {
		t.Run(tc.label, func(t *testing.T) {
			c, _ := controllerOverExampleEnv(t)
			for name, value := range tc.vars {
				require.NoError(t, c.SetVar(name, value))
			}

			assert.Equal(t, tc.isConfigured, c.GetEmbedderConfig().Configured)
		})
	}
}

func TestController_GetLLMProviders_ConfiguresEveryProductDoor(t *testing.T) {
	c, _ := controllerOverExampleEnv(t)
	configured := c.GetLLMProviders()

	for _, door := range provider.AllProviderTypes {
		cfg, ok := configured[string(door)]
		if assert.True(t, ok, "the product has a %s provider the wizard cannot configure", door) {
			assert.NotEqual(t, "Unknown", cfg.Name, "the wizard reads no variables for %s", door)
		}
	}
	assert.Len(t, configured, len(provider.AllProviderTypes))
}

func TestController_GetLLMProviderConfig_CountsAnthropicConfiguredByAKeyOrACompleteFederation(t *testing.T) {
	federated := func(unset ...string) map[string]string {
		vars := map[string]string{
			"ANTHROPIC_API_KEY":             "",
			"ANTHROPIC_FEDERATION_RULE_ID":  "fdrl_01",
			"ANTHROPIC_ORGANIZATION_ID":     "00000000-0000-0000-0000-000000000000",
			"ANTHROPIC_SERVICE_ACCOUNT_ID":  "svac_01",
			"ANTHROPIC_IDENTITY_TOKEN_FILE": "/var/run/secrets/anthropic.com/token",
			"ANTHROPIC_IDENTITY_TOKEN":      "eyJhbGciOiJSUzI1NiJ9.e30.sig",
		}
		for _, name := range unset {
			vars[name] = ""
		}
		return vars
	}

	for _, tc := range []struct {
		label        string
		vars         map[string]string
		isConfigured bool
	}{
		{"an API key alone", map[string]string{"ANTHROPIC_API_KEY": "sk-ant-api03-k"}, true},
		{"every federation ID and a token file", federated("ANTHROPIC_IDENTITY_TOKEN"), true},
		{"every federation ID and a literal token", federated("ANTHROPIC_IDENTITY_TOKEN_FILE"), true},
		{"federation without a rule", federated("ANTHROPIC_FEDERATION_RULE_ID"), false},
		{"federation without an organization", federated("ANTHROPIC_ORGANIZATION_ID"), false},
		{"federation without a service account", federated("ANTHROPIC_SERVICE_ACCOUNT_ID"), false},
		{"federation IDs without an identity token", federated("ANTHROPIC_IDENTITY_TOKEN_FILE", "ANTHROPIC_IDENTITY_TOKEN"), false},
		{"neither a key nor federation", federated("ANTHROPIC_FEDERATION_RULE_ID", "ANTHROPIC_ORGANIZATION_ID",
			"ANTHROPIC_SERVICE_ACCOUNT_ID", "ANTHROPIC_IDENTITY_TOKEN_FILE", "ANTHROPIC_IDENTITY_TOKEN"), false},
	} {
		t.Run(tc.label, func(t *testing.T) {
			c, _ := controllerOverExampleEnv(t)
			for name, value := range tc.vars {
				require.NoError(t, c.SetVar(name, value))
			}

			assert.Equal(t, tc.isConfigured, c.GetLLMProviderConfig("anthropic").Configured)
		})
	}
}

func TestController_UpdateLLMProviderConfig_PersistsEachCompatDoorAndResetRestoresIt(t *testing.T) {
	for _, door := range openAICompatDoors {
		t.Run("the "+door.name+" door", func(t *testing.T) {
			c, envPath := controllerOverExampleEnv(t)
			require.Contains(t, c.GetLLMProviders(), door.id)

			cfg := c.GetLLMProviderConfig(door.id)
			require.Equal(t, door.name, cfg.Name)
			require.False(t, cfg.Configured)
			require.NotEmpty(t, cfg.BaseURL.Value, "the example env ships no default URL for %s", door.id)

			cfg.APIKey.Value = "key-" + door.id
			cfg.BaseURL.Value = "https://" + door.id + ".example/v1"
			cfg.ProviderName.Value = door.id
			require.NoError(t, c.UpdateLLMProviderConfig(door.id, cfg))
			require.NoError(t, c.Commit())

			saved := openController(t, envPath).GetLLMProviderConfig(door.id)
			assert.True(t, saved.Configured)
			assert.Equal(t, door.prefix+"_API_KEY", saved.APIKey.Name)
			assert.Equal(t, "key-"+door.id, saved.APIKey.Value)
			assert.Equal(t, "https://"+door.id+".example/v1", saved.BaseURL.Value)
			assert.Equal(t, door.id, saved.ProviderName.Value)

			c2 := openController(t, envPath)
			pending := c2.GetLLMProviderConfig(door.id)
			pending.APIKey.Value, pending.BaseURL.Value, pending.ProviderName.Value = "unsaved", "https://unsaved", "unsaved"
			require.NoError(t, c2.UpdateLLMProviderConfig(door.id, pending))
			c2.ResetLLMProviderConfig(door.id)

			discarded := c2.GetLLMProviderConfig(door.id)
			assert.Equal(t, "key-"+door.id, discarded.APIKey.Value)
			assert.Equal(t, "https://"+door.id+".example/v1", discarded.BaseURL.Value)
			assert.Equal(t, door.id, discarded.ProviderName.Value)
		})
	}
}

func TestController_UpdateLLMProviderConfig_PersistsTheCustomAzureFieldsAndResetRestoresThem(t *testing.T) {
	c, envPath := controllerOverExampleEnv(t)

	cfg := c.GetLLMProviderConfig("custom")
	cfg.BaseURL.Value = "https://example.openai.azure.com"
	cfg.APIKey.Value = "k"
	cfg.Model.Value = "gpt-5.6-terra"
	cfg.APIType.Value = "azure"
	cfg.APIVersion.Value = "2024-10-21"
	require.NoError(t, c.UpdateLLMProviderConfig("custom", cfg))
	require.NoError(t, c.Commit())

	saved := openController(t, envPath).GetLLMProviderConfig("custom")
	assert.Equal(t, "azure", saved.APIType.Value)
	assert.Equal(t, "2024-10-21", saved.APIVersion.Value)

	c2 := openController(t, envPath)
	pending := c2.GetLLMProviderConfig("custom")
	pending.APIType.Value, pending.APIVersion.Value = "azure_ad", "2025-01-01-preview"
	require.NoError(t, c2.UpdateLLMProviderConfig("custom", pending))
	c2.ResetLLMProviderConfig("custom")

	discarded := c2.GetLLMProviderConfig("custom")
	assert.Equal(t, "azure", discarded.APIType.Value)
	assert.Equal(t, "2024-10-21", discarded.APIVersion.Value)
}

func TestController_CompatDoorVariablesAreDescribedMaskedAndCritical(t *testing.T) {
	c, _ := controllerOverExampleEnv(t)

	for _, door := range openAICompatDoors {
		for _, name := range []string{door.prefix + "_API_KEY", door.prefix + "_SERVER_URL", door.prefix + "_PROVIDER"} {
			assert.NotEqual(t, name, c.getVariableDescription(name), "%s has no description", name)
			assert.True(t, c.isCriticalVariable(name), "changing %s does not restart the service", name)
		}
		assert.True(t, c.isVariableMasked(door.prefix+"_API_KEY"), "%s_API_KEY is shown in clear", door.prefix)
	}
}
