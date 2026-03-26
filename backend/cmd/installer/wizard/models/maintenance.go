package models

import (
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// MaintenanceHandler handles the maintenance operations list
type MaintenanceHandler struct {
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
}

// NewMaintenanceHandler creates a new maintenance operations handler
func NewMaintenanceHandler(c controller.Controller, s styles.Styles, w window.Window) *MaintenanceHandler {
	return &MaintenanceHandler{
		controller: c,
		styles:     s,
		window:     w,
	}
}

// LoadItems loads the available maintenance operations based on system state
func (h *MaintenanceHandler) LoadItems() []ListItem {
	items := []ListItem{}
	checker := h.controller.GetChecker()

	// determine which operations to show based on checker consolidated helpers
	showStart := checker.CanStartAll()

	if showStart {
		items = append(items, ListItem{
			ID: StartPentagiScreen,
		})
	}

	// stop PentAGI - show if any stack is running
	showStop := checker.CanStopAll()
	if showStop {
		items = append(items, ListItem{
			ID: StopPentagiScreen,
		})
	}

	// restart PentAGI - show if any stack is running
	if showStop {
		items = append(items, ListItem{
			ID: RestartPentagiScreen,
		})
	}

	// download Worker Image - show if worker image doesn't exist
	if checker.CanDownloadWorker() {
		items = append(items, ListItem{
			ID: DownloadWorkerImageScreen,
		})
	}

	// update Worker Image - show if worker image exists and updates available
	if checker.CanUpdateWorker() {
		items = append(items, ListItem{
			ID:          UpdateWorkerImageScreen,
			Highlighted: true,
		})
	}

	// update PentAGI - show if updates are available for any stack
	showUpdatePentagi := checker.CanUpdateAll()

	if showUpdatePentagi {
		items = append(items, ListItem{
			ID:          UpdatePentagiScreen,
			Highlighted: true,
		})
	}

	// update Installer - show if installer updates are available
	if checker.CanUpdateInstaller() {
		items = append(items, ListItem{
			ID:          UpdateInstallerScreen,
			Highlighted: true,
		})
	}

	// factory Reset - always show if anything is installed
	if checker.CanFactoryReset() {
		items = append(items, ListItem{
			ID: FactoryResetScreen,
		})
	}

	// remove PentAGI - show if any stack is installed
	if checker.CanRemoveAll() {
		items = append(items, ListItem{
			ID: RemovePentagiScreen,
		})
	}

	// purge PentAGI - show if any stack is installed
	if checker.CanPurgeAll() {
		items = append(items, ListItem{
			ID: PurgePentagiScreen,
		})
	}

	// reset admin password - show if PentAGI is running
	if checker.CanResetPassword() {
		items = append(items, ListItem{
			ID: ResetPasswordScreen,
		})
	}

	return items
}

// HandleSelection handles maintenance operation selection
func (h *MaintenanceHandler) HandleSelection(item ListItem) tea.Cmd {
	// navigate to the selected operation form
	return func() tea.Msg {
		return NavigationMsg{Target: item.ID}
	}
}

// GetFormTitle returns the title for the maintenance screen
func (h *MaintenanceHandler) GetFormTitle() string {
	return locale.MaintenanceTitle
}

// GetFormDescription returns the description for the maintenance screen
func (h *MaintenanceHandler) GetFormDescription() string {
	return locale.MaintenanceDescription
}

// GetFormName returns the name for the maintenance screen
func (h *MaintenanceHandler) GetFormName() string {
	return locale.MaintenanceName
}

// GetOverview returns the overview content for the maintenance screen
func (h *MaintenanceHandler) GetOverview() string {
	var sections []string

	sections = append(sections, h.styles.Subtitle.Render(locale.MaintenanceTitle))
	sections = append(sections, "")
	sections = append(sections, h.styles.Paragraph.Bold(true).Render(locale.MaintenanceDescription))
	sections = append(sections, "")
	sections = append(sections, locale.MaintenanceOverview)

	return strings.Join(sections, "\n")
}

// ShowConfiguredStatus returns whether to show configuration status
func (h *MaintenanceHandler) ShowConfiguredStatus() bool {
	return false
}

// MaintenanceModel represents the maintenance operations list screen
type MaintenanceModel struct {
	*ListScreen
	*MaintenanceHandler
}

// NewMaintenanceModel creates a new maintenance operations model
func NewMaintenanceModel(
	c controller.Controller, s styles.Styles, w window.Window, r Registry,
) *MaintenanceModel {
	handler := NewMaintenanceHandler(c, s, w)
	listScreen := NewListScreen(c, s, w, r, handler)

	return &MaintenanceModel{
		ListScreen:         listScreen,
		MaintenanceHandler: handler,
	}
}
