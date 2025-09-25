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
	LangfuseBaseURLPlaceholder       = "https://cloud.langfuse.com"
	LangfuseProjectIDPlaceholder     = "cm000000000000000000000000"
	LangfusePublicKeyPlaceholder     = "pk-lf-00000000-0000-0000-0000-000000000000"
	LangfuseSecretKeyPlaceholder     = ""
	LangfuseAdminEmailPlaceholder    = "admin@pentagi.com"
	LangfuseAdminPasswordPlaceholder = ""
	LangfuseAdminNamePlaceholder     = "admin"
	LangfuseLicenseKeyPlaceholder    = "sk-lf-ee-xxxxxxxxxxxxxxxxxxxxxxxx"
)

// LangfuseFormModel represents the Langfuse configuration form
type LangfuseFormModel struct {
	*BaseScreen

	// screen-specific components
	deploymentList     list.Model
	deploymentDelegate *BaseListDelegate
}

// NewLangfuseFormModel creates a new Langfuse form model
func NewLangfuseFormModel(c controller.Controller, s styles.Styles, w window.Window) *LangfuseFormModel {
	m := &LangfuseFormModel{}

	m.BaseScreen = NewBaseScreen(c, s, w, m, m)
	m.initializeDeploymentList(s)

	return m
}

// initializeDeploymentList sets up the deployment type selection list
func (m *LangfuseFormModel) initializeDeploymentList(styles styles.Styles) {
	options := []BaseListOption{
		{Value: "embedded", Display: locale.MonitoringLangfuseEmbedded},
		{Value: "external", Display: locale.MonitoringLangfuseExternal},
		{Value: "disabled", Display: locale.MonitoringLangfuseDisabled},
	}

	m.deploymentDelegate = NewBaseListDelegate(
		styles.FormLabel.Align(lipgloss.Center),
		MinMenuWidth-6,
	)

	m.deploymentList = m.GetListHelper().CreateList(options, m.deploymentDelegate, MinMenuWidth-6, 3)

	config := m.GetController().GetLangfuseConfig()

	m.GetListHelper().SelectByValue(&m.deploymentList, config.DeploymentType)
}

// getSelectedDeploymentType returns the currently selected deployment type using the helper
func (m *LangfuseFormModel) getSelectedDeploymentType() string {
	selectedValue := m.GetListHelper().GetSelectedValue(&m.deploymentList)
	if selectedValue == "" {
		return "disabled"
	}

	return selectedValue
}

// BaseScreenHandler interface implementation

func (m *LangfuseFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetLangfuseConfig()
	fields := []FormField{}
	deploymentType := m.getSelectedDeploymentType()

	switch deploymentType {
	case "embedded":
		// Embedded mode - requires all fields including admin credentials
		fields = append(fields, m.createTextField(config, "listen_ip",
			locale.MonitoringLangfuseListenIP, locale.MonitoringLangfuseListenIPDesc, false, "",
		))
		fields = append(fields, m.createTextField(config, "listen_port",
			locale.MonitoringLangfuseListenPort, locale.MonitoringLangfuseListenPortDesc, false, "",
		))
		fields = append(fields, m.createTextField(config, "project_id",
			locale.MonitoringLangfuseProjectID, locale.MonitoringLangfuseProjectIDDesc, false, LangfuseProjectIDPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "public_key",
			locale.MonitoringLangfusePublicKey, locale.MonitoringLangfusePublicKeyDesc, true, LangfusePublicKeyPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "secret_key",
			locale.MonitoringLangfuseSecretKey, locale.MonitoringLangfuseSecretKeyDesc, true, LangfuseSecretKeyPlaceholder,
		))
		if !config.Installed {
			fields = append(fields, m.createTextField(config, "admin_email",
				locale.MonitoringLangfuseAdminEmail, locale.MonitoringLangfuseAdminEmailDesc, false, LangfuseAdminEmailPlaceholder,
			))
			fields = append(fields, m.createTextField(config, "admin_password",
				locale.MonitoringLangfuseAdminPassword, locale.MonitoringLangfuseAdminPasswordDesc, true, LangfuseAdminPasswordPlaceholder,
			))
			fields = append(fields, m.createTextField(config, "admin_name",
				locale.MonitoringLangfuseAdminName, locale.MonitoringLangfuseAdminNameDesc, false, LangfuseAdminNamePlaceholder,
			))
		}
		fields = append(fields, m.createTextField(config, "license_key",
			locale.MonitoringLangfuseLicenseKey, locale.MonitoringLangfuseLicenseKeyDesc, true, LangfuseLicenseKeyPlaceholder,
		))

	case "external":
		// External mode - requires connection details only
		fields = append(fields, m.createTextField(config, "base_url",
			locale.MonitoringLangfuseBaseURL, locale.MonitoringLangfuseBaseURLDesc, false, LangfuseBaseURLPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "project_id",
			locale.MonitoringLangfuseProjectID, locale.MonitoringLangfuseProjectIDDesc, false, LangfuseProjectIDPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "public_key",
			locale.MonitoringLangfusePublicKey, locale.MonitoringLangfusePublicKeyDesc, true, LangfusePublicKeyPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "secret_key",
			locale.MonitoringLangfuseSecretKey, locale.MonitoringLangfuseSecretKeyDesc, true, LangfuseSecretKeyPlaceholder,
		))

	case "disabled":
		// Disabled mode has no additional fields
	}

	m.SetFormFields(fields)
	return nil
}

func (m *LangfuseFormModel) createTextField(
	config *controller.LangfuseConfig, key, title, description string, masked bool, placeholder string,
) FormField {
	var envVar loader.EnvVar
	switch key {
	case "listen_ip":
		envVar = config.ListenIP
	case "listen_port":
		envVar = config.ListenPort
	case "base_url":
		envVar = config.BaseURL
	case "project_id":
		envVar = config.ProjectID
	case "public_key":
		envVar = config.PublicKey
	case "secret_key":
		envVar = config.SecretKey
	case "admin_email":
		envVar = config.AdminEmail
	case "admin_password":
		envVar = config.AdminPassword
	case "admin_name":
		envVar = config.AdminName
	case "license_key":
		envVar = config.LicenseKey
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

func (m *LangfuseFormModel) GetFormTitle() string {
	return locale.MonitoringLangfuseFormTitle
}

func (m *LangfuseFormModel) GetFormDescription() string {
	return locale.MonitoringLangfuseFormDescription
}

func (m *LangfuseFormModel) GetFormName() string {
	return locale.MonitoringLangfuseFormName
}

func (m *LangfuseFormModel) GetFormSummary() string {
	return ""
}

func (m *LangfuseFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MonitoringLangfuseFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.MonitoringLangfuseFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.MonitoringLangfuseFormOverview))

	return strings.Join(sections, "\n")
}

func (m *LangfuseFormModel) GetCurrentConfiguration() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	config := m.GetController().GetLangfuseConfig()

	getMaskedValue := func(value string) string {
		maskedValue := strings.Repeat("*", len(value))
		if len(value) > 15 {
			maskedValue = maskedValue[:15] + "..."
		}
		return maskedValue
	}

	switch config.DeploymentType {
	case "embedded":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.MonitoringLangfuseEmbedded))
		if listenIP := config.ListenIP.Value; listenIP != "" {
			listenIP = m.GetStyles().Info.Render(listenIP)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringLangfuseListenIP, listenIP))
		} else if listenIP := config.ListenIP.Default; listenIP != "" {
			listenIP = m.GetStyles().Muted.Render(listenIP)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringLangfuseListenIP, listenIP))
		}

		if listenPort := config.ListenPort.Value; listenPort != "" {
			listenPort = m.GetStyles().Info.Render(listenPort)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringLangfuseListenPort, listenPort))
		} else if listenPort := config.ListenPort.Default; listenPort != "" {
			listenPort = m.GetStyles().Muted.Render(listenPort)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringLangfuseListenPort, listenPort))
		}
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseBaseURL, m.GetStyles().Info.Render(config.BaseURL.Value)))
		}
		if config.ProjectID.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseProjectID, m.GetStyles().Info.Render(config.ProjectID.Value)))
		}
		if publicKey := config.PublicKey.Value; publicKey != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfusePublicKey, m.GetStyles().Muted.Render(getMaskedValue(publicKey))))
		}
		if secretKey := config.SecretKey.Value; secretKey != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseSecretKey, m.GetStyles().Muted.Render(getMaskedValue(secretKey))))
		}
		if config.AdminEmail.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseAdminEmail, m.GetStyles().Info.Render(config.AdminEmail.Value)))
		}
		if adminPassword := config.AdminPassword.Value; adminPassword != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseAdminPassword, m.GetStyles().Muted.Render(getMaskedValue(adminPassword))))
		}
		if config.AdminName.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseAdminName, m.GetStyles().Info.Render(config.AdminName.Value)))
		}

	case "external":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.MonitoringLangfuseExternal))
		if config.BaseURL.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseBaseURL, m.GetStyles().Info.Render(config.BaseURL.Value)))
		}
		if config.ProjectID.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseProjectID, m.GetStyles().Info.Render(config.ProjectID.Value)))
		}
		if publicKey := config.PublicKey.Value; publicKey != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfusePublicKey, m.GetStyles().Muted.Render(getMaskedValue(publicKey))))
		}
		if secretKey := config.SecretKey.Value; secretKey != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringLangfuseSecretKey, m.GetStyles().Muted.Render(getMaskedValue(secretKey))))
		}

	case "disabled":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Warning.Render(locale.MonitoringLangfuseDisabled))
	}

	return strings.Join(sections, "\n")
}

func (m *LangfuseFormModel) IsConfigured() bool {
	config := m.GetController().GetLangfuseConfig()
	return config.DeploymentType != "disabled"
}

func (m *LangfuseFormModel) GetHelpContent() string {
	var sections []string
	deploymentType := m.getSelectedDeploymentType()

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MonitoringLangfuseFormTitle))
	sections = append(sections, "")
	sections = append(sections, locale.MonitoringLangfuseModeGuide)
	sections = append(sections, "")

	switch deploymentType {
	case "embedded":
		sections = append(sections, locale.MonitoringLangfuseEmbeddedHelp)
	case "external":
		sections = append(sections, locale.MonitoringLangfuseExternalHelp)
	case "disabled":
		sections = append(sections, locale.MonitoringLangfuseDisabledHelp)
	}

	return strings.Join(sections, "\n")
}

func (m *LangfuseFormModel) HandleSave() error {
	config := m.GetController().GetLangfuseConfig()
	deploymentType := m.getSelectedDeploymentType()
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.LangfuseConfig{
		DeploymentType: deploymentType,
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		ListenIP:      config.ListenIP,
		ListenPort:    config.ListenPort,
		BaseURL:       config.BaseURL,
		ProjectID:     config.ProjectID,
		PublicKey:     config.PublicKey,
		SecretKey:     config.SecretKey,
		AdminEmail:    config.AdminEmail,
		AdminPassword: config.AdminPassword,
		AdminName:     config.AdminName,
		Installed:     config.Installed,
		LicenseKey:    config.LicenseKey,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "listen_ip":
			newConfig.ListenIP.Value = value
		case "listen_port":
			newConfig.ListenPort.Value = value
		case "base_url":
			newConfig.BaseURL.Value = value
		case "project_id":
			newConfig.ProjectID.Value = value
		case "public_key":
			newConfig.PublicKey.Value = value
		case "secret_key":
			newConfig.SecretKey.Value = value
		case "admin_email":
			newConfig.AdminEmail.Value = value
		case "admin_password":
			newConfig.AdminPassword.Value = value
		case "admin_name":
			newConfig.AdminName.Value = value
		case "license_key":
			newConfig.LicenseKey.Value = value
		}
	}

	// save the configuration
	if err := m.GetController().UpdateLangfuseConfig(newConfig); err != nil {
		logger.Errorf("[LangfuseFormModel] SAVE: error updating langfuse config: %v", err)
		return err
	}

	logger.Log("[LangfuseFormModel] SAVE: success")
	return nil
}

func (m *LangfuseFormModel) HandleReset() {
	// reset config to defaults
	config := m.GetController().ResetLangfuseConfig()

	// reset deployment selection
	m.GetListHelper().SelectByValue(&m.deploymentList, config.DeploymentType)

	// rebuild form with reset deployment type
	m.BuildForm()
}

func (m *LangfuseFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *LangfuseFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *LangfuseFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// BaseListHandler interface implementation

func (m *LangfuseFormModel) GetList() *list.Model {
	return &m.deploymentList
}

func (m *LangfuseFormModel) GetListDelegate() *BaseListDelegate {
	return m.deploymentDelegate
}

func (m *LangfuseFormModel) OnListSelectionChanged(oldSelection, newSelection string) {
	// rebuild form when deployment type changes
	m.BuildForm()
}

func (m *LangfuseFormModel) GetListTitle() string {
	return locale.MonitoringLangfuseDeploymentType
}

func (m *LangfuseFormModel) GetListDescription() string {
	return locale.MonitoringLangfuseDeploymentTypeDesc
}

// Update method - handle screen-specific input
func (m *LangfuseFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
var _ BaseScreenModel = (*LangfuseFormModel)(nil)
var _ BaseScreenHandler = (*LangfuseFormModel)(nil)
var _ BaseListHandler = (*LangfuseFormModel)(nil)
