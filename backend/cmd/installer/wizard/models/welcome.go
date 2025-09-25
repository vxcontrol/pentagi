package models

import (
	"fmt"
	"strings"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	MinTerminalWidth   = 80 // Minimum width for horizontal layout
	MinPanelWidth      = 25 // Minimum width for left/right panels
	MinRightPanelWidth = 35 // Minimum width for info panel
)

type WelcomeModel struct {
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
	viewport   viewport.Model
	ready      bool
}

func NewWelcomeModel(c controller.Controller, s styles.Styles, w window.Window) *WelcomeModel {
	return &WelcomeModel{
		controller: c,
		styles:     s,
		window:     w,
	}
}

func (m *WelcomeModel) Init() tea.Cmd {
	m.ready = false // to fit viewport to the window size with correct header height
	return nil
}

func (m *WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateViewport()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.controller.GetChecker().IsReadyToContinue() {
				if m.controller.GetEulaConsent() {
					return m, func() tea.Msg { return NavigationMsg{Target: MainMenuScreen} }
				} else {
					return m, func() tea.Msg { return NavigationMsg{Target: EULAScreen} }
				}
			}
		default:
			if !m.ready {
				break
			}
			switch msg.String() {
			case "up":
				m.viewport.ScrollUp(1)
			case "down":
				m.viewport.ScrollDown(1)
			case "left":
				m.viewport.ScrollLeft(2)
			case "right":
				m.viewport.ScrollRight(2)
			case "pgup":
				m.viewport.PageUp()
			case "pgdown":
				m.viewport.PageDown()
			}
		}
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *WelcomeModel) updateViewport() {
	// Use window manager for content dimensions
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return
	}

	if !m.ready {
		m.viewport = viewport.New(contentWidth, contentHeight)
		m.viewport.Style = lipgloss.NewStyle()
		m.ready = true
	} else {
		m.viewport.Width = contentWidth
		m.viewport.Height = contentHeight
	}

	m.viewport.SetContent(m.renderContent())
}

func (m *WelcomeModel) View() string {
	// Use window manager for content dimensions
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return locale.UILoading
	}

	// Ensure viewport is ready
	if !m.ready {
		m.updateViewport()
	}

	if !m.ready {
		return locale.UILoading
	}

	return m.viewport.View()
}

// renderContent prepares all content for the viewport
func (m *WelcomeModel) renderContent() string {
	leftPanel := m.renderSystemChecks()
	rightPanel := m.renderInfoPanel()

	if m.isVerticalLayout() {
		return m.renderVerticalLayout(leftPanel, rightPanel)
	}

	return m.renderHorizontalLayout(leftPanel, rightPanel)
}

func (m *WelcomeModel) renderVerticalLayout(leftPanel, rightPanel string) string {
	contentWidth := m.window.GetContentWidth()
	verticalStyle := lipgloss.NewStyle().Width(contentWidth).Padding(0, 2, 0, 2)

	return lipgloss.JoinVertical(lipgloss.Left,
		verticalStyle.Render(leftPanel),
		verticalStyle.Height(1).Render(""),
		verticalStyle.Render(rightPanel),
	)
}

func (m *WelcomeModel) renderHorizontalLayout(leftPanel, rightPanel string) string {
	contentWidth := m.window.GetContentWidth()
	leftWidth := contentWidth / 3
	rightWidth := contentWidth - leftWidth - 4
	leftWidth = max(leftWidth, MinPanelWidth)
	rightWidth = max(rightWidth, MinRightPanelWidth)

	leftStyled := lipgloss.NewStyle().Width(leftWidth).Padding(0, 2, 0, 2).Render(leftPanel)
	rightStyled := lipgloss.NewStyle().Width(rightWidth).PaddingLeft(2).Render(rightPanel)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)
}

func (m *WelcomeModel) renderSystemChecks() string {
	var sections []string

	c := m.controller.GetChecker()

	sections = append(sections, m.styles.Subtitle.Render(locale.ChecksTitle))
	sections = append(sections, "")

	coreChecks := []struct {
		label string
		value bool
	}{
		{locale.CheckEnvironmentFile, c.EnvFileExists},
		{locale.CheckWritePermissions, c.EnvDirWritable},
		{locale.CheckDockerAPI, c.DockerApiAccessible},
		{locale.CheckDockerVersion, c.DockerVersionOK},
		{locale.CheckDockerCompose, c.DockerComposeInstalled},
		{locale.CheckWorkerEnvironment, c.WorkerEnvApiAccessible},
		{locale.CheckSystemResources, c.SysCPUOK && c.SysMemoryOK && c.SysDiskFreeSpaceOK},
		{locale.CheckNetworkConnectivity, c.SysNetworkOK},
	}

	for _, check := range coreChecks {
		sections = append(sections, m.styles.RenderStatusText(check.label, check.value))
	}

	if !c.IsReadyToContinue() {
		sections = append(sections, "")
		sections = append(sections, m.styles.Warning.Render(locale.ChecksWarningFailed))
	}

	return strings.Join(sections, "\n")
}

func (m *WelcomeModel) renderInfoPanel() string {
	if !m.controller.GetChecker().IsReadyToContinue() {
		return m.renderTroubleshootingInfo()
	}
	return m.renderInstallerInfo()
}

func (m *WelcomeModel) renderTroubleshootingInfo() string {
	var sections []string

	c := m.controller.GetChecker()

	sections = append(sections, m.styles.Error.Render(locale.TroubleshootTitle))
	sections = append(sections, "")

	// Environment file check - this is critical and should be shown first
	if !c.EnvFileExists {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootEnvFileTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootEnvFileDesc))
		sections = append(sections, "")
		sections = append(sections, m.styles.Info.Render(locale.TroubleshootEnvFileFix))
		sections = append(sections, "")
	}

	// Write permissions check
	if c.EnvFileExists && !c.EnvDirWritable {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootWritePermTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootWritePermDesc))
		sections = append(sections, "")
		sections = append(sections, m.styles.Info.Render(locale.TroubleshootWritePermFix))
		sections = append(sections, "")
	}

	// Docker API accessibility with specific error types
	if !c.DockerApiAccessible {
		switch c.DockerErrorType {
		case checker.DockerErrorNotInstalled:
			sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootDockerNotInstalledTitle))
			sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootDockerNotInstalledDesc))
			sections = append(sections, "")
			sections = append(sections, m.styles.Info.Render(locale.TroubleshootDockerNotInstalledFix))
		case checker.DockerErrorNotRunning:
			sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootDockerNotRunningTitle))
			sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootDockerNotRunningDesc))
			sections = append(sections, "")
			sections = append(sections, m.styles.Info.Render(locale.TroubleshootDockerNotRunningFix))
		case checker.DockerErrorPermission:
			sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootDockerPermissionTitle))
			sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootDockerPermissionDesc))
			sections = append(sections, "")
			sections = append(sections, m.styles.Info.Render(locale.TroubleshootDockerPermissionFix))
		default:
			sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootDockerAPITitle))
			sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootDockerAPIDesc))
			sections = append(sections, "")
			sections = append(sections, m.styles.Info.Render(locale.TroubleshootDockerAPIFix))
		}
		sections = append(sections, "")
	}

	// Docker version check
	if !c.DockerVersionOK {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootDockerVersionTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootDockerVersionDesc))
		sections = append(sections, "")
		versionFix := fmt.Sprintf(locale.TroubleshootDockerVersionFix, c.DockerVersion)
		sections = append(sections, m.styles.Info.Render(versionFix))
		sections = append(sections, "")
	}

	// Docker Compose check
	if !c.DockerComposeInstalled {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootComposeTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootComposeDesc))
		sections = append(sections, "")
		sections = append(sections, m.styles.Info.Render(locale.TroubleshootComposeFix))
		sections = append(sections, "")
	}

	// Docker Compose version check
	if c.DockerComposeInstalled && !c.DockerComposeVersionOK {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootComposeVersionTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootComposeVersionDesc))
		sections = append(sections, "")
		composeFix := fmt.Sprintf(locale.TroubleshootComposeVersionFix, c.DockerComposeVersion)
		sections = append(sections, m.styles.Info.Render(composeFix))
		sections = append(sections, "")
	}

	// Worker environment check (only if configured)
	if !c.WorkerEnvApiAccessible {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootWorkerTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootWorkerDesc))
		sections = append(sections, "")
		sections = append(sections, m.styles.Info.Render(locale.TroubleshootWorkerFix))
		sections = append(sections, "")
	}

	// System resource checks
	if !c.SysCPUOK {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootCPUTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootCPUDesc))
		sections = append(sections, "")
		cpuFix := fmt.Sprintf(locale.TroubleshootCPUFix, c.SysCPUCount)
		sections = append(sections, m.styles.Info.Render(cpuFix))
		sections = append(sections, "")
	}

	if !c.SysMemoryOK {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootMemoryTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootMemoryDesc))
		sections = append(sections, "")
		memoryFix := fmt.Sprintf(locale.TroubleshootMemoryFix, c.SysMemoryRequired, c.SysMemoryAvailable)
		sections = append(sections, m.styles.Info.Render(memoryFix))
		sections = append(sections, "")
	}

	if !c.SysDiskFreeSpaceOK {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootDiskTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootDiskDesc))
		sections = append(sections, "")
		diskFix := fmt.Sprintf(locale.TroubleshootDiskFix, c.SysDiskRequired, c.SysDiskAvailable)
		sections = append(sections, m.styles.Info.Render(diskFix))
		sections = append(sections, "")
	}

	// Network connectivity check
	if !c.SysNetworkOK {
		sections = append(sections, m.styles.Subtitle.Render(locale.TroubleshootNetworkTitle))
		sections = append(sections, m.styles.Paragraph.Render(locale.TroubleshootNetworkDesc))
		sections = append(sections, "")
		networkFailures := strings.Join(c.SysNetworkFailures, "\n")
		networkFix := fmt.Sprintf(locale.TroubleshootNetworkFix, networkFailures)
		sections = append(sections, m.styles.Info.Render(networkFix))
		sections = append(sections, "")
	}

	sections = append(sections, m.styles.Warning.Render(locale.TroubleshootFixHint))

	return strings.Join(sections, "\n")
}

func (m *WelcomeModel) renderInstallerInfo() string {
	var sections []string

	sections = append(sections, m.styles.Success.Render(locale.WelcomeFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.styles.Paragraph.Render(locale.WelcomeFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.styles.Subtitle.Render(locale.WelcomeWorkflowTitle))
	sections = append(sections, "")

	steps := []string{
		locale.WelcomeWorkflowStep1,
		locale.WelcomeWorkflowStep2,
		locale.WelcomeWorkflowStep3,
		locale.WelcomeWorkflowStep4,
		locale.WelcomeWorkflowStep5,
	}

	for _, step := range steps {
		sections = append(sections, m.styles.Info.Render(step))
	}

	sections = append(sections, "")
	sections = append(sections, m.styles.Success.Render(locale.WelcomeSystemReady))

	return strings.Join(sections, "\n")
}

// isVerticalLayout determines if content should be stacked vertically
func (m *WelcomeModel) isVerticalLayout() bool {
	contentWidth := m.window.GetContentWidth()
	return contentWidth < MinTerminalWidth
}

// Public methods for app.go integration

// IsReadyToContinue returns if system checks are passing
func (m *WelcomeModel) IsReadyToContinue() bool {
	return m.controller.GetChecker().IsReadyToContinue()
}

// HasScrollableContent returns if content is scrollable using viewport methods
func (m *WelcomeModel) HasScrollableContent() bool {
	if !m.ready {
		return false
	}
	// Content is scrollable if we're not at both top and bottom simultaneously
	return !(m.viewport.AtTop() && m.viewport.AtBottom())
}

// BaseScreenModel interface implementation

// GetFormTitle returns the title for the form (layout header)
func (m *WelcomeModel) GetFormTitle() string {
	return locale.WelcomeFormTitle
}

// GetFormDescription returns the description for the form (right panel)
func (m *WelcomeModel) GetFormDescription() string {
	return locale.WelcomeFormDescription
}

// GetFormName returns the name for the form (right panel)
func (m *WelcomeModel) GetFormName() string {
	return locale.WelcomeFormName
}

// GetFormOverview returns form overview for list screens (right panel)
func (m *WelcomeModel) GetFormOverview() string {
	return locale.WelcomeFormOverview
}

// GetCurrentConfiguration returns text with current configuration for the list screens
func (m *WelcomeModel) GetCurrentConfiguration() string {
	c := m.controller.GetChecker()

	if !c.IsReadyToContinue() {
		var failedChecks []string

		if !c.EnvFileExists {
			failedChecks = append(failedChecks, locale.CheckEnvironmentFile)
		}
		if !c.EnvDirWritable {
			failedChecks = append(failedChecks, locale.CheckWritePermissions)
		}
		if !c.DockerApiAccessible {
			failedChecks = append(failedChecks, locale.CheckDockerAPI)
		}
		if !c.DockerVersionOK {
			failedChecks = append(failedChecks, locale.CheckDockerVersion)
		}
		if !c.DockerComposeInstalled {
			failedChecks = append(failedChecks, locale.CheckDockerCompose)
		}
		if c.DockerComposeInstalled && !c.DockerComposeVersionOK {
			failedChecks = append(failedChecks, locale.CheckDockerComposeVersion)
		}
		if !c.WorkerEnvApiAccessible {
			failedChecks = append(failedChecks, locale.CheckWorkerEnvironment)
		}
		if !c.SysCPUOK || !c.SysMemoryOK || !c.SysDiskFreeSpaceOK {
			failedChecks = append(failedChecks, locale.CheckSystemResources)
		}
		if !c.SysNetworkOK {
			failedChecks = append(failedChecks, locale.CheckNetworkConnectivity)
		}

		if len(failedChecks) > 0 {
			return fmt.Sprintf(locale.WelcomeConfigurationFailed, strings.Join(failedChecks, ", "))
		}
	}

	return locale.WelcomeConfigurationPassed
}

// IsConfigured returns true if the form is configured
func (m *WelcomeModel) IsConfigured() bool {
	return m.controller.GetChecker().IsReadyToContinue()
}

// GetFormHotKeys returns the hotkeys for the form (layout footer)
func (m *WelcomeModel) GetFormHotKeys() []string {
	hotkeys := []string{"up|down"}

	if m.HasScrollableContent() {
		hotkeys = append(hotkeys, "left|right", "pgup|pgdown")
	}

	if m.controller.GetChecker().IsReadyToContinue() {
		hotkeys = append(hotkeys, "enter")
	}

	return hotkeys
}

// Compile-time interface validation
var _ BaseScreenModel = (*WelcomeModel)(nil)
