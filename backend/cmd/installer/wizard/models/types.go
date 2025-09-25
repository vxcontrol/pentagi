package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	MinMenuWidth  = 38 // Minimum width for left menu panel
	MaxMenuWidth  = 88 // Maximum width for left menu panel
	MinInfoWidth  = 34 // Minimum width for right info panel
	PaddingWidth  = 8  // Total padding width for horizontal layout (left + right)
	PaddingHeight = 2  // Top padding only
)

type Registry interface {
	GetScreen(id ScreenID) BaseScreenModel
}

// RestoreModel restores the model to the BaseScreenModel interface
func RestoreModel(model tea.Model) BaseScreenModel {
	switch m := model.(type) {
	case *WelcomeModel:
		return m
	case *EULAModel:
		return m
	case *MainMenuModel:
		return m
	case *LLMProvidersModel:
		return m
	case *LLMProviderFormModel:
		return m
	case *SummarizerModel:
		return m
	case *SummarizerFormModel:
		return m
	case *MonitoringModel:
		return m
	case *LangfuseFormModel:
		return m
	case *ObservabilityFormModel:
		return m
	case *ToolsModel:
		return m
	case *AIAgentsSettingsFormModel:
		return m
	case *SearchEnginesFormModel:
		return m
	case *ScraperFormModel:
		return m
	case *DockerFormModel:
		return m
	case *EmbedderFormModel:
		return m
	case *ApplyChangesFormModel:
		return m
	case *MaintenanceModel:
		return m
	case *ProcessorOperationFormModel:
		return m
	case *ResetPasswordModel:
		return m
	case *MockFormModel:
		return m
	default:
		return nil
	}
}

// ScreenID represents unique screen identifiers for type-safe navigation
// Format: "screen" or "screen§arg1§arg2§..." for parameterized screens
type ScreenID string

const (
	// Core navigation screens
	WelcomeScreen  ScreenID = "welcome"
	EULAScreen     ScreenID = "eula"
	MainMenuScreen ScreenID = "main_menu"

	// LLM Provider screens
	LLMProvidersScreen         ScreenID = "llm_providers"
	LLMProviderOpenAIScreen    ScreenID = "llm_provider_form§openai"
	LLMProviderAnthropicScreen ScreenID = "llm_provider_form§anthropic"
	LLMProviderGeminiScreen    ScreenID = "llm_provider_form§gemini"
	LLMProviderBedrockScreen   ScreenID = "llm_provider_form§bedrock"
	LLMProviderOllamaScreen    ScreenID = "llm_provider_form§ollama"
	LLMProviderCustomScreen    ScreenID = "llm_provider_form§custom"

	// Summarizer screens
	SummarizerScreen          ScreenID = "summarizer"
	SummarizerGeneralScreen   ScreenID = "summarizer_form§general"
	SummarizerAssistantScreen ScreenID = "summarizer_form§assistant"

	// Integration screens
	MonitoringScreen     ScreenID = "monitoring"
	LangfuseScreen       ScreenID = "langfuse_form"
	ObservabilityScreen  ScreenID = "observability_form"
	EmbedderFormScreen   ScreenID = "embedder_form"
	ServerSettingsScreen ScreenID = "server_settings_form"

	// Tools screens
	ToolsScreen                ScreenID = "tools"
	AIAgentsSettingsFormScreen ScreenID = "ai_agents_settings_form"
	SearchEnginesFormScreen    ScreenID = "search_engines_form"
	ScraperFormScreen          ScreenID = "scraper_form"
	DockerFormScreen           ScreenID = "docker_form"

	// Management screens
	ApplyChangesScreen        ScreenID = "apply_changes"
	InstallPentagiScreen      ScreenID = "processor_operation_form§all§install"
	StartPentagiScreen        ScreenID = "processor_operation_form§all§start"
	StopPentagiScreen         ScreenID = "processor_operation_form§all§stop"
	RestartPentagiScreen      ScreenID = "processor_operation_form§all§restart"
	DownloadWorkerImageScreen ScreenID = "processor_operation_form§worker§download"
	UpdateWorkerImageScreen   ScreenID = "processor_operation_form§worker§update"
	UpdatePentagiScreen       ScreenID = "processor_operation_form§compose§update"
	UpdateInstallerScreen     ScreenID = "processor_operation_form§installer§update"
	FactoryResetScreen        ScreenID = "processor_operation_form§all§factory_reset"
	RemovePentagiScreen       ScreenID = "processor_operation_form§all§remove"
	PurgePentagiScreen        ScreenID = "processor_operation_form§all§purge"
	ResetPasswordScreen       ScreenID = "reset_password"
	MaintenanceScreen         ScreenID = "maintenance"
)

type LLMProviderID string

const (
	LLMProviderOpenAI    LLMProviderID = "openai"
	LLMProviderAnthropic LLMProviderID = "anthropic"
	LLMProviderGemini    LLMProviderID = "gemini"
	LLMProviderBedrock   LLMProviderID = "bedrock"
	LLMProviderOllama    LLMProviderID = "ollama"
	LLMProviderCustom    LLMProviderID = "custom"
)

// NavigationMsg represents screen navigation requests
type NavigationMsg struct {
	Target ScreenID
	GoBack bool
}

// MenuState represents main menu state and selection
type MenuState struct {
	SelectedIndex int
	Items         []MenuItem
	InfoContent   string
}

// MenuItem represents a menu item with availability and styling
type MenuItem struct {
	ID          string
	Title       string
	Description string
	Available   bool
	Enabled     bool
	Hidden      bool
}

// StatusInfo represents system status information for display
type StatusInfo struct {
	Label   string
	Value   bool
	Details string
}

// GetScreen returns the base screen identifier without arguments
func (s ScreenID) GetScreen() string {
	parts := strings.Split(string(s), "§")
	return parts[0]
}

// GetArgs returns the arguments for parameterized screens
func (s ScreenID) GetArgs() []string {
	parts := strings.Split(string(s), "§")
	if len(parts) <= 1 {
		return []string{}
	}
	return parts[1:]
}

// CreateScreenID creates a ScreenID with arguments
func CreateScreenID(screen string, args ...string) ScreenID {
	if len(args) == 0 {
		return ScreenID(screen)
	}
	parts := append([]string{screen}, args...)
	return ScreenID(strings.Join(parts, "§"))
}
