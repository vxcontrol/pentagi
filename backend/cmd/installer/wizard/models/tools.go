package models

import (
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// ToolsHandler implements ListScreenHandler for tools
type ToolsHandler struct {
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
}

// NewToolsHandler creates a new tools handler
func NewToolsHandler(c controller.Controller, s styles.Styles, w window.Window) *ToolsHandler {
	return &ToolsHandler{
		controller: c,
		styles:     s,
		window:     w,
	}
}

// ListScreenHandler interface implementation

func (h *ToolsHandler) LoadItems() []ListItem {
	items := []ListItem{
		{ID: AIAgentsSettingsFormScreen},
		{ID: SearchEnginesFormScreen},
		{ID: ScraperFormScreen},
		{ID: DockerFormScreen},
	}

	return items
}

func (h *ToolsHandler) HandleSelection(item ListItem) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{
			Target: item.ID,
		}
	}
}

func (h *ToolsHandler) GetFormTitle() string {
	return locale.ToolsTitle
}

func (h *ToolsHandler) GetFormDescription() string {
	return locale.ToolsDescription
}

func (h *ToolsHandler) GetFormName() string {
	return locale.ToolsName
}

func (h *ToolsHandler) GetOverview() string {
	var sections []string

	sections = append(sections, h.styles.Subtitle.Render(locale.ToolsTitle))
	sections = append(sections, "")
	sections = append(sections, h.styles.Paragraph.Bold(true).Render(locale.ToolsDescription))
	sections = append(sections, "")
	sections = append(sections, locale.ToolsOverview)

	return strings.Join(sections, "\n")
}

func (h *ToolsHandler) ShowConfiguredStatus() bool {
	return false // tools don't show configuration status icons
}

// ToolsModel represents the tools menu screen using ListScreen
type ToolsModel struct {
	*ListScreen
	*ToolsHandler
}

// NewToolsModel creates a new tools model
func NewToolsModel(c controller.Controller, s styles.Styles, w window.Window, r Registry) *ToolsModel {
	handler := NewToolsHandler(c, s, w)
	listScreen := NewListScreen(c, s, w, r, handler)

	return &ToolsModel{
		ListScreen:   listScreen,
		ToolsHandler: handler,
	}
}

// Compile-time interface validation
var _ BaseScreenModel = (*ToolsModel)(nil)
