package models

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Anthropic authentication modes shown in the mode selector.
const (
	anthropicModeAPIKey    = "api_key"
	anthropicModeFederated = "federated"
)

// LLMProviderFormModel represents the LLM Provider configuration form
type LLMProviderFormModel struct {
	*BaseScreen

	// screen-specific components
	providerID   LLMProviderID
	providerName string

	// Anthropic-only: authentication mode selector (nil list for other providers)
	modeList     list.Model
	modeDelegate *BaseListDelegate
}

// NewLLMProviderFormModel creates a new LLM Provider form model
func NewLLMProviderFormModel(
	c controller.Controller, s styles.Styles, w window.Window, pid LLMProviderID,
) *LLMProviderFormModel {
	m := &LLMProviderFormModel{
		providerID:   pid,
		providerName: c.GetLLMProviderConfig(string(pid)).Name,
	}

	// Anthropic gets an authentication mode selector (like the Scraper form);
	// every other provider is a plain field form with no list.
	var listHandler BaseListHandler
	if pid == LLMProviderAnthropic {
		listHandler = m
	}
	m.BaseScreen = NewBaseScreen(c, s, w, m, listHandler)
	if pid == LLMProviderAnthropic {
		m.initializeModeList(s)
	}

	return m
}

// initializeModeList sets up the Anthropic authentication mode selector.
func (m *LLMProviderFormModel) initializeModeList(styles styles.Styles) {
	options := []BaseListOption{
		{Value: anthropicModeAPIKey, Display: locale.LLMAnthropicAuthModeAPIKey},
		{Value: anthropicModeFederated, Display: locale.LLMAnthropicAuthModeFederated},
	}

	m.modeDelegate = NewBaseListDelegate(
		styles.FormLabel.Align(lipgloss.Center),
		MinMenuWidth-6,
	)
	m.modeList = m.GetListHelper().CreateList(options, m.modeDelegate, MinMenuWidth-6, 2)

	config := m.GetController().GetLLMProviderConfig(string(m.providerID))
	mode := anthropicModeAPIKey
	if config.AnthropicOrganizationID.Value != "" ||
		config.AnthropicWorkspaceID.Value != "" ||
		config.AnthropicServiceAccountID.Value != "" ||
		config.AnthropicIdentityToken.Value != "" ||
		config.AnthropicIdentityTokenFile.Value != "" ||
		config.AnthropicFederationRuleID.Value != "" {
		mode = anthropicModeFederated
	}
	m.GetListHelper().SelectByValue(&m.modeList, mode)
}

// getAnthropicMode returns the selected Anthropic authentication mode.
func (m *LLMProviderFormModel) getAnthropicMode() string {
	if m.providerID != LLMProviderAnthropic {
		return anthropicModeAPIKey
	}
	if v := m.GetListHelper().GetSelectedValue(&m.modeList); v != "" {
		return v
	}
	return anthropicModeAPIKey
}

// BaseScreenHandler interface implementation

func (m *LLMProviderFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetLLMProviderConfig(string(m.providerID))
	fields := []FormField{}

	// Add fields based on provider type
	switch m.providerID {
	case LLMProviderOpenAI, LLMProviderGemini:
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createAPIKeyField(config))

	case LLMProviderAnthropic:
		fields = append(fields, m.createBaseURLField(config))
		if m.getAnthropicMode() == anthropicModeFederated {
			fields = append(fields, m.createAnthropicField(config, "anthropic_org_id",
				locale.LLMFormFieldAnthropicOrgID, locale.LLMFormAnthropicOrgIDDesc, config.AnthropicOrganizationID, false))
			fields = append(fields, m.createAnthropicField(config, "anthropic_workspace_id",
				locale.LLMFormFieldAnthropicWorkspace, locale.LLMFormAnthropicWorkspaceDesc, config.AnthropicWorkspaceID, false))
			fields = append(fields, m.createAnthropicField(config, "anthropic_service_account_id",
				locale.LLMFormFieldAnthropicServiceAcc, locale.LLMFormAnthropicServiceAccDesc, config.AnthropicServiceAccountID, false))
			fields = append(fields, m.createAnthropicField(config, "anthropic_identity_token",
				locale.LLMFormFieldAnthropicIDToken, locale.LLMFormAnthropicIDTokenDesc, config.AnthropicIdentityToken, true))
			fields = append(fields, m.createAnthropicField(config, "anthropic_identity_token_file",
				locale.LLMFormFieldAnthropicIDTokenF, locale.LLMFormAnthropicIDTokenFDesc, config.AnthropicIdentityTokenFile, false))
			fields = append(fields, m.createAnthropicField(config, "anthropic_federation_rule_id",
				locale.LLMFormFieldAnthropicFedRule, locale.LLMFormAnthropicFedRuleDesc, config.AnthropicFederationRuleID, false))
		} else {
			fields = append(fields, m.createAPIKeyField(config))
		}

	case LLMProviderBedrock:
		fields = append(fields, m.createRegionField(config))
		fields = append(fields, m.createDefaultAuthField(config))
		fields = append(fields, m.createBearerTokenField(config))
		fields = append(fields, m.createAccessKeyField(config))
		fields = append(fields, m.createSecretKeyField(config))
		fields = append(fields, m.createSessionTokenField(config))
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createConfigPathField(config))

	case LLMProviderOllama:
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createOllamaAPIKeyField(config))
		fields = append(fields, m.createModelField(config))
		fields = append(fields, m.createConfigPathField(config))
		fields = append(fields, m.createPullTimeoutField(config))
		fields = append(fields, m.createPullEnabledField(config))
		fields = append(fields, m.createLoadModelsEnabledField(config))

	case LLMProviderDeepSeek, LLMProviderGLM, LLMProviderKimi, LLMProviderQwen, LLMProviderMiniMax, LLMProviderMistral, LLMProviderXAI:
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createAPIKeyField(config))
		fields = append(fields, m.createProviderNameField(config))

	case LLMProviderCustom:
		fields = append(fields, m.createBaseURLField(config))
		fields = append(fields, m.createAPIKeyField(config))
		fields = append(fields, m.createModelField(config))
		fields = append(fields, m.createConfigPathField(config))
		fields = append(fields, m.createPreserveReasoningField(config))
		fields = append(fields, m.createAPITypeField(config))
		fields = append(fields, m.createAPIVersionField(config))
		fields = append(fields, m.createProviderNameField(config))
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

func (m *LLMProviderFormModel) createAnthropicField(
	config *controller.LLMProviderConfig, key, title, description string, envVar loader.EnvVar, masked bool,
) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), envVar)

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      masked,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createDefaultAuthField(config *controller.LLMProviderConfig) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), config.DefaultAuth)

	return FormField{
		Key:         "default_auth",
		Title:       locale.LLMFormFieldDefaultAuth,
		Description: locale.LLMFormDefaultAuthDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

func (m *LLMProviderFormModel) createBearerTokenField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.BearerToken)

	return FormField{
		Key:         "bearer_token",
		Title:       locale.LLMFormFieldBearerToken,
		Description: locale.LLMFormBearerTokenDesc,
		Required:    false,
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
		Required:    false,
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
		Required:    false,
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
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.HostConfigPath)
	if config.HostConfigPath.Default == "" {
		input.Placeholder = controller.LLMConfigMountPath(string(m.providerID))
	}
	// the path shown may come from the container variable while the host one is neither in
	// the env file nor changed, and NewTextInput leaves the input of such a variable empty
	input.SetValue(config.HostConfigPath.Value)

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

func (m *LLMProviderFormModel) createAPITypeField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.APIType)
	input.ShowSuggestions = true
	input.SetSuggestions([]string{"openai", "azure", "azure_ad"})

	return FormField{
		Description: locale.LLMFormAPITypeDesc,
		Input:       input,
		Key:         "api_type",
		Required:    false,
		Suggestions: input.AvailableSuggestions(),
		Title:       locale.LLMFormFieldAPIType,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createAPIVersionField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.APIVersion)

	return FormField{
		Description: locale.LLMFormAPIVersionDesc,
		Input:       input,
		Key:         "api_version",
		Required:    false,
		Title:       locale.LLMFormFieldAPIVersion,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createPreserveReasoningField(config *controller.LLMProviderConfig) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), config.PreserveReasoning)

	return FormField{
		Key:         "preserve_reasoning",
		Title:       locale.LLMFormFieldPreserveReasoning,
		Description: locale.LLMFormPreserveReasoningDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

func (m *LLMProviderFormModel) createProviderNameField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.ProviderName)
	input.Placeholder = "openrouter"

	return FormField{
		Key:         "provider_name",
		Title:       locale.LLMFormFieldProviderName,
		Description: locale.LLMFormProviderNameDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createPullTimeoutField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.PullTimeout)
	input.Placeholder = config.PullTimeout.Default

	return FormField{
		Key:         "pull_timeout",
		Title:       locale.LLMFormFieldPullTimeout,
		Description: locale.LLMFormPullTimeoutDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *LLMProviderFormModel) createPullEnabledField(config *controller.LLMProviderConfig) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), config.PullEnabled)
	input.Placeholder = config.PullEnabled.Default

	return FormField{
		Key:         "pull_enabled",
		Title:       locale.LLMFormFieldPullEnabled,
		Description: locale.LLMFormPullEnabledDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

func (m *LLMProviderFormModel) createLoadModelsEnabledField(config *controller.LLMProviderConfig) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), config.LoadModelsEnabled)
	input.Placeholder = config.LoadModelsEnabled.Default

	return FormField{
		Key:         "load_models_enabled",
		Title:       locale.LLMFormFieldLoadModelsEnabled,
		Description: locale.LLMFormLoadModelsEnabledDesc,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

func (m *LLMProviderFormModel) createOllamaAPIKeyField(config *controller.LLMProviderConfig) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), config.APIKey)

	return FormField{
		Key:         "ollama_api_key",
		Title:       locale.LLMFormFieldAPIKey,
		Description: locale.LLMFormOllamaAPIKeyDesc,
		Required:    false,
		Masked:      true,
		Input:       input,
		Value:       input.Value(),
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
	case LLMProviderDeepSeek:
		return locale.LLMProviderDeepSeekDesc
	case LLMProviderGLM:
		return locale.LLMProviderGLMDesc
	case LLMProviderKimi:
		return locale.LLMProviderKimiDesc
	case LLMProviderQwen:
		return locale.LLMProviderQwenDesc
	case LLMProviderMiniMax:
		return locale.LLMProviderMiniMaxDesc
	case LLMProviderMistral:
		return locale.LLMProviderMistralDesc
	case LLMProviderXAI:
		return locale.LLMProviderXAIDesc
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
	case LLMProviderDeepSeek:
		return locale.LLMProviderDeepSeek
	case LLMProviderGLM:
		return locale.LLMProviderGLM
	case LLMProviderKimi:
		return locale.LLMProviderKimi
	case LLMProviderQwen:
		return locale.LLMProviderQwen
	case LLMProviderMiniMax:
		return locale.LLMProviderMiniMax
	case LLMProviderMistral:
		return locale.LLMProviderMistral
	case LLMProviderXAI:
		return locale.LLMProviderXAI
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
	case LLMProviderOpenAI, LLMProviderGemini:
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBaseURL, m.GetStyles().Info.Render(locale.StatusConfigured)))
		}
		if config.APIKey.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldAPIKey, m.GetStyles().Muted.Render(getMaskedValue(config.APIKey.Value))))
		}

	case LLMProviderAnthropic:
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBaseURL, m.GetStyles().Info.Render(locale.StatusConfigured)))
		}
		if m.getAnthropicMode() == anthropicModeFederated {
			sections = append(sections, "• "+locale.LLMAnthropicAuthModeTitle+": "+
				m.GetStyles().Success.Render(locale.LLMAnthropicAuthModeFederated))
			showText := func(label string, envVar loader.EnvVar) {
				if envVar.Value != "" {
					sections = append(sections, fmt.Sprintf("• %s: %s", label, m.GetStyles().Info.Render(envVar.Value)))
				}
			}
			showText(locale.LLMFormFieldAnthropicOrgID, config.AnthropicOrganizationID)
			showText(locale.LLMFormFieldAnthropicWorkspace, config.AnthropicWorkspaceID)
			showText(locale.LLMFormFieldAnthropicServiceAcc, config.AnthropicServiceAccountID)
			if config.AnthropicIdentityToken.Value != "" {
				sections = append(sections, fmt.Sprintf("• %s: %s",
					locale.LLMFormFieldAnthropicIDToken, m.GetStyles().Muted.Render(getMaskedValue(config.AnthropicIdentityToken.Value))))
			}
			showText(locale.LLMFormFieldAnthropicIDTokenF, config.AnthropicIdentityTokenFile)
			showText(locale.LLMFormFieldAnthropicFedRule, config.AnthropicFederationRuleID)
		} else if config.APIKey.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldAPIKey, m.GetStyles().Muted.Render(getMaskedValue(config.APIKey.Value))))
		}

	case LLMProviderBedrock:
		if config.Region.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldRegion, m.GetStyles().Info.Render(config.Region.Value)))
		}
		if config.DefaultAuth.Value == "true" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldDefaultAuth, m.GetStyles().Success.Render("enabled")))
		}
		if config.BearerToken.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBearerToken, m.GetStyles().Muted.Render(getMaskedValue(config.BearerToken.Value))))
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
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBaseURL, m.GetStyles().Info.Render(locale.StatusConfigured)))
		}
		if config.HostConfigPath.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldConfigPath, m.GetStyles().Info.Render(config.HostConfigPath.Value)))
		}

	case LLMProviderOllama:
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
		if config.HostConfigPath.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldConfigPath, m.GetStyles().Info.Render(config.HostConfigPath.Value)))
		}
		if config.PullTimeout.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldPullTimeout, m.GetStyles().Info.Render(config.PullTimeout.Value)))
		}
		if config.PullEnabled.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldPullEnabled, m.GetStyles().Info.Render(config.PullEnabled.Value)))
		}
		if config.LoadModelsEnabled.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldLoadModelsEnabled, m.GetStyles().Info.Render(config.LoadModelsEnabled.Value)))
		}

	case LLMProviderDeepSeek, LLMProviderGLM, LLMProviderKimi, LLMProviderQwen, LLMProviderMiniMax, LLMProviderMistral, LLMProviderXAI:
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldBaseURL, m.GetStyles().Info.Render(locale.StatusConfigured)))
		}
		if config.APIKey.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldAPIKey, m.GetStyles().Muted.Render(getMaskedValue(config.APIKey.Value))))
		}
		if config.ProviderName.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldProviderName, m.GetStyles().Info.Render(config.ProviderName.Value)))
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
		if config.HostConfigPath.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldConfigPath, m.GetStyles().Info.Render(config.HostConfigPath.Value)))
		}
		if config.PreserveReasoning.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldPreserveReasoning, m.GetStyles().Info.Render(config.PreserveReasoning.Value)))
		}
		if config.ProviderName.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.LLMFormFieldProviderName, m.GetStyles().Info.Render(config.ProviderName.Value)))
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
		if m.getAnthropicMode() == anthropicModeFederated {
			sections = append(sections, locale.LLMFormAnthropicFederatedHelp)
		} else {
			sections = append(sections, locale.LLMFormAnthropicHelp)
		}
	case LLMProviderGemini:
		sections = append(sections, locale.LLMFormGeminiHelp)
	case LLMProviderBedrock:
		sections = append(sections, locale.LLMFormBedrockHelp)
	case LLMProviderOllama:
		sections = append(sections, locale.LLMFormOllamaHelp)
	case LLMProviderDeepSeek:
		sections = append(sections, locale.LLMFormDeepSeekHelp)
	case LLMProviderGLM:
		sections = append(sections, locale.LLMFormGLMHelp)
	case LLMProviderKimi:
		sections = append(sections, locale.LLMFormKimiHelp)
	case LLMProviderQwen:
		sections = append(sections, locale.LLMFormQwenHelp)
	case LLMProviderMiniMax:
		sections = append(sections, locale.LLMFormMiniMaxHelp)
	case LLMProviderMistral:
		sections = append(sections, locale.LLMFormMistralHelp)
	case LLMProviderXAI:
		sections = append(sections, locale.LLMFormXAIHelp)
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
		BaseURL:                    config.BaseURL,
		APIKey:                     config.APIKey,
		Model:                      config.Model,
		DefaultAuth:                config.DefaultAuth,
		BearerToken:                config.BearerToken,
		AccessKey:                  config.AccessKey,
		SecretKey:                  config.SecretKey,
		SessionToken:               config.SessionToken,
		Region:                     config.Region,
		AnthropicOrganizationID:    config.AnthropicOrganizationID,
		AnthropicWorkspaceID:       config.AnthropicWorkspaceID,
		AnthropicServiceAccountID:  config.AnthropicServiceAccountID,
		AnthropicIdentityToken:     config.AnthropicIdentityToken,
		AnthropicIdentityTokenFile: config.AnthropicIdentityTokenFile,
		AnthropicFederationRuleID:  config.AnthropicFederationRuleID,
		ConfigPath:                 config.ConfigPath,
		HostConfigPath:             config.HostConfigPath,
		PreserveReasoning:          config.PreserveReasoning,
		APIType:                    config.APIType,
		APIVersion:                 config.APIVersion,
		ProviderName:               config.ProviderName,
		PullTimeout:                config.PullTimeout,
		PullEnabled:                config.PullEnabled,
		LoadModelsEnabled:          config.LoadModelsEnabled,
		EmbeddedLLMConfigsPath:     config.EmbeddedLLMConfigsPath,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "base_url":
			newConfig.BaseURL.Value = value
		case "api_key":
			newConfig.APIKey.Value = value
		case "anthropic_org_id":
			newConfig.AnthropicOrganizationID.Value = value
		case "anthropic_workspace_id":
			newConfig.AnthropicWorkspaceID.Value = value
		case "anthropic_service_account_id":
			newConfig.AnthropicServiceAccountID.Value = value
		case "anthropic_identity_token":
			newConfig.AnthropicIdentityToken.Value = value
		case "anthropic_identity_token_file":
			newConfig.AnthropicIdentityTokenFile.Value = value
		case "anthropic_federation_rule_id":
			newConfig.AnthropicFederationRuleID.Value = value
		case "model":
			newConfig.Model.Value = value
		case "default_auth":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for default auth: %s (must be 'true' or 'false')", value)
			}
			newConfig.DefaultAuth.Value = value
		case "bearer_token":
			newConfig.BearerToken.Value = value
		case "access_key":
			newConfig.AccessKey.Value = value
		case "secret_key":
			newConfig.SecretKey.Value = value
		case "session_token":
			newConfig.SessionToken.Value = value
		case "region":
			newConfig.Region.Value = value
		case "ollama_api_key":
			newConfig.APIKey.Value = value
		case "config_path":
			// User edits HostConfigPath, ConfigPath is auto-generated on save
			// validate config path if provided (skip validation for embedded configs)
			if value != "" {
				// embedded configs don't need validation (they're inside the docker image),
				// and neither does the path the default example is mounted at
				isEmbedded := slices.Contains(newConfig.EmbeddedLLMConfigsPath, value) ||
					value == controller.LLMConfigMountPath(string(m.providerID))

				// only validate custom (non-embedded) configs on host filesystem
				if !isEmbedded {
					info, err := os.Stat(value)
					if err != nil {
						if os.IsNotExist(err) {
							return fmt.Errorf("config file does not exist: %s", value)
						}
						return fmt.Errorf("cannot access config file %s: %v", value, err)
					}
					if info.IsDir() {
						return fmt.Errorf("config path must be a file, not a directory: %s", value)
					}
				}
			}
			newConfig.HostConfigPath.Value = value
		case "preserve_reasoning":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for preserve reasoning: %s (must be 'true' or 'false')", value)
			}
			newConfig.PreserveReasoning.Value = value
		case "api_type":
			newConfig.APIType.Value = value
		case "api_version":
			newConfig.APIVersion.Value = value
		case "provider_name":
			newConfig.ProviderName.Value = value
		case "pull_timeout":
			newConfig.PullTimeout.Value = value
		case "pull_enabled":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for pull enabled: %s (must be 'true' or 'false')", value)
			}
			newConfig.PullEnabled.Value = value
		case "load_models_enabled":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for load models enabled: %s (must be 'true' or 'false')", value)
			}
			newConfig.LoadModelsEnabled.Value = value
		}
	}

	// determine if configured based on provider type
	switch m.providerID {
	case LLMProviderAnthropic:
		// API key, or a complete federated identity (mirrors controller.anthropicFederated)
		federated := newConfig.AnthropicFederationRuleID.Value != "" &&
			newConfig.AnthropicOrganizationID.Value != "" &&
			newConfig.AnthropicServiceAccountID.Value != "" &&
			(newConfig.AnthropicIdentityToken.Value != "" || newConfig.AnthropicIdentityTokenFile.Value != "")
		newConfig.Configured = newConfig.APIKey.Value != "" || federated
	case LLMProviderBedrock:
		// Configured if any of three auth methods is set: DefaultAuth, BearerToken, or AccessKey+SecretKey
		newConfig.Configured = newConfig.DefaultAuth.Value == "true" ||
			newConfig.BearerToken.Value != "" ||
			(newConfig.AccessKey.Value != "" && newConfig.SecretKey.Value != "")
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

	// Anthropic: a reset clears the federated fields, so return the selector to API key.
	if m.providerID == LLMProviderAnthropic {
		m.GetListHelper().SelectByValue(&m.modeList, anthropicModeAPIKey)
	}

	// rebuild form with reset values
	m.BuildForm()
}

// BaseListHandler interface implementation (Anthropic authentication mode selector)

func (m *LLMProviderFormModel) GetList() *list.Model {
	return &m.modeList
}

func (m *LLMProviderFormModel) GetListDelegate() *BaseListDelegate {
	return m.modeDelegate
}

func (m *LLMProviderFormModel) OnListSelectionChanged(oldSelection, newSelection string) {
	// rebuild the form so the fields match the chosen authentication mode
	m.BuildForm()
}

func (m *LLMProviderFormModel) GetListTitle() string {
	return locale.LLMAnthropicAuthModeTitle
}

func (m *LLMProviderFormModel) GetListDescription() string {
	return locale.LLMAnthropicAuthModeDesc
}

func (m *LLMProviderFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *LLMProviderFormModel) GetFormFields() []FormField {
	return m.fields
}

func (m *LLMProviderFormModel) SetFormFields(fields []FormField) {
	m.fields = fields
}

// Update method - handle screen-specific input
func (m *LLMProviderFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// handle the Anthropic mode selector first (no-op when there is no list)
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
	case LLMProviderDeepSeek:
		return "https://api.deepseek.com"
	case LLMProviderGLM:
		return "https://api.z.ai/api/paas/v4"
	case LLMProviderKimi:
		return "https://api.moonshot.ai/v1"
	case LLMProviderQwen:
		return "https://dashscope-us.aliyuncs.com/compatible-mode/v1"
	case LLMProviderMiniMax:
		return "https://api.minimax.io/v1"
	case LLMProviderMistral:
		return "https://api.mistral.ai/v1"
	case LLMProviderXAI:
		return "https://api.x.ai/v1"
	case LLMProviderCustom:
		return "http://llm-server:8000"
	default:
		return ""
	}
}

// Compile-time interface validation
var _ BaseScreenModel = (*LLMProviderFormModel)(nil)
var _ BaseScreenHandler = (*LLMProviderFormModel)(nil)
var _ BaseListHandler = (*LLMProviderFormModel)(nil)
