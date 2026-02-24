package models

import (
	"fmt"
	"strings"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// SearchEnginesFormModel represents the Search Engines configuration form
type SearchEnginesFormModel struct {
	*BaseScreen
}

// NewSearchEnginesFormModel creates a new Search Engines form model
func NewSearchEnginesFormModel(c controller.Controller, s styles.Styles, w window.Window) *SearchEnginesFormModel {
	m := &SearchEnginesFormModel{}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// BaseScreenHandler interface implementation

func (m *SearchEnginesFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetSearchEnginesConfig()
	fields := []FormField{}

	// DuckDuckGo (boolean)
	fields = append(fields, m.createBooleanField("duckduckgo_enabled",
		locale.ToolsSearchEnginesDuckDuckGo,
		locale.ToolsSearchEnginesDuckDuckGoDesc,
		config.DuckDuckGoEnabled,
	))

	// Sploitus (boolean)
	fields = append(fields, m.createBooleanField("sploitus_enabled",
		locale.ToolsSearchEnginesSploitus,
		locale.ToolsSearchEnginesSploitusDesc,
		config.SploitusEnabled,
	))

	// Perplexity API Key
	fields = append(fields, m.createAPIKeyField("perplexity_api_key",
		locale.ToolsSearchEnginesPerplexityKey,
		locale.ToolsSearchEnginesPerplexityKeyDesc,
		config.PerplexityAPIKey,
	))

	// Perplexity Model (suggestions)
	fields = append(fields, m.createSelectTextField(
		"perplexity_model",
		"Perplexity Model",
		"Select Perplexity model",
		config.PerplexityModel,
		[]string{"sonar", "sonar-pro", "sonar-reasoning", "sonar-reasoning-pro", "sonar-deep-research"},
		false,
	))

	// Perplexity Context Size (suggestions)
	fields = append(fields, m.createSelectTextField(
		"perplexity_context_size",
		"Perplexity Context Size",
		"Select Perplexity context size",
		config.PerplexityContextSize,
		[]string{"low", "medium", "high"},
		false,
	))

	// Tavily API Key
	fields = append(fields, m.createAPIKeyField("tavily_api_key",
		locale.ToolsSearchEnginesTavilyKey,
		locale.ToolsSearchEnginesTavilyKeyDesc,
		config.TavilyAPIKey,
	))

	// Traversaal API Key
	fields = append(fields, m.createAPIKeyField("traversaal_api_key",
		locale.ToolsSearchEnginesTraversaalKey,
		locale.ToolsSearchEnginesTraversaalKeyDesc,
		config.TraversaalAPIKey,
	))

	// Google API Key
	fields = append(fields, m.createAPIKeyField("google_api_key",
		locale.ToolsSearchEnginesGoogleKey,
		locale.ToolsSearchEnginesGoogleKeyDesc,
		config.GoogleAPIKey,
	))

	// Google CX Key
	fields = append(fields, m.createAPIKeyField("google_cx_key",
		locale.ToolsSearchEnginesGoogleCX,
		locale.ToolsSearchEnginesGoogleCXDesc,
		config.GoogleCXKey,
	))

	// Google LR Key
	fields = append(fields, m.createAPIKeyField("google_lr_key",
		locale.ToolsSearchEnginesGoogleLR,
		locale.ToolsSearchEnginesGoogleLRDesc,
		config.GoogleLRKey,
	))

	// Searxng URL
	fields = append(fields, m.createTextField("searxng_url",
		locale.ToolsSearchEnginesSearxngURL,
		locale.ToolsSearchEnginesSearxngURLDesc,
		config.SearxngURL,
		false,
	))

	// Searxng Categories
	fields = append(fields, m.createTextField("searxng_categories",
		locale.ToolsSearchEnginesSearxngCategories,
		locale.ToolsSearchEnginesSearxngCategoriesDesc,
		config.SearxngCategories,
		false,
	))

	// Searxng Language
	fields = append(fields, m.createSelectTextField("searxng_language",
		locale.ToolsSearchEnginesSearxngLanguage,
		locale.ToolsSearchEnginesSearxngLanguageDesc,
		config.SearxngLanguage,
		[]string{"en", "ch", "fr", "de", "it", "es", "pt", "ru", "zh"},
		false,
	))

	// Searxng Safe Search
	fields = append(fields, m.createSelectTextField("searxng_safe_search",
		locale.ToolsSearchEnginesSearxngSafeSearch,
		locale.ToolsSearchEnginesSearxngSafeSearchDesc,
		config.SearxngSafeSearch,
		[]string{"0", "1", "2"},
		false,
	))

	// Searxng Time Range
	fields = append(fields, m.createSelectTextField("searxng_time_range",
		locale.ToolsSearchEnginesSearxngTimeRange,
		locale.ToolsSearchEnginesSearxngTimeRangeDesc,
		config.SearxngTimeRange,
		[]string{"day", "month", "year"},
		false,
	))

	m.SetFormFields(fields)
	return nil
}

func (m *SearchEnginesFormModel) createBooleanField(key, title, description string, envVar loader.EnvVar) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), envVar)

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

func (m *SearchEnginesFormModel) createAPIKeyField(key, title, description string, envVar loader.EnvVar) FormField {
	return m.createTextField(key, title, description, envVar, true)
}

func (m *SearchEnginesFormModel) createTextField(key, title, description string, envVar loader.EnvVar, masked bool) FormField {
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

func (m *SearchEnginesFormModel) createSelectTextField(key, title, description string, envVar loader.EnvVar, suggestions []string, masked bool) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), envVar)
	input.ShowSuggestions = true
	input.SetSuggestions(suggestions)

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      masked,
		Input:       input,
		Value:       input.Value(),
		Suggestions: suggestions,
	}
}

func (m *SearchEnginesFormModel) GetFormTitle() string {
	return locale.ToolsSearchEnginesFormTitle
}

func (m *SearchEnginesFormModel) GetFormDescription() string {
	return locale.ToolsSearchEnginesFormDescription
}

func (m *SearchEnginesFormModel) GetFormName() string {
	return locale.ToolsSearchEnginesFormName
}

func (m *SearchEnginesFormModel) GetFormSummary() string {
	return ""
}

func (m *SearchEnginesFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ToolsSearchEnginesFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ToolsSearchEnginesFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ToolsSearchEnginesFormOverview))

	return strings.Join(sections, "\n")
}

func (m *SearchEnginesFormModel) GetCurrentConfiguration() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	config := m.GetController().GetSearchEnginesConfig()

	// DuckDuckGo
	duckduckgoEnabled := config.DuckDuckGoEnabled.Value
	if duckduckgoEnabled == "" {
		duckduckgoEnabled = config.DuckDuckGoEnabled.Default
	}
	if duckduckgoEnabled == "true" {
		sections = append(sections, fmt.Sprintf("• DuckDuckGo: %s",
			m.GetStyles().Success.Render(locale.StatusEnabled)))
	} else {
		sections = append(sections, fmt.Sprintf("• DuckDuckGo: %s",
			m.GetStyles().Warning.Render(locale.StatusDisabled)))
	}

	// Sploitus
	sploitusEnabled := config.SploitusEnabled.Value
	if sploitusEnabled == "" {
		sploitusEnabled = config.SploitusEnabled.Default
	}
	if sploitusEnabled == "true" {
		sections = append(sections, fmt.Sprintf("• Sploitus: %s",
			m.GetStyles().Success.Render(locale.StatusEnabled)))
	} else {
		sections = append(sections, fmt.Sprintf("• Sploitus: %s",
			m.GetStyles().Warning.Render(locale.StatusDisabled)))
	}

	// Perplexity
	if config.PerplexityAPIKey.Value != "" {
		sections = append(sections, fmt.Sprintf("• Perplexity: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Perplexity: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	// Tavily
	if config.TavilyAPIKey.Value != "" {
		sections = append(sections, fmt.Sprintf("• Tavily: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Tavily: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	// Traversaal
	if config.TraversaalAPIKey.Value != "" {
		sections = append(sections, fmt.Sprintf("• Traversaal: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Traversaal: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	// Google Search
	if config.GoogleAPIKey.Value != "" && config.GoogleCXKey.Value != "" {
		sections = append(sections, fmt.Sprintf("• Google Search: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Google Search: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	// Searxng
	if config.SearxngURL.Value != "" {
		sections = append(sections, fmt.Sprintf("• Searxng: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Searxng: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	sections = append(sections, "")
	if config.ConfiguredCount > 0 {
		sections = append(sections, m.GetStyles().Success.Render(
			fmt.Sprintf(locale.MessageSearchEnginesConfigured, config.ConfiguredCount)))
	} else {
		sections = append(sections, m.GetStyles().Warning.Render(locale.MessageSearchEnginesNone))
	}

	return strings.Join(sections, "\n")
}

func (m *SearchEnginesFormModel) IsConfigured() bool {
	return m.GetController().GetSearchEnginesConfig().ConfiguredCount > 0
}

func (m *SearchEnginesFormModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ToolsSearchEnginesFormTitle))
	sections = append(sections, "")
	sections = append(sections, locale.ToolsSearchEnginesFormOverview)

	return strings.Join(sections, "\n")
}

func (m *SearchEnginesFormModel) HandleSave() error {
	config := m.GetController().GetSearchEnginesConfig()
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.SearchEnginesConfig{
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		DuckDuckGoEnabled:     config.DuckDuckGoEnabled,
		SploitusEnabled:       config.SploitusEnabled,
		PerplexityAPIKey:      config.PerplexityAPIKey,
		PerplexityModel:       config.PerplexityModel,
		PerplexityContextSize: config.PerplexityContextSize,
		TavilyAPIKey:          config.TavilyAPIKey,
		TraversaalAPIKey:      config.TraversaalAPIKey,
		GoogleAPIKey:          config.GoogleAPIKey,
		GoogleCXKey:           config.GoogleCXKey,
		GoogleLRKey:           config.GoogleLRKey,
		SearxngURL:            config.SearxngURL,
		SearxngCategories:     config.SearxngCategories,
		SearxngLanguage:       config.SearxngLanguage,
		SearxngSafeSearch:     config.SearxngSafeSearch,
		SearxngTimeRange:      config.SearxngTimeRange,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "duckduckgo_enabled":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for DuckDuckGo: %s (must be 'true' or 'false')", value)
			}
			newConfig.DuckDuckGoEnabled.Value = value
		case "sploitus_enabled":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for Sploitus: %s (must be 'true' or 'false')", value)
			}
			newConfig.SploitusEnabled.Value = value
		case "perplexity_api_key":
			newConfig.PerplexityAPIKey.Value = value
		case "perplexity_model":
			newConfig.PerplexityModel.Value = value
		case "perplexity_context_size":
			newConfig.PerplexityContextSize.Value = value
		case "tavily_api_key":
			newConfig.TavilyAPIKey.Value = value
		case "traversaal_api_key":
			newConfig.TraversaalAPIKey.Value = value
		case "google_api_key":
			newConfig.GoogleAPIKey.Value = value
		case "google_cx_key":
			newConfig.GoogleCXKey.Value = value
		case "google_lr_key":
			newConfig.GoogleLRKey.Value = value
		case "searxng_url":
			newConfig.SearxngURL.Value = value
		case "searxng_categories":
			newConfig.SearxngCategories.Value = value
		case "searxng_language":
			newConfig.SearxngLanguage.Value = value
		case "searxng_safe_search":
			newConfig.SearxngSafeSearch.Value = value
		case "searxng_time_range":
			newConfig.SearxngTimeRange.Value = value
		}
	}

	// save the configuration
	if err := m.GetController().UpdateSearchEnginesConfig(newConfig); err != nil {
		logger.Errorf("[SearchEnginesFormModel] SAVE: error updating search engines config: %v", err)
		return err
	}

	logger.Log("[SearchEnginesFormModel] SAVE: success")
	return nil
}

func (m *SearchEnginesFormModel) HandleReset() {
	// reset config to defaults
	m.GetController().ResetSearchEnginesConfig()

	// rebuild form with reset values
	m.BuildForm()
}

func (m *SearchEnginesFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *SearchEnginesFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *SearchEnginesFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// Update method - handle screen-specific input
func (m *SearchEnginesFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// Compile-time interface validation
var _ BaseScreenModel = (*SearchEnginesFormModel)(nil)
var _ BaseScreenHandler = (*SearchEnginesFormModel)(nil)
