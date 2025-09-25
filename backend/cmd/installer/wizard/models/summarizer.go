package models

import (
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// SummarizerHandler implements ListScreenHandler for summarizer types
type SummarizerHandler struct {
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
}

// NewSummarizerHandler creates a new summarizer types handler
func NewSummarizerHandler(c controller.Controller, s styles.Styles, w window.Window) *SummarizerHandler {
	return &SummarizerHandler{
		controller: c,
		styles:     s,
		window:     w,
	}
}

// ListScreenHandler interface implementation

func (h *SummarizerHandler) LoadItems() []ListItem {
	items := []ListItem{
		{ID: SummarizerGeneralScreen},
		{ID: SummarizerAssistantScreen},
	}

	return items
}

func (h *SummarizerHandler) HandleSelection(item ListItem) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{
			Target: item.ID,
		}
	}
}

func (h *SummarizerHandler) GetFormTitle() string {
	return locale.SummarizerTitle
}

func (h *SummarizerHandler) GetFormDescription() string {
	return locale.SummarizerDescription
}

func (h *SummarizerHandler) GetFormName() string {
	return locale.SummarizerName
}

func (h *SummarizerHandler) GetOverview() string {
	var sections []string

	sections = append(sections, h.styles.Subtitle.Render(locale.SummarizerTitle))
	sections = append(sections, "")
	sections = append(sections, h.styles.Paragraph.Bold(true).Render(locale.SummarizerDescription))
	sections = append(sections, "")
	sections = append(sections, locale.SummarizerOverview)

	return strings.Join(sections, "\n")
}

func (h *SummarizerHandler) ShowConfiguredStatus() bool {
	return false // always configured and not shown
}

// SummarizerModel represents the summarizer types menu screen using ListScreen
type SummarizerModel struct {
	*ListScreen
	*SummarizerHandler
}

// NewSummarizerModel creates a new summarizer types model
func NewSummarizerModel(c controller.Controller, s styles.Styles, w window.Window, r Registry) *SummarizerModel {
	handler := NewSummarizerHandler(c, s, w)
	listScreen := NewListScreen(c, s, w, r, handler)

	return &SummarizerModel{
		ListScreen:        listScreen,
		SummarizerHandler: handler,
	}
}

// Compile-time interface validation
var _ BaseScreenModel = (*SummarizerModel)(nil)
