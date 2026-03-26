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
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScraperFormModel represents the Scraper configuration form
type ScraperFormModel struct {
	*BaseScreen

	// screen-specific components
	modeList     list.Model
	modeDelegate *BaseListDelegate
}

// NewScraperFormModel creates a new Scraper form model
func NewScraperFormModel(c controller.Controller, s styles.Styles, w window.Window) *ScraperFormModel {
	m := &ScraperFormModel{}

	m.BaseScreen = NewBaseScreen(c, s, w, m, m)
	m.initializeModeList(s)

	return m
}

// initializeModeList sets up the mode selection list
func (m *ScraperFormModel) initializeModeList(styles styles.Styles) {
	options := []BaseListOption{
		{Value: "embedded", Display: locale.ToolsScraperEmbedded},
		{Value: "external", Display: locale.ToolsScraperExternal},
		{Value: "disabled", Display: locale.ToolsScraperDisabled},
	}

	m.modeDelegate = NewBaseListDelegate(
		styles.FormLabel.Align(lipgloss.Center),
		MinMenuWidth-6,
	)

	m.modeList = m.GetListHelper().CreateList(options, m.modeDelegate, MinMenuWidth-6, 3)

	config := m.GetController().GetScraperConfig()

	m.GetListHelper().SelectByValue(&m.modeList, config.Mode)
}

// getSelectedMode returns the currently selected scraper mode using the helper
func (m *ScraperFormModel) getSelectedMode() string {
	selectedValue := m.GetListHelper().GetSelectedValue(&m.modeList)
	if selectedValue == "" {
		return "disabled"
	}

	return selectedValue
}

// BaseScreenHandler interface implementation

func (m *ScraperFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetScraperConfig()
	fields := []FormField{}
	mode := m.getSelectedMode()

	switch mode {
	case "external":
		// External mode - public and private URLs with credentials
		fields = append(fields, m.createURLField(config, "public_url",
			locale.ToolsScraperPublicURL,
			locale.ToolsScraperPublicURLDesc,
			"https://scraper.example.com",
		))
		fields = append(fields, m.createCredentialField(config, "public_username",
			locale.ToolsScraperPublicUsername,
			locale.ToolsScraperPublicUsernameDesc,
		))
		fields = append(fields, m.createCredentialField(config, "public_password",
			locale.ToolsScraperPublicPassword,
			locale.ToolsScraperPublicPasswordDesc,
		))

		fields = append(fields, m.createURLField(config, "private_url",
			locale.ToolsScraperPrivateURL,
			locale.ToolsScraperPrivateURLDesc,
			"https://scraper-internal.example.com",
		))
		fields = append(fields, m.createCredentialField(config, "private_username",
			locale.ToolsScraperPrivateUsername,
			locale.ToolsScraperPrivateUsernameDesc,
		))
		fields = append(fields, m.createCredentialField(config, "private_password",
			locale.ToolsScraperPrivatePassword,
			locale.ToolsScraperPrivatePasswordDesc,
		))

	case "embedded":
		// Embedded mode - optional public URL override and local settings
		fields = append(fields, m.createURLField(config, "public_url",
			locale.ToolsScraperPublicURL,
			locale.ToolsScraperPublicURLEmbeddedDesc,
			controller.DefaultScraperBaseURL,
		))
		fields = append(fields, m.createCredentialField(config, "public_username",
			locale.ToolsScraperPublicUsername,
			locale.ToolsScraperPublicUsernameDesc,
		))
		fields = append(fields, m.createCredentialField(config, "public_password",
			locale.ToolsScraperPublicPassword,
			locale.ToolsScraperPublicPasswordDesc,
		))
		fields = append(fields, m.createCredentialField(config, "private_username",
			locale.ToolsScraperLocalUsername,
			locale.ToolsScraperLocalUsernameDesc,
		))
		fields = append(fields, m.createCredentialField(config, "private_password",
			locale.ToolsScraperLocalPassword,
			locale.ToolsScraperLocalPasswordDesc,
		))
		fields = append(fields, m.createSessionsField(config, "max_sessions",
			locale.ToolsScraperMaxConcurrentSessions,
			locale.ToolsScraperMaxConcurrentSessionsDesc,
			config.MaxConcurrentSessions.Default,
		))

	case "disabled":
		// Disabled mode has no additional fields
	}

	m.SetFormFields(fields)
	return nil
}

func (m *ScraperFormModel) createURLField(
	config *controller.ScraperConfig, key, title, description, placeholder string,
) FormField {
	input := textinput.New()
	input.Prompt = ""
	input.PlaceholderStyle = m.GetStyles().FormPlaceholder
	input.Placeholder = placeholder

	var value string
	switch key {
	case "public_url":
		value = config.PublicURL.Value
	case "private_url":
		value = config.PrivateURL.Value
	}

	if value != "" {
		input.SetValue(value)
	}

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *ScraperFormModel) createCredentialField(
	config *controller.ScraperConfig, key, title, description string,
) FormField {
	input := textinput.New()
	input.Prompt = ""
	input.PlaceholderStyle = m.GetStyles().FormPlaceholder

	var value string
	switch key {
	case "public_username":
		value = config.PublicUsername
	case "public_password":
		value = config.PublicPassword
	case "private_username":
		value = config.PrivateUsername
	case "private_password":
		value = config.PrivatePassword
	}

	if value != "" {
		input.SetValue(value)
	}

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      true,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *ScraperFormModel) createSessionsField(
	config *controller.ScraperConfig, key, title, description, placeholder string,
) FormField {
	input := textinput.New()
	input.Prompt = ""
	input.PlaceholderStyle = m.GetStyles().FormPlaceholder
	input.Placeholder = placeholder

	if config.MaxConcurrentSessions.Value != "" {
		input.SetValue(config.MaxConcurrentSessions.Value)
	}

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *ScraperFormModel) GetFormTitle() string {
	return locale.ToolsScraperFormTitle
}

func (m *ScraperFormModel) GetFormDescription() string {
	return locale.ToolsScraperFormDescription
}

func (m *ScraperFormModel) GetFormName() string {
	return locale.ToolsScraperFormName
}

func (m *ScraperFormModel) GetFormSummary() string {
	return ""
}

func (m *ScraperFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ToolsScraperFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ToolsScraperFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ToolsScraperFormOverview))

	return strings.Join(sections, "\n")
}

func (m *ScraperFormModel) GetCurrentConfiguration() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	getMaskedValue := func(value string) string {
		maskedValue := strings.Repeat("*", len(value))
		if len(value) > 15 {
			maskedValue = maskedValue[:15] + "..."
		}
		return maskedValue
	}

	config := m.GetController().GetScraperConfig()

	switch config.Mode {
	case "embedded":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.StatusEmbedded))
		if privateURL := config.PrivateURL.Value; privateURL != "" {
			cleanURL := controller.RemoveCredentialsFromURL(privateURL)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPrivateURL, m.GetStyles().Info.Render(cleanURL)))
		}
		if publicUsername := config.PublicUsername; publicUsername != "" {
			maskedPublicUsername := getMaskedValue(publicUsername)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPublicUsername, m.GetStyles().Muted.Render(maskedPublicUsername)))
		}
		if publicPassword := config.PublicPassword; publicPassword != "" {
			maskedPublicPassword := getMaskedValue(publicPassword)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPublicPassword, m.GetStyles().Muted.Render(maskedPublicPassword)))
		}
		if publicURL := config.PublicURL.Value; publicURL != "" {
			cleanURL := controller.RemoveCredentialsFromURL(publicURL)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPublicURL, m.GetStyles().Info.Render(cleanURL)))
		}
		if localUsername := config.LocalUsername.Value; localUsername != "" {
			maskedLocalUsername := getMaskedValue(localUsername)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperLocalUsername, m.GetStyles().Muted.Render(maskedLocalUsername)))
		}
		if localPassword := config.LocalPassword.Value; localPassword != "" {
			maskedLocalPassword := getMaskedValue(localPassword)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperLocalPassword, m.GetStyles().Muted.Render(maskedLocalPassword)))
		}
		if maxConcurrentSessions := config.MaxConcurrentSessions.Value; maxConcurrentSessions != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperMaxConcurrentSessions, m.GetStyles().Info.Render(maxConcurrentSessions)))
		}

	case "external":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.StatusExternal))
		if publicURL := config.PublicURL.Value; publicURL != "" {
			cleanURL := controller.RemoveCredentialsFromURL(publicURL)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPublicURL, m.GetStyles().Info.Render(cleanURL)))
		}
		if publicUsername := config.PublicUsername; publicUsername != "" {
			maskedPublicUsername := getMaskedValue(publicUsername)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPublicUsername, m.GetStyles().Muted.Render(maskedPublicUsername)))
		}
		if publicPassword := config.PublicPassword; publicPassword != "" {
			maskedPublicPassword := getMaskedValue(publicPassword)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPublicPassword, m.GetStyles().Muted.Render(maskedPublicPassword)))
		}
		if privateURL := config.PrivateURL.Value; privateURL != "" {
			cleanURL := controller.RemoveCredentialsFromURL(privateURL)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPrivateURL, m.GetStyles().Info.Render(cleanURL)))
		}
		if privateUsername := config.PrivateUsername; privateUsername != "" {
			maskedPrivateUsername := getMaskedValue(privateUsername)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPrivateUsername, m.GetStyles().Muted.Render(maskedPrivateUsername)))
		}
		if privatePassword := config.PrivatePassword; privatePassword != "" {
			maskedPrivatePassword := getMaskedValue(privatePassword)
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.ToolsScraperPrivatePassword, m.GetStyles().Muted.Render(maskedPrivatePassword)))
		}

	case "disabled":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Warning.Render(locale.StatusDisabled))
	}

	return strings.Join(sections, "\n")
}

func (m *ScraperFormModel) IsConfigured() bool {
	return m.GetController().GetScraperConfig().Mode != "disabled"
}

func (m *ScraperFormModel) GetHelpContent() string {
	var sections []string
	mode := m.getSelectedMode()

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ToolsScraperFormTitle))
	sections = append(sections, "")

	switch mode {
	case "embedded":
		sections = append(sections, locale.ToolsScraperEmbeddedHelp)
	case "external":
		sections = append(sections, locale.ToolsScraperExternalHelp)
	case "disabled":
		sections = append(sections, locale.ToolsScraperDisabledHelp)
	}

	return strings.Join(sections, "\n")
}

func (m *ScraperFormModel) HandleSave() error {
	config := m.GetController().GetScraperConfig()
	mode := m.getSelectedMode()
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.ScraperConfig{
		Mode: mode,
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		PublicURL:             config.PublicURL,
		PrivateURL:            config.PrivateURL,
		LocalUsername:         config.LocalUsername,
		LocalPassword:         config.LocalPassword,
		MaxConcurrentSessions: config.MaxConcurrentSessions,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "public_url":
			newConfig.PublicURL.Value = value
		case "private_url":
			newConfig.PrivateURL.Value = value
		case "public_username":
			newConfig.PublicUsername = value
		case "public_password":
			newConfig.PublicPassword = value
		case "private_username":
			newConfig.PrivateUsername = value
			newConfig.LocalUsername.Value = value
		case "private_password":
			newConfig.PrivatePassword = value
			newConfig.LocalPassword.Value = value
		case "max_sessions":
			// validate numeric input
			if value != "" {
				if _, err := strconv.Atoi(value); err != nil {
					return fmt.Errorf("invalid number for max concurrent sessions: %s", value)
				}
			}
			newConfig.MaxConcurrentSessions.Value = value
		}
	}

	// save the configuration
	if err := m.GetController().UpdateScraperConfig(newConfig); err != nil {
		logger.Errorf("[ScraperFormModel] SAVE: error updating scraper config: %v", err)
		return err
	}

	logger.Log("[ScraperFormModel] SAVE: success")
	return nil
}

func (m *ScraperFormModel) HandleReset() {
	// reset config to defaults
	config := m.GetController().ResetScraperConfig()

	// reset mode selection
	m.GetListHelper().SelectByValue(&m.modeList, config.Mode)

	// rebuild form with reset mode
	m.BuildForm()
}

func (m *ScraperFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *ScraperFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *ScraperFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// BaseListHandler interface implementation

func (m *ScraperFormModel) GetList() *list.Model {
	return &m.modeList
}

func (m *ScraperFormModel) GetListDelegate() *BaseListDelegate {
	return m.modeDelegate
}

func (m *ScraperFormModel) OnListSelectionChanged(oldSelection, newSelection string) {
	// rebuild form when mode changes
	m.BuildForm()
}

func (m *ScraperFormModel) GetListTitle() string {
	return locale.ToolsScraperModeTitle
}

func (m *ScraperFormModel) GetListDescription() string {
	return locale.ToolsScraperModeDesc
}

// Update method - handle screen-specific input
func (m *ScraperFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
var _ BaseScreenModel = (*ScraperFormModel)(nil)
var _ BaseScreenHandler = (*ScraperFormModel)(nil)
var _ BaseListHandler = (*ScraperFormModel)(nil)
