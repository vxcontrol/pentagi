package models

import (
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// MonitoringHandler implements ListScreenHandler for monitoring platforms
type MonitoringHandler struct {
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
}

// NewMonitoringHandler creates a new monitoring platforms handler
func NewMonitoringHandler(c controller.Controller, s styles.Styles, w window.Window) *MonitoringHandler {
	return &MonitoringHandler{
		controller: c,
		styles:     s,
		window:     w,
	}
}

// ListScreenHandler interface implementation

func (h *MonitoringHandler) LoadItems() []ListItem {
	items := []ListItem{
		{ID: LangfuseScreen},
		{ID: ObservabilityScreen},
	}

	return items
}

func (h *MonitoringHandler) HandleSelection(item ListItem) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{
			Target: item.ID,
		}
	}
}

func (h *MonitoringHandler) GetFormTitle() string {
	return locale.MonitoringTitle
}

func (h *MonitoringHandler) GetFormDescription() string {
	return locale.MonitoringDescription
}

func (h *MonitoringHandler) GetFormName() string {
	return locale.MonitoringName
}

func (h *MonitoringHandler) GetOverview() string {
	var sections []string

	sections = append(sections, h.styles.Subtitle.Render(locale.MonitoringTitle))
	sections = append(sections, "")
	sections = append(sections, h.styles.Paragraph.Bold(true).Render(locale.MonitoringDescription))
	sections = append(sections, "")
	sections = append(sections, locale.MonitoringOverview)

	return strings.Join(sections, "\n")
}

func (h *MonitoringHandler) ShowConfiguredStatus() bool {
	return true // show configuration status for monitoring platforms
}

// MonitoringModel represents the monitoring platforms menu screen using ListScreen
type MonitoringModel struct {
	*ListScreen
	*MonitoringHandler
}

// NewMonitoringModel creates a new monitoring platforms model
func NewMonitoringModel(c controller.Controller, s styles.Styles, w window.Window, r Registry) *MonitoringModel {
	handler := NewMonitoringHandler(c, s, w)
	listScreen := NewListScreen(c, s, w, r, handler)

	return &MonitoringModel{
		ListScreen:        listScreen,
		MonitoringHandler: handler,
	}
}

// Compile-time interface validation
var _ BaseScreenModel = (*MonitoringModel)(nil)
