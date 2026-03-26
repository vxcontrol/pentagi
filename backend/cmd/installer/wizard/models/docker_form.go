package models

import (
	"fmt"
	"os"
	"strings"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// DockerFormModel represents the Docker Environment configuration form
type DockerFormModel struct {
	*BaseScreen
}

// NewDockerFormModel creates a new Docker Environment form model
func NewDockerFormModel(c controller.Controller, s styles.Styles, w window.Window) *DockerFormModel {
	m := &DockerFormModel{}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// BaseScreenHandler interface implementation

func (m *DockerFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetDockerConfig()
	fields := []FormField{}

	// Container capabilities
	fields = append(fields, m.createBooleanField("docker_inside",
		locale.ToolsDockerInside,
		locale.ToolsDockerInsideDesc,
		config.DockerInside,
	))

	fields = append(fields, m.createBooleanField("docker_net_admin",
		locale.ToolsDockerNetAdmin,
		locale.ToolsDockerNetAdminDesc,
		config.DockerNetAdmin,
	))

	// Connection settings
	fields = append(fields, m.createTextField("docker_socket",
		locale.ToolsDockerSocket,
		locale.ToolsDockerSocketDesc,
		config.DockerSocket,
		false,
	))

	fields = append(fields, m.createTextField("docker_network",
		locale.ToolsDockerNetwork,
		locale.ToolsDockerNetworkDesc,
		config.DockerNetwork,
		false,
	))

	fields = append(fields, m.createTextField("docker_public_ip",
		locale.ToolsDockerPublicIP,
		locale.ToolsDockerPublicIPDesc,
		config.DockerPublicIP,
		false,
	))

	// Storage configuration
	fields = append(fields, m.createTextField("docker_work_dir",
		locale.ToolsDockerWorkDir,
		locale.ToolsDockerWorkDirDesc,
		config.DockerWorkDir,
		false,
	))

	// Default images
	fields = append(fields, m.createTextField("docker_default_image",
		locale.ToolsDockerDefaultImage,
		locale.ToolsDockerDefaultImageDesc,
		config.DockerDefaultImage,
		false,
	))

	fields = append(fields, m.createTextField("docker_default_image_for_pentest",
		locale.ToolsDockerDefaultImageForPentest,
		locale.ToolsDockerDefaultImageForPentestDesc,
		config.DockerDefaultImageForPentest,
		false,
	))

	// TLS connection settings (optional)
	fields = append(fields, m.createTextField("docker_host",
		locale.ToolsDockerHost,
		locale.ToolsDockerHostDesc,
		config.DockerHost,
		false,
	))

	fields = append(fields, m.createBooleanField("docker_tls_verify",
		locale.ToolsDockerTLSVerify,
		locale.ToolsDockerTLSVerifyDesc,
		config.DockerTLSVerify,
	))

	fields = append(fields, m.createTextField("docker_cert_path",
		locale.ToolsDockerCertPath,
		locale.ToolsDockerCertPathDesc,
		config.HostDockerCertPath,
		false,
	))

	m.SetFormFields(fields)
	return nil
}

func (m *DockerFormModel) createBooleanField(key, title, description string, envVar loader.EnvVar) FormField {
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

func (m *DockerFormModel) createTextField(key, title, description string, envVar loader.EnvVar, masked bool) FormField {
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

func (m *DockerFormModel) GetFormTitle() string {
	return locale.ToolsDockerFormTitle
}

func (m *DockerFormModel) GetFormDescription() string {
	return locale.ToolsDockerFormDescription
}

func (m *DockerFormModel) GetFormName() string {
	return locale.ToolsDockerFormName
}

func (m *DockerFormModel) GetFormSummary() string {
	return ""
}

func (m *DockerFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ToolsDockerFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ToolsDockerFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ToolsDockerFormOverview))

	return strings.Join(sections, "\n")
}

func (m *DockerFormModel) GetCurrentConfiguration() string {
	var sections []string

	config := m.GetController().GetDockerConfig()

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	// Container capabilities
	dockerInside := config.DockerInside.Value
	if dockerInside == "" {
		dockerInside = config.DockerInside.Default
	}
	if dockerInside == "true" {
		sections = append(sections, fmt.Sprintf("• Docker Access: %s",
			m.GetStyles().Success.Render(locale.StatusEnabled)))
	} else {
		sections = append(sections, fmt.Sprintf("• Docker Access: %s",
			m.GetStyles().Warning.Render(locale.StatusDisabled)))
	}

	dockerNetAdmin := config.DockerNetAdmin.Value
	if dockerNetAdmin == "" {
		dockerNetAdmin = config.DockerNetAdmin.Default
	}
	if dockerNetAdmin == "true" {
		sections = append(sections, fmt.Sprintf("• Network Admin: %s",
			m.GetStyles().Success.Render(locale.StatusEnabled)))
	} else {
		sections = append(sections, fmt.Sprintf("• Network Admin: %s",
			m.GetStyles().Warning.Render(locale.StatusDisabled)))
	}

	// Connection settings
	if config.DockerNetwork.Value != "" {
		sections = append(sections, fmt.Sprintf("• Custom Network: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Custom Network: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	if config.DockerPublicIP.Value != "" && config.DockerPublicIP.Value != "0.0.0.0" {
		sections = append(sections, fmt.Sprintf("• Public IP: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Public IP: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	// Default images
	if config.DockerDefaultImage.Value != "" {
		sections = append(sections, fmt.Sprintf("• Default Image: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Default Image: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	if config.DockerDefaultImageForPentest.Value != "" {
		sections = append(sections, fmt.Sprintf("• Pentest Image: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else {
		sections = append(sections, fmt.Sprintf("• Pentest Image: %s",
			m.GetStyles().Warning.Render(locale.StatusNotConfigured)))
	}

	// TLS settings
	if config.DockerHost.Value != "" && config.DockerTLSVerify.Value == "1" && config.HostDockerCertPath.Value != "" {
		sections = append(sections, fmt.Sprintf("• TLS Connection: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	} else if config.DockerHost.Value != "" {
		sections = append(sections, fmt.Sprintf("• Remote Connection: %s",
			m.GetStyles().Success.Render(locale.StatusConfigured)))
	}

	sections = append(sections, "")
	if config.Configured {
		sections = append(sections, m.GetStyles().Success.Render(locale.MessageDockerConfigured))
	} else {
		sections = append(sections, m.GetStyles().Warning.Render(locale.MessageDockerNotConfigured))
	}

	return strings.Join(sections, "\n")
}

func (m *DockerFormModel) IsConfigured() bool {
	return m.GetController().GetDockerConfig().Configured
}

func (m *DockerFormModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ToolsDockerFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ToolsDockerFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ToolsDockerGeneralHelp))
	sections = append(sections, "")

	// Show field-specific help based on focused field
	fieldIndex := m.GetFocusedIndex()
	fields := m.GetFormFields()

	if fieldIndex >= 0 && fieldIndex < len(fields) {
		field := fields[fieldIndex]

		switch field.Key {
		case "docker_inside":
			sections = append(sections, locale.ToolsDockerInsideHelp)
		case "docker_net_admin":
			sections = append(sections, locale.ToolsDockerNetAdminHelp)
		case "docker_socket":
			sections = append(sections, locale.ToolsDockerSocketHelp)
		case "docker_network":
			sections = append(sections, locale.ToolsDockerNetworkHelp)
		case "docker_public_ip":
			sections = append(sections, locale.ToolsDockerPublicIPHelp)
		case "docker_work_dir":
			sections = append(sections, locale.ToolsDockerWorkDirHelp)
		case "docker_default_image":
			sections = append(sections, locale.ToolsDockerDefaultImageHelp)
		case "docker_default_image_for_pentest":
			sections = append(sections, locale.ToolsDockerDefaultImageForPentestHelp)
		case "docker_host":
			sections = append(sections, locale.ToolsDockerHostHelp)
		case "docker_tls_verify":
			sections = append(sections, locale.ToolsDockerTLSVerifyHelp)
		case "docker_cert_path":
			sections = append(sections, locale.ToolsDockerCertPathHelp)
		default:
			sections = append(sections, locale.ToolsDockerFormOverview)
		}
	}

	return strings.Join(sections, "\n")
}

func (m *DockerFormModel) HandleSave() error {
	config := m.GetController().GetDockerConfig()
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.DockerConfig{
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		DockerInside:                 config.DockerInside,
		DockerNetAdmin:               config.DockerNetAdmin,
		DockerSocket:                 config.DockerSocket,
		DockerNetwork:                config.DockerNetwork,
		DockerPublicIP:               config.DockerPublicIP,
		DockerWorkDir:                config.DockerWorkDir,
		DockerDefaultImage:           config.DockerDefaultImage,
		DockerDefaultImageForPentest: config.DockerDefaultImageForPentest,
		DockerHost:                   config.DockerHost,
		DockerTLSVerify:              config.DockerTLSVerify,
		HostDockerCertPath:           config.HostDockerCertPath,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "docker_inside":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for Docker Access: %s (must be 'true' or 'false')", value)
			}
			newConfig.DockerInside.Value = value
		case "docker_net_admin":
			// validate boolean input
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid boolean value for Network Admin: %s (must be 'true' or 'false')", value)
			}
			newConfig.DockerNetAdmin.Value = value
		case "docker_socket":
			newConfig.DockerSocket.Value = value
		case "docker_network":
			newConfig.DockerNetwork.Value = value
		case "docker_public_ip":
			newConfig.DockerPublicIP.Value = value
		case "docker_work_dir":
			newConfig.DockerWorkDir.Value = value
		case "docker_default_image":
			newConfig.DockerDefaultImage.Value = value
		case "docker_default_image_for_pentest":
			newConfig.DockerDefaultImageForPentest.Value = value
		case "docker_host":
			newConfig.DockerHost.Value = value
		case "docker_tls_verify":
			// validate boolean input for TLS verification
			if value != "" && value != "true" && value != "false" && value != "1" && value != "0" {
				return fmt.Errorf("invalid boolean value for TLS Verification: %s (must be 'true', 'false', '1', or '0')", value)
			}
			// normalize to "1" or "" for TLS verification
			switch value {
			case "true":
				value = "1"
			case "false":
				value = ""
			}
			newConfig.DockerTLSVerify.Value = value
		case "docker_cert_path":
			// validate cert path if provided
			if value != "" {
				info, err := os.Stat(value)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("docker cert path does not exist: %s", value)
					}
					return fmt.Errorf("cannot access docker cert path %s: %v", value, err)
				}
				if !info.IsDir() {
					return fmt.Errorf("docker cert path must be a directory, not a file: %s", value)
				}
			}
			newConfig.HostDockerCertPath.Value = value
		}
	}

	// save the configuration
	if err := m.GetController().UpdateDockerConfig(newConfig); err != nil {
		logger.Errorf("[DockerFormModel] SAVE: error updating Docker config: %v", err)
		return err
	}

	logger.Log("[DockerFormModel] SAVE: success")
	return nil
}

func (m *DockerFormModel) HandleReset() {
	// reset config to defaults
	m.GetController().ResetDockerConfig()

	// rebuild form with reset values
	m.BuildForm()
}

func (m *DockerFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *DockerFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *DockerFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// Update method - handle screen-specific input
func (m *DockerFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
var _ BaseScreenModel = (*DockerFormModel)(nil)
var _ BaseScreenHandler = (*DockerFormModel)(nil)
