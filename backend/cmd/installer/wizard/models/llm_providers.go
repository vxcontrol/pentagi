package models

import (
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// LLMProvidersHandler implements ListScreenHandler for LLM providers
type LLMProvidersHandler struct {
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
}

// NewLLMProvidersHandler creates a new LLM providers handler
func NewLLMProvidersHandler(c controller.Controller, s styles.Styles, w window.Window) *LLMProvidersHandler {
	return &LLMProvidersHandler{
		controller: c,
		styles:     s,
		window:     w,
	}
}

// ListScreenHandler interface implementation

func (h *LLMProvidersHandler) LoadItems() []ListItem {
	items := []ListItem{
		{ID: LLMProviderOpenAIScreen},
		{ID: LLMProviderAnthropicScreen},
		{ID: LLMProviderGeminiScreen},
		{ID: LLMProviderBedrockScreen},
		{ID: LLMProviderOllamaScreen},
		{ID: LLMProviderCustomScreen},
	}

	return items
}

func (h *LLMProvidersHandler) HandleSelection(item ListItem) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{
			Target: item.ID,
		}
	}
}

func (h *LLMProvidersHandler) GetFormTitle() string {
	return locale.LLMProvidersTitle
}

func (h *LLMProvidersHandler) GetFormDescription() string {
	return locale.LLMProvidersDescription
}

func (h *LLMProvidersHandler) GetFormName() string {
	return locale.LLMProvidersName
}

func (h *LLMProvidersHandler) GetOverview() string {
	var sections []string

	sections = append(sections, h.styles.Subtitle.Render(locale.LLMProvidersTitle))
	sections = append(sections, "")
	sections = append(sections, h.styles.Paragraph.Bold(true).Render(locale.LLMProvidersDescription))
	sections = append(sections, "")
	sections = append(sections, locale.LLMProvidersOverview)

	return strings.Join(sections, "\n")
}

func (h *LLMProvidersHandler) ShowConfiguredStatus() bool {
	return true // show configuration status for LLM providers
}

// LLMProvidersModel represents the LLM providers menu screen using ListScreen
type LLMProvidersModel struct {
	*ListScreen
	*LLMProvidersHandler
}

// NewLLMProvidersModel creates a new LLM providers model
func NewLLMProvidersModel(c controller.Controller, s styles.Styles, w window.Window, r Registry) *LLMProvidersModel {
	handler := NewLLMProvidersHandler(c, s, w)
	listScreen := NewListScreen(c, s, w, r, handler)

	return &LLMProvidersModel{
		ListScreen:          listScreen,
		LLMProvidersHandler: handler,
	}
}

// Compile-time interface validation
var _ BaseScreenModel = (*LLMProvidersModel)(nil)
