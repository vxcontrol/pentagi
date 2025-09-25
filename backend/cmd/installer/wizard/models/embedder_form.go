package models

import (
	"fmt"
	"strconv"
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EmbeddingProviderInfo contains information about an embedding provider
type EmbeddingProviderInfo struct {
	ID                string
	Name              string
	Description       string
	URLPlaceholder    string
	APIKeyPlaceholder string
	ModelPlaceholder  string
	RequiresAPIKey    bool
	SupportsURL       bool
	SupportsModel     bool
	HelpText          string
}

// EmbedderFormModel represents the Embedder configuration form
type EmbedderFormModel struct {
	*BaseScreen

	// screen-specific components
	providerList     list.Model
	providerDelegate *BaseListDelegate

	// provider information
	providers map[string]*EmbeddingProviderInfo
}

// NewEmbedderFormModel creates a new Embedder form model
func NewEmbedderFormModel(c controller.Controller, s styles.Styles, w window.Window) *EmbedderFormModel {
	m := &EmbedderFormModel{
		providers: initEmbeddingProviders(),
	}

	m.BaseScreen = NewBaseScreen(c, s, w, m, m)
	m.initializeProviderList(s)

	return m
}

// initEmbeddingProviders initializes the provider information
func initEmbeddingProviders() map[string]*EmbeddingProviderInfo {
	return map[string]*EmbeddingProviderInfo{
		locale.EmbedderProviderIDDefault: {
			ID:                locale.EmbedderProviderIDDefault,
			Name:              locale.EmbedderProviderDefault,
			Description:       locale.EmbedderProviderDefaultDesc,
			URLPlaceholder:    "",
			APIKeyPlaceholder: "",
			ModelPlaceholder:  "",
			RequiresAPIKey:    false,
			SupportsURL:       false,
			SupportsModel:     false,
			HelpText:          locale.EmbedderHelpDefault,
		},
		locale.EmbedderProviderIDOpenAI: {
			ID:                locale.EmbedderProviderIDOpenAI,
			Name:              locale.EmbedderProviderOpenAI,
			Description:       locale.EmbedderProviderOpenAIDesc,
			URLPlaceholder:    locale.EmbedderURLPlaceholderOpenAI,
			APIKeyPlaceholder: locale.EmbedderAPIKeyPlaceholderDefault,
			ModelPlaceholder:  locale.EmbedderModelPlaceholderOpenAI,
			RequiresAPIKey:    true,
			SupportsURL:       true,
			SupportsModel:     true,
			HelpText:          locale.EmbedderHelpOpenAI,
		},
		locale.EmbedderProviderIDOllama: {
			ID:                locale.EmbedderProviderIDOllama,
			Name:              locale.EmbedderProviderOllama,
			Description:       locale.EmbedderProviderOllamaDesc,
			URLPlaceholder:    locale.EmbedderURLPlaceholderOllama,
			APIKeyPlaceholder: locale.EmbedderAPIKeyPlaceholderOllama,
			ModelPlaceholder:  locale.EmbedderModelPlaceholderOllama,
			RequiresAPIKey:    false,
			SupportsURL:       true,
			SupportsModel:     true,
			HelpText:          locale.EmbedderHelpOllama,
		},
		locale.EmbedderProviderIDMistral: {
			ID:                locale.EmbedderProviderIDMistral,
			Name:              locale.EmbedderProviderMistral,
			Description:       locale.EmbedderProviderMistralDesc,
			URLPlaceholder:    locale.EmbedderURLPlaceholderMistral,
			APIKeyPlaceholder: locale.EmbedderAPIKeyPlaceholderMistral,
			ModelPlaceholder:  locale.EmbedderModelPlaceholderMistral,
			RequiresAPIKey:    true,
			SupportsURL:       true,
			SupportsModel:     false,
			HelpText:          locale.EmbedderHelpMistral,
		},
		locale.EmbedderProviderIDJina: {
			ID:                locale.EmbedderProviderIDJina,
			Name:              locale.EmbedderProviderJina,
			Description:       locale.EmbedderProviderJinaDesc,
			URLPlaceholder:    locale.EmbedderURLPlaceholderJina,
			APIKeyPlaceholder: locale.EmbedderAPIKeyPlaceholderJina,
			ModelPlaceholder:  locale.EmbedderModelPlaceholderJina,
			RequiresAPIKey:    true,
			SupportsURL:       true,
			SupportsModel:     true,
			HelpText:          locale.EmbedderHelpJina,
		},
		locale.EmbedderProviderIDHuggingFace: {
			ID:                locale.EmbedderProviderIDHuggingFace,
			Name:              locale.EmbedderProviderHuggingFace,
			Description:       locale.EmbedderProviderHuggingFaceDesc,
			URLPlaceholder:    locale.EmbedderURLPlaceholderHuggingFace,
			APIKeyPlaceholder: locale.EmbedderAPIKeyPlaceholderHuggingFace,
			ModelPlaceholder:  locale.EmbedderModelPlaceholderHuggingFace,
			RequiresAPIKey:    true,
			SupportsURL:       true,
			SupportsModel:     true,
			HelpText:          locale.EmbedderHelpHuggingFace,
		},
		locale.EmbedderProviderIDGoogleAI: {
			ID:                locale.EmbedderProviderIDGoogleAI,
			Name:              locale.EmbedderProviderGoogleAI,
			Description:       locale.EmbedderProviderGoogleAIDesc,
			URLPlaceholder:    locale.EmbedderURLPlaceholderGoogleAI,
			APIKeyPlaceholder: locale.EmbedderAPIKeyPlaceholderGoogleAI,
			ModelPlaceholder:  locale.EmbedderModelPlaceholderGoogleAI,
			RequiresAPIKey:    true,
			SupportsURL:       false,
			SupportsModel:     true,
			HelpText:          locale.EmbedderHelpGoogleAI,
		},
		locale.EmbedderProviderIDVoyageAI: {
			ID:                locale.EmbedderProviderIDVoyageAI,
			Name:              locale.EmbedderProviderVoyageAI,
			Description:       locale.EmbedderProviderVoyageAIDesc,
			URLPlaceholder:    locale.EmbedderURLPlaceholderVoyageAI,
			APIKeyPlaceholder: locale.EmbedderAPIKeyPlaceholderVoyageAI,
			ModelPlaceholder:  locale.EmbedderModelPlaceholderVoyageAI,
			RequiresAPIKey:    true,
			SupportsURL:       false,
			SupportsModel:     true,
			HelpText:          locale.EmbedderHelpVoyageAI,
		},
		locale.EmbedderProviderIDDisabled: {
			ID:                locale.EmbedderProviderIDDisabled,
			Name:              locale.EmbedderProviderDisabled,
			Description:       locale.EmbedderProviderDisabledDesc,
			URLPlaceholder:    "",
			APIKeyPlaceholder: "",
			ModelPlaceholder:  "",
			RequiresAPIKey:    false,
			SupportsURL:       false,
			SupportsModel:     false,
			HelpText:          locale.EmbedderHelpDisabled,
		},
	}
}

// initializeProviderList sets up the provider selection list
func (m *EmbedderFormModel) initializeProviderList(styles styles.Styles) {
	options := []BaseListOption{
		{Value: locale.EmbedderProviderIDDefault, Display: locale.EmbedderProviderDefault},
		{Value: locale.EmbedderProviderIDOpenAI, Display: locale.EmbedderProviderOpenAI},
		{Value: locale.EmbedderProviderIDOllama, Display: locale.EmbedderProviderOllama},
		{Value: locale.EmbedderProviderIDMistral, Display: locale.EmbedderProviderMistral},
		{Value: locale.EmbedderProviderIDJina, Display: locale.EmbedderProviderJina},
		{Value: locale.EmbedderProviderIDHuggingFace, Display: locale.EmbedderProviderHuggingFace},
		{Value: locale.EmbedderProviderIDGoogleAI, Display: locale.EmbedderProviderGoogleAI},
		{Value: locale.EmbedderProviderIDVoyageAI, Display: locale.EmbedderProviderVoyageAI},
		{Value: locale.EmbedderProviderIDDisabled, Display: locale.EmbedderProviderDisabled},
	}

	m.providerDelegate = NewBaseListDelegate(
		styles.FormLabel.Align(lipgloss.Center),
		MinMenuWidth-6,
	)

	m.providerList = m.GetListHelper().CreateList(options, m.providerDelegate, MinMenuWidth-6, 3)

	// set current selection
	config := m.GetController().GetEmbedderConfig()
	selectedProvider := m.getProviderID(config.Provider.Value)
	m.GetListHelper().SelectByValue(&m.providerList, selectedProvider)
}

// getProviderID converts provider value to ID
func (m *EmbedderFormModel) getProviderID(provider string) string {
	switch provider {
	case "":
		return locale.EmbedderProviderIDDefault
	case locale.EmbedderProviderIDDisabled:
		return locale.EmbedderProviderIDDisabled
	default:
		// check if it's a known provider
		if _, exists := m.providers[provider]; exists {
			return provider
		}
		// fallback to default for unknown providers
		return locale.EmbedderProviderIDDefault
	}
}

// getSelectedProvider returns the currently selected provider ID
func (m *EmbedderFormModel) getSelectedProvider() string {
	selectedValue := m.GetListHelper().GetSelectedValue(&m.providerList)
	if selectedValue == "" {
		return locale.EmbedderProviderIDDefault
	}
	return selectedValue
}

// getCurrentProviderInfo returns information about the currently selected provider
func (m *EmbedderFormModel) getCurrentProviderInfo() *EmbeddingProviderInfo {
	providerID := m.getSelectedProvider()
	if info, exists := m.providers[providerID]; exists {
		return info
	}
	return m.providers[locale.EmbedderProviderIDDefault]
}

// BaseScreenHandler interface implementation

func (m *EmbedderFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetEmbedderConfig()
	fields := []FormField{}
	providerInfo := m.getCurrentProviderInfo()

	// URL field (if supported)
	if providerInfo.SupportsURL {
		fields = append(fields, m.createURLField(config, providerInfo))
	}

	// API Key field (if required)
	if providerInfo.RequiresAPIKey {
		fields = append(fields, m.createAPIKeyField(config, providerInfo))
	}

	// Model field (if supported)
	if providerInfo.SupportsModel {
		fields = append(fields, m.createModelField(config, providerInfo))
	}

	// Batch size field (always show except for disabled)
	if providerInfo.ID != locale.EmbedderProviderIDDisabled {
		fields = append(fields, m.createBatchSizeField(config))
	}

	// Strip newlines field (always show except for disabled)
	if providerInfo.ID != locale.EmbedderProviderIDDisabled {
		fields = append(fields, m.createStripNewLinesField(config))
	}

	m.SetFormFields(fields)
	return nil
}

func (m *EmbedderFormModel) createURLField(
	config *controller.EmbedderConfig, providerInfo *EmbeddingProviderInfo,
) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.URL)
	if providerInfo.URLPlaceholder != "" {
		input.Placeholder = providerInfo.URLPlaceholder
	}

	return FormField{
		Key:         "url",
		Title:       locale.EmbedderFormURL,
		Description: locale.EmbedderFormURLDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *EmbedderFormModel) createAPIKeyField(
	config *controller.EmbedderConfig, providerInfo *EmbeddingProviderInfo,
) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.APIKey)
	if providerInfo.APIKeyPlaceholder != "" {
		input.Placeholder = providerInfo.APIKeyPlaceholder
	}

	return FormField{
		Key:         "api_key",
		Title:       locale.EmbedderFormAPIKey,
		Description: locale.EmbedderFormAPIKeyDesc,
		Required:    false,
		Masked:      true,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *EmbedderFormModel) createModelField(
	config *controller.EmbedderConfig, providerInfo *EmbeddingProviderInfo,
) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.Model)
	if providerInfo.ModelPlaceholder != "" {
		input.Placeholder = providerInfo.ModelPlaceholder
	}

	return FormField{
		Key:         "model",
		Title:       locale.EmbedderFormModel,
		Description: locale.EmbedderFormModelDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *EmbedderFormModel) createBatchSizeField(config *controller.EmbedderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.BatchSize)

	return FormField{
		Key:         "batch_size",
		Title:       locale.EmbedderFormBatchSize,
		Description: locale.EmbedderFormBatchSizeDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *EmbedderFormModel) createStripNewLinesField(config *controller.EmbedderConfig) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), config.StripNewLines)

	return FormField{
		Key:         "strip_newlines",
		Title:       locale.EmbedderFormStripNewLines,
		Description: locale.EmbedderFormStripNewLinesDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

func (m *EmbedderFormModel) GetFormTitle() string {
	return locale.EmbedderFormTitle
}

func (m *EmbedderFormModel) GetFormDescription() string {
	return locale.EmbedderFormDescription
}

func (m *EmbedderFormModel) GetFormName() string {
	return locale.EmbedderFormName
}

func (m *EmbedderFormModel) GetFormSummary() string {
	return ""
}

func (m *EmbedderFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.EmbedderFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.EmbedderFormDescription))
	sections = append(sections, "")
	sections = append(sections, locale.EmbedderFormOverview)

	return strings.Join(sections, "\n")
}

func (m *EmbedderFormModel) GetCurrentConfiguration() string {
	var sections []string

	config := m.GetController().GetEmbedderConfig()

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	providerID := m.getProviderID(config.Provider.Value)
	providerInfo := m.providers[providerID]

	if config.Configured {
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.EmbedderFormProvider, m.GetStyles().Success.Render(providerInfo.Name)))
	} else {
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.EmbedderFormProvider, m.GetStyles().Warning.Render(providerInfo.Name+
				" ("+locale.StatusNotConfigured+")")))
	}

	if config.URL.Value != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.EmbedderFormURL, m.GetStyles().Info.Render(config.URL.Value)))
	}

	if config.APIKey.Value != "" {
		maskedKey := strings.Repeat("*", len(config.APIKey.Value))
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.EmbedderFormAPIKey, m.GetStyles().Muted.Render(maskedKey)))
	}

	if config.Model.Value != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.EmbedderFormModel, m.GetStyles().Info.Render(config.Model.Value)))
	}

	if config.BatchSize.Value != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.EmbedderFormBatchSize, m.GetStyles().Info.Render(config.BatchSize.Value)))
	} else if config.BatchSize.Default != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.EmbedderFormBatchSize, m.GetStyles().Info.Render(config.BatchSize.Default)))
	}

	stripNewLines := config.StripNewLines.Value
	if stripNewLines == "" {
		stripNewLines = config.StripNewLines.Default
	}
	if stripNewLines != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.EmbedderFormStripNewLines, m.GetStyles().Info.Render(stripNewLines)))
	}

	return strings.Join(sections, "\n")
}

func (m *EmbedderFormModel) IsConfigured() bool {
	return m.GetController().GetEmbedderConfig().Configured
}

func (m *EmbedderFormModel) GetHelpContent() string {
	var sections []string
	providerInfo := m.getCurrentProviderInfo()

	config := m.GetController().GetEmbedderConfig()

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.EmbedderFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.EmbedderFormDescription))
	sections = append(sections, "")
	if config.Configured {
		sections = append(sections, fmt.Sprintf("%s %s\n%s",
			m.GetStyles().Warning.Bold(true).Render(locale.EmbedderHelpAttentionPrefix),
			m.GetStyles().Paragraph.Render(locale.EmbedderHelpAttention),
			m.GetStyles().Warning.Render(locale.EmbedderHelpAttentionSuffix)))
		sections = append(sections, "")
	}
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.EmbedderHelpGeneral))
	sections = append(sections, "")
	sections = append(sections, providerInfo.HelpText)

	return strings.Join(sections, "\n")
}

func (m *EmbedderFormModel) HandleSave() error {
	config := m.GetController().GetEmbedderConfig()
	selectedProvider := m.getSelectedProvider()
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.EmbedderConfig{
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		Provider:      config.Provider,
		URL:           config.URL,
		APIKey:        config.APIKey,
		Model:         config.Model,
		BatchSize:     config.BatchSize,
		StripNewLines: config.StripNewLines,
	}

	// set provider
	switch selectedProvider {
	case locale.EmbedderProviderIDDefault:
		newConfig.Provider.Value = "" // empty means use default (openai)
	case locale.EmbedderProviderIDDisabled:
		newConfig.Provider.Value = locale.EmbedderProviderIDDisabled
	default:
		newConfig.Provider.Value = selectedProvider
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "url":
			newConfig.URL.Value = value
		case "api_key":
			newConfig.APIKey.Value = value
		case "model":
			newConfig.Model.Value = value
		case "batch_size":
			// validate numeric input
			if value != "" {
				if intVal, err := strconv.Atoi(value); err != nil || intVal <= 0 || intVal > 10000 {
					return fmt.Errorf("invalid batch size: %s (must be a number between 1 and 10000)", value)
				}
			}
			newConfig.BatchSize.Value = value
		case "strip_newlines":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for strip newlines: %s (must be 'true' or 'false')", value)
			}
			newConfig.StripNewLines.Value = value
		}
	}

	// save the configuration
	if err := m.GetController().UpdateEmbedderConfig(newConfig); err != nil {
		logger.Errorf("[EmbedderFormModel] SAVE: error updating embedder config: %v", err)
		return err
	}

	logger.Log("[EmbedderFormModel] SAVE: success")
	return nil
}

func (m *EmbedderFormModel) HandleReset() {
	// reset config to defaults
	config := m.GetController().ResetEmbedderConfig()

	// reset provider selection
	selectedProvider := m.getProviderID(config.Provider.Value)
	m.GetListHelper().SelectByValue(&m.providerList, selectedProvider)

	// rebuild form with reset values
	m.BuildForm()
}

func (m *EmbedderFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *EmbedderFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *EmbedderFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// BaseListHandler interface implementation

func (m *EmbedderFormModel) GetList() *list.Model {
	return &m.providerList
}

func (m *EmbedderFormModel) GetListDelegate() *BaseListDelegate {
	return m.providerDelegate
}

func (m *EmbedderFormModel) OnListSelectionChanged(oldSelection, newSelection string) {
	// rebuild form when provider changes
	m.BuildForm()
}

func (m *EmbedderFormModel) GetListTitle() string {
	return locale.EmbedderFormProvider
}

func (m *EmbedderFormModel) GetListDescription() string {
	return locale.EmbedderFormProviderDesc
}

// Update method - handle screen-specific input
func (m *EmbedderFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// handle list input first (if focused on list)
		if cmd := m.HandleListInput(msg); cmd != nil {
			return m, cmd
		}

		// then handle field input
		if cmd := m.HandleFieldInput(msg); cmd != nil {
			return m, cmd
		}
	}

	// delegate to base screen for common handling
	cmd := m.BaseScreen.Update(msg)
	return m, cmd
}

// Compile-time interface validation
var _ BaseScreenModel = (*EmbedderFormModel)(nil)
var _ BaseScreenHandler = (*EmbedderFormModel)(nil)
var _ BaseListHandler = (*EmbedderFormModel)(nil)
