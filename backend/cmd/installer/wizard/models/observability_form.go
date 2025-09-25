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

// ObservabilityFormModel represents the Observability configuration form
type ObservabilityFormModel struct {
	*BaseScreen

	// screen-specific components
	deploymentList     list.Model
	deploymentDelegate *BaseListDelegate
}

// NewObservabilityFormModel creates a new Observability form model
func NewObservabilityFormModel(c controller.Controller, s styles.Styles, w window.Window) *ObservabilityFormModel {
	m := &ObservabilityFormModel{}

	m.BaseScreen = NewBaseScreen(c, s, w, m, m)
	m.initializeDeploymentList(s)

	return m
}

// initializeDeploymentList sets up the deployment type selection list
func (m *ObservabilityFormModel) initializeDeploymentList(styles styles.Styles) {
	options := []BaseListOption{
		{Value: "embedded", Display: locale.MonitoringObservabilityEmbedded},
		{Value: "external", Display: locale.MonitoringObservabilityExternal},
		{Value: "disabled", Display: locale.MonitoringObservabilityDisabled},
	}

	m.deploymentDelegate = NewBaseListDelegate(
		styles.FormLabel.Align(lipgloss.Center),
		MinMenuWidth-6,
	)

	m.deploymentList = m.GetListHelper().CreateList(options, m.deploymentDelegate, MinMenuWidth-6, 3)

	config := m.GetController().GetObservabilityConfig()

	m.GetListHelper().SelectByValue(&m.deploymentList, config.DeploymentType)
}

// getSelectedDeploymentType returns the currently selected deployment type using the helper
func (m *ObservabilityFormModel) getSelectedDeploymentType() string {
	selectedValue := m.GetListHelper().GetSelectedValue(&m.deploymentList)
	if selectedValue == "" {
		return "disabled"
	}

	return selectedValue
}

// BaseScreenHandler interface implementation

func (m *ObservabilityFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetObservabilityConfig()
	deploymentType := m.getSelectedDeploymentType()
	fields := []FormField{}

	switch deploymentType {
	case "external":
		// External mode - requires external OpenTelemetry host
		fields = append(fields, m.createTextField(config, "otel_host",
			locale.MonitoringObservabilityOTelHost,
			locale.MonitoringObservabilityOTelHostDesc,
			"external-collector:8148",
		))

	case "embedded":
		// Embedded mode - expose listen settings for Grafana and OTel collector
		fields = append(fields, m.createTextField(config, "grafana_listen_ip",
			locale.MonitoringObservabilityGrafanaListenIP, locale.MonitoringObservabilityGrafanaListenIPDesc, ""))
		fields = append(fields, m.createTextField(config, "grafana_listen_port",
			locale.MonitoringObservabilityGrafanaListenPort, locale.MonitoringObservabilityGrafanaListenPortDesc, ""))
		fields = append(fields, m.createTextField(config, "otel_grpc_listen_ip",
			locale.MonitoringObservabilityOTelGrpcListenIP, locale.MonitoringObservabilityOTelGrpcListenIPDesc, ""))
		fields = append(fields, m.createTextField(config, "otel_grpc_listen_port",
			locale.MonitoringObservabilityOTelGrpcListenPort, locale.MonitoringObservabilityOTelGrpcListenPortDesc, ""))
		fields = append(fields, m.createTextField(config, "otel_http_listen_ip",
			locale.MonitoringObservabilityOTelHttpListenIP, locale.MonitoringObservabilityOTelHttpListenIPDesc, ""))
		fields = append(fields, m.createTextField(config, "otel_http_listen_port",
			locale.MonitoringObservabilityOTelHttpListenPort, locale.MonitoringObservabilityOTelHttpListenPortDesc, ""))

	case "disabled":
		// Disabled mode has no additional fields
	}

	m.SetFormFields(fields)
	return nil
}

func (m *ObservabilityFormModel) createTextField(
	config *controller.ObservabilityConfig, key, title, description string, placeholder string,
) FormField {
	var envVar loader.EnvVar
	switch key {
	case "otel_host":
		envVar = config.OTelHost
	case "grafana_listen_ip":
		envVar = config.GrafanaListenIP
	case "grafana_listen_port":
		envVar = config.GrafanaListenPort
	case "otel_grpc_listen_ip":
		envVar = config.OTelGrpcListenIP
	case "otel_grpc_listen_port":
		envVar = config.OTelGrpcListenPort
	case "otel_http_listen_ip":
		envVar = config.OTelHttpListenIP
	case "otel_http_listen_port":
		envVar = config.OTelHttpListenPort
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
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *ObservabilityFormModel) GetFormTitle() string {
	return locale.MonitoringObservabilityFormTitle
}

func (m *ObservabilityFormModel) GetFormDescription() string {
	return locale.MonitoringObservabilityFormDescription
}

func (m *ObservabilityFormModel) GetFormName() string {
	return locale.MonitoringObservabilityFormName
}

func (m *ObservabilityFormModel) GetFormSummary() string {
	return ""
}

func (m *ObservabilityFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MonitoringObservabilityFormName))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.MonitoringObservabilityFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.MonitoringObservabilityFormOverview))

	return strings.Join(sections, "\n")
}

func (m *ObservabilityFormModel) GetCurrentConfiguration() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	config := m.GetController().GetObservabilityConfig()

	switch config.DeploymentType {
	case "embedded":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.StatusEmbedded))

		if listenIP := config.GrafanaListenIP.Value; listenIP != "" {
			listenIP = m.GetStyles().Info.Render(listenIP)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityGrafanaListenIP, listenIP))
		} else if listenIP := config.GrafanaListenIP.Default; listenIP != "" {
			listenIP = m.GetStyles().Muted.Render(listenIP)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityGrafanaListenIP, listenIP))
		}

		if listenPort := config.GrafanaListenPort.Value; listenPort != "" {
			listenPort = m.GetStyles().Info.Render(listenPort)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityGrafanaListenPort, listenPort))
		} else if listenPort := config.GrafanaListenPort.Default; listenPort != "" {
			listenPort = m.GetStyles().Muted.Render(listenPort)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityGrafanaListenPort, listenPort))
		}

		if listenIP := config.OTelGrpcListenIP.Value; listenIP != "" {
			listenIP = m.GetStyles().Info.Render(listenIP)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityOTelGrpcListenIP, listenIP))
		} else if listenIP := config.OTelGrpcListenIP.Default; listenIP != "" {
			listenIP = m.GetStyles().Muted.Render(listenIP)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityOTelGrpcListenIP, listenIP))
		}

		if listenPort := config.OTelGrpcListenPort.Value; listenPort != "" {
			listenPort = m.GetStyles().Info.Render(listenPort)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityOTelGrpcListenPort, listenPort))
		} else if listenPort := config.OTelGrpcListenPort.Default; listenPort != "" {
			listenPort = m.GetStyles().Muted.Render(listenPort)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityOTelGrpcListenPort, listenPort))
		}

		if listenIP := config.OTelHttpListenIP.Value; listenIP != "" {
			listenIP = m.GetStyles().Info.Render(listenIP)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityOTelHttpListenIP, listenIP))
		} else if listenIP := config.OTelHttpListenIP.Default; listenIP != "" {
			listenIP = m.GetStyles().Muted.Render(listenIP)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityOTelHttpListenIP, listenIP))
		}

		if listenPort := config.OTelHttpListenPort.Value; listenPort != "" {
			listenPort = m.GetStyles().Info.Render(listenPort)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityOTelHttpListenPort, listenPort))
		} else if listenPort := config.OTelHttpListenPort.Default; listenPort != "" {
			listenPort = m.GetStyles().Muted.Render(listenPort)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.MonitoringObservabilityOTelHttpListenPort, listenPort))
		}

	case "external":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.StatusExternal))
		if config.OTelHost.Value != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s",
				locale.MonitoringObservabilityOTelHost, m.GetStyles().Info.Render(config.OTelHost.Value)))
		}

	case "disabled":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Warning.Render(locale.StatusDisabled))
	}

	return strings.Join(sections, "\n")
}

func (m *ObservabilityFormModel) IsConfigured() bool {
	return m.GetController().GetObservabilityConfig().DeploymentType != "disabled"
}

func (m *ObservabilityFormModel) GetHelpContent() string {
	var sections []string
	deploymentType := m.getSelectedDeploymentType()

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MonitoringObservabilityFormTitle))
	sections = append(sections, "")
	sections = append(sections, locale.MonitoringObservabilityModeGuide)
	sections = append(sections, "")

	switch deploymentType {
	case "embedded":
		sections = append(sections, locale.MonitoringObservabilityEmbeddedHelp)
	case "external":
		sections = append(sections, locale.MonitoringObservabilityExternalHelp)
	case "disabled":
		sections = append(sections, locale.MonitoringObservabilityDisabledHelp)
	}

	return strings.Join(sections, "\n")
}

func (m *ObservabilityFormModel) HandleSave() error {
	config := m.GetController().GetObservabilityConfig()
	deploymentType := m.getSelectedDeploymentType()
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.ObservabilityConfig{
		DeploymentType: deploymentType,
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		OTelHost:           config.OTelHost,
		GrafanaListenIP:    config.GrafanaListenIP,
		GrafanaListenPort:  config.GrafanaListenPort,
		OTelGrpcListenIP:   config.OTelGrpcListenIP,
		OTelGrpcListenPort: config.OTelGrpcListenPort,
		OTelHttpListenIP:   config.OTelHttpListenIP,
		OTelHttpListenPort: config.OTelHttpListenPort,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "otel_host":
			newConfig.OTelHost.Value = value
		case "grafana_listen_ip":
			newConfig.GrafanaListenIP.Value = value
		case "grafana_listen_port":
			newConfig.GrafanaListenPort.Value = value
		case "otel_grpc_listen_ip":
			newConfig.OTelGrpcListenIP.Value = value
		case "otel_grpc_listen_port":
			newConfig.OTelGrpcListenPort.Value = value
		case "otel_http_listen_ip":
			newConfig.OTelHttpListenIP.Value = value
		case "otel_http_listen_port":
			newConfig.OTelHttpListenPort.Value = value
		}
	}

	// save the configuration
	if err := m.GetController().UpdateObservabilityConfig(newConfig); err != nil {
		logger.Errorf("[ObservabilityFormModel] SAVE: error updating observability config: %v", err)
		return err
	}

	logger.Log("[ObservabilityFormModel] SAVE: success")
	return nil
}

func (m *ObservabilityFormModel) HandleReset() {
	// reset config to defaults
	config := m.GetController().ResetObservabilityConfig()

	// reset deployment selection
	m.GetListHelper().SelectByValue(&m.deploymentList, config.DeploymentType)

	// rebuild form with reset deployment type
	m.BuildForm()
}

func (m *ObservabilityFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *ObservabilityFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *ObservabilityFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// BaseListHandler interface implementation

func (m *ObservabilityFormModel) GetList() *list.Model {
	return &m.deploymentList
}

func (m *ObservabilityFormModel) GetListDelegate() *BaseListDelegate {
	return m.deploymentDelegate
}

func (m *ObservabilityFormModel) OnListSelectionChanged(oldSelection, newSelection string) {
	// rebuild form when deployment type changes
	m.BuildForm()
}

func (m *ObservabilityFormModel) GetListTitle() string {
	return locale.MonitoringObservabilityDeploymentType
}

func (m *ObservabilityFormModel) GetListDescription() string {
	return locale.MonitoringObservabilityDeploymentTypeDesc
}

// Update method - handle screen-specific input
func (m *ObservabilityFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
var _ BaseScreenModel = (*ObservabilityFormModel)(nil)
var _ BaseScreenHandler = (*ObservabilityFormModel)(nil)
var _ BaseListHandler = (*ObservabilityFormModel)(nil)
