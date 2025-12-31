package models

import (
	"fmt"
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// LLMProviderFormModel represents the LLM Provider configuration form
type LLMProviderFormModel struct {
	*BaseScreen

	// screen-specific components
	providerID   LLMProviderID
	providerName string
}

// NewLLMProviderFormModel creates a new LLM Provider form model
func NewLLMProviderFormModel(
	c controller.Controller, s styles.Styles, w window.Window, pid LLMProviderID,
) *LLMProviderFormModel {
	m := &LLMProviderFormModel{
		providerID:   pid,
		providerName: c.GetLLMProviderConfig(string(pid)).Name,
	}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// BaseScreenHandler interface implementation

func (m *LLMProviderFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetLLMProviderConfig(string(m.providerID))
	fields := []FormField{}

	// Add fields based on provider type
	switch m.providerID {
	case LLMProviderOpenAI, LLMProviderAnthropic, LLMProviderGemini:
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createAPIKeyField(config))

	case LLMProviderBedrock:
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createAccessKeyField(config))
		fields = append(fields, m.createSecretKeyField(config))
		fields = append(fields, m.createSessionTokenField(config))
		fields = append(fields, m.createRegionField(config))

	case LLMProviderOllama:
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createConfigPathField(config))

	case LLMProviderCustom:
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createAPIKeyField(config))
		fields = append(fields, m.createModelField(config))
		fields = append(fields, m.createConfigPathField(config))
		fields = append(fields, m.createLegacyReasoningField(config))
	}

	m.SetFormFields(fields)
	return nil
}

func (m *LLMProviderFormModel) createBaseURLField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.BaseURL)
	input.Placeholder = m.getDefaultBaseURL()

	return FormField{
		Key:         "base_url",
		Title:       locale.LLMFormFieldBaseURL,
		Description: locale.LLMFormBaseURLDesc,
		Required:    true,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createAPIKeyField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.APIKey)

	return FormField{
		Key:         "api_key",
		Title:       locale.LLMFormFieldAPIKey,
		Description: locale.LLMFormAPIKeyDesc,
		Required:    true,
		Masked:      true,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createAccessKeyField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.AccessKey)

	return FormField{
		Key:         "access_key",
		Title:       locale.LLMFormFieldAccessKey,
		Description: locale.LLMFormAccessKeyDesc,
		Required:    true,
		Masked:      true,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createSecretKeyField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.SecretKey)

	return FormField{
		Key:         "secret_key",
		Title:       locale.LLMFormFieldSecretKey,
		Description: locale.LLMFormSecretKeyDesc,
		Required:    true,
		Masked:      true,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createSessionTokenField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.SessionToken)

	return FormField{
		Key:         "session_token",
		Title:       locale.LLMFormFieldSessionToken,
		Description: locale.LLMFormSessionTokenDesc,
		Required:    false,
		Masked:      true,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createRegionField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.Region)
	input.Placeholder = "us-east-1"

	return FormField{
		Key:         "region",
		Title:       locale.LLMFormFieldRegion,
		Description: locale.LLMFormRegionDesc,
		Required:    true,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createModelField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.Model)

	return FormField{
		Key:         "model",
		Title:       locale.LLMFormFieldModel,
		Description: locale.LLMFormModelDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createConfigPathField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.ConfigPath)
	if config.ConfigPath.Default == "" {
		input.Placeholder = "/opt/pentagi/conf/config.yml"
	}

	return FormField{
		Key:         "config_path",
		Title:       locale.LLMFormFieldConfigPath,
		Description: locale.LLMFormConfigPathDesc,
		Suggestions: config.EmbeddedLLMConfigsPath,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createLegacyReasoningField(config *controller.LLMProviderConfig) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), config.LegacyReasoning)

	return FormField{
		Key:         "legacy_reasoning",
		Title:       locale.LLMFormFieldLegacyReasoning,
		Description: locale.LLMFormLegacyReasoningDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

func (m *LLMProviderFormModel) GetFormTitle() string {
	return fmt.Sprintf(locale.LLMProviderFormTitle, m.providerName)
}

func (m *LLMProviderFormModel) GetFormDescription() string {
	switch m.providerID {
	case LLMProviderOpenAI:
		return locale.LLMProviderOpenAIDesc
	case LLMProviderAnthropic:
		return locale.LLMProviderAnthropicDesc
	case LLMProviderGemini:
		return locale.LLMProviderGeminiDesc
	case LLMProviderBedrock:
		return locale.LLMProviderBedrockDesc
	case LLMProviderOllama:
		return locale.LLMProviderOllamaDesc
	case LLMProviderCustom:
		return locale.LLMProviderCustomDesc
	default:
		return locale.LLMProviderFormDescription
	}
}

func (m *LLMProviderFormModel) GetFormName() string {
	switch m.providerID {
	case LLMProviderOpenAI:
		return locale.LLMProviderOpenAI
	case LLMProviderAnthropic:
		return locale.LLMProviderAnthropic
	case LLMProviderGemini:
		return locale.LLMProviderGemini
	case LLMProviderBedrock:
		return locale.LLMProviderBedrock
	case LLMProviderOllama:
		return locale.LLMProviderOllama
	case LLMProviderCustom:
		return locale.LLMProviderCustom
	default:
		return fmt.Sprintf(locale.LLMProviderFormName, m.providerName)
	}
}

func (m *LLMProviderFormModel) GetFormSummary() string {
	return ""
}

func (m *LLMProviderFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(fmt.Sprintf(locale.LLMProviderFormTitle, m.providerName)))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.LLMProviderFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.LLMProviderFormOverview))

	return strings.Join(sections, "\n")
}

func (m *LLMProviderFormModel) GetCurrentConfiguration() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.providerName))

	config := m.GetController().GetLLMProviderConfig(string(m.providerID))

	if config.Configured {
		sections = append(sections, fmt.Sprintf("• %s%s",
			locale.UIStatus, m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• %s%s",
			locale.UIStatus, m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	getMaskedValue := func(value string) string {
		maskedValue := strings.Repeat("*", len(value))
		if len(value) > 15 {
			maskedValue = maskedValue[:15] + "..."
		}
		return maskedValue
	}

	// Show configured fields (without values for security)
	switch m.providerID {
	case LLMProviderOpenAI, LLMProviderAnthropic, LLMProviderGemini:
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBaseURL, m.GetStyles().Info.Render(locale.StatusConfigured)))
		}
		if config.APIKey.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldAPIKey, m.GetStyles().Muted.Render(getMaskedValue(config.APIKey.Value))))
		}

	case LLMProviderBedrock:
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBaseURL, m.GetStyles().Info.Render(locale.StatusConfigured)))
		}
		if config.AccessKey.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldAccessKey, m.GetStyles().Muted.Render(getMaskedValue(config.AccessKey.Value))))
		}
		if config.SecretKey.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldSecretKey, m.GetStyles().Muted.Render(getMaskedValue(config.SecretKey.Value))))
		}
		if config.SessionToken.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldSessionToken, m.GetStyles().Muted.Render(getMaskedValue(config.SessionToken.Value))))
		}
		if config.Region.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldRegion, m.GetStyles().Info.Render(config.Region.Value)))
		}

	case LLMProviderOllama:
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBaseURL, m.GetStyles().Info.Render(config.BaseURL.Value)))
		}
		if config.ConfigPath.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldConfigPath, m.GetStyles().Info.Render(config.ConfigPath.Value)))
		}

	case LLMProviderCustom:
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBaseURL, m.GetStyles().Info.Render(config.BaseURL.Value)))
		}
		if config.APIKey.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldAPIKey, m.GetStyles().Muted.Render(getMaskedValue(config.APIKey.Value))))
		}
		if config.Model.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldModel, m.GetStyles().Info.Render(config.Model.Value)))
		}
		if config.ConfigPath.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldConfigPath, m.GetStyles().Info.Render(config.ConfigPath.Value)))
		}
		if config.LegacyReasoning.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldLegacyReasoning, m.GetStyles().Info.Render(config.LegacyReasoning.Value)))
		}
	}

	return strings.Join(sections, "\n")
}

func (m *LLMProviderFormModel) IsConfigured() bool {
	return m.GetController().GetLLMProviderConfig(string(m.providerID)).Configured
}

func (m *LLMProviderFormModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(fmt.Sprintf(locale.LLMProviderFormTitle, m.providerName)))
	sections = append(sections, "")

	switch m.providerID {
	case LLMProviderOpenAI:
		sections = append(sections, locale.LLMFormOpenAIHelp)
	case LLMProviderAnthropic:
		sections = append(sections, locale.LLMFormAnthropicHelp)
	case LLMProviderGemini:
		sections = append(sections, locale.LLMFormGeminiHelp)
	case LLMProviderBedrock:
		sections = append(sections, locale.LLMFormBedrockHelp)
	case LLMProviderOllama:
		sections = append(sections, locale.LLMFormOllamaHelp)
	case LLMProviderCustom:
		sections = append(sections, locale.LLMFormCustomHelp)
	}

	return strings.Join(sections, "\n")
}

func (m *LLMProviderFormModel) HandleSave() error {
	config := m.GetController().GetLLMProviderConfig(string(m.providerID))
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.LLMProviderConfig{
		Name: config.Name,
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		BaseURL:                config.BaseURL,
		APIKey:                 config.APIKey,
		Model:                  config.Model,
		AccessKey:              config.AccessKey,
		SecretKey:              config.SecretKey,
		SessionToken:           config.SessionToken,
		Region:                 config.Region,
		ConfigPath:             config.ConfigPath,
		LegacyReasoning:        config.LegacyReasoning,
		EmbeddedLLMConfigsPath: config.EmbeddedLLMConfigsPath,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "base_url":
			newConfig.BaseURL.Value = value
		case "api_key":
			newConfig.APIKey.Value = value
		case "model":
			newConfig.Model.Value = value
		case "access_key":
			newConfig.AccessKey.Value = value
		case "secret_key":
			newConfig.SecretKey.Value = value
		case "session_token":
			newConfig.SessionToken.Value = value
		case "region":
			newConfig.Region.Value = value
		case "config_path":
			newConfig.ConfigPath.Value = value
		case "legacy_reasoning":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for legacy reasoning: %s (must be 'true' or 'false')", value)
			}
			newConfig.LegacyReasoning.Value = value
		}
	}

	// determine if configured based on provider type
	switch m.providerID {
	case LLMProviderBedrock:
		newConfig.Configured = (newConfig.AccessKey.Value != "" && newConfig.SecretKey.Value != "") || newConfig.SessionToken.Value != ""
	case LLMProviderOllama:
		newConfig.Configured = newConfig.BaseURL.Value != ""
	default:
		newConfig.Configured = newConfig.APIKey.Value != ""
	}

	// save the configuration
	if err := m.GetController().UpdateLLMProviderConfig(string(m.providerID), newConfig); err != nil {
		logger.Errorf("[LLMProviderFormModel] SAVE: error updating LLM provider config: %v", err)
		return err
	}

	logger.Log("[LLMProviderFormModel] SAVE: success for provider %s", m.providerID)
	return nil
}

func (m *LLMProviderFormModel) HandleReset() {
	// reset config to defaults
	m.GetController().ResetLLMProviderConfig(string(m.providerID))

	// rebuild form with reset values
	m.BuildForm()
}

func (m *LLMProviderFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *LLMProviderFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *LLMProviderFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// Update method - handle screen-specific input
func (m *LLMProviderFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// then handle field input
		if cmd := m.HandleFieldInput(msg); cmd != nil {
			return m, cmd
		}
	}

	// delegate to base screen for common handling
	cmd := m.BaseScreen.Update(msg)
	return m, cmd
}

// Helper methods

func (m *LLMProviderFormModel) getDefaultBaseURL() string {
	switch m.providerID {
	case LLMProviderOpenAI:
		return "https://api.openai.com/v1"
	case LLMProviderAnthropic:
		return "https://api.anthropic.com/v1"
	case LLMProviderGemini:
		return "https://generativelanguage.googleapis.com/v1beta"
	case LLMProviderBedrock:
		return "" // Bedrock uses regional endpoints
	case LLMProviderOllama:
		return "http://ollama-server:11434"
	case LLMProviderCustom:
		return "http://llm-server:8000"
	default:
		return ""
	}
}

// Compile-time interface validation
var _ BaseScreenModel = (*LLMProviderFormModel)(nil)
var _ BaseScreenHandler = (*LLMProviderFormModel)(nil)
