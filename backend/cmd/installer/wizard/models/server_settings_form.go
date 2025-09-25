package models

import (
	"fmt"
	"strconv"
	"strings"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// ServerSettingsFormModel represents the PentAGI server settings configuration form
type ServerSettingsFormModel struct {
	*BaseScreen
}

// NewServerSettingsFormModel creates a new server settings form model
func NewServerSettingsFormModel(c controller.Controller, s styles.Styles, w window.Window) *ServerSettingsFormModel {
	m := &ServerSettingsFormModel{}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// BuildForm constructs the fields for server settings
func (m *ServerSettingsFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetServerSettingsConfig()
	fields := []FormField{}

	// host and port
	fields = append(fields, m.createTextField("pentagi_server_host",
		locale.ServerSettingsHost,
		locale.ServerSettingsHostDesc,
		config.ListenIP,
		false,
	))

	fields = append(fields, m.createTextField("pentagi_server_port",
		locale.ServerSettingsPort,
		locale.ServerSettingsPortDesc,
		config.ListenPort,
		false,
	))

	// public url
	fields = append(fields, m.createTextField("pentagi_public_url",
		locale.ServerSettingsPublicURL,
		locale.ServerSettingsPublicURLDesc,
		config.PublicURL,
		false,
	))

	// cors origins
	fields = append(fields, m.createTextField("pentagi_cors_origins",
		locale.ServerSettingsCORSOrigins,
		locale.ServerSettingsCORSOriginsDesc,
		config.CorsOrigins,
		false,
	))

	// proxy: url, username, password
	fields = append(fields, m.createTextField("proxy_url",
		locale.ServerSettingsProxyURL,
		locale.ServerSettingsProxyURLDesc,
		config.ProxyURL,
		false,
	))
	fields = append(fields, m.createRawField("proxy_username",
		locale.ServerSettingsProxyUsername,
		locale.ServerSettingsProxyUsernameDesc,
		config.ProxyUsername,
		true,
	))
	fields = append(fields, m.createRawField("proxy_password",
		locale.ServerSettingsProxyPassword,
		locale.ServerSettingsProxyPasswordDesc,
		config.ProxyPassword,
		true,
	))

	// ssl dir
	fields = append(fields, m.createTextField("pentagi_ssl_dir",
		locale.ServerSettingsSSLDir,
		locale.ServerSettingsSSLDirDesc,
		config.SSLDir,
		false,
	))

	// data dir
	fields = append(fields, m.createTextField("pentagi_data_dir",
		locale.ServerSettingsDataDir,
		locale.ServerSettingsDataDirDesc,
		config.DataDir,
		false,
	))

	// cookie signing salt (masked)
	fields = append(fields, m.createTextField("pentagi_cookie_signing_salt",
		locale.ServerSettingsCookieSigningSalt,
		locale.ServerSettingsCookieSigningSaltDesc,
		config.CookieSigningSalt,
		true,
	))

	m.SetFormFields(fields)
	return nil
}

func (m *ServerSettingsFormModel) createTextField(key, title, description string, envVar loader.EnvVar, masked bool) FormField {
	// reuse generic text input builder
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

// createRawField is used for non-env raw values (like usernames/passwords parsed from URLs)
func (m *ServerSettingsFormModel) createRawField(key, title, description, value string, masked bool) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), loader.EnvVar{Value: value})
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

func (m *ServerSettingsFormModel) GetFormTitle() string {
	return locale.ServerSettingsFormTitle
}

func (m *ServerSettingsFormModel) GetFormDescription() string {
	return locale.ServerSettingsFormDescription
}

func (m *ServerSettingsFormModel) GetFormName() string {
	return locale.ServerSettingsFormName
}

func (m *ServerSettingsFormModel) GetFormSummary() string {
	return ""
}

func (m *ServerSettingsFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ServerSettingsFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ServerSettingsFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ServerSettingsFormOverview))

	return strings.Join(sections, "\n")
}

func (m *ServerSettingsFormModel) GetCurrentConfiguration() string {
	var sections []string
	cfg := m.GetController().GetServerSettingsConfig()

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	getMaskedValue := func(value string) string {
		maskedValue := strings.Repeat("*", len(value))
		if len(value) > 15 {
			maskedValue = maskedValue[:15] + "..."
		}
		return maskedValue
	}

	if listenIP := cfg.ListenIP.Value; listenIP != "" {
		listenIP = m.GetStyles().Info.Render(listenIP)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsHostHint, listenIP))
	} else if listenIP := cfg.ListenIP.Default; listenIP != "" {
		listenIP = m.GetStyles().Muted.Render(listenIP)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsHostHint, listenIP))
	}

	if listenPort := cfg.ListenPort.Value; listenPort != "" {
		listenPort = m.GetStyles().Info.Render(listenPort)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPortHint, listenPort))
	} else if listenPort := cfg.ListenPort.Default; listenPort != "" {
		listenPort = m.GetStyles().Muted.Render(listenPort)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPortHint, listenPort))
	}

	if publicURL := cfg.PublicURL.Value; publicURL != "" {
		publicURL = m.GetStyles().Info.Render(publicURL)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPublicURLHint, publicURL))
	} else if publicURL := cfg.PublicURL.Default; publicURL != "" {
		publicURL = m.GetStyles().Muted.Render(publicURL)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPublicURLHint, publicURL))
	}

	if cors := cfg.CorsOrigins.Value; cors != "" {
		cors = m.GetStyles().Info.Render(cors)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsCORSOriginsHint, cors))
	} else if cors := cfg.CorsOrigins.Default; cors != "" {
		cors = m.GetStyles().Muted.Render(cors)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsCORSOriginsHint, cors))
	}

	if proxyURL := cfg.ProxyURL.Value; proxyURL != "" {
		proxyURL = m.GetStyles().Info.Render(proxyURL)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsProxyURLHint, proxyURL))
	} else {
		proxyURL = locale.StatusNotConfigured
		proxyURL = m.GetStyles().Muted.Render(proxyURL)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsProxyURLHint, proxyURL))
	}

	if proxyUsername := getMaskedValue(cfg.ProxyUsername); proxyUsername != "" {
		proxyUsername = m.GetStyles().Muted.Render(proxyUsername)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsProxyUsernameHint, proxyUsername))
	}

	if proxyPassword := getMaskedValue(cfg.ProxyPassword); proxyPassword != "" {
		proxyPassword = m.GetStyles().Muted.Render(proxyPassword)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsProxyPasswordHint, proxyPassword))
	}

	if sslDir := cfg.SSLDir.Value; sslDir != "" {
		sslDir = m.GetStyles().Info.Render(sslDir)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsSSLDirHint, sslDir))
	} else if sslDir := cfg.SSLDir.Default; sslDir != "" {
		sslDir = m.GetStyles().Muted.Render(sslDir)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsSSLDirHint, sslDir))
	}

	if dataDir := cfg.DataDir.Value; dataDir != "" {
		dataDir = m.GetStyles().Info.Render(dataDir)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsDataDirHint, dataDir))
	} else if dataDir := cfg.DataDir.Default; dataDir != "" {
		dataDir = m.GetStyles().Muted.Render(dataDir)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsDataDirHint, dataDir))
	}

	return strings.Join(sections, "\n")
}

func (m *ServerSettingsFormModel) IsConfigured() bool {
	cfg := m.GetController().GetServerSettingsConfig()
	return cfg.ListenIP.Value != "" && cfg.ListenPort.Value != ""
}

func (m *ServerSettingsFormModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ServerSettingsFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ServerSettingsFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ServerSettingsGeneralHelp))
	sections = append(sections, "")

	fieldIndex := m.GetFocusedIndex()
	fields := m.GetFormFields()

	if fieldIndex >= 0 && fieldIndex < len(fields) {
		field := fields[fieldIndex]
		switch field.Key {
		case "pentagi_server_host":
			sections = append(sections, locale.ServerSettingsHostHelp)
		case "pentagi_server_port":
			sections = append(sections, locale.ServerSettingsPortHelp)
		case "pentagi_public_url":
			sections = append(sections, locale.ServerSettingsPublicURLHelp)
		case "pentagi_cors_origins":
			sections = append(sections, locale.ServerSettingsCORSOriginsHelp)
		case "proxy_url":
			sections = append(sections, locale.ServerSettingsProxyURLHelp)
		case "pentagi_ssl_dir":
			sections = append(sections, locale.ServerSettingsSSLDirHelp)
		case "pentagi_data_dir":
			sections = append(sections, locale.ServerSettingsDataDirHelp)
		case "pentagi_cookie_signing_salt":
			sections = append(sections, locale.ServerSettingsCookieSigningSaltHelp)
		default:
			sections = append(sections, locale.ServerSettingsFormOverview)
		}
	}

	return strings.Join(sections, "\n")
}

func (m *ServerSettingsFormModel) HandleSave() error {
	cfg := m.GetController().GetServerSettingsConfig()
	fields := m.GetFormFields()

	newCfg := &controller.ServerSettingsConfig{
		ListenIP:          cfg.ListenIP,
		ListenPort:        cfg.ListenPort,
		CorsOrigins:       cfg.CorsOrigins,
		CookieSigningSalt: cfg.CookieSigningSalt,
		ProxyURL:          cfg.ProxyURL,
		SSLDir:            cfg.SSLDir,
		DataDir:           cfg.DataDir,
		PublicURL:         cfg.PublicURL,
	}

	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "pentagi_server_host":
			newCfg.ListenIP.Value = value
		case "pentagi_server_port":
			if value != "" {
				if _, err := strconv.Atoi(value); err != nil {
					return fmt.Errorf("invalid port: %s", value)
				}
			}
			newCfg.ListenPort.Value = value
		case "pentagi_public_url":
			newCfg.PublicURL.Value = value
		case "pentagi_cors_origins":
			newCfg.CorsOrigins.Value = value
		case "proxy_url":
			newCfg.ProxyURL.Value = value
		case "proxy_username":
			newCfg.ProxyUsername = value
		case "proxy_password":
			newCfg.ProxyPassword = value
		case "pentagi_ssl_dir":
			newCfg.SSLDir.Value = value
		case "pentagi_data_dir":
			newCfg.DataDir.Value = value
		case "pentagi_cookie_signing_salt":
			newCfg.CookieSigningSalt.Value = value
		}
	}

	if err := m.GetController().UpdateServerSettingsConfig(newCfg); err != nil {
		logger.Errorf("[ServerSettingsFormModel] SAVE: error updating server settings: %v", err)
		return err
	}

	logger.Log("[ServerSettingsFormModel] SAVE: success")
	return nil
}

func (m *ServerSettingsFormModel) HandleReset() {
	m.GetController().ResetServerSettingsConfig()
	m.BuildForm()
}

func (m *ServerSettingsFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// no-op for now
}

func (m *ServerSettingsFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *ServerSettingsFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// Update handles screen-specific input, then delegates to base screen
func (m *ServerSettingsFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := m.HandleFieldInput(msg); cmd != nil {
			return m, cmd
		}
	}

	cmd := m.BaseScreen.Update(msg)
	return m, cmd
}

// compile-time interface validation
var _ BaseScreenModel = (*ServerSettingsFormModel)(nil)
var _ BaseScreenHandler = (*ServerSettingsFormModel)(nil)
