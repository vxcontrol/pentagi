package controller

import (
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/locale"
)

const (
	EmbeddedLLMConfigsPath   = "providers-configs"
	DefaultDockerCertPath    = "/opt/pentagi/docker/ssl"
	DefaultCustomConfigsPath = "/opt/pentagi/conf/custom.provider.yml"
	DefaultOllamaConfigsPath = "/opt/pentagi/conf/ollama.provider.yml"
	DefaultLLMConfigsPath    = "/opt/pentagi/conf/"
	DefaultScraperBaseURL    = "https://scraper/"
	DefaultScraperDomain     = "scraper"
	DefaultScraperSchema     = "https"
)

type Controller interface {
	GetState() state.State
	GetChecker() *checker.CheckResult

	state.State

	LLMProviderConfigController
	LangfuseConfigController
	GraphitiConfigController
	ObservabilityConfigController
	SummarizerConfigController
	EmbedderConfigController
	AIAgentsConfigController
	ScraperConfigController
	SearchEnginesConfigController
	DockerConfigController
	ChangesConfigController
	ServerSettingsConfigController
}

type LLMProviderConfigController interface {
	GetLLMProviders() map[string]*LLMProviderConfig
	GetLLMProviderConfig(providerID string) *LLMProviderConfig
	UpdateLLMProviderConfig(providerID string, config *LLMProviderConfig) error
	ResetLLMProviderConfig(providerID string) map[string]*LLMProviderConfig
}

type LangfuseConfigController interface {
	GetLangfuseConfig() *LangfuseConfig
	UpdateLangfuseConfig(config *LangfuseConfig) error
	ResetLangfuseConfig() *LangfuseConfig
}

type GraphitiConfigController interface {
	GetGraphitiConfig() *GraphitiConfig
	UpdateGraphitiConfig(config *GraphitiConfig) error
	ResetGraphitiConfig() *GraphitiConfig
}

type ObservabilityConfigController interface {
	GetObservabilityConfig() *ObservabilityConfig
	UpdateObservabilityConfig(config *ObservabilityConfig) error
	ResetObservabilityConfig() *ObservabilityConfig
}

type SummarizerConfigController interface {
	GetSummarizerConfig(summarizerType SummarizerType) *SummarizerConfig
	UpdateSummarizerConfig(config *SummarizerConfig) error
	ResetSummarizerConfig(summarizerType SummarizerType) *SummarizerConfig
}

type EmbedderConfigController interface {
	GetEmbedderConfig() *EmbedderConfig
	UpdateEmbedderConfig(config *EmbedderConfig) error
	ResetEmbedderConfig() *EmbedderConfig
}

type AIAgentsConfigController interface {
	GetAIAgentsConfig() *AIAgentsConfig
	UpdateAIAgentsConfig(config *AIAgentsConfig) error
	ResetAIAgentsConfig() *AIAgentsConfig
}

type ScraperConfigController interface {
	GetScraperConfig() *ScraperConfig
	UpdateScraperConfig(config *ScraperConfig) error
	ResetScraperConfig() *ScraperConfig
}

type SearchEnginesConfigController interface {
	GetSearchEnginesConfig() *SearchEnginesConfig
	UpdateSearchEnginesConfig(config *SearchEnginesConfig) error
	ResetSearchEnginesConfig() *SearchEnginesConfig
}

type DockerConfigController interface {
	GetDockerConfig() *DockerConfig
	UpdateDockerConfig(config *DockerConfig) error
	ResetDockerConfig() *DockerConfig
}

type ChangesConfigController interface {
	GetApplyChangesConfig() *ApplyChangesConfig
}

type ServerSettingsConfigController interface {
	GetServerSettingsConfig() *ServerSettingsConfig
	UpdateServerSettingsConfig(config *ServerSettingsConfig) error
	ResetServerSettingsConfig() *ServerSettingsConfig
}

// controller bridges TUI models with the state package
type controller struct {
	files   files.Files
	checker checker.CheckResult
	state.State
}

func NewController(state state.State, files files.Files, checker checker.CheckResult) Controller {
	return &controller{
		files:   files,
		checker: checker,
		State:   state,
	}
}

// GetState returns the underlying state interface for processor integration
func (c *controller) GetState() state.State {
	return c.State
}

// GetChecker returns the checker result for processor integration
func (c *controller) GetChecker() *checker.CheckResult {
	return &c.checker
}

// LLMProviderConfig represents LLM provider configuration
type LLMProviderConfig struct {
	// dependent on the provider type
	Name string

	// direct form field mappings using loader.EnvVar
	// these fields directly correspond to environment variables and form inputs (not computed)
	BaseURL loader.EnvVar // OPEN_AI_SERVER_URL | ANTHROPIC_SERVER_URL | GEMINI_SERVER_URL | BEDROCK_SERVER_URL | OLLAMA_SERVER_URL | LLM_SERVER_URL
	APIKey  loader.EnvVar // OPEN_AI_KEY | ANTHROPIC_API_KEY | GEMINI_API_KEY | BEDROCK_ACCESS_KEY_ID | OLLAMA_SERVER_CONFIG_PATH | LLM_SERVER_KEY
	Model   loader.EnvVar // LLM_SERVER_MODEL
	// AWS Bedrock specific fields
	AccessKey    loader.EnvVar // BEDROCK_ACCESS_KEY_ID
	SecretKey    loader.EnvVar // BEDROCK_SECRET_ACCESS_KEY
	SessionToken loader.EnvVar // BEDROCK_SESSION_TOKEN
	Region       loader.EnvVar // BEDROCK_REGION
	// Ollama and Custom specific fields
	ConfigPath        loader.EnvVar // OLLAMA_SERVER_CONFIG_PATH | LLM_SERVER_CONFIG_PATH
	HostConfigPath    loader.EnvVar // PENTAGI_OLLAMA_SERVER_CONFIG_PATH | PENTAGI_LLM_SERVER_CONFIG_PATH
	LegacyReasoning   loader.EnvVar // LLM_SERVER_LEGACY_REASONING
	PreserveReasoning loader.EnvVar // LLM_SERVER_PRESERVE_REASONING
	// Custom specific fields
	ProviderName loader.EnvVar // LLM_SERVER_PROVIDER
	// Ollama specific fields
	PullTimeout       loader.EnvVar // OLLAMA_SERVER_PULL_MODELS_TIMEOUT
	PullEnabled       loader.EnvVar // OLLAMA_SERVER_PULL_MODELS_ENABLED
	LoadModelsEnabled loader.EnvVar // OLLAMA_SERVER_LOAD_MODELS_ENABLED

	// computed fields (not directly mapped to env vars)
	Configured bool

	// local path to the embedded LLM config files inside the container
	EmbeddedLLMConfigsPath []string
}

func GetEmbeddedLLMConfigsPath(files files.Files) []string {
	providersConfigsPath := make([]string, 0)
	if confFiles, err := files.List(EmbeddedLLMConfigsPath); err == nil {
		for _, confFile := range confFiles {
			confPath := DefaultLLMConfigsPath + strings.TrimPrefix(confFile, EmbeddedLLMConfigsPath+"/")
			providersConfigsPath = append(providersConfigsPath, confPath)
		}
		sort.Strings(providersConfigsPath)
	}

	return providersConfigsPath
}

// GetLLMProviders returns configured LLM providers
func (c *controller) GetLLMProviders() map[string]*LLMProviderConfig {
	return map[string]*LLMProviderConfig{
		"openai":    c.GetLLMProviderConfig("openai"),
		"anthropic": c.GetLLMProviderConfig("anthropic"),
		"gemini":    c.GetLLMProviderConfig("gemini"),
		"bedrock":   c.GetLLMProviderConfig("bedrock"),
		"ollama":    c.GetLLMProviderConfig("ollama"),
		"custom":    c.GetLLMProviderConfig("custom"),
	}
}

// GetLLMProviderConfig returns the current LLM provider configuration
func (c *controller) GetLLMProviderConfig(providerID string) *LLMProviderConfig {
	providersConfigsPath := GetEmbeddedLLMConfigsPath(c.files)
	providerConfig := &LLMProviderConfig{
		Name:                   "Unknown",
		EmbeddedLLMConfigsPath: providersConfigsPath,
	}

	switch providerID {
	case "openai":
		providerConfig.Name = "OpenAI"
		providerConfig.APIKey, _ = c.GetVar("OPEN_AI_KEY")
		providerConfig.BaseURL, _ = c.GetVar("OPEN_AI_SERVER_URL")
		providerConfig.Configured = providerConfig.APIKey.Value != ""

	case "anthropic":
		providerConfig.Name = "Anthropic"
		providerConfig.APIKey, _ = c.GetVar("ANTHROPIC_API_KEY")
		providerConfig.BaseURL, _ = c.GetVar("ANTHROPIC_SERVER_URL")
		providerConfig.Configured = providerConfig.APIKey.Value != ""

	case "gemini":
		providerConfig.Name = "Google Gemini"
		providerConfig.APIKey, _ = c.GetVar("GEMINI_API_KEY")
		providerConfig.BaseURL, _ = c.GetVar("GEMINI_SERVER_URL")
		providerConfig.Configured = providerConfig.APIKey.Value != ""

	case "bedrock":
		providerConfig.Name = "AWS Bedrock"
		providerConfig.Region, _ = c.GetVar("BEDROCK_REGION")
		providerConfig.AccessKey, _ = c.GetVar("BEDROCK_ACCESS_KEY_ID")
		providerConfig.SecretKey, _ = c.GetVar("BEDROCK_SECRET_ACCESS_KEY")
		providerConfig.SessionToken, _ = c.GetVar("BEDROCK_SESSION_TOKEN")
		providerConfig.BaseURL, _ = c.GetVar("BEDROCK_SERVER_URL")
		providerConfig.Configured = (providerConfig.AccessKey.Value != "" && providerConfig.SecretKey.Value != "") || providerConfig.SessionToken.Value != ""

	case "ollama":
		providerConfig.Name = "Ollama"
		providerConfig.BaseURL, _ = c.GetVar("OLLAMA_SERVER_URL")
		providerConfig.ConfigPath, _ = c.GetVar("OLLAMA_SERVER_CONFIG_PATH")
		providerConfig.HostConfigPath, _ = c.GetVar("PENTAGI_OLLAMA_SERVER_CONFIG_PATH")
		if slices.Contains(providersConfigsPath, providerConfig.ConfigPath.Value) {
			providerConfig.HostConfigPath.Value = providerConfig.ConfigPath.Value
		}
		providerConfig.Model, _ = c.GetVar("OLLAMA_SERVER_MODEL")
		providerConfig.PullTimeout, _ = c.GetVar("OLLAMA_SERVER_PULL_MODELS_TIMEOUT")
		providerConfig.PullEnabled, _ = c.GetVar("OLLAMA_SERVER_PULL_MODELS_ENABLED")
		providerConfig.LoadModelsEnabled, _ = c.GetVar("OLLAMA_SERVER_LOAD_MODELS_ENABLED")
		providerConfig.Configured = providerConfig.BaseURL.Value != ""

	case "custom":
		providerConfig.Name = "Custom"
		providerConfig.BaseURL, _ = c.GetVar("LLM_SERVER_URL")
		providerConfig.APIKey, _ = c.GetVar("LLM_SERVER_KEY")
		providerConfig.Model, _ = c.GetVar("LLM_SERVER_MODEL")
		providerConfig.ConfigPath, _ = c.GetVar("LLM_SERVER_CONFIG_PATH")
		providerConfig.HostConfigPath, _ = c.GetVar("PENTAGI_LLM_SERVER_CONFIG_PATH")
		if slices.Contains(providersConfigsPath, providerConfig.ConfigPath.Value) {
			providerConfig.HostConfigPath.Value = providerConfig.ConfigPath.Value
		}
		providerConfig.LegacyReasoning, _ = c.GetVar("LLM_SERVER_LEGACY_REASONING")
		providerConfig.PreserveReasoning, _ = c.GetVar("LLM_SERVER_PRESERVE_REASONING")
		providerConfig.ProviderName, _ = c.GetVar("LLM_SERVER_PROVIDER")
		providerConfig.Configured = providerConfig.BaseURL.Value != "" && providerConfig.APIKey.Value != "" &&
			(providerConfig.Model.Value != "" || providerConfig.ConfigPath.Value != "")
	}

	return providerConfig
}

// UpdateLLMProviderConfig updates a specific LLM provider configuration
func (c *controller) UpdateLLMProviderConfig(providerID string, config *LLMProviderConfig) error {
	switch providerID {
	case "openai":
		if err := c.SetVar(config.APIKey.Name, config.APIKey.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.APIKey.Name, err)
		}
		if err := c.SetVar(config.BaseURL.Name, config.BaseURL.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.BaseURL.Name, err)
		}

	case "anthropic":
		if err := c.SetVar(config.APIKey.Name, config.APIKey.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.APIKey.Name, err)
		}
		if err := c.SetVar(config.BaseURL.Name, config.BaseURL.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.BaseURL.Name, err)
		}

	case "gemini":
		if err := c.SetVar(config.APIKey.Name, config.APIKey.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.APIKey.Name, err)
		}
		if err := c.SetVar(config.BaseURL.Name, config.BaseURL.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.BaseURL.Name, err)
		}

	case "bedrock":
		if err := c.SetVar(config.Region.Name, config.Region.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.Region.Name, err)
		}
		if err := c.SetVar(config.AccessKey.Name, config.AccessKey.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.AccessKey.Name, err)
		}
		if err := c.SetVar(config.SecretKey.Name, config.SecretKey.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.SecretKey.Name, err)
		}
		if err := c.SetVar(config.SessionToken.Name, config.SessionToken.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.SessionToken.Name, err)
		}
		if err := c.SetVar(config.BaseURL.Name, config.BaseURL.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.BaseURL.Name, err)
		}

	case "ollama":
		if err := c.SetVar(config.BaseURL.Name, config.BaseURL.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.BaseURL.Name, err)
		}
		if err := c.SetVar(config.Model.Name, config.Model.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.Model.Name, err)
		}
		if err := c.SetVar(config.PullTimeout.Name, config.PullTimeout.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.PullTimeout.Name, err)
		}
		if err := c.SetVar(config.PullEnabled.Name, config.PullEnabled.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.PullEnabled.Name, err)
		}
		if err := c.SetVar(config.LoadModelsEnabled.Name, config.LoadModelsEnabled.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.LoadModelsEnabled.Name, err)
		}

		var containerPath, hostPath string
		if config.HostConfigPath.Value != "" {
			if slices.Contains(config.EmbeddedLLMConfigsPath, config.HostConfigPath.Value) {
				containerPath = config.HostConfigPath.Value
				hostPath = ""
			} else {
				containerPath = DefaultOllamaConfigsPath
				hostPath = config.HostConfigPath.Value
			}
		}

		if err := c.SetVar(config.ConfigPath.Name, containerPath); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.ConfigPath.Name, err)
		}
		if err := c.SetVar(config.HostConfigPath.Name, hostPath); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.HostConfigPath.Name, err)
		}

	case "custom":
		if err := c.SetVar(config.BaseURL.Name, config.BaseURL.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.BaseURL.Name, err)
		}
		if err := c.SetVar(config.APIKey.Name, config.APIKey.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.APIKey.Name, err)
		}
		if err := c.SetVar(config.Model.Name, config.Model.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.Model.Name, err)
		}
		if err := c.SetVar(config.LegacyReasoning.Name, config.LegacyReasoning.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.LegacyReasoning.Name, err)
		}
		if err := c.SetVar(config.PreserveReasoning.Name, config.PreserveReasoning.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.PreserveReasoning.Name, err)
		}
		if err := c.SetVar(config.ProviderName.Name, config.ProviderName.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.ProviderName.Name, err)
		}

		var containerPath, hostPath string
		if config.HostConfigPath.Value != "" {
			if slices.Contains(config.EmbeddedLLMConfigsPath, config.HostConfigPath.Value) {
				containerPath = config.HostConfigPath.Value
				hostPath = ""
			} else {
				containerPath = DefaultCustomConfigsPath
				hostPath = config.HostConfigPath.Value
			}
		}

		if err := c.SetVar(config.ConfigPath.Name, containerPath); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.ConfigPath.Name, err)
		}
		if err := c.SetVar(config.HostConfigPath.Name, hostPath); err != nil {
			return fmt.Errorf("failed to set %s: %w", config.HostConfigPath.Name, err)
		}
	}

	return nil
}

// ResetLLMProviderConfig resets a specific LLM provider configuration
func (c *controller) ResetLLMProviderConfig(providerID string) map[string]*LLMProviderConfig {
	var vars []string
	switch providerID {
	case "openai":
		vars = []string{"OPEN_AI_KEY", "OPEN_AI_SERVER_URL"}
	case "anthropic":
		vars = []string{"ANTHROPIC_API_KEY", "ANTHROPIC_SERVER_URL"}
	case "gemini":
		vars = []string{"GEMINI_API_KEY", "GEMINI_SERVER_URL"}
	case "bedrock":
		vars = []string{
			"BEDROCK_ACCESS_KEY_ID", "BEDROCK_SECRET_ACCESS_KEY", "BEDROCK_SESSION_TOKEN",
			"BEDROCK_REGION", "BEDROCK_SERVER_URL",
		}
	case "ollama":
		vars = []string{
			"OLLAMA_SERVER_URL",
			"OLLAMA_SERVER_MODEL",
			"OLLAMA_SERVER_CONFIG_PATH",
			"OLLAMA_SERVER_PULL_MODELS_TIMEOUT",
			"OLLAMA_SERVER_PULL_MODELS_ENABLED",
			"OLLAMA_SERVER_LOAD_MODELS_ENABLED",
			"PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
		}
	case "custom":
		vars = []string{
			"LLM_SERVER_URL", "LLM_SERVER_KEY", "LLM_SERVER_MODEL",
			"LLM_SERVER_CONFIG_PATH", "LLM_SERVER_LEGACY_REASONING",
			"LLM_SERVER_PRESERVE_REASONING", "LLM_SERVER_PROVIDER",
			"PENTAGI_LLM_SERVER_CONFIG_PATH", // local path to the LLM config file
		}
	}

	if len(vars) != 0 {
		if err := c.ResetVars(vars); err != nil {
			return nil
		}
	}

	return c.GetLLMProviders()
}

// LangfuseConfig represents Langfuse configuration
type LangfuseConfig struct {
	// deployment configuration
	DeploymentType string // "embedded" or "external" or "disabled"

	// embedded listen settings
	ListenIP   loader.EnvVar // LANGFUSE_LISTEN_IP
	ListenPort loader.EnvVar // LANGFUSE_LISTEN_PORT

	// integration settings (always required)
	BaseURL   loader.EnvVar // LANGFUSE_BASE_URL
	ProjectID loader.EnvVar // LANGFUSE_PROJECT_ID | LANGFUSE_INIT_PROJECT_ID
	PublicKey loader.EnvVar // LANGFUSE_PUBLIC_KEY | LANGFUSE_INIT_PROJECT_PUBLIC_KEY
	SecretKey loader.EnvVar // LANGFUSE_SECRET_KEY | LANGFUSE_INIT_PROJECT_SECRET_KEY

	// embedded instance settings (only for embedded mode)
	AdminEmail    loader.EnvVar // LANGFUSE_INIT_USER_EMAIL
	AdminPassword loader.EnvVar // LANGFUSE_INIT_USER_PASSWORD
	AdminName     loader.EnvVar // LANGFUSE_INIT_USER_NAME

	// enterprise license (optional for embedded mode)
	LicenseKey loader.EnvVar // LANGFUSE_EE_LICENSE_KEY

	// computed fields (not directly mapped to env vars)
	Installed bool
}

// GetLangfuseConfig returns the current Langfuse configuration
func (c *controller) GetLangfuseConfig() *LangfuseConfig {
	vars, _ := c.GetVars([]string{
		"LANGFUSE_LISTEN_IP",
		"LANGFUSE_LISTEN_PORT",
		"LANGFUSE_BASE_URL",
		"LANGFUSE_PROJECT_ID",
		"LANGFUSE_PUBLIC_KEY",
		"LANGFUSE_SECRET_KEY",
		"LANGFUSE_INIT_USER_EMAIL",
		"LANGFUSE_INIT_USER_PASSWORD",
		"LANGFUSE_INIT_USER_NAME",
		"LANGFUSE_EE_LICENSE_KEY",
	})
	// defaults
	if v := vars["LANGFUSE_LISTEN_IP"]; v.Default == "" {
		v.Default = "127.0.0.1"
		vars["LANGFUSE_LISTEN_IP"] = v
	}
	if v := vars["LANGFUSE_LISTEN_PORT"]; v.Default == "" {
		v.Default = "4000"
		vars["LANGFUSE_LISTEN_PORT"] = v
	}

	// Determine deployment type based on endpoint value
	var deploymentType string
	baseURL := vars["LANGFUSE_BASE_URL"]
	projectID := vars["LANGFUSE_PROJECT_ID"]
	publicKey := vars["LANGFUSE_PUBLIC_KEY"]
	secretKey := vars["LANGFUSE_SECRET_KEY"]
	adminEmail := vars["LANGFUSE_INIT_USER_EMAIL"]
	adminPassword := vars["LANGFUSE_INIT_USER_PASSWORD"]
	adminName := vars["LANGFUSE_INIT_USER_NAME"]
	licenseKey := vars["LANGFUSE_EE_LICENSE_KEY"]

	switch baseURL.Value {
	case "":
		deploymentType = "disabled"
	case checker.DefaultLangfuseEndpoint:
		deploymentType = "embedded"
		if projectID.Value == "" && !projectID.IsChanged {
			if initProjectID, ok := c.GetVar("LANGFUSE_INIT_PROJECT_ID"); ok {
				projectID.Value = initProjectID.Value
				projectID.IsChanged = true
			}
		}
		if publicKey.Value == "" && !publicKey.IsChanged {
			if initPublicKey, ok := c.GetVar("LANGFUSE_INIT_PROJECT_PUBLIC_KEY"); ok {
				publicKey.Value = initPublicKey.Value
				publicKey.IsChanged = true
			}
		}
		if secretKey.Value == "" && !secretKey.IsChanged {
			if initSecretKey, ok := c.GetVar("LANGFUSE_INIT_PROJECT_SECRET_KEY"); ok {
				secretKey.Value = initSecretKey.Value
				secretKey.IsChanged = true
			}
		}
	default:
		deploymentType = "external"
	}

	return &LangfuseConfig{
		DeploymentType: deploymentType,
		ListenIP:       vars["LANGFUSE_LISTEN_IP"],
		ListenPort:     vars["LANGFUSE_LISTEN_PORT"],
		BaseURL:        baseURL,
		ProjectID:      projectID,
		PublicKey:      publicKey,
		SecretKey:      secretKey,
		AdminEmail:     adminEmail,
		AdminPassword:  adminPassword,
		AdminName:      adminName,
		Installed:      c.checker.LangfuseInstalled,
		LicenseKey:     licenseKey,
	}
}

// UpdateLangfuseConfig updates Langfuse configuration with proper endpoint handling
func (c *controller) UpdateLangfuseConfig(config *LangfuseConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Set deployment type based configuration
	switch config.DeploymentType {
	case "embedded":
		// for embedded mode, use default endpoint and sync with docker-compose settings
		config.BaseURL.Value = checker.DefaultLangfuseEndpoint

		if err := c.SetVar("LANGFUSE_LISTEN_IP", config.ListenIP.Value); err != nil {
			return fmt.Errorf("failed to set LANGFUSE_LISTEN_IP: %w", err)
		}
		if err := c.SetVar("LANGFUSE_LISTEN_PORT", config.ListenPort.Value); err != nil {
			return fmt.Errorf("failed to set LANGFUSE_LISTEN_PORT: %w", err)
		}

		// update enterprise license key if provided
		if err := c.SetVar("LANGFUSE_EE_LICENSE_KEY", config.LicenseKey.Value); err != nil {
			return fmt.Errorf("failed to set LANGFUSE_EE_LICENSE_KEY: %w", err)
		}

		// Sync with docker-compose environment variables
		if !config.Installed {
			if err := c.SetVar("LANGFUSE_INIT_PROJECT_ID", config.ProjectID.Value); err != nil {
				return fmt.Errorf("failed to set LANGFUSE_INIT_PROJECT_ID: %w", err)
			}
			if err := c.SetVar("LANGFUSE_INIT_PROJECT_PUBLIC_KEY", config.PublicKey.Value); err != nil {
				return fmt.Errorf("failed to set LANGFUSE_INIT_PROJECT_PUBLIC_KEY: %w", err)
			}
			if err := c.SetVar("LANGFUSE_INIT_PROJECT_SECRET_KEY", config.SecretKey.Value); err != nil {
				return fmt.Errorf("failed to set LANGFUSE_INIT_PROJECT_SECRET_KEY: %w", err)
			}
			if err := c.SetVar("LANGFUSE_INIT_USER_EMAIL", config.AdminEmail.Value); err != nil {
				return fmt.Errorf("failed to set LANGFUSE_INIT_USER_EMAIL: %w", err)
			}
			if err := c.SetVar("LANGFUSE_INIT_USER_NAME", config.AdminName.Value); err != nil {
				return fmt.Errorf("failed to set LANGFUSE_INIT_USER_NAME: %w", err)
			}
			if err := c.SetVar("LANGFUSE_INIT_USER_PASSWORD", config.AdminPassword.Value); err != nil {
				return fmt.Errorf("failed to set LANGFUSE_INIT_USER_PASSWORD: %w", err)
			}
		}

	case "external":
		// for external mode, use provided endpoint

	case "disabled":
		// for disabled mode, clear endpoint and disable
		config.BaseURL.Value = ""
	}

	// update integration environment variables
	if err := c.SetVar("LANGFUSE_BASE_URL", config.BaseURL.Value); err != nil {
		return fmt.Errorf("failed to set LANGFUSE_BASE_URL: %w", err)
	}
	if err := c.SetVar("LANGFUSE_PROJECT_ID", config.ProjectID.Value); err != nil {
		return fmt.Errorf("failed to set LANGFUSE_PROJECT_ID: %w", err)
	}
	if err := c.SetVar("LANGFUSE_PUBLIC_KEY", config.PublicKey.Value); err != nil {
		return fmt.Errorf("failed to set LANGFUSE_PUBLIC_KEY: %w", err)
	}
	if err := c.SetVar("LANGFUSE_SECRET_KEY", config.SecretKey.Value); err != nil {
		return fmt.Errorf("failed to set LANGFUSE_SECRET_KEY: %w", err)
	}

	return nil
}

func (c *controller) ResetLangfuseConfig() *LangfuseConfig {
	vars := []string{
		"LANGFUSE_BASE_URL",
		"LANGFUSE_PROJECT_ID",
		"LANGFUSE_PUBLIC_KEY",
		"LANGFUSE_SECRET_KEY",
		"LANGFUSE_LISTEN_IP",
		"LANGFUSE_LISTEN_PORT",
		"LANGFUSE_EE_LICENSE_KEY",
	}

	if !c.checker.LangfuseInstalled {
		vars = append(vars,
			"LANGFUSE_INIT_USER_EMAIL",
			"LANGFUSE_INIT_USER_NAME",
			"LANGFUSE_INIT_USER_PASSWORD",
			"LANGFUSE_INIT_PROJECT_ID",
			"LANGFUSE_INIT_PROJECT_PUBLIC_KEY",
			"LANGFUSE_INIT_PROJECT_SECRET_KEY",
		)
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetLangfuseConfig()
}

// GraphitiConfig represents Graphiti knowledge graph configuration
type GraphitiConfig struct {
	// deployment configuration
	DeploymentType string // "embedded" or "external" or "disabled"

	// integration settings (always)
	GraphitiURL loader.EnvVar // GRAPHITI_URL
	Timeout     loader.EnvVar // GRAPHITI_TIMEOUT
	ModelName   loader.EnvVar // GRAPHITI_MODEL_NAME

	// neo4j settings (embedded only)
	Neo4jUser     loader.EnvVar // NEO4J_USER
	Neo4jPassword loader.EnvVar // NEO4J_PASSWORD
	Neo4jDatabase loader.EnvVar // NEO4J_DATABASE
	Neo4jURI      loader.EnvVar // NEO4J_URI

	// computed fields (not directly mapped to env vars)
	Installed bool
}

// GetGraphitiConfig returns the current Graphiti configuration
func (c *controller) GetGraphitiConfig() *GraphitiConfig {
	vars, _ := c.GetVars([]string{
		"GRAPHITI_URL",
		"GRAPHITI_TIMEOUT",
		"GRAPHITI_MODEL_NAME",
		"NEO4J_USER",
		"NEO4J_PASSWORD",
		"NEO4J_DATABASE",
		"NEO4J_URI",
	})

	// set defaults if missing
	if v := vars["GRAPHITI_TIMEOUT"]; v.Default == "" {
		v.Default = "30"
		vars["GRAPHITI_TIMEOUT"] = v
	}
	if v := vars["GRAPHITI_MODEL_NAME"]; v.Default == "" {
		v.Default = "gpt-5-mini"
		vars["GRAPHITI_MODEL_NAME"] = v
	}
	if v := vars["NEO4J_USER"]; v.Default == "" {
		v.Default = "neo4j"
		vars["NEO4J_USER"] = v
	}
	if v := vars["NEO4J_PASSWORD"]; v.Default == "" {
		v.Default = "devpassword"
		vars["NEO4J_PASSWORD"] = v
	}
	if v := vars["NEO4J_DATABASE"]; v.Default == "" {
		v.Default = "neo4j"
		vars["NEO4J_DATABASE"] = v
	}
	if v := vars["NEO4J_URI"]; v.Default == "" {
		v.Default = "bolt://neo4j:7687"
		vars["NEO4J_URI"] = v
	}

	graphitiURL := vars["GRAPHITI_URL"]

	// determine deployment type based on GRAPHITI_ENABLED and GRAPHITI_URL
	graphitiEnabled, _ := c.GetVar("GRAPHITI_ENABLED")

	var deploymentType string
	if graphitiEnabled.Value != "true" || graphitiURL.Value == "" {
		deploymentType = "disabled"
	} else if graphitiURL.Value == checker.DefaultGraphitiEndpoint {
		deploymentType = "embedded"
	} else {
		deploymentType = "external"
	}

	return &GraphitiConfig{
		DeploymentType: deploymentType,
		GraphitiURL:    graphitiURL,
		Timeout:        vars["GRAPHITI_TIMEOUT"],
		ModelName:      vars["GRAPHITI_MODEL_NAME"],
		Neo4jUser:      vars["NEO4J_USER"],
		Neo4jPassword:  vars["NEO4J_PASSWORD"],
		Neo4jDatabase:  vars["NEO4J_DATABASE"],
		Neo4jURI:       vars["NEO4J_URI"],
		Installed:      c.checker.GraphitiInstalled,
	}
}

// UpdateGraphitiConfig updates Graphiti configuration
func (c *controller) UpdateGraphitiConfig(config *GraphitiConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// set deployment type based configuration
	switch config.DeploymentType {
	case "embedded":
		// for embedded mode, use default endpoint
		config.GraphitiURL.Value = checker.DefaultGraphitiEndpoint

		// enable Graphiti
		if err := c.SetVar("GRAPHITI_ENABLED", "true"); err != nil {
			return fmt.Errorf("failed to set GRAPHITI_ENABLED: %w", err)
		}

		// update timeout, model, and neo4j settings
		if err := c.SetVar("GRAPHITI_TIMEOUT", config.Timeout.Value); err != nil {
			return fmt.Errorf("failed to set GRAPHITI_TIMEOUT: %w", err)
		}
		if err := c.SetVar("GRAPHITI_MODEL_NAME", config.ModelName.Value); err != nil {
			return fmt.Errorf("failed to set GRAPHITI_MODEL_NAME: %w", err)
		}
		if err := c.SetVar("NEO4J_USER", config.Neo4jUser.Value); err != nil {
			return fmt.Errorf("failed to set NEO4J_USER: %w", err)
		}
		if err := c.SetVar("NEO4J_PASSWORD", config.Neo4jPassword.Value); err != nil {
			return fmt.Errorf("failed to set NEO4J_PASSWORD: %w", err)
		}
		if err := c.SetVar("NEO4J_DATABASE", config.Neo4jDatabase.Value); err != nil {
			return fmt.Errorf("failed to set NEO4J_DATABASE: %w", err)
		}

	case "external":
		// for external mode, use provided endpoint
		// enable Graphiti
		if err := c.SetVar("GRAPHITI_ENABLED", "true"); err != nil {
			return fmt.Errorf("failed to set GRAPHITI_ENABLED: %w", err)
		}

		// update timeout only (model is configured on external server)
		if err := c.SetVar("GRAPHITI_TIMEOUT", config.Timeout.Value); err != nil {
			return fmt.Errorf("failed to set GRAPHITI_TIMEOUT: %w", err)
		}

	case "disabled":
		// for disabled mode, disable Graphiti
		if err := c.SetVar("GRAPHITI_ENABLED", "false"); err != nil {
			return fmt.Errorf("failed to set GRAPHITI_ENABLED: %w", err)
		}
		config.GraphitiURL.Value = ""
	}

	// update integration environment variables
	if err := c.SetVar("GRAPHITI_URL", config.GraphitiURL.Value); err != nil {
		return fmt.Errorf("failed to set GRAPHITI_URL: %w", err)
	}

	return nil
}

func (c *controller) ResetGraphitiConfig() *GraphitiConfig {
	vars := []string{
		"GRAPHITI_ENABLED",
		"GRAPHITI_URL",
		"GRAPHITI_TIMEOUT",
		"GRAPHITI_MODEL_NAME",
		"NEO4J_USER",
		"NEO4J_PASSWORD",
		"NEO4J_DATABASE",
		"NEO4J_URI",
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetGraphitiConfig()
}

// ObservabilityConfig represents observability configuration
type ObservabilityConfig struct {
	// deployment configuration
	DeploymentType string // "embedded" or "external" or "disabled"

	// embedded listen settings
	GrafanaListenIP    loader.EnvVar // GRAFANA_LISTEN_IP
	GrafanaListenPort  loader.EnvVar // GRAFANA_LISTEN_PORT
	OTelGrpcListenIP   loader.EnvVar // OTEL_GRPC_LISTEN_IP
	OTelGrpcListenPort loader.EnvVar // OTEL_GRPC_LISTEN_PORT
	OTelHttpListenIP   loader.EnvVar // OTEL_HTTP_LISTEN_IP
	OTelHttpListenPort loader.EnvVar // OTEL_HTTP_LISTEN_PORT

	// integration settings
	OTelHost loader.EnvVar // OTEL_HOST
}

// GetObservabilityConfig returns the current observability configuration
func (c *controller) GetObservabilityConfig() *ObservabilityConfig {
	vars, _ := c.GetVars([]string{
		"OTEL_HOST",
		"GRAFANA_LISTEN_IP",
		"GRAFANA_LISTEN_PORT",
		"OTEL_GRPC_LISTEN_IP",
		"OTEL_GRPC_LISTEN_PORT",
		"OTEL_HTTP_LISTEN_IP",
		"OTEL_HTTP_LISTEN_PORT",
	})

	// set defaults if missing
	defaults := map[string]string{
		"GRAFANA_LISTEN_IP":     "127.0.0.1",
		"GRAFANA_LISTEN_PORT":   "3000",
		"OTEL_GRPC_LISTEN_IP":   "127.0.0.1",
		"OTEL_GRPC_LISTEN_PORT": "8148",
		"OTEL_HTTP_LISTEN_IP":   "127.0.0.1",
		"OTEL_HTTP_LISTEN_PORT": "4318",
	}
	for k, def := range defaults {
		if v := vars[k]; v.Default == "" {
			v.Default = def
			vars[k] = v
		}
	}

	otelHost := vars["OTEL_HOST"]

	// determine deployment type based on endpoint value
	var deploymentType string
	switch otelHost.Value {
	case "":
		deploymentType = "disabled"
	case checker.DefaultObservabilityEndpoint:
		deploymentType = "embedded"
	default:
		deploymentType = "external"
	}

	return &ObservabilityConfig{
		DeploymentType:     deploymentType,
		OTelHost:           otelHost,
		GrafanaListenIP:    vars["GRAFANA_LISTEN_IP"],
		GrafanaListenPort:  vars["GRAFANA_LISTEN_PORT"],
		OTelGrpcListenIP:   vars["OTEL_GRPC_LISTEN_IP"],
		OTelGrpcListenPort: vars["OTEL_GRPC_LISTEN_PORT"],
		OTelHttpListenIP:   vars["OTEL_HTTP_LISTEN_IP"],
		OTelHttpListenPort: vars["OTEL_HTTP_LISTEN_PORT"],
	}
}

// UpdateObservabilityConfig updates the observability configuration
func (c *controller) UpdateObservabilityConfig(config *ObservabilityConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	langfuseOtelEnvVar, _ := c.GetVar("LANGFUSE_OTEL_EXPORTER_OTLP_ENDPOINT")

	// set deployment type based configuration
	switch config.DeploymentType {
	case "embedded":
		// for embedded mode, use default endpoints
		config.OTelHost.Value = checker.DefaultObservabilityEndpoint
		langfuseOtelEnvVar.Value = checker.DefaultLangfuseOtelEndpoint

		// update listen settings for embedded mode
		updates := map[string]string{
			"GRAFANA_LISTEN_IP":     config.GrafanaListenIP.Value,
			"GRAFANA_LISTEN_PORT":   config.GrafanaListenPort.Value,
			"OTEL_GRPC_LISTEN_IP":   config.OTelGrpcListenIP.Value,
			"OTEL_GRPC_LISTEN_PORT": config.OTelGrpcListenPort.Value,
			"OTEL_HTTP_LISTEN_IP":   config.OTelHttpListenIP.Value,
			"OTEL_HTTP_LISTEN_PORT": config.OTelHttpListenPort.Value,
		}
		if err := c.SetVars(updates); err != nil {
			return fmt.Errorf("failed to set embedded listen vars: %w", err)
		}

		// note: langfuse listen vars are set in UpdateLangfuseConfig
	case "external":
		// for external mode, use provided endpoint
		langfuseOtelEnvVar.Value = ""

	case "disabled":
		// for disabled mode, clear endpoint and disable
		config.OTelHost.Value = ""
		langfuseOtelEnvVar.Value = ""
	}

	// update Langfuse and Observability integration if it's enabled
	if err := c.SetVar(langfuseOtelEnvVar.Name, langfuseOtelEnvVar.Value); err != nil {
		return fmt.Errorf("failed to set LANGFUSE_OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
	}

	// update integration environment variables
	if err := c.SetVar(config.OTelHost.Name, config.OTelHost.Value); err != nil {
		return fmt.Errorf("failed to set OTEL_HOST: %w", err)
	}

	return nil
}

func (c *controller) ResetObservabilityConfig() *ObservabilityConfig {
	vars := []string{
		"OTEL_HOST",
		"GRAFANA_LISTEN_IP",
		"GRAFANA_LISTEN_PORT",
		"OTEL_GRPC_LISTEN_IP",
		"OTEL_GRPC_LISTEN_PORT",
		"OTEL_HTTP_LISTEN_IP",
		"OTEL_HTTP_LISTEN_PORT",
		// langfuse integration with observability
		"LANGFUSE_OTEL_EXPORTER_OTLP_ENDPOINT",
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetObservabilityConfig()
}

type SummarizerType string

const (
	SummarizerTypeGeneral   SummarizerType = "general"
	SummarizerTypeAssistant SummarizerType = "assistant"
)

// SummarizerConfig represents summarizer configuration settings
type SummarizerConfig struct {
	// type identifier ("general" or "assistant")
	Type SummarizerType

	// common boolean settings
	PreserveLast loader.EnvVar // PREFIX + PRESERVE_LAST
	UseQA        loader.EnvVar // PREFIX + USE_QA
	SumHumanInQA loader.EnvVar // PREFIX + SUM_MSG_HUMAN_IN_QA

	// size settings (in bytes)
	LastSecBytes loader.EnvVar // PREFIX + LAST_SEC_BYTES
	MaxBPBytes   loader.EnvVar // PREFIX + MAX_BP_BYTES
	MaxQABytes   loader.EnvVar // PREFIX + MAX_QA_BYTES

	// count settings
	MaxQASections  loader.EnvVar // PREFIX + MAX_QA_SECTIONS
	KeepQASections loader.EnvVar // PREFIX + KEEP_QA_SECTIONS
}

// GetSummarizerConfig returns summarizer configuration for specified type
func (c *controller) GetSummarizerConfig(summarizerType SummarizerType) *SummarizerConfig {
	var prefix string
	if summarizerType == SummarizerTypeAssistant {
		prefix = "ASSISTANT_SUMMARIZER_"
	} else {
		prefix = "SUMMARIZER_"
	}

	config := &SummarizerConfig{
		Type: summarizerType,
	}

	// Read variables directly from state (defaults handled by loader)
	config.PreserveLast, _ = c.GetVar(prefix + "PRESERVE_LAST")

	if summarizerType == SummarizerTypeGeneral {
		config.UseQA, _ = c.GetVar(prefix + "USE_QA")
		config.SumHumanInQA, _ = c.GetVar(prefix + "SUM_MSG_HUMAN_IN_QA")
	}

	// Size settings
	config.LastSecBytes, _ = c.GetVar(prefix + "LAST_SEC_BYTES")
	config.MaxBPBytes, _ = c.GetVar(prefix + "MAX_BP_BYTES")
	config.MaxQABytes, _ = c.GetVar(prefix + "MAX_QA_BYTES")

	// Count settings
	config.MaxQASections, _ = c.GetVar(prefix + "MAX_QA_SECTIONS")
	config.KeepQASections, _ = c.GetVar(prefix + "KEEP_QA_SECTIONS")

	return config
}

// UpdateSummarizerConfig updates summarizer configuration
func (c *controller) UpdateSummarizerConfig(config *SummarizerConfig) error {
	var prefix string
	if config.Type == SummarizerTypeAssistant {
		prefix = "ASSISTANT_SUMMARIZER_"
	} else {
		prefix = "SUMMARIZER_"
	}

	// Update boolean settings
	if err := c.SetVar(prefix+"PRESERVE_LAST", config.PreserveLast.Value); err != nil {
		return fmt.Errorf("failed to set %s: %w", prefix+"PRESERVE_LAST", err)
	}

	// General-specific boolean settings
	if config.Type == SummarizerTypeGeneral {
		if err := c.SetVar(prefix+"USE_QA", config.UseQA.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", prefix+"USE_QA", err)
		}
		if err := c.SetVar(prefix+"SUM_MSG_HUMAN_IN_QA", config.SumHumanInQA.Value); err != nil {
			return fmt.Errorf("failed to set %s: %w", prefix+"SUM_MSG_HUMAN_IN_QA", err)
		}
	}

	// Update size settings
	if err := c.SetVar(prefix+"LAST_SEC_BYTES", config.LastSecBytes.Value); err != nil {
		return fmt.Errorf("failed to set %s: %w", prefix+"LAST_SEC_BYTES", err)
	}
	if err := c.SetVar(prefix+"MAX_BP_BYTES", config.MaxBPBytes.Value); err != nil {
		return fmt.Errorf("failed to set %s: %w", prefix+"MAX_BP_BYTES", err)
	}
	if err := c.SetVar(prefix+"MAX_QA_BYTES", config.MaxQABytes.Value); err != nil {
		return fmt.Errorf("failed to set %s: %w", prefix+"MAX_QA_BYTES", err)
	}

	// Update count settings
	if err := c.SetVar(prefix+"MAX_QA_SECTIONS", config.MaxQASections.Value); err != nil {
		return fmt.Errorf("failed to set %s: %w", prefix+"MAX_QA_SECTIONS", err)
	}
	if err := c.SetVar(prefix+"KEEP_QA_SECTIONS", config.KeepQASections.Value); err != nil {
		return fmt.Errorf("failed to set %s: %w", prefix+"KEEP_QA_SECTIONS", err)
	}

	return nil
}

func (c *controller) ResetSummarizerConfig(summarizerType SummarizerType) *SummarizerConfig {
	var prefix string
	if summarizerType == SummarizerTypeAssistant {
		prefix = "ASSISTANT_SUMMARIZER_"
	} else {
		prefix = "SUMMARIZER_"
	}

	vars := []string{
		prefix + "PRESERVE_LAST",
		prefix + "LAST_SEC_BYTES",
		prefix + "MAX_BP_BYTES",
		prefix + "MAX_QA_BYTES",
		prefix + "MAX_QA_SECTIONS",
		prefix + "KEEP_QA_SECTIONS",
	}

	if summarizerType == SummarizerTypeGeneral {
		vars = append(vars,
			prefix+"USE_QA",
			prefix+"SUM_MSG_HUMAN_IN_QA",
		)
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetSummarizerConfig(summarizerType)
}

// EmbedderConfig represents embedder configuration settings
type EmbedderConfig struct {
	// direct form field mappings using loader.EnvVar
	// these fields directly correspond to environment variables and form inputs (not computed)
	Provider      loader.EnvVar // EMBEDDING_PROVIDER
	URL           loader.EnvVar // EMBEDDING_URL
	APIKey        loader.EnvVar // EMBEDDING_KEY
	Model         loader.EnvVar // EMBEDDING_MODEL
	BatchSize     loader.EnvVar // EMBEDDING_BATCH_SIZE
	StripNewLines loader.EnvVar // EMBEDDING_STRIP_NEW_LINES

	// computed fields (not directly mapped to env vars)
	Configured bool
	Installed  bool
}

// GetEmbedderConfig returns current embedder configuration
func (c *controller) GetEmbedderConfig() *EmbedderConfig {
	config := &EmbedderConfig{}
	config.Provider, _ = c.GetVar("EMBEDDING_PROVIDER")
	config.URL, _ = c.GetVar("EMBEDDING_URL")
	config.APIKey, _ = c.GetVar("EMBEDDING_KEY")
	config.Model, _ = c.GetVar("EMBEDDING_MODEL")
	config.BatchSize, _ = c.GetVar("EMBEDDING_BATCH_SIZE")
	config.StripNewLines, _ = c.GetVar("EMBEDDING_STRIP_NEW_LINES")
	config.Installed = c.checker.PentagiInstalled

	// Determine if configured based on provider requirements
	switch config.Provider.Value {
	case "openai", "":
		// For OpenAI, check if we have API key either in EMBEDDING_KEY or OPEN_AI_KEY
		openaiKey, _ := c.GetVar("OPEN_AI_KEY")
		config.Configured = config.APIKey.Value != "" || openaiKey.Value != ""
	case "ollama":
		// for Ollama, no API key required, but URL must be provided
		config.Configured = config.URL.Value != ""
	case "huggingface", "googleai":
		// These require API key
		config.Configured = config.APIKey.Value != ""
	default:
		// Others are configured if API key is present
		config.Configured = config.APIKey.Value != ""
	}

	return config
}

// UpdateEmbedderConfig updates embedder configuration
func (c *controller) UpdateEmbedderConfig(config *EmbedderConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Update environment variables
	if err := c.SetVar("EMBEDDING_PROVIDER", config.Provider.Value); err != nil {
		return fmt.Errorf("failed to set EMBEDDING_PROVIDER: %w", err)
	}
	if err := c.SetVar("EMBEDDING_URL", config.URL.Value); err != nil {
		return fmt.Errorf("failed to set EMBEDDING_URL: %w", err)
	}
	if err := c.SetVar("EMBEDDING_KEY", config.APIKey.Value); err != nil {
		return fmt.Errorf("failed to set EMBEDDING_KEY: %w", err)
	}
	if err := c.SetVar("EMBEDDING_MODEL", config.Model.Value); err != nil {
		return fmt.Errorf("failed to set EMBEDDING_MODEL: %w", err)
	}
	if err := c.SetVar("EMBEDDING_BATCH_SIZE", config.BatchSize.Value); err != nil {
		return fmt.Errorf("failed to set EMBEDDING_BATCH_SIZE: %w", err)
	}
	if err := c.SetVar("EMBEDDING_STRIP_NEW_LINES", config.StripNewLines.Value); err != nil {
		return fmt.Errorf("failed to set EMBEDDING_STRIP_NEW_LINES: %w", err)
	}

	return nil
}

func (c *controller) ResetEmbedderConfig() *EmbedderConfig {
	vars := []string{
		"EMBEDDING_PROVIDER",
		"EMBEDDING_URL",
		"EMBEDDING_KEY",
		"EMBEDDING_MODEL",
		"EMBEDDING_BATCH_SIZE",
		"EMBEDDING_STRIP_NEW_LINES",
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetEmbedderConfig()
}

// AIAgentsConfig represents extra AI agents configuration
type AIAgentsConfig struct {
	// direct form field mappings using loader.EnvVar
	// these fields directly correspond to environment variables and form inputs (not computed)
	HumanInTheLoop     loader.EnvVar // ASK_USER
	AssistantUseAgents loader.EnvVar // ASSISTANT_USE_AGENTS
}

func (c *controller) GetAIAgentsConfig() *AIAgentsConfig {
	config := &AIAgentsConfig{}

	config.HumanInTheLoop, _ = c.GetVar("ASK_USER")
	config.AssistantUseAgents, _ = c.GetVar("ASSISTANT_USE_AGENTS")

	return config
}

func (c *controller) UpdateAIAgentsConfig(config *AIAgentsConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := c.SetVar("ASK_USER", config.HumanInTheLoop.Value); err != nil {
		return fmt.Errorf("failed to set ASK_USER: %w", err)
	}
	if err := c.SetVar("ASSISTANT_USE_AGENTS", config.AssistantUseAgents.Value); err != nil {
		return fmt.Errorf("failed to set ASSISTANT_USE_AGENTS: %w", err)
	}

	return nil
}

func (c *controller) ResetAIAgentsConfig() *AIAgentsConfig {
	vars := []string{
		"ASK_USER",
		"ASSISTANT_USE_AGENTS",
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetAIAgentsConfig()
}

// ScraperConfig represents scraper configuration settings
type ScraperConfig struct {
	// direct form field mappings using loader.EnvVar
	// these fields directly correspond to environment variables and form inputs (not computed)
	PublicURL             loader.EnvVar // SCRAPER_PUBLIC_URL
	PrivateURL            loader.EnvVar // SCRAPER_PRIVATE_URL
	LocalUsername         loader.EnvVar // LOCAL_SCRAPER_USERNAME
	LocalPassword         loader.EnvVar // LOCAL_SCRAPER_PASSWORD
	MaxConcurrentSessions loader.EnvVar // LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS

	// computed fields (not directly mapped to env vars)
	// these are derived from the above EnvVar fields
	Mode string // "embedded", "external", "disabled" - computed from PrivateURL

	// parsed credentials for external mode (extracted from URLs)
	PublicUsername  string
	PublicPassword  string
	PrivateUsername string
	PrivatePassword string
}

// GetScraperConfig returns current scraper configuration
func (c *controller) GetScraperConfig() *ScraperConfig {
	// get all environment variables using the state controller
	publicURL, _ := c.GetVar("SCRAPER_PUBLIC_URL")
	privateURL, _ := c.GetVar("SCRAPER_PRIVATE_URL")
	localUsername, _ := c.GetVar("LOCAL_SCRAPER_USERNAME")
	localPassword, _ := c.GetVar("LOCAL_SCRAPER_PASSWORD")
	maxSessions, _ := c.GetVar("LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS")

	config := &ScraperConfig{
		PublicURL:             publicURL,
		PrivateURL:            privateURL,
		LocalUsername:         localUsername,
		LocalPassword:         localPassword,
		MaxConcurrentSessions: maxSessions,
	}

	config.Mode = c.determineScraperMode(privateURL.Value, publicURL.Value)

	if config.Mode == "external" || config.Mode == "embedded" {
		config.PublicUsername, config.PublicPassword = c.extractCredentialsFromURL(publicURL.Value)
		config.PrivateUsername, config.PrivatePassword = c.extractCredentialsFromURL(privateURL.Value)
		config.PublicURL.Value = RemoveCredentialsFromURL(publicURL.Value)
		config.PrivateURL.Value = RemoveCredentialsFromURL(privateURL.Value)
	}

	return config
}

// determineScraperMode determines scraper mode based on private URL
func (c *controller) determineScraperMode(privateURL, publicURL string) string {
	if privateURL == "" && publicURL == "" {
		return "disabled"
	}

	// parse URL to check if this is embedded mode (domain "scraper" and schema "https")
	parsedURL, err := url.Parse(privateURL)
	if err != nil {
		// if URL is malformed, treat as external
		return "external"
	}

	if parsedURL.Scheme == DefaultScraperSchema && parsedURL.Hostname() == DefaultScraperDomain {
		return "embedded"
	}

	return "external"
}

// extractCredentialsFromURL extracts username and password from URL
func (c *controller) extractCredentialsFromURL(urlStr string) (username, password string) {
	if urlStr == "" {
		return "", ""
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", ""
	}

	if parsedURL.User == nil {
		return "", ""
	}

	username = parsedURL.User.Username()
	password, _ = parsedURL.User.Password()

	return username, password
}

// UpdateScraperConfig updates scraper configuration
func (c *controller) UpdateScraperConfig(config *ScraperConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	switch config.Mode {
	case "disabled":
		// clear scraper URLs, preserve local settings
		if err := c.SetVar("SCRAPER_PUBLIC_URL", ""); err != nil {
			return fmt.Errorf("failed to clear SCRAPER_PUBLIC_URL: %w", err)
		}
		if err := c.SetVar("SCRAPER_PRIVATE_URL", ""); err != nil {
			return fmt.Errorf("failed to clear SCRAPER_PRIVATE_URL: %w", err)
		}
		// local settings remain unchanged

	case "external":
		// construct URLs with credentials if provided
		privateURL := config.PrivateURL.Value
		if config.PrivateUsername != "" && config.PrivatePassword != "" {
			privateURL = c.addCredentialsToURL(config.PrivateURL.Value, config.PrivateUsername, config.PrivatePassword)
		}

		publicURL := config.PublicURL.Value
		if config.PublicUsername != "" && config.PublicPassword != "" {
			publicURL = c.addCredentialsToURL(config.PublicURL.Value, config.PublicUsername, config.PublicPassword)
		}

		if err := c.SetVar("SCRAPER_PUBLIC_URL", publicURL); err != nil {
			return fmt.Errorf("failed to set SCRAPER_PUBLIC_URL: %w", err)
		}
		if err := c.SetVar("SCRAPER_PRIVATE_URL", privateURL); err != nil {
			return fmt.Errorf("failed to set SCRAPER_PRIVATE_URL: %w", err)
		}
		// local settings remain unchanged

	case "embedded":
		// handle embedded mode
		privateURL := DefaultScraperBaseURL
		if config.PrivateUsername != "" && config.PrivatePassword != "" {
			privateURL = c.addCredentialsToURL(privateURL, config.PrivateUsername, config.PrivatePassword)
		}

		publicURL := config.PublicURL.Value
		if config.PublicUsername != "" && config.PublicPassword != "" {
			// fallback to private URL if public URL is not set
			if publicURL == "" {
				publicURL = privateURL
			}
			publicURL = c.addCredentialsToURL(publicURL, config.PublicUsername, config.PublicPassword)
		}

		// update all relevant variables
		if err := c.SetVar("SCRAPER_PUBLIC_URL", publicURL); err != nil {
			return fmt.Errorf("failed to set SCRAPER_PUBLIC_URL: %w", err)
		}
		if err := c.SetVar("SCRAPER_PRIVATE_URL", privateURL); err != nil {
			return fmt.Errorf("failed to set SCRAPER_PRIVATE_URL: %w", err)
		}
		if err := c.SetVar("LOCAL_SCRAPER_USERNAME", config.PrivateUsername); err != nil {
			return fmt.Errorf("failed to set LOCAL_SCRAPER_USERNAME: %w", err)
		}
		if err := c.SetVar("LOCAL_SCRAPER_PASSWORD", config.PrivatePassword); err != nil {
			return fmt.Errorf("failed to set LOCAL_SCRAPER_PASSWORD: %w", err)
		}
		if err := c.SetVar("LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS", config.MaxConcurrentSessions.Value); err != nil {
			return fmt.Errorf("failed to set LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS: %w", err)
		}
	}

	return nil
}

// addCredentialsToURL adds username and password to URL
func (c *controller) addCredentialsToURL(urlStr, username, password string) string {
	if username == "" || password == "" {
		return urlStr
	}

	if urlStr == "" {
		// do not implicitly set a default base url for non-scraper contexts
		// caller must provide a valid base url; return empty to avoid crafting invalid urls
		return ""
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	// set user info
	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String()
}

// ResetScraperConfig resets scraper configuration to defaults
func (c *controller) ResetScraperConfig() *ScraperConfig {
	// reset all scraper-related environment variables to their defaults
	vars := []string{
		"SCRAPER_PUBLIC_URL",
		"SCRAPER_PRIVATE_URL",
		"LOCAL_SCRAPER_USERNAME",
		"LOCAL_SCRAPER_PASSWORD",
		"LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS",
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetScraperConfig()
}

// SearchEnginesConfig represents search engines configuration settings
type SearchEnginesConfig struct {
	// direct form field mappings using loader.EnvVar
	// these fields directly correspond to environment variables and form inputs (not computed)
	DuckDuckGoEnabled loader.EnvVar // DUCKDUCKGO_ENABLED
	SploitusEnabled   loader.EnvVar // SPLOITUS_ENABLED
	PerplexityAPIKey  loader.EnvVar // PERPLEXITY_API_KEY
	TavilyAPIKey      loader.EnvVar // TAVILY_API_KEY
	TraversaalAPIKey  loader.EnvVar // TRAVERSAAL_API_KEY
	GoogleAPIKey      loader.EnvVar // GOOGLE_API_KEY
	GoogleCXKey       loader.EnvVar // GOOGLE_CX_KEY
	GoogleLRKey       loader.EnvVar // GOOGLE_LR_KEY

	// perplexity extra settings
	PerplexityModel       loader.EnvVar // PERPLEXITY_MODEL
	PerplexityContextSize loader.EnvVar // PERPLEXITY_CONTEXT_SIZE

	// searxng extra settings
	SearxngURL        loader.EnvVar // SEARXNG_URL
	SearxngCategories loader.EnvVar // SEARXNG_CATEGORIES
	SearxngLanguage   loader.EnvVar // SEARXNG_LANGUAGE
	SearxngSafeSearch loader.EnvVar // SEARXNG_SAFESEARCH
	SearxngTimeRange  loader.EnvVar // SEARXNG_TIME_RANGE

	// computed fields (not directly mapped to env vars)
	ConfiguredCount int // number of configured engines
}

// GetSearchEnginesConfig returns current search engines configuration
func (c *controller) GetSearchEnginesConfig() *SearchEnginesConfig {
	// get all environment variables using the state controller
	duckduckgoEnabled, _ := c.GetVar("DUCKDUCKGO_ENABLED")
	sploitusEnabled, _ := c.GetVar("SPLOITUS_ENABLED")
	perplexityAPIKey, _ := c.GetVar("PERPLEXITY_API_KEY")
	tavilyAPIKey, _ := c.GetVar("TAVILY_API_KEY")
	traversaalAPIKey, _ := c.GetVar("TRAVERSAAL_API_KEY")
	googleAPIKey, _ := c.GetVar("GOOGLE_API_KEY")
	googleCXKey, _ := c.GetVar("GOOGLE_CX_KEY")
	googleLRKey, _ := c.GetVar("GOOGLE_LR_KEY")
	perplexityModel, _ := c.GetVar("PERPLEXITY_MODEL")
	perplexityContextSize, _ := c.GetVar("PERPLEXITY_CONTEXT_SIZE")
	searxngURL, _ := c.GetVar("SEARXNG_URL")
	searxngCategories, _ := c.GetVar("SEARXNG_CATEGORIES")
	searxngLanguage, _ := c.GetVar("SEARXNG_LANGUAGE")
	searxngSafeSearch, _ := c.GetVar("SEARXNG_SAFESEARCH")
	searxngTimeRange, _ := c.GetVar("SEARXNG_TIME_RANGE")

	config := &SearchEnginesConfig{
		DuckDuckGoEnabled:     duckduckgoEnabled,
		SploitusEnabled:       sploitusEnabled,
		PerplexityAPIKey:      perplexityAPIKey,
		PerplexityModel:       perplexityModel,
		PerplexityContextSize: perplexityContextSize,
		TavilyAPIKey:          tavilyAPIKey,
		TraversaalAPIKey:      traversaalAPIKey,
		GoogleAPIKey:          googleAPIKey,
		GoogleCXKey:           googleCXKey,
		GoogleLRKey:           googleLRKey,
		SearxngURL:            searxngURL,
		SearxngCategories:     searxngCategories,
		SearxngLanguage:       searxngLanguage,
		SearxngSafeSearch:     searxngSafeSearch,
		SearxngTimeRange:      searxngTimeRange,
	}

	// compute configured count
	configuredCount := 0
	if duckduckgoEnabled.Value == "true" {
		configuredCount++
	} else if duckduckgoEnabled.Value == "" && duckduckgoEnabled.Default == "true" {
		configuredCount++
	}
	if sploitusEnabled.Value == "true" {
		configuredCount++
	} else if sploitusEnabled.Value == "" && sploitusEnabled.Default == "true" {
		configuredCount++
	}
	if perplexityAPIKey.Value != "" {
		configuredCount++
	}
	if tavilyAPIKey.Value != "" {
		configuredCount++
	}
	if traversaalAPIKey.Value != "" {
		configuredCount++
	}
	if googleAPIKey.Value != "" && googleCXKey.Value != "" {
		configuredCount++
	}
	if searxngURL.Value != "" {
		configuredCount++
	}
	config.ConfiguredCount = configuredCount

	return config
}

// UpdateSearchEnginesConfig updates search engines configuration
func (c *controller) UpdateSearchEnginesConfig(config *SearchEnginesConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// update environment variables
	if err := c.SetVar("DUCKDUCKGO_ENABLED", config.DuckDuckGoEnabled.Value); err != nil {
		return fmt.Errorf("failed to set DUCKDUCKGO_ENABLED: %w", err)
	}
	if err := c.SetVar("SPLOITUS_ENABLED", config.SploitusEnabled.Value); err != nil {
		return fmt.Errorf("failed to set SPLOITUS_ENABLED: %w", err)
	}
	if err := c.SetVar("PERPLEXITY_API_KEY", config.PerplexityAPIKey.Value); err != nil {
		return fmt.Errorf("failed to set PERPLEXITY_API_KEY: %w", err)
	}
	if err := c.SetVar("PERPLEXITY_MODEL", config.PerplexityModel.Value); err != nil {
		return fmt.Errorf("failed to set PERPLEXITY_MODEL: %w", err)
	}
	if err := c.SetVar("PERPLEXITY_CONTEXT_SIZE", config.PerplexityContextSize.Value); err != nil {
		return fmt.Errorf("failed to set PERPLEXITY_CONTEXT_SIZE: %w", err)
	}
	if err := c.SetVar("TAVILY_API_KEY", config.TavilyAPIKey.Value); err != nil {
		return fmt.Errorf("failed to set TAVILY_API_KEY: %w", err)
	}
	if err := c.SetVar("TRAVERSAAL_API_KEY", config.TraversaalAPIKey.Value); err != nil {
		return fmt.Errorf("failed to set TRAVERSAAL_API_KEY: %w", err)
	}
	if err := c.SetVar("GOOGLE_API_KEY", config.GoogleAPIKey.Value); err != nil {
		return fmt.Errorf("failed to set GOOGLE_API_KEY: %w", err)
	}
	if err := c.SetVar("GOOGLE_CX_KEY", config.GoogleCXKey.Value); err != nil {
		return fmt.Errorf("failed to set GOOGLE_CX_KEY: %w", err)
	}
	if err := c.SetVar("GOOGLE_LR_KEY", config.GoogleLRKey.Value); err != nil {
		return fmt.Errorf("failed to set GOOGLE_LR_KEY: %w", err)
	}
	if err := c.SetVar("SEARXNG_URL", config.SearxngURL.Value); err != nil {
		return fmt.Errorf("failed to set SEARXNG_URL: %w", err)
	}
	if err := c.SetVar("SEARXNG_CATEGORIES", config.SearxngCategories.Value); err != nil {
		return fmt.Errorf("failed to set SEARXNG_CATEGORIES: %w", err)
	}
	if err := c.SetVar("SEARXNG_LANGUAGE", config.SearxngLanguage.Value); err != nil {
		return fmt.Errorf("failed to set SEARXNG_LANGUAGE: %w", err)
	}
	if err := c.SetVar("SEARXNG_SAFESEARCH", config.SearxngSafeSearch.Value); err != nil {
		return fmt.Errorf("failed to set SEARXNG_SAFESEARCH: %w", err)
	}
	if err := c.SetVar("SEARXNG_TIME_RANGE", config.SearxngTimeRange.Value); err != nil {
		return fmt.Errorf("failed to set SEARXNG_TIME_RANGE: %w", err)
	}

	return nil
}

// ResetSearchEnginesConfig resets search engines configuration to defaults
func (c *controller) ResetSearchEnginesConfig() *SearchEnginesConfig {
	// reset all search engines-related environment variables to their defaults
	vars := []string{
		"DUCKDUCKGO_ENABLED",
		"SPLOITUS_ENABLED",
		"PERPLEXITY_API_KEY",
		"PERPLEXITY_MODEL",
		"PERPLEXITY_CONTEXT_SIZE",
		"TAVILY_API_KEY",
		"TRAVERSAAL_API_KEY",
		"GOOGLE_API_KEY",
		"GOOGLE_CX_KEY",
		"GOOGLE_LR_KEY",
		"SEARXNG_URL",
		"SEARXNG_CATEGORIES",
		"SEARXNG_LANGUAGE",
		"SEARXNG_SAFESEARCH",
		"SEARXNG_TIME_RANGE",
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetSearchEnginesConfig()
}

// DockerConfig represents Docker environment configuration
type DockerConfig struct {
	// direct form field mappings using loader.EnvVar
	// these fields directly correspond to environment variables and form inputs (not computed)
	DockerInside                 loader.EnvVar // DOCKER_INSIDE
	DockerNetAdmin               loader.EnvVar // DOCKER_NET_ADMIN
	DockerSocket                 loader.EnvVar // DOCKER_SOCKET
	DockerNetwork                loader.EnvVar // DOCKER_NETWORK
	DockerPublicIP               loader.EnvVar // DOCKER_PUBLIC_IP
	DockerWorkDir                loader.EnvVar // DOCKER_WORK_DIR
	DockerDefaultImage           loader.EnvVar // DOCKER_DEFAULT_IMAGE
	DockerDefaultImageForPentest loader.EnvVar // DOCKER_DEFAULT_IMAGE_FOR_PENTEST

	// TLS connection settings (optional)
	DockerHost         loader.EnvVar // DOCKER_HOST
	DockerTLSVerify    loader.EnvVar // DOCKER_TLS_VERIFY
	HostDockerCertPath loader.EnvVar // PENTAGI_DOCKER_CERT_PATH

	// computed fields (not directly mapped to env vars)
	Configured bool
}

// GetDockerConfig returns the current Docker configuration
func (c *controller) GetDockerConfig() *DockerConfig {
	vars, _ := c.GetVars([]string{
		"DOCKER_INSIDE",
		"DOCKER_NET_ADMIN",
		"DOCKER_SOCKET",
		"DOCKER_NETWORK",
		"DOCKER_PUBLIC_IP",
		"DOCKER_WORK_DIR",
		"DOCKER_DEFAULT_IMAGE",
		"DOCKER_DEFAULT_IMAGE_FOR_PENTEST",
		"DOCKER_HOST",
		"DOCKER_TLS_VERIFY",
		"PENTAGI_DOCKER_CERT_PATH",
	})

	config := &DockerConfig{
		DockerInside:                 vars["DOCKER_INSIDE"],
		DockerNetAdmin:               vars["DOCKER_NET_ADMIN"],
		DockerSocket:                 vars["DOCKER_SOCKET"],
		DockerNetwork:                vars["DOCKER_NETWORK"],
		DockerPublicIP:               vars["DOCKER_PUBLIC_IP"],
		DockerWorkDir:                vars["DOCKER_WORK_DIR"],
		DockerDefaultImage:           vars["DOCKER_DEFAULT_IMAGE"],
		DockerDefaultImageForPentest: vars["DOCKER_DEFAULT_IMAGE_FOR_PENTEST"],
		DockerHost:                   vars["DOCKER_HOST"],
		DockerTLSVerify:              vars["DOCKER_TLS_VERIFY"],
		HostDockerCertPath:           vars["PENTAGI_DOCKER_CERT_PATH"],
	}

	// patch docker host default value
	if config.DockerHost.Default == "" {
		config.DockerHost.Default = "unix:///var/run/docker.sock"
	}

	// determine if Docker is configured
	// basic configuration is considered complete if DOCKER_INSIDE is set or default images are specified
	config.Configured = config.DockerInside.Value != "" ||
		config.DockerDefaultImage.Value != "" ||
		config.DockerDefaultImageForPentest.Value != ""

	return config
}

// UpdateDockerConfig updates the Docker configuration
func (c *controller) UpdateDockerConfig(config *DockerConfig) error {
	updates := map[string]string{
		"DOCKER_INSIDE":                    config.DockerInside.Value,
		"DOCKER_NET_ADMIN":                 config.DockerNetAdmin.Value,
		"DOCKER_SOCKET":                    config.DockerSocket.Value,
		"DOCKER_NETWORK":                   config.DockerNetwork.Value,
		"DOCKER_PUBLIC_IP":                 config.DockerPublicIP.Value,
		"DOCKER_WORK_DIR":                  config.DockerWorkDir.Value,
		"DOCKER_DEFAULT_IMAGE":             config.DockerDefaultImage.Value,
		"DOCKER_DEFAULT_IMAGE_FOR_PENTEST": config.DockerDefaultImageForPentest.Value,
		"DOCKER_HOST":                      config.DockerHost.Value,
		"DOCKER_TLS_VERIFY":                config.DockerTLSVerify.Value,
		"PENTAGI_DOCKER_CERT_PATH":         config.HostDockerCertPath.Value,
	}

	dockerHost := config.DockerHost.Value
	if strings.HasPrefix(dockerHost, "unix://") && !config.DockerHost.IsDefault() {
		// mount custom docker socket to the pentagi container
		updates["PENTAGI_DOCKER_SOCKET"] = strings.TrimPrefix(dockerHost, "unix://")
	} else {
		// ensure previous custom socket mapping is cleared when not using unix socket
		updates["PENTAGI_DOCKER_SOCKET"] = ""
	}

	if config.HostDockerCertPath.Value != "" {
		updates["DOCKER_CERT_PATH"] = DefaultDockerCertPath
	} else {
		updates["DOCKER_CERT_PATH"] = ""
	}

	if err := c.SetVars(updates); err != nil {
		return err
	}

	return nil
}

// ResetDockerConfig resets the Docker configuration to defaults
func (c *controller) ResetDockerConfig() *DockerConfig {
	vars := []string{
		"DOCKER_INSIDE",
		"DOCKER_NET_ADMIN",
		"DOCKER_SOCKET",
		"DOCKER_NETWORK",
		"DOCKER_PUBLIC_IP",
		"DOCKER_WORK_DIR",
		"DOCKER_DEFAULT_IMAGE",
		"DOCKER_DEFAULT_IMAGE_FOR_PENTEST",
		"DOCKER_HOST",
		"DOCKER_TLS_VERIFY",
		"DOCKER_CERT_PATH",
		// Volume mapping for docker socket
		"PENTAGI_DOCKER_SOCKET",
		"PENTAGI_DOCKER_CERT_PATH",
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetDockerConfig()
}

// ServerSettingsConfig represents PentAGI server settings configuration
type ServerSettingsConfig struct {
	// direct form field mappings using loader.EnvVar
	LicenseKey          loader.EnvVar // LICENSE_KEY
	ListenIP            loader.EnvVar // PENTAGI_LISTEN_IP
	ListenPort          loader.EnvVar // PENTAGI_LISTEN_PORT
	PublicURL           loader.EnvVar // PUBLIC_URL
	CorsOrigins         loader.EnvVar // CORS_ORIGINS
	CookieSigningSalt   loader.EnvVar // COOKIE_SIGNING_SALT
	ProxyURL            loader.EnvVar // PROXY_URL
	ExternalSSLCAPath   loader.EnvVar // EXTERNAL_SSL_CA_PATH
	ExternalSSLInsecure loader.EnvVar // EXTERNAL_SSL_INSECURE
	SSLDir              loader.EnvVar // PENTAGI_SSL_DIR
	DataDir             loader.EnvVar // PENTAGI_DATA_DIR

	// parsed credentials for proxy server (extracted from URLs)
	ProxyUsername string
	ProxyPassword string
}

// GetServerSettingsConfig returns current server settings
func (c *controller) GetServerSettingsConfig() *ServerSettingsConfig {
	vars, _ := c.GetVars([]string{
		"LICENSE_KEY",
		"PENTAGI_LISTEN_IP",
		"PENTAGI_LISTEN_PORT",
		"PUBLIC_URL",
		"CORS_ORIGINS",
		"COOKIE_SIGNING_SALT",
		"PROXY_URL",
		"EXTERNAL_SSL_CA_PATH",
		"EXTERNAL_SSL_INSECURE",
		"PENTAGI_SSL_DIR",
		"PENTAGI_DATA_DIR",
	})

	defaults := map[string]string{
		"LICENSE_KEY":           "",
		"PENTAGI_LISTEN_IP":     "127.0.0.1",
		"PENTAGI_LISTEN_PORT":   "8443",
		"PUBLIC_URL":            "https://localhost:8443",
		"CORS_ORIGINS":          "https://localhost:8443",
		"PENTAGI_DATA_DIR":      "pentagi-data",
		"PENTAGI_SSL_DIR":       "pentagi-ssl",
		"EXTERNAL_SSL_INSECURE": "false",
	}

	for varName, defaultValue := range defaults {
		if v := vars[varName]; v.Default == "" {
			v.Default = defaultValue
			vars[varName] = v
		}
	}

	cfg := &ServerSettingsConfig{
		LicenseKey:          vars["LICENSE_KEY"],
		ListenIP:            vars["PENTAGI_LISTEN_IP"],
		ListenPort:          vars["PENTAGI_LISTEN_PORT"],
		PublicURL:           vars["PUBLIC_URL"],
		CorsOrigins:         vars["CORS_ORIGINS"],
		CookieSigningSalt:   vars["COOKIE_SIGNING_SALT"],
		ProxyURL:            vars["PROXY_URL"],
		ExternalSSLCAPath:   vars["EXTERNAL_SSL_CA_PATH"],
		ExternalSSLInsecure: vars["EXTERNAL_SSL_INSECURE"],
		SSLDir:              vars["PENTAGI_SSL_DIR"],
		DataDir:             vars["PENTAGI_DATA_DIR"],
	}

	// split proxy URL into credentials + naked URL for UI
	if cfg.ProxyURL.Value != "" {
		user, pass := c.extractCredentialsFromURL(cfg.ProxyURL.Value)
		cfg.ProxyUsername = user
		cfg.ProxyPassword = pass
		cfg.ProxyURL.Value = RemoveCredentialsFromURL(cfg.ProxyURL.Value)
	}

	return cfg
}

// UpdateServerSettingsConfig updates server settings
func (c *controller) UpdateServerSettingsConfig(config *ServerSettingsConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// build proxy URL with credentials if provided
	proxyURL := config.ProxyURL.Value
	if proxyURL != "" && config.ProxyUsername != "" && config.ProxyPassword != "" {
		// add credentials to proxy URL only if proxy URL is provided
		proxyURL = c.addCredentialsToURL(proxyURL, config.ProxyUsername, config.ProxyPassword)
	}

	updates := map[string]string{
		"LICENSE_KEY":           config.LicenseKey.Value,
		"PENTAGI_LISTEN_IP":     config.ListenIP.Value,
		"PENTAGI_LISTEN_PORT":   config.ListenPort.Value,
		"PUBLIC_URL":            config.PublicURL.Value,
		"CORS_ORIGINS":          config.CorsOrigins.Value,
		"COOKIE_SIGNING_SALT":   config.CookieSigningSalt.Value,
		"PROXY_URL":             proxyURL,
		"EXTERNAL_SSL_CA_PATH":  config.ExternalSSLCAPath.Value,
		"EXTERNAL_SSL_INSECURE": config.ExternalSSLInsecure.Value,
		"PENTAGI_SSL_DIR":       config.SSLDir.Value,
		"PENTAGI_DATA_DIR":      config.DataDir.Value,
	}

	if err := c.SetVars(updates); err != nil {
		return err
	}

	return nil
}

// ResetServerSettingsConfig resets server settings to defaults
func (c *controller) ResetServerSettingsConfig() *ServerSettingsConfig {
	vars := []string{
		"LICENSE_KEY",
		"PENTAGI_LISTEN_IP",
		"PENTAGI_LISTEN_PORT",
		"PUBLIC_URL",
		"CORS_ORIGINS",
		"COOKIE_SIGNING_SALT",
		"PROXY_URL",
		"EXTERNAL_SSL_CA_PATH",
		"EXTERNAL_SSL_INSECURE",
		"PENTAGI_SSL_DIR",
		"PENTAGI_DATA_DIR",
	}

	if err := c.ResetVars(vars); err != nil {
		return nil
	}

	return c.GetServerSettingsConfig()
}

// ChangeInfo represents information about a single environment variable change
type ChangeInfo struct {
	Variable    string // environment variable name
	Description string // localized description
	NewValue    string // new value being set
	Masked      bool   // whether the value should be masked in display
}

// ApplyChangesConfig contains information about pending changes and installation status
type ApplyChangesConfig struct {
	// installation state
	IsInstalled bool // whether PentAGI is currently installed

	// deployment selections
	LangfuseEnabled      bool // whether Langfuse embedded deployment is selected
	ObservabilityEnabled bool // whether Observability embedded deployment is selected

	// changes information
	Changes      []ChangeInfo // list of pending environment variable changes
	ChangesCount int          // total number of changes
	HasCritical  bool         // whether there are critical changes requiring restart
	HasSecrets   bool         // whether there are secret/sensitive changes
}

// GetApplyChangesConfig returns the current apply changes configuration
func (c *controller) GetApplyChangesConfig() *ApplyChangesConfig {
	config := &ApplyChangesConfig{
		IsInstalled: c.checker.PentagiInstalled,
		Changes:     []ChangeInfo{},
	}

	// check deployment selections
	langfuseConfig := c.GetLangfuseConfig()
	config.LangfuseEnabled = langfuseConfig.DeploymentType == "embedded"

	observabilityConfig := c.GetObservabilityConfig()
	config.ObservabilityEnabled = observabilityConfig.DeploymentType == "embedded"

	// collect all changed variables
	allVars := c.GetAllVars()
	for varName, envVar := range allVars {
		if envVar.IsChanged {
			description := c.getVariableDescription(varName)
			masked := c.isVariableMasked(varName)
			value := envVar.Value
			if value == "" {
				value = "{EMPTY}"
			}

			config.Changes = append(config.Changes, ChangeInfo{
				Variable:    varName,
				Description: description,
				NewValue:    value,
				Masked:      masked,
			})

			// mark critical and secret changes
			if c.isCriticalVariable(varName) {
				config.HasCritical = true
			}
			if masked {
				config.HasSecrets = true
			}
		}
	}

	slices.SortFunc(config.Changes, func(a, b ChangeInfo) int {
		return strings.Compare(a.Description, b.Description)
	})

	config.ChangesCount = len(config.Changes)
	return config
}

// getVariableDescription returns a user-friendly description for an environment variable
func (c *controller) getVariableDescription(varName string) string {
	// map of environment variable name -> description
	envVarDescriptions := map[string]string{
		"OPEN_AI_KEY":                       locale.EnvDesc_OPEN_AI_KEY,
		"OPEN_AI_SERVER_URL":                locale.EnvDesc_OPEN_AI_SERVER_URL,
		"ANTHROPIC_API_KEY":                 locale.EnvDesc_ANTHROPIC_API_KEY,
		"ANTHROPIC_SERVER_URL":              locale.EnvDesc_ANTHROPIC_SERVER_URL,
		"GEMINI_API_KEY":                    locale.EnvDesc_GEMINI_API_KEY,
		"GEMINI_SERVER_URL":                 locale.EnvDesc_GEMINI_SERVER_URL,
		"BEDROCK_ACCESS_KEY_ID":             locale.EnvDesc_BEDROCK_ACCESS_KEY_ID,
		"BEDROCK_SECRET_ACCESS_KEY":         locale.EnvDesc_BEDROCK_SECRET_ACCESS_KEY,
		"BEDROCK_SESSION_TOKEN":             locale.EnvDesc_BEDROCK_SESSION_TOKEN,
		"BEDROCK_REGION":                    locale.EnvDesc_BEDROCK_REGION,
		"BEDROCK_SERVER_URL":                locale.EnvDesc_BEDROCK_SERVER_URL,
		"OLLAMA_SERVER_URL":                 locale.EnvDesc_OLLAMA_SERVER_URL,
		"OLLAMA_SERVER_MODEL":               locale.EnvDesc_OLLAMA_SERVER_MODEL,
		"OLLAMA_SERVER_CONFIG_PATH":         locale.EnvDesc_OLLAMA_SERVER_CONFIG_PATH,
		"OLLAMA_SERVER_PULL_MODELS_TIMEOUT": locale.EnvDesc_OLLAMA_SERVER_PULL_MODELS_TIMEOUT,
		"OLLAMA_SERVER_PULL_MODELS_ENABLED": locale.EnvDesc_OLLAMA_SERVER_PULL_MODELS_ENABLED,
		"OLLAMA_SERVER_LOAD_MODELS_ENABLED": locale.EnvDesc_OLLAMA_SERVER_LOAD_MODELS_ENABLED,
		"LLM_SERVER_URL":                    locale.EnvDesc_LLM_SERVER_URL,
		"LLM_SERVER_KEY":                    locale.EnvDesc_LLM_SERVER_KEY,
		"LLM_SERVER_MODEL":                  locale.EnvDesc_LLM_SERVER_MODEL,
		"LLM_SERVER_CONFIG_PATH":            locale.EnvDesc_LLM_SERVER_CONFIG_PATH,
		"LLM_SERVER_LEGACY_REASONING":       locale.EnvDesc_LLM_SERVER_LEGACY_REASONING,
		"LLM_SERVER_PRESERVE_REASONING":     locale.EnvDesc_LLM_SERVER_PRESERVE_REASONING,
		"LLM_SERVER_PROVIDER":               locale.EnvDesc_LLM_SERVER_PROVIDER,

		"LANGFUSE_LISTEN_IP":   locale.EnvDesc_LANGFUSE_LISTEN_IP,
		"LANGFUSE_LISTEN_PORT": locale.EnvDesc_LANGFUSE_LISTEN_PORT,
		"LANGFUSE_BASE_URL":    locale.EnvDesc_LANGFUSE_BASE_URL,
		"LANGFUSE_PROJECT_ID":  locale.EnvDesc_LANGFUSE_PROJECT_ID,
		"LANGFUSE_PUBLIC_KEY":  locale.EnvDesc_LANGFUSE_PUBLIC_KEY,
		"LANGFUSE_SECRET_KEY":  locale.EnvDesc_LANGFUSE_SECRET_KEY,

		// langfuse init variables
		"LANGFUSE_INIT_PROJECT_ID":         locale.EnvDesc_LANGFUSE_INIT_PROJECT_ID,
		"LANGFUSE_INIT_PROJECT_PUBLIC_KEY": locale.EnvDesc_LANGFUSE_INIT_PROJECT_PUBLIC_KEY,
		"LANGFUSE_INIT_PROJECT_SECRET_KEY": locale.EnvDesc_LANGFUSE_INIT_PROJECT_SECRET_KEY,
		"LANGFUSE_INIT_USER_EMAIL":         locale.EnvDesc_LANGFUSE_INIT_USER_EMAIL,
		"LANGFUSE_INIT_USER_NAME":          locale.EnvDesc_LANGFUSE_INIT_USER_NAME,
		"LANGFUSE_INIT_USER_PASSWORD":      locale.EnvDesc_LANGFUSE_INIT_USER_PASSWORD,

		"LANGFUSE_OTEL_EXPORTER_OTLP_ENDPOINT": locale.EnvDesc_LANGFUSE_OTEL_EXPORTER_OTLP_ENDPOINT,

		"GRAFANA_LISTEN_IP":     locale.EnvDesc_GRAFANA_LISTEN_IP,
		"GRAFANA_LISTEN_PORT":   locale.EnvDesc_GRAFANA_LISTEN_PORT,
		"OTEL_GRPC_LISTEN_IP":   locale.EnvDesc_OTEL_GRPC_LISTEN_IP,
		"OTEL_GRPC_LISTEN_PORT": locale.EnvDesc_OTEL_GRPC_LISTEN_PORT,
		"OTEL_HTTP_LISTEN_IP":   locale.EnvDesc_OTEL_HTTP_LISTEN_IP,
		"OTEL_HTTP_LISTEN_PORT": locale.EnvDesc_OTEL_HTTP_LISTEN_PORT,
		"OTEL_HOST":             locale.EnvDesc_OTEL_HOST,

		"SUMMARIZER_PRESERVE_LAST":       locale.EnvDesc_SUMMARIZER_PRESERVE_LAST,
		"SUMMARIZER_USE_QA":              locale.EnvDesc_SUMMARIZER_USE_QA,
		"SUMMARIZER_SUM_MSG_HUMAN_IN_QA": locale.EnvDesc_SUMMARIZER_SUM_MSG_HUMAN_IN_QA,
		"SUMMARIZER_LAST_SEC_BYTES":      locale.EnvDesc_SUMMARIZER_LAST_SEC_BYTES,
		"SUMMARIZER_MAX_BP_BYTES":        locale.EnvDesc_SUMMARIZER_MAX_BP_BYTES,
		"SUMMARIZER_MAX_QA_BYTES":        locale.EnvDesc_SUMMARIZER_MAX_QA_BYTES,
		"SUMMARIZER_MAX_QA_SECTIONS":     locale.EnvDesc_SUMMARIZER_MAX_QA_SECTIONS,
		"SUMMARIZER_KEEP_QA_SECTIONS":    locale.EnvDesc_SUMMARIZER_KEEP_QA_SECTIONS,

		"ASSISTANT_SUMMARIZER_PRESERVE_LAST":    locale.EnvDesc_ASSISTANT_SUMMARIZER_PRESERVE_LAST,
		"ASSISTANT_SUMMARIZER_LAST_SEC_BYTES":   locale.EnvDesc_ASSISTANT_SUMMARIZER_LAST_SEC_BYTES,
		"ASSISTANT_SUMMARIZER_MAX_BP_BYTES":     locale.EnvDesc_ASSISTANT_SUMMARIZER_MAX_BP_BYTES,
		"ASSISTANT_SUMMARIZER_MAX_QA_BYTES":     locale.EnvDesc_ASSISTANT_SUMMARIZER_MAX_QA_BYTES,
		"ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS":  locale.EnvDesc_ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS,
		"ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS": locale.EnvDesc_ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS,

		"EMBEDDING_PROVIDER":        locale.EnvDesc_EMBEDDING_PROVIDER,
		"EMBEDDING_URL":             locale.EnvDesc_EMBEDDING_URL,
		"EMBEDDING_KEY":             locale.EnvDesc_EMBEDDING_KEY,
		"EMBEDDING_MODEL":           locale.EnvDesc_EMBEDDING_MODEL,
		"EMBEDDING_BATCH_SIZE":      locale.EnvDesc_EMBEDDING_BATCH_SIZE,
		"EMBEDDING_STRIP_NEW_LINES": locale.EnvDesc_EMBEDDING_STRIP_NEW_LINES,

		"ASK_USER": locale.EnvDesc_ASK_USER,

		"ASSISTANT_USE_AGENTS": locale.EnvDesc_ASSISTANT_USE_AGENTS,

		"SCRAPER_PUBLIC_URL":                    locale.EnvDesc_SCRAPER_PUBLIC_URL,
		"SCRAPER_PRIVATE_URL":                   locale.EnvDesc_SCRAPER_PRIVATE_URL,
		"LOCAL_SCRAPER_USERNAME":                locale.EnvDesc_LOCAL_SCRAPER_USERNAME,
		"LOCAL_SCRAPER_PASSWORD":                locale.EnvDesc_LOCAL_SCRAPER_PASSWORD,
		"LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS": locale.EnvDesc_LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS,

		"DUCKDUCKGO_ENABLED": locale.EnvDesc_DUCKDUCKGO_ENABLED,
		"SPLOITUS_ENABLED":   locale.EnvDesc_SPLOITUS_ENABLED,
		"PERPLEXITY_API_KEY": locale.EnvDesc_PERPLEXITY_API_KEY,
		"TAVILY_API_KEY":     locale.EnvDesc_TAVILY_API_KEY,
		"TRAVERSAAL_API_KEY": locale.EnvDesc_TRAVERSAAL_API_KEY,
		"GOOGLE_API_KEY":     locale.EnvDesc_GOOGLE_API_KEY,
		"GOOGLE_CX_KEY":      locale.EnvDesc_GOOGLE_CX_KEY,
		"GOOGLE_LR_KEY":      locale.EnvDesc_GOOGLE_LR_KEY,

		"PERPLEXITY_MODEL":        locale.EnvDesc_PERPLEXITY_MODEL,
		"PERPLEXITY_CONTEXT_SIZE": locale.EnvDesc_PERPLEXITY_CONTEXT_SIZE,

		"SEARXNG_URL":        locale.EnvDesc_SEARXNG_URL,
		"SEARXNG_CATEGORIES": locale.EnvDesc_SEARXNG_CATEGORIES,
		"SEARXNG_LANGUAGE":   locale.EnvDesc_SEARXNG_LANGUAGE,
		"SEARXNG_SAFESEARCH": locale.EnvDesc_SEARXNG_SAFESEARCH,
		"SEARXNG_TIME_RANGE": locale.EnvDesc_SEARXNG_TIME_RANGE,

		"DOCKER_INSIDE":                    locale.EnvDesc_DOCKER_INSIDE,
		"DOCKER_NET_ADMIN":                 locale.EnvDesc_DOCKER_NET_ADMIN,
		"DOCKER_SOCKET":                    locale.EnvDesc_DOCKER_SOCKET,
		"DOCKER_NETWORK":                   locale.EnvDesc_DOCKER_NETWORK,
		"DOCKER_PUBLIC_IP":                 locale.EnvDesc_DOCKER_PUBLIC_IP,
		"DOCKER_WORK_DIR":                  locale.EnvDesc_DOCKER_WORK_DIR,
		"DOCKER_DEFAULT_IMAGE":             locale.EnvDesc_DOCKER_DEFAULT_IMAGE,
		"DOCKER_DEFAULT_IMAGE_FOR_PENTEST": locale.EnvDesc_DOCKER_DEFAULT_IMAGE_FOR_PENTEST,
		"DOCKER_HOST":                      locale.EnvDesc_DOCKER_HOST,
		"DOCKER_TLS_VERIFY":                locale.EnvDesc_DOCKER_TLS_VERIFY,
		"DOCKER_CERT_PATH":                 locale.EnvDesc_DOCKER_CERT_PATH,

		"LICENSE_KEY":                       locale.EnvDesc_LICENSE_KEY,
		"PENTAGI_LISTEN_IP":                 locale.EnvDesc_PENTAGI_LISTEN_IP,
		"PENTAGI_LISTEN_PORT":               locale.EnvDesc_PENTAGI_LISTEN_PORT,
		"PUBLIC_URL":                        locale.EnvDesc_PUBLIC_URL,
		"CORS_ORIGINS":                      locale.EnvDesc_CORS_ORIGINS,
		"COOKIE_SIGNING_SALT":               locale.EnvDesc_COOKIE_SIGNING_SALT,
		"PROXY_URL":                         locale.EnvDesc_PROXY_URL,
		"EXTERNAL_SSL_CA_PATH":              locale.EnvDesc_EXTERNAL_SSL_CA_PATH,
		"EXTERNAL_SSL_INSECURE":             locale.EnvDesc_EXTERNAL_SSL_INSECURE,
		"PENTAGI_SSL_DIR":                   locale.EnvDesc_PENTAGI_SSL_DIR,
		"PENTAGI_DATA_DIR":                  locale.EnvDesc_PENTAGI_DATA_DIR,
		"PENTAGI_DOCKER_SOCKET":             locale.EnvDesc_PENTAGI_DOCKER_SOCKET,
		"PENTAGI_DOCKER_CERT_PATH":          locale.EnvDesc_PENTAGI_DOCKER_CERT_PATH,
		"PENTAGI_LLM_SERVER_CONFIG_PATH":    locale.EnvDesc_PENTAGI_LLM_SERVER_CONFIG_PATH,
		"PENTAGI_OLLAMA_SERVER_CONFIG_PATH": locale.EnvDesc_PENTAGI_OLLAMA_SERVER_CONFIG_PATH,

		"STATIC_DIR":     locale.EnvDesc_STATIC_DIR,
		"STATIC_URL":     locale.EnvDesc_STATIC_URL,
		"SERVER_PORT":    locale.EnvDesc_SERVER_PORT,
		"SERVER_HOST":    locale.EnvDesc_SERVER_HOST,
		"SERVER_SSL_CRT": locale.EnvDesc_SERVER_SSL_CRT,
		"SERVER_SSL_KEY": locale.EnvDesc_SERVER_SSL_KEY,
		"SERVER_USE_SSL": locale.EnvDesc_SERVER_USE_SSL,

		"OAUTH_GOOGLE_CLIENT_ID":     locale.EnvDesc_OAUTH_GOOGLE_CLIENT_ID,
		"OAUTH_GOOGLE_CLIENT_SECRET": locale.EnvDesc_OAUTH_GOOGLE_CLIENT_SECRET,
		"OAUTH_GITHUB_CLIENT_ID":     locale.EnvDesc_OAUTH_GITHUB_CLIENT_ID,
		"OAUTH_GITHUB_CLIENT_SECRET": locale.EnvDesc_OAUTH_GITHUB_CLIENT_SECRET,

		"LANGFUSE_EE_LICENSE_KEY": locale.EnvDesc_LANGFUSE_EE_LICENSE_KEY,

		"GRAPHITI_URL":        locale.EnvDesc_GRAPHITI_URL,
		"GRAPHITI_TIMEOUT":    locale.EnvDesc_GRAPHITI_TIMEOUT,
		"GRAPHITI_MODEL_NAME": locale.EnvDesc_GRAPHITI_MODEL_NAME,
		"NEO4J_USER":          locale.EnvDesc_NEO4J_USER,
		"NEO4J_DATABASE":      locale.EnvDesc_NEO4J_DATABASE,

		"PENTAGI_POSTGRES_PASSWORD": locale.EnvDesc_PENTAGI_POSTGRES_PASSWORD,
		"NEO4J_PASSWORD":            locale.EnvDesc_NEO4J_PASSWORD,
	}
	if desc, ok := envVarDescriptions[varName]; ok {
		return desc
	}
	return varName
}

// maskedVariables contains environment variable names that should be masked in display
var maskedVariables = map[string]bool{
	// API keys and Secrets
	"OPEN_AI_KEY":               true,
	"ANTHROPIC_API_KEY":         true,
	"GEMINI_API_KEY":            true,
	"BEDROCK_ACCESS_KEY_ID":     true,
	"BEDROCK_SECRET_ACCESS_KEY": true,
	"BEDROCK_SESSION_TOKEN":     true,
	"LLM_SERVER_KEY":            true,
	"LANGFUSE_PUBLIC_KEY":       true,
	"LANGFUSE_SECRET_KEY":       true,
	"EMBEDDING_KEY":             true,
	"LOCAL_SCRAPER_PASSWORD":    true,
	"PERPLEXITY_API_KEY":        true,
	"TAVILY_API_KEY":            true,
	"TRAVERSAAL_API_KEY":        true,
	"GOOGLE_API_KEY":            true,
	"GOOGLE_CX_KEY":             true,

	// oauth client secrets
	"OAUTH_GOOGLE_CLIENT_SECRET": true,
	"OAUTH_GITHUB_CLIENT_SECRET": true,

	// urls can embed credentials; mask to avoid leaking secrets
	"PROXY_URL":           true,
	"SCRAPER_PUBLIC_URL":  true,
	"SCRAPER_PRIVATE_URL": true,

	// langfuse init secrets
	"LANGFUSE_INIT_PROJECT_PUBLIC_KEY": true,
	"LANGFUSE_INIT_PROJECT_SECRET_KEY": true,
	"LANGFUSE_INIT_USER_PASSWORD":      true,

	// langfuse license key
	"LANGFUSE_EE_LICENSE_KEY": true,

	// postgres password for pentagi service (pgvector binds on localhost)
	"PENTAGI_POSTGRES_PASSWORD": true,

	// neo4j password for graphiti service (neo4j binds on localhost)
	"NEO4J_PASSWORD": true,

	// langfuse stack secrets (compose-managed)
	"LANGFUSE_SALT":                      true,
	"LANGFUSE_ENCRYPTION_KEY":            true,
	"LANGFUSE_NEXTAUTH_SECRET":           true,
	"LANGFUSE_CLICKHOUSE_PASSWORD":       true,
	"LANGFUSE_S3_ACCESS_KEY_ID":          true,
	"LANGFUSE_S3_SECRET_ACCESS_KEY":      true,
	"LANGFUSE_REDIS_AUTH":                true,
	"LANGFUSE_AUTH_CUSTOM_CLIENT_SECRET": true,

	// server settings
	"COOKIE_SIGNING_SALT": true,
}

// isVariableMasked returns true if the variable should be masked in display
func (c *controller) isVariableMasked(varName string) bool {
	return maskedVariables[varName]
}

// criticalVariables contains environment variable names that require service restart
var criticalVariables = map[string]bool{
	// LLM Provider changes
	"OPEN_AI_KEY":                       true,
	"OPEN_AI_SERVER_URL":                true,
	"ANTHROPIC_API_KEY":                 true,
	"ANTHROPIC_SERVER_URL":              true,
	"GEMINI_API_KEY":                    true,
	"GEMINI_SERVER_URL":                 true,
	"BEDROCK_ACCESS_KEY_ID":             true,
	"BEDROCK_SECRET_ACCESS_KEY":         true,
	"BEDROCK_SESSION_TOKEN":             true,
	"BEDROCK_REGION":                    true,
	"OLLAMA_SERVER_URL":                 true,
	"OLLAMA_SERVER_MODEL":               true,
	"OLLAMA_SERVER_CONFIG_PATH":         true,
	"OLLAMA_SERVER_PULL_MODELS_TIMEOUT": true,
	"OLLAMA_SERVER_PULL_MODELS_ENABLED": true,
	"OLLAMA_SERVER_LOAD_MODELS_ENABLED": true,
	"LLM_SERVER_URL":                    true,
	"LLM_SERVER_KEY":                    true,
	"LLM_SERVER_MODEL":                  true,
	"LLM_SERVER_CONFIG_PATH":            true,
	"LLM_SERVER_LEGACY_REASONING":       true,
	"LLM_SERVER_PRESERVE_REASONING":     true,
	"LLM_SERVER_PROVIDER":               true,

	// tools changes
	"DUCKDUCKGO_ENABLED":      true,
	"SPLOITUS_ENABLED":        true,
	"PERPLEXITY_API_KEY":      true,
	"PERPLEXITY_MODEL":        true,
	"PERPLEXITY_CONTEXT_SIZE": true,
	"TAVILY_API_KEY":          true,
	"TRAVERSAAL_API_KEY":      true,
	"GOOGLE_API_KEY":          true,
	"GOOGLE_CX_KEY":           true,
	"GOOGLE_LR_KEY":           true,
	"SEARXNG_URL":             true,
	"SEARXNG_CATEGORIES":      true,
	"SEARXNG_LANGUAGE":        true,
	"SEARXNG_SAFESEARCH":      true,
	"SEARXNG_TIME_RANGE":      true,

	// mounting custom LLM server config into pentagi container changes volume mapping
	"PENTAGI_LLM_SERVER_CONFIG_PATH":    true,
	"PENTAGI_OLLAMA_SERVER_CONFIG_PATH": true,

	// Embedding provider changes
	"EMBEDDING_PROVIDER":        true,
	"EMBEDDING_URL":             true,
	"EMBEDDING_KEY":             true,
	"EMBEDDING_MODEL":           true,
	"EMBEDDING_BATCH_SIZE":      true,
	"EMBEDDING_STRIP_NEW_LINES": true,

	// Docker configuration changes
	"DOCKER_INSIDE":                    true,
	"DOCKER_NET_ADMIN":                 true,
	"DOCKER_SOCKET":                    true,
	"DOCKER_NETWORK":                   true,
	"DOCKER_PUBLIC_IP":                 true,
	"DOCKER_DEFAULT_IMAGE":             true,
	"DOCKER_DEFAULT_IMAGE_FOR_PENTEST": true,
	"DOCKER_HOST":                      true,
	"DOCKER_TLS_VERIFY":                true,
	"DOCKER_CERT_PATH":                 true,
	"PENTAGI_DOCKER_SOCKET":            true,

	// observability changes
	"OTEL_HOST": true,

	// graphiti changes
	"GRAPHITI_URL":        true,
	"GRAPHITI_TIMEOUT":    true,
	"GRAPHITI_MODEL_NAME": true,

	// server settings changes
	"ASK_USER":              true,
	"LICENSE_KEY":           true,
	"PENTAGI_LISTEN_IP":     true,
	"PENTAGI_LISTEN_PORT":   true,
	"PUBLIC_URL":            true,
	"CORS_ORIGINS":          true,
	"COOKIE_SIGNING_SALT":   true,
	"PROXY_URL":             true,
	"EXTERNAL_SSL_CA_PATH":  true,
	"EXTERNAL_SSL_INSECURE": true,
	"STATIC_DIR":            true,
	"STATIC_URL":            true,
	"SERVER_PORT":           true,
	"SERVER_HOST":           true,
	"SERVER_SSL_CRT":        true,
	"SERVER_SSL_KEY":        true,
	"SERVER_USE_SSL":        true,
	"PENTAGI_SSL_DIR":       true,
	"PENTAGI_DATA_DIR":      true,

	// scraper settings
	"SCRAPER_PUBLIC_URL":  true,
	"SCRAPER_PRIVATE_URL": true,

	// oauth settings
	"OAUTH_GOOGLE_CLIENT_ID":     true,
	"OAUTH_GOOGLE_CLIENT_SECRET": true,
	"OAUTH_GITHUB_CLIENT_ID":     true,
	"OAUTH_GITHUB_CLIENT_SECRET": true,

	// langfuse integration settings passed to pentagi
	"LANGFUSE_BASE_URL":   true,
	"LANGFUSE_PROJECT_ID": true,
	"LANGFUSE_PUBLIC_KEY": true,
	"LANGFUSE_SECRET_KEY": true,

	// summarizer settings (general)
	"SUMMARIZER_PRESERVE_LAST":       true,
	"SUMMARIZER_USE_QA":              true,
	"SUMMARIZER_SUM_MSG_HUMAN_IN_QA": true,
	"SUMMARIZER_LAST_SEC_BYTES":      true,
	"SUMMARIZER_MAX_BP_BYTES":        true,
	"SUMMARIZER_MAX_QA_SECTIONS":     true,
	"SUMMARIZER_MAX_QA_BYTES":        true,
	"SUMMARIZER_KEEP_QA_SECTIONS":    true,

	// assistant-level settings
	"ASSISTANT_USE_AGENTS":                  true,
	"ASSISTANT_SUMMARIZER_PRESERVE_LAST":    true,
	"ASSISTANT_SUMMARIZER_LAST_SEC_BYTES":   true,
	"ASSISTANT_SUMMARIZER_MAX_BP_BYTES":     true,
	"ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS":  true,
	"ASSISTANT_SUMMARIZER_MAX_QA_BYTES":     true,
	"ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS": true,
}

// isCriticalVariable returns true if changing this variable requires service restart
func (c *controller) isCriticalVariable(varName string) bool {
	return criticalVariables[varName]
}

// RemoveCredentialsFromURL removes credentials from URL - public method for form display
func RemoveCredentialsFromURL(urlStr string) string {
	if urlStr == "" {
		return urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	parsedURL.User = nil

	return parsedURL.String()
}
