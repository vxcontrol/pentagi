package models

import (
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// MainMenuHandler implements ListScreenHandler for main menu items
type MainMenuHandler struct {
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
}

// NewMainMenuHandler creates a new main menu handler
func NewMainMenuHandler(c controller.Controller, s styles.Styles, w window.Window) *MainMenuHandler {
	return &MainMenuHandler{
		controller: c,
		styles:     s,
		window:     w,
	}
}

// ListScreenHandler interface implementation

func (h *MainMenuHandler) LoadItems() []ListItem {
	items := []ListItem{
		{ID: LLMProvidersScreen},
		{ID: EmbedderFormScreen},
		{ID: SummarizerScreen},
		{ID: ToolsScreen},
		{ID: MonitoringScreen},
		{ID: ServerSettingsScreen},
		{ID: ApplyChangesScreen, Highlighted: true},
		{ID: InstallPentagiScreen, Highlighted: true},
		{ID: MaintenanceScreen},
	}

	// filter out disabled items
	var enabledItems []ListItem
	for _, item := range items {
		if h.isItemEnabled(item) {
			enabledItems = append(enabledItems, item)
		}
	}

	return enabledItems
}

func (h *MainMenuHandler) HandleSelection(item ListItem) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{
			Target: item.ID,
		}
	}
}

func (h *MainMenuHandler) GetOverview() string {
	var sections []string

	checker := h.controller.GetChecker()

	sections = append(sections, h.styles.Subtitle.Render(locale.MainMenuTitle))
	sections = append(sections, "")
	sections = append(sections, h.styles.Paragraph.Bold(true).Render(locale.MainMenuDescription))
	sections = append(sections, "")
	sections = append(sections, locale.MainMenuOverview)

	// system status section
	sections = append(sections, h.styles.Subtitle.Render(locale.MenuSystemStatus))
	sections = append(sections, "")

	statusItems := []struct {
		Label string
		Value bool
	}{
		{"Docker", checker.DockerApiAccessible},
		{"PentAGI", checker.PentagiRunning},
		{"Langfuse", checker.LangfuseRunning},
		{"Observability", checker.ObservabilityRunning},
	}

	for _, status := range statusItems {
		sections = append(sections, h.styles.RenderStatusText(status.Label, status.Value))
	}

	sections = append(sections, "")
	sections = append(sections, locale.MainMenuOverview)

	return strings.Join(sections, "\n")
}

func (h *MainMenuHandler) ShowConfiguredStatus() bool {
	return false // main menu doesn't show configured status icons
}

func (h *MainMenuHandler) GetFormTitle() string {
	return locale.MainMenuTitle
}

func (h *MainMenuHandler) GetFormDescription() string {
	return locale.MainMenuDescription
}

func (h *MainMenuHandler) GetFormName() string {
	return locale.MainMenuName
}

// Helper methods

func (h *MainMenuHandler) isItemEnabled(item ListItem) bool {
	checker := h.controller.GetChecker()
	switch item.ID {
	case ApplyChangesScreen:
		// show apply changes only when there are pending changes
		return h.controller.IsDirty()
	case InstallPentagiScreen:
		// show install pentagi only when no pending changes and pentagi not installed yet
		return !h.controller.IsDirty() && checker.CanInstallAll()
	case MaintenanceScreen:
		// mirror maintenance screen visibility logic: show only when at least one operation is applicable
		return checker.CanStartAll() || checker.CanStopAll() || checker.CanRestartAll() ||
			checker.CanDownloadWorker() || checker.CanUpdateWorker() || checker.CanUpdateAll() ||
			checker.CanUpdateInstaller() || checker.CanFactoryReset() || checker.CanRemoveAll() || checker.CanPurgeAll()
	default:
		return true
	}
}

// MainMenuModel represents the main configuration menu screen using ListScreen
type MainMenuModel struct {
	*ListScreen
	*MainMenuHandler
}

// NewMainMenuModel creates a new main menu model
func NewMainMenuModel(
	c controller.Controller, s styles.Styles, w window.Window, r Registry,
) *MainMenuModel {
	handler := NewMainMenuHandler(c, s, w)
	listScreen := NewListScreen(c, s, w, r, handler)

	return &MainMenuModel{
		ListScreen:      listScreen,
		MainMenuHandler: handler,
	}
}

// Compile-time interface validation
var _ BaseScreenModel = (*MainMenuModel)(nil)
