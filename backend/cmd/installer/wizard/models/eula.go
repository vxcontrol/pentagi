package models

import (
	"fmt"

	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EULAModel represents the EULA agreement screen
type EULAModel struct {
	styles        styles.Styles
	window        window.Window
	viewport      viewport.Model
	files         files.Files
	controller    controller.Controller
	content       string
	ready         bool
	scrolled      bool
	scrolledToEnd bool
}

// NewEULAModel creates a new EULA screen model
func NewEULAModel(c controller.Controller, s styles.Styles, w window.Window, f files.Files) *EULAModel {
	return &EULAModel{
		styles:     s,
		window:     w,
		files:      f,
		controller: c,
	}
}

// Init implements tea.Model
func (m *EULAModel) Init() tea.Cmd {
	m.resetForm()
	return m.loadEULA
}

// loadEULA loads the EULA content from files
func (m *EULAModel) loadEULA() tea.Msg {
	content, err := m.files.GetContent("EULA.md")
	if err != nil {
		logger.Errorf("[EULAModel] LOAD: file error: %v", err)
		return EULALoadedMsg{
			Content: fmt.Sprintf(locale.EULAErrorLoadingTitle, err),
			Error:   err,
		}
	}

	rendered, err := m.styles.GetRenderer().Render(string(content))
	if err != nil {
		logger.Errorf("[EULAModel] LOAD: render error: %v", err)
		rendered = fmt.Sprintf(locale.EULAContentFallback, string(content), err)
	}

	return EULALoadedMsg{
		Content: rendered,
		Error:   nil,
	}
}

func (m *EULAModel) resetForm() {
	m.content = ""
	m.ready = false
	m.scrolled = false
	m.scrolledToEnd = false
}

// Update implements tea.Model
func (m *EULAModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateViewport()

	case EULALoadedMsg:
		m.content = msg.Content
		m.updateViewport()
		m.updateScrollStatus()
		return m, func() tea.Msg { return nil }

	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if m.scrolledToEnd {
				logger.Log("[EULAModel] ACCEPT")
				if err := m.controller.SetEulaConsent(); err != nil {
					logger.Errorf("[EULAModel] CONSENT: error: %v", err)
					return m, func() tea.Msg { return nil }
				}
				// skip eula screen write to stack and go to main menu screen straight away
				return m, func() tea.Msg {
					m.resetForm()
					return NavigationMsg{GoBack: true, Target: MainMenuScreen}
				}
			}
		case "n", "N":
			if m.scrolledToEnd {
				logger.Log("[EULAModel] REJECT")
				return m, func() tea.Msg { return NavigationMsg{GoBack: true} }
			}
		default:
			if !m.ready {
				break
			}
			switch msg.String() {
			case "enter":
				m.viewport.PageDown()
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
			case "home":
				m.viewport.GotoTop()
			case "end":
				m.viewport.GotoBottom()
			}
			m.updateScrollStatus()
		}
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		m.updateScrollStatus()
		return m, cmd
	}

	return m, nil
}

// updateViewport sets up the viewport with proper dimensions
func (m *EULAModel) updateViewport() {
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 || m.content == "" {
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

	m.viewport.SetContent(m.content)
	m.updateScrollStatus()
}

// updateScrollStatus checks if user has scrolled to the end
func (m *EULAModel) updateScrollStatus() {
	if m.ready {
		m.scrolled = m.viewport.ScrollPercent() > 0
		m.scrolledToEnd = m.viewport.AtBottom()
	}
}

// View implements tea.Model using proper lipgloss layout
func (m *EULAModel) View() string {
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return locale.EULALoading
	}

	if !m.ready || m.content == "" {
		return m.renderLoading()
	}

	content := m.viewport.View()

	return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Top, content)
}

// renderLoading renders loading state
func (m *EULAModel) renderLoading() string {
	contentWidth, contentHeight := m.window.GetContentSize()
	loading := m.styles.Info.Render(locale.EULALoading)
	return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, loading)
}

// GetScrollInfo returns scroll information for footer display
func (m *EULAModel) GetScrollInfo() (scrolled bool, atEnd bool, percent int) {
	if !m.ready {
		return false, false, 0
	}

	percent = int(m.viewport.ScrollPercent() * 100)

	return m.scrolled, m.scrolledToEnd, percent
}

// EULALoadedMsg represents successful EULA loading
type EULALoadedMsg struct {
	Content string
	Error   error
}

// BaseScreenModel interface implementation

// GetFormTitle returns empty title for the form (glamour renders from the top)
func (m *EULAModel) GetFormTitle() string {
	return ""
}

// GetFormDescription returns the description for the form (right panel)
func (m *EULAModel) GetFormDescription() string {
	return locale.EULAFormDescription
}

// GetFormName returns the name for the form (right panel)
func (m *EULAModel) GetFormName() string {
	return locale.EULAFormName
}

// GetFormOverview returns form overview for list screens (right panel)
func (m *EULAModel) GetFormOverview() string {
	return locale.EULAFormOverview
}

// GetCurrentConfiguration returns text with current configuration for the list screens
func (m *EULAModel) GetCurrentConfiguration() string {
	if m.controller.GetEulaConsent() {
		return locale.EULAConfigurationAccepted
	}

	if m.scrolledToEnd {
		return locale.EULAConfigurationRead
	}

	return locale.EULAConfigurationPending
}

// IsConfigured returns true if eula consent is set
func (m *EULAModel) IsConfigured() bool {
	return m.controller.GetEulaConsent()
}

// GetFormHotKeys returns the hotkeys for the form (layout footer)
func (m *EULAModel) GetFormHotKeys() []string {
	var hotkeys []string

	if m.ready && m.content != "" {
		hotkeys = append(hotkeys, "up|down", "pgup|pgdown", "home|end")
	}

	if m.scrolledToEnd {
		hotkeys = append(hotkeys, "y|n")
	}

	return hotkeys
}

// Compile-time interface validation
var _ BaseScreenModel = (*EULAModel)(nil)
