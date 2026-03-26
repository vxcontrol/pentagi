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

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	GraphitiURLPlaceholder       = "http://graphiti:8000"
	GraphitiTimeoutPlaceholder   = "30"
	GraphitiModelNamePlaceholder = "gpt-5-mini"
	GraphitiNeo4jUserPlaceholder = "neo4j"
)

// GraphitiFormModel represents the Graphiti configuration form
type GraphitiFormModel struct {
	*BaseScreen

	// screen-specific components
	deploymentList     list.Model
	deploymentDelegate *BaseListDelegate
}

// NewGraphitiFormModel creates a new Graphiti form model
func NewGraphitiFormModel(c controller.Controller, s styles.Styles, w window.Window) *GraphitiFormModel {
	m := &GraphitiFormModel{}

	m.BaseScreen = NewBaseScreen(c, s, w, m, m)
	m.initializeDeploymentList(s)

	return m
}

// initializeDeploymentList sets up the deployment type selection list
func (m *GraphitiFormModel) initializeDeploymentList(styles styles.Styles) {
	options := []BaseListOption{
		{Value: "embedded", Display: locale.MonitoringGraphitiEmbedded},
		{Value: "external", Display: locale.MonitoringGraphitiExternal},
		{Value: "disabled", Display: locale.MonitoringGraphitiDisabled},
	}

	m.deploymentDelegate = NewBaseListDelegate(
		styles.FormLabel.Align(lipgloss.Center),
		MinMenuWidth-6,
	)

	m.deploymentList = m.GetListHelper().CreateList(options, m.deploymentDelegate, MinMenuWidth-6, 3)

	config := m.GetController().GetGraphitiConfig()

	m.GetListHelper().SelectByValue(&m.deploymentList, config.DeploymentType)
}

// getSelectedDeploymentType returns the currently selected deployment type using the helper
func (m *GraphitiFormModel) getSelectedDeploymentType() string {
	selectedValue := m.GetListHelper().GetSelectedValue(&m.deploymentList)
	if selectedValue == "" {
		return "disabled"
	}

	return selectedValue
}

// BaseScreenHandler interface implementation

func (m *GraphitiFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetGraphitiConfig()
	fields := []FormField{}
	deploymentType := m.getSelectedDeploymentType()

	switch deploymentType {
	case "embedded":
		// Embedded mode - requires timeout, model, and neo4j credentials
		fields = append(fields, m.createTextField(config, "timeout",
			locale.MonitoringGraphitiTimeout, locale.MonitoringGraphitiTimeoutDesc, false, GraphitiTimeoutPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "model_name",
			locale.MonitoringGraphitiModelName, locale.MonitoringGraphitiModelNameDesc, false, GraphitiModelNamePlaceholder,
		))
		fields = append(fields, m.createTextField(config, "neo4j_user",
			locale.MonitoringGraphitiNeo4jUser, locale.MonitoringGraphitiNeo4jUserDesc, false, GraphitiNeo4jUserPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "neo4j_password",
			locale.MonitoringGraphitiNeo4jPassword, locale.MonitoringGraphitiNeo4jPasswordDesc, true, "",
		))
		fields = append(fields, m.createTextField(config, "neo4j_database",
			locale.MonitoringGraphitiNeo4jDatabase, locale.MonitoringGraphitiNeo4jDatabaseDesc, false, GraphitiNeo4jUserPlaceholder,
		))

	case "external":
		// External mode - requires connection details only
		fields = append(fields, m.createTextField(config, "url",
			locale.MonitoringGraphitiURL, locale.MonitoringGraphitiURLDesc, false, GraphitiURLPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "timeout",
			locale.MonitoringGraphitiTimeout, locale.MonitoringGraphitiTimeoutDesc, false, GraphitiTimeoutPlaceholder,
		))

	case "disabled":
		// Disabled mode has no additional fields
	}

	m.SetFormFields(fields)
	return nil
}

func (m *GraphitiFormModel) createTextField(
	config *controller.GraphitiConfig, key, title, description string, masked bool, placeholder string,
) FormField {
	var envVar loader.EnvVar
	switch key {
	case "url":
		envVar = config.GraphitiURL
	case "timeout":
		envVar = config.Timeout
	case "model_name":
		envVar = config.ModelName
	case "neo4j_user":
		envVar = config.Neo4jUser
	case "neo4j_password":
		envVar = config.Neo4jPassword
	case "neo4j_database":
		envVar = config.Neo4jDatabase
	}

	input := NewTextInput(m.GetStyles(), m.GetWindow(), envVar)
	if placeholder != "" {
		input.Placeholder = placeholder
	}

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

func (m *GraphitiFormModel) GetFormTitle() string {
	return locale.MonitoringGraphitiFormTitle
}

func (m *GraphitiFormModel) GetFormDescription() string {
	return locale.MonitoringGraphitiFormDescription
}

func (m *GraphitiFormModel) GetFormName() string {
	return locale.MonitoringGraphitiFormName
}

func (m *GraphitiFormModel) GetFormSummary() string {
	return ""
}

func (m *GraphitiFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MonitoringGraphitiFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.MonitoringGraphitiFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.MonitoringGraphitiFormOverview))

	return strings.Join(sections, "\n")
}

func (m *GraphitiFormModel) GetCurrentConfiguration() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	config := m.GetController().GetGraphitiConfig()

	getMaskedValue := func(value string) string {
		maskedValue := strings.Repeat("*", len(value))
		if len(value) > 15 {
			maskedValue = maskedValue[:15] + "..."
		}
		return maskedValue
	}

	switch config.DeploymentType {
	case "embedded":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.MonitoringGraphitiEmbedded))
		if config.GraphitiURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiURL, m.GetStyles().Info.Render(config.GraphitiURL.Value)))
		}
		if timeout := config.Timeout.Value; timeout != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiTimeout, m.GetStyles().Info.Render(timeout)))
		} else if timeout := config.Timeout.Default; timeout != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiTimeout, m.GetStyles().Muted.Render(timeout)))
		}
		if modelName := config.ModelName.Value; modelName != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiModelName, m.GetStyles().Info.Render(modelName)))
		} else if modelName := config.ModelName.Default; modelName != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiModelName, m.GetStyles().Muted.Render(modelName)))
		}
		if neo4jUser := config.Neo4jUser.Value; neo4jUser != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiNeo4jUser, m.GetStyles().Info.Render(neo4jUser)))
		} else if neo4jUser := config.Neo4jUser.Default; neo4jUser != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiNeo4jUser, m.GetStyles().Muted.Render(neo4jUser)))
		}
		if neo4jPassword := config.Neo4jPassword.Value; neo4jPassword != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiNeo4jPassword, m.GetStyles().Muted.Render(getMaskedValue(neo4jPassword))))
		}
		if neo4jDatabase := config.Neo4jDatabase.Value; neo4jDatabase != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiNeo4jDatabase, m.GetStyles().Info.Render(neo4jDatabase)))
		} else if neo4jDatabase := config.Neo4jDatabase.Default; neo4jDatabase != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiNeo4jDatabase, m.GetStyles().Muted.Render(neo4jDatabase)))
		}

	case "external":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.MonitoringGraphitiExternal))
		if config.GraphitiURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiURL, m.GetStyles().Info.Render(config.GraphitiURL.Value)))
		}
		if timeout := config.Timeout.Value; timeout != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiTimeout, m.GetStyles().Info.Render(timeout)))
		} else if timeout := config.Timeout.Default; timeout != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringGraphitiTimeout, m.GetStyles().Muted.Render(timeout)))
		}

	case "disabled":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Warning.Render(locale.MonitoringGraphitiDisabled))
	}

	return strings.Join(sections, "\n")
}

func (m *GraphitiFormModel) IsConfigured() bool {
	config := m.GetController().GetGraphitiConfig()
	return config.DeploymentType != "disabled"
}

func (m *GraphitiFormModel) GetHelpContent() string {
	var sections []string
	deploymentType := m.getSelectedDeploymentType()

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MonitoringGraphitiFormTitle))
	sections = append(sections, "")
	sections = append(sections, locale.MonitoringGraphitiModeGuide)
	sections = append(sections, "")

	switch deploymentType {
	case "embedded":
		sections = append(sections, locale.MonitoringGraphitiEmbeddedHelp)
	case "external":
		sections = append(sections, locale.MonitoringGraphitiExternalHelp)
	case "disabled":
		sections = append(sections, locale.MonitoringGraphitiDisabledHelp)
	}

	return strings.Join(sections, "\n")
}

func (m *GraphitiFormModel) HandleSave() error {
	config := m.GetController().GetGraphitiConfig()
	deploymentType := m.getSelectedDeploymentType()
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.GraphitiConfig{
		DeploymentType: deploymentType,
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		GraphitiURL:   config.GraphitiURL,
		Timeout:       config.Timeout,
		ModelName:     config.ModelName,
		Neo4jUser:     config.Neo4jUser,
		Neo4jPassword: config.Neo4jPassword,
		Neo4jDatabase: config.Neo4jDatabase,
		Neo4jURI:      config.Neo4jURI,
		Installed:     config.Installed,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "url":
			newConfig.GraphitiURL.Value = value
		case "timeout":
			newConfig.Timeout.Value = value
		case "model_name":
			newConfig.ModelName.Value = value
		case "neo4j_user":
			newConfig.Neo4jUser.Value = value
		case "neo4j_password":
			newConfig.Neo4jPassword.Value = value
		case "neo4j_database":
			newConfig.Neo4jDatabase.Value = value
		}
	}

	// save the configuration
	if err := m.GetController().UpdateGraphitiConfig(newConfig); err != nil {
		logger.Errorf("[GraphitiFormModel] SAVE: error updating graphiti config: %v", err)
		return err
	}

	logger.Log("[GraphitiFormModel] SAVE: success")
	return nil
}

func (m *GraphitiFormModel) HandleReset() {
	// reset config to defaults
	config := m.GetController().ResetGraphitiConfig()

	// reset deployment selection
	m.GetListHelper().SelectByValue(&m.deploymentList, config.DeploymentType)

	// rebuild form with reset deployment type
	m.BuildForm()
}

func (m *GraphitiFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *GraphitiFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *GraphitiFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// BaseListHandler interface implementation

func (m *GraphitiFormModel) GetList() *list.Model {
	return &m.deploymentList
}

func (m *GraphitiFormModel) GetListDelegate() *BaseListDelegate {
	return m.deploymentDelegate
}

func (m *GraphitiFormModel) OnListSelectionChanged(oldSelection, newSelection string) {
	// rebuild form when deployment type changes
	m.BuildForm()
}

func (m *GraphitiFormModel) GetListTitle() string {
	return locale.MonitoringGraphitiDeploymentType
}

func (m *GraphitiFormModel) GetListDescription() string {
	return locale.MonitoringGraphitiDeploymentTypeDesc
}

// Update method - handle screen-specific input
func (m *GraphitiFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
var _ BaseScreenModel = (*GraphitiFormModel)(nil)
var _ BaseScreenHandler = (*GraphitiFormModel)(nil)
var _ BaseListHandler = (*GraphitiFormModel)(nil)
