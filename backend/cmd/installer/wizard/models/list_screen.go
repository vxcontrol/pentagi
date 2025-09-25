package models

import (
	"fmt"
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListScreenHandler defines methods that concrete list screens must implement
type ListScreenHandler interface {
	// LoadItems loads the list items
	LoadItems() []ListItem

	// HandleSelection handles item selection, returns navigation command
	HandleSelection(item ListItem) tea.Cmd

	// GetOverview returns general overview content
	GetOverview() string

	// ShowConfiguredStatus returns whether to show configuration status icons
	ShowConfiguredStatus() bool
}

// ListItem represents a single item in the list
type ListItem struct {
	ID          ScreenID
	Model       BaseScreenModel
	Highlighted bool // if true, the item is the most important action for user
}

// ListScreen provides common functionality for menu/list screens
type ListScreen struct {
	// Dependencies
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
	registry   Registry

	// State
	selectedIndex int
	items         []ListItem

	// Handler
	handler ListScreenHandler
}

// NewListScreen creates a new list screen instance
func NewListScreen(
	c controller.Controller, s styles.Styles, w window.Window, r Registry, h ListScreenHandler,
) *ListScreen {
	return &ListScreen{
		controller: c,
		styles:     s,
		window:     w,
		registry:   r,
		handler:    h,
	}
}

// BaseScreenModel interface partial implementation

func (l *ListScreen) GetFormOverview() string {
	var sections []string

	// general overview
	if overview := l.handler.GetOverview(); overview != "" {
		sections = append(sections, overview)
		sections = append(sections, "")
	}

	// statistics
	if len(l.items) > 0 {
		if l.handler.ShowConfiguredStatus() {
			configuredCount := 0
			for _, item := range l.items {
				if item.Model.IsConfigured() {
					configuredCount++
				}
			}

			sections = append(sections, l.styles.Subtitle.Render(locale.UIStatistics))

			configuredText := l.styles.Success.Render(fmt.Sprintf("%s: %d", locale.StatusConfigured, configuredCount))
			sections = append(sections, "• "+configuredText)

			notConfiguredCount := len(l.items) - configuredCount
			notConfiguredText := l.styles.Warning.Render(fmt.Sprintf("%s: %d", locale.StatusNotConfigured, notConfiguredCount))
			sections = append(sections, "• "+notConfiguredText)

			sections = append(sections, "")
		}
	}

	return strings.Join(sections, "\n")
}

func (l *ListScreen) GetCurrentConfiguration() string {
	var sections []string

	for idx, item := range l.items {
		sections = append(sections, item.Model.GetCurrentConfiguration())

		if idx < len(l.items)-1 {
			sections = append(sections, "")
		}
	}

	return strings.Join(sections, "\n")
}

func (l *ListScreen) IsConfigured() bool {
	var configuredCount int

	for _, item := range l.items {
		if item.Model.IsConfigured() {
			configuredCount++
		}
	}

	return configuredCount == len(l.items)
}

func (l *ListScreen) GetFormHotKeys() []string {
	if len(l.items) > 0 {
		return []string{"up|down", "enter"}
	}
	return []string{"enter"}
}

// tea.Model interface implementation

func (l *ListScreen) Init() tea.Cmd {
	l.items = l.handler.LoadItems()
	for i, item := range l.items {
		if item.Model == nil {
			l.items[i].Model = l.registry.GetScreen(item.ID)
		}
	}

	if l.selectedIndex < 0 || l.selectedIndex >= len(l.items) {
		l.selectedIndex = 0
	}

	return nil
}

func (l *ListScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// window resize handled by app.go

	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if l.selectedIndex > 0 {
				l.selectedIndex--
			}

		case "down":
			if l.selectedIndex < len(l.items)-1 {
				l.selectedIndex++
			}

		case "enter":
			return l.handleSelection()
		}
	}

	return l, nil
}

func (l *ListScreen) View() string {
	contentWidth, contentHeight := l.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return locale.UILoading
	}

	leftPanel := l.renderItemsList()
	rightPanel := l.renderItemInfo()

	if l.isVerticalLayout() {
		return l.renderVerticalLayout(leftPanel, rightPanel, contentWidth, contentHeight)
	}

	return l.renderHorizontalLayout(leftPanel, rightPanel, contentWidth, contentHeight)
}

// Helper methods for concrete implementations

// GetScreen returns the screen model for the given ID (implement Registry interface)
func (l *ListScreen) GetScreen(id ScreenID) BaseScreenModel {
	return l.registry.GetScreen(id)
}

// GetController returns the state controller
func (l *ListScreen) GetController() controller.Controller {
	return l.controller
}

// GetStyles returns the styles
func (l *ListScreen) GetStyles() styles.Styles {
	return l.styles
}

// GetWindow returns the window
func (l *ListScreen) GetWindow() window.Window {
	return l.window
}

// Internal methods

// handleSelection processes item selection
func (l *ListScreen) handleSelection() (tea.Model, tea.Cmd) {
	if l.selectedIndex >= len(l.items) {
		return l, nil
	}

	selectedItem := l.items[l.selectedIndex]
	return l, l.handler.HandleSelection(selectedItem)
}

// getItemInfo is used as fallback only, most info comes from GetConfigScreen
func (l *ListScreen) getItemInfo(item ListItem) string {
	var sections []string

	sections = append(sections, l.styles.Subtitle.Render(item.Model.GetFormName()))
	sections = append(sections, "")
	sections = append(sections, l.styles.Paragraph.Render(item.Model.GetFormDescription()))

	return strings.Join(sections, "\n")
}

// renderItemsList creates the left panel with items list
func (l *ListScreen) renderItemsList() string {
	var sections []string

	for i, item := range l.items {
		selected := i == l.selectedIndex

		var itemText string
		if item.Model != nil {
			if l.handler.ShowConfiguredStatus() {
				statusIcon := l.styles.RenderStatusIcon(item.Model.IsConfigured()) + " "
				itemText = statusIcon + item.Model.GetFormName()
			} else {
				itemText = item.Model.GetFormName()
			}
		} else {
			// fallback to registry to resolve model for label
			model := l.registry.GetScreen(item.ID)
			if l.handler.ShowConfiguredStatus() {
				statusIcon := l.styles.RenderStatusIcon(model.IsConfigured()) + " "
				itemText = statusIcon + model.GetFormName()
			} else {
				itemText = model.GetFormName()
			}
		}

		rendered := l.styles.RenderMenuItem(itemText, selected, false, item.Highlighted)
		sections = append(sections, rendered)
	}

	if l.handler.ShowConfiguredStatus() {
		sections = append(sections, "")
		sections = append(sections, l.styles.Muted.Render(locale.LegendConfigured))
		sections = append(sections, l.styles.Muted.Render(locale.LegendNotConfigured))
	}

	return strings.Join(sections, "\n")
}

// renderItemInfo creates the right panel with item details
func (l *ListScreen) renderItemInfo() string {
	if len(l.items) == 0 || l.selectedIndex >= len(l.items) {
		return l.styles.Info.Render(locale.UINoConfigSelected)
	}

	selectedItem := l.items[l.selectedIndex]

	// try to get config screen overview first
	if overview := selectedItem.Model.GetFormOverview(); overview != "" {
		currentConfiguration := selectedItem.Model.GetCurrentConfiguration()
		if currentConfiguration == "" {
			return overview
		}

		wholeContent := overview + "\n" + currentConfiguration
		if l.getContentTrueHeight(wholeContent)+PaddingHeight < l.window.GetContentHeight() {
			return wholeContent
		}

		return overview
	}

	// fallback to handler's item info
	return l.getItemInfo(selectedItem)
}

// Layout methods

func (l *ListScreen) getContentTrueHeight(content string) int {
	contentWidth := l.window.GetContentWidth()

	if l.isVerticalLayout() {
		verticalStyle := lipgloss.NewStyle().Width(contentWidth).Padding(verticalLayoutPaddings...)
		contentStyled := verticalStyle.Render(content)
		return lipgloss.Height(contentStyled)
	}

	leftWidth, rightWidth := MinMenuWidth, MinInfoWidth
	extraWidth := contentWidth - leftWidth - rightWidth - PaddingWidth
	if extraWidth > 0 {
		leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)
		rightWidth = contentWidth - leftWidth - PaddingWidth/2
	}

	contentStyled := lipgloss.NewStyle().Width(rightWidth).PaddingLeft(2).Render(content)

	return lipgloss.Height(contentStyled)
}

// isVerticalLayout determines if vertical layout should be used
func (l *ListScreen) isVerticalLayout() bool {
	contentWidth := l.window.GetContentWidth()
	return contentWidth < (MinMenuWidth + MinInfoWidth + PaddingWidth)
}

// renderVerticalLayout renders content in vertical layout
func (l *ListScreen) renderVerticalLayout(leftPanel, rightPanel string, width, height int) string {
	verticalStyle := lipgloss.NewStyle().Width(width).Padding(verticalLayoutPaddings...)

	leftStyled := verticalStyle.Render(leftPanel)
	rightStyled := verticalStyle.Render(rightPanel)
	if lipgloss.Height(leftStyled)+lipgloss.Height(rightStyled)+3 < height {
		return lipgloss.JoinVertical(lipgloss.Left,
			verticalStyle.Render(leftPanel),
			verticalStyle.Height(2).Render("\n"),
			verticalStyle.Render(rightPanel),
		)
	}

	return verticalStyle.Render(leftPanel)
}

// renderHorizontalLayout renders content in horizontal layout
func (l *ListScreen) renderHorizontalLayout(leftPanel, rightPanel string, width, height int) string {
	leftWidth, rightWidth := MinMenuWidth, MinInfoWidth
	extraWidth := width - leftWidth - rightWidth - PaddingWidth
	if extraWidth > 0 {
		leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)
		rightWidth = width - leftWidth - PaddingWidth/2
	}

	leftStyled := lipgloss.NewStyle().Width(leftWidth).Padding(horizontalLayoutPaddings...).Render(leftPanel)
	rightStyled := lipgloss.NewStyle().Width(rightWidth).PaddingLeft(2).Render(rightPanel)

	viewport := viewport.New(width, height-PaddingHeight)
	viewport.SetContent(lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled))

	return viewport.View()
}
