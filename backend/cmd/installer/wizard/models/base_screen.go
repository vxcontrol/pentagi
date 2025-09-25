package models

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	verticalLayoutPaddings   = []int{0, 4, 0, 2}
	horizontalLayoutPaddings = []int{0, 2, 0, 2}
)

// BaseListOption represents a generic list option that can be used in any list
type BaseListOption struct {
	Value   string // the actual value
	Display string // the display text (can be different from value)
}

func (d BaseListOption) FilterValue() string { return d.Value }

// BaseListDelegate handles rendering of generic list options
type BaseListDelegate struct {
	style      lipgloss.Style
	width      int
	selectedFg lipgloss.Color
	normalFg   lipgloss.Color
}

// NewBaseListDelegate creates a new generic list delegate
func NewBaseListDelegate(style lipgloss.Style, width int) *BaseListDelegate {
	return &BaseListDelegate{
		style:      style,
		width:      width,
		selectedFg: styles.Primary,
		normalFg:   lipgloss.Color(""),
	}
}

// SetColors allows customizing the colors
func (d *BaseListDelegate) SetColors(selectedFg, normalFg lipgloss.Color) {
	d.selectedFg = selectedFg
	d.normalFg = normalFg
}

func (d *BaseListDelegate) SetWidth(width int)                     { d.width = width }
func (d BaseListDelegate) Height() int                             { return 1 }
func (d BaseListDelegate) Spacing() int                            { return 0 }
func (d BaseListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d BaseListDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	option, ok := listItem.(BaseListOption)
	if !ok {
		return
	}

	str := option.Display
	if index == m.Index() {
		str = d.style.Width(d.width).Foreground(d.selectedFg).Render(str)
	} else {
		str = d.style.Width(d.width).Foreground(d.normalFg).Render(str)
	}

	fmt.Fprint(w, str)
}

// BaseListHelper provides utility functions for working with lists
type BaseListHelper struct{}

// CreateList creates a new list with the given options and delegate
func (h BaseListHelper) CreateList(options []BaseListOption, delegate list.ItemDelegate, width, height int) list.Model {
	items := make([]list.Item, len(options))
	for i, option := range options {
		items[i] = option
	}

	listModel := list.New(items, delegate, width, height)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(false)
	listModel.SetShowHelp(false)
	listModel.SetShowTitle(false)

	return listModel
}

// SelectByValue selects the list item that matches the given value
func (h BaseListHelper) SelectByValue(listModel *list.Model, value string) {
	items := listModel.Items()
	for i, item := range items {
		if option, ok := item.(BaseListOption); ok && option.Value == value {
			listModel.Select(i)
			break
		}
	}
}

// GetSelectedValue returns the value of the currently selected item
func (h BaseListHelper) GetSelectedValue(listModel *list.Model) string {
	selectedItem := listModel.SelectedItem()
	if selectedItem == nil {
		return ""
	}

	if option, ok := selectedItem.(BaseListOption); ok {
		return option.Value
	}

	return ""
}

// GetSelectedDisplay returns the display text of the currently selected item
func (h BaseListHelper) GetSelectedDisplay(listModel *list.Model) string {
	selectedItem := listModel.SelectedItem()
	if selectedItem == nil {
		return ""
	}

	if option, ok := selectedItem.(BaseListOption); ok {
		return option.Display
	}

	return ""
}

// BaseScreenModel defines methods that concrete screens must implement
type BaseScreenModel interface {
	// GetFormTitle returns the title for the form (layout header)
	GetFormTitle() string

	// GetFormDescription returns the description for the form (right panel)
	GetFormDescription() string

	// GetFormName returns the name for the form (right panel)
	GetFormName() string

	// GetFormOverview returns form overview for list screens (right panel)
	GetFormOverview() string

	// GetCurrentConfiguration returns text with current configuration for the list screens
	GetCurrentConfiguration() string

	// IsConfigured returns true if the form is configured
	IsConfigured() bool

	// GetFormHotKeys returns the hotkeys for the form (layout footer)
	GetFormHotKeys() []string

	tea.Model // for common interface logic
}

// BaseScreenHandler defines methods that concrete screens must implement
type BaseScreenHandler interface {
	// BuildForm builds the specific form fields for this screen
	BuildForm() tea.Cmd

	// GetFormSummary returns optional summary for the form bottom
	GetFormSummary() string

	// GetHelpContent returns the right panel help content
	GetHelpContent() string

	// HandleSave handles saving the form data
	HandleSave() error

	// HandleReset handles resetting the form to default values
	HandleReset()

	// OnFieldChanged is called when a form field value changes
	OnFieldChanged(fieldIndex int, oldValue, newValue string)

	// GetFormFields returns the current form fields
	GetFormFields() []FormField

	// SetFormFields sets the form fields
	SetFormFields(fields []FormField)
}

// BaseListHandler defines methods for screens that use lists (optional)
type BaseListHandler interface {
	// GetList returns the list model if this screen uses a list
	GetList() *list.Model

	// GetListDelegate returns the list delegate if this screen uses a list
	GetListDelegate() *BaseListDelegate

	// OnListSelectionChanged is called when list selection changes
	OnListSelectionChanged(oldSelection, newSelection string)

	// GetListTitle returns the title of the list
	GetListTitle() string

	// GetListDescription returns the description of the list
	GetListDescription() string
}

// FormField represents a single form field
type FormField struct {
	Key         string
	Title       string
	Description string
	Placeholder string
	Required    bool
	Masked      bool
	Input       textinput.Model
	Value       string
	Suggestions []string
}

// BaseScreen provides common functionality for installer form screens
type BaseScreen struct {
	// Dependencies
	controller controller.Controller
	styles     styles.Styles
	window     window.Window

	// State
	initialized  bool
	hasChanges   bool
	focusedIndex int
	showValues   bool

	// Form data
	fields       []FormField
	fieldHeights []int
	bottomHeight int

	// UI components
	viewportForm viewport.Model
	viewportHelp viewport.Model
	formContent  string

	// Handlers - must be set by concrete implementations
	handler     BaseScreenHandler
	listHandler BaseListHandler // optional, can be nil

	// Common utilities
	listHelper BaseListHelper
}

// NewBaseScreen creates a new base screen instance
func NewBaseScreen(
	c controller.Controller, s styles.Styles, w window.Window,
	h BaseScreenHandler, lh BaseListHandler, // can be nil
) *BaseScreen {
	return &BaseScreen{
		controller:   c,
		styles:       s,
		window:       w,
		showValues:   false,
		viewportForm: viewport.New(w.GetContentSize()),
		viewportHelp: viewport.New(w.GetContentSize()),
		handler:      h,
		listHandler:  lh,
		fieldHeights: []int{},
		listHelper:   BaseListHelper{},
	}
}

// Init initializes the base screen
func (b *BaseScreen) Init() tea.Cmd {
	cmd := b.handler.BuildForm()
	b.fields = b.handler.GetFormFields()
	b.updateViewports()
	return cmd
}

// Update handles common update logic and returns commands only
// Concrete implementations should call this and return themselves as the model
func (b *BaseScreen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.updateViewports()

	case tea.KeyMsg:
		return b.handleKeyPress(msg)
	}

	return nil
}

// GetFormHotKeys returns the hotkeys for the form
func (b *BaseScreen) GetFormHotKeys() []string {
	haveMaskedFields := false
	for _, field := range b.fields {
		if field.Masked {
			haveMaskedFields = true
			break
		}
	}

	haveFieldsWithSuggestions := false
	for _, field := range b.fields {
		if len(field.Suggestions) > 0 {
			haveFieldsWithSuggestions = true
			break
		}
	}

	hasList := b.listHandler != nil && b.listHandler.GetList() != nil

	var hotkeys []string
	if len(b.fields) > 0 || hasList {
		hotkeys = append(hotkeys, "down|up")
		if hasList {
			hotkeys = append(hotkeys, "left|right")
		}
		hotkeys = append(hotkeys, "ctrl+s")
		hotkeys = append(hotkeys, "ctrl+r")

	}
	if haveMaskedFields {
		hotkeys = append(hotkeys, "ctrl+h")
	}
	if haveFieldsWithSuggestions {
		hotkeys = append(hotkeys, "tab")
	}
	hotkeys = append(hotkeys, "enter")

	return hotkeys
}

// handleKeyPress handles common keyboard interactions
func (b *BaseScreen) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "down":
		b.focusNext()
		b.updateViewports()
		b.ensureFocusVisible()

	case "up":
		b.focusPrev()
		b.updateViewports()
		b.ensureFocusVisible()

	case "ctrl+s":
		return b.saveConfiguration()

	case "ctrl+r":
		b.resetForm()
		b.updateViewports()

	case "ctrl+h":
		b.toggleShowValues()
		b.updateViewports()

	case "tab":
		b.handleTabCompletion()

	case "enter":
		return b.saveAndReturn()
	}

	return nil
}

// View renders the screen
func (b *BaseScreen) View() string {
	contentWidth, contentHeight := b.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return locale.UILoading
	}

	if !b.initialized {
		b.handler.BuildForm()
		b.fields = b.handler.GetFormFields()
		b.updateViewports()
		b.initialized = true
	}

	leftPanel := b.renderForm()
	rightPanel := b.renderHelp()

	if b.isVerticalLayout() {
		return b.renderVerticalLayout(leftPanel, rightPanel, contentWidth, contentHeight)
	}

	return b.renderHorizontalLayout(leftPanel, rightPanel, contentWidth, contentHeight)
}

// Common form methods

// GetInputWidth calculates the appropriate input width
func (b *BaseScreen) GetInputWidth() int {
	viewportWidth, _ := b.getViewportFormSize()
	inputWidth := viewportWidth - 6
	if b.isVerticalLayout() {
		inputWidth = viewportWidth - 4
	}
	return inputWidth
}

// getViewportFormSize calculates viewport left panel dimensions
func (b *BaseScreen) getViewportFormSize() (int, int) {
	contentWidth, contentHeight := b.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return 0, 0
	}

	if b.isVerticalLayout() {
		return contentWidth - PaddingWidth/2, contentHeight - PaddingHeight
	} else {
		leftWidth, rightWidth := MinMenuWidth, MinInfoWidth
		extraWidth := contentWidth - leftWidth - rightWidth - PaddingWidth
		if extraWidth > 0 {
			leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)
		}
		return leftWidth, contentHeight - PaddingHeight
	}
}

// updateViewports updates the viewports with current content
func (b *BaseScreen) updateViewports() {
	contentWidth, contentHeight := b.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return
	}

	b.updateFormContent()

	viewportWidth, viewportHeight := b.getViewportFormSize()
	formContentHeight := lipgloss.Height(b.formContent)

	b.viewportForm.Width = viewportWidth
	b.viewportForm.Height = min(viewportHeight, formContentHeight)
	b.viewportForm.SetContent(b.formContent) // force update of the viewport content

	helpContent := b.renderHelpContent()
	b.viewportHelp.Width = lipgloss.Width(helpContent)
	b.viewportHelp.Height = lipgloss.Height(helpContent)
	b.viewportHelp.SetContent(helpContent) // force update of the viewport content
}

// updateFormContent renders form content and calculates field heights
func (b *BaseScreen) updateFormContent() {
	var sections []string
	b.fieldHeights = []int{}
	inputWidth := b.GetInputWidth()

	if b.listHandler != nil {
		if listModel := b.listHandler.GetList(); listModel != nil {
			listStyle := b.styles.FormInput.Width(inputWidth)
			if b.focusedIndex == 0 {
				listStyle = listStyle.BorderForeground(styles.Primary)
			}

			listModel.SetWidth(inputWidth - 4)
			renderedList := listStyle.Render(listModel.View())

			// field title
			titleStyle := b.styles.FormLabel
			if b.getFieldIndex() == -1 {
				titleStyle = titleStyle.Foreground(styles.Primary)
			}
			title := titleStyle.Render(b.listHandler.GetListTitle())
			sections = append(sections, title)

			// field description
			description := b.styles.FormHelp.Render(b.listHandler.GetListDescription())
			sections = append(sections, description)

			sections = append(sections, renderedList)
			sections = append(sections, "")

			listHeight := lipgloss.Height(b.renderFormContent(sections[:3]))
			b.fieldHeights = append(b.fieldHeights, listHeight)
		}
	}

	for i, field := range b.fields {
		// check if this field is focused
		focused := b.getFieldIndex() == i

		// field title
		titleStyle := b.styles.FormLabel
		if focused {
			titleStyle = titleStyle.Foreground(styles.Primary)
		}
		title := titleStyle.Render(field.Title)
		sections = append(sections, title)

		// field description
		description := b.styles.FormHelp.Render(field.Description)
		sections = append(sections, description)

		// input field
		inputStyle := b.styles.FormInput.Width(inputWidth)
		if focused {
			inputStyle = inputStyle.BorderForeground(styles.Primary)
		}

		// configure input
		input := field.Input
		input.Width = inputWidth - 3
		input.SetValue(input.Value()) // force update of the input value

		// set up suggestions for tab completion
		if len(field.Suggestions) > 0 {
			input.ShowSuggestions = true
			input.SetSuggestions(field.Suggestions)
		}

		// apply masking if needed and not showing values
		if field.Masked && !b.showValues {
			input.EchoMode = textinput.EchoPassword
		} else {
			input.EchoMode = textinput.EchoNormal
		}

		// ensure focus state is correct
		if focused {
			input.Focus()
		} else {
			input.Blur()
		}

		renderedInput := inputStyle.Render(input.View())
		sections = append(sections, renderedInput)
		sections = append(sections, "")

		// update the field with configured input
		b.fields[i].Input = input

		// calculate field height
		renderedField := b.renderFormContent([]string{title, description, renderedInput})
		b.fieldHeights = append(b.fieldHeights, lipgloss.Height(renderedField))
	}

	// update list styles
	if b.listHandler != nil {
		if listModel := b.listHandler.GetList(); listModel != nil {
			listModel.Styles.PaginationStyle = b.styles.FormPagination.Width(inputWidth)
		}
		if listDelegate := b.listHandler.GetListDelegate(); listDelegate != nil {
			listDelegate.SetWidth(inputWidth)
		}
	}

	statusMessage := ""
	if b.hasChanges {
		statusMessage = b.styles.Warning.Render(locale.UIUnsavedChanges)
	} else {
		statusMessage = b.styles.Success.Render(locale.UIConfigSaved)
	}

	sections = append(sections, statusMessage)
	bottomSections := []string{statusMessage}

	if summary := b.handler.GetFormSummary(); summary != "" {
		sections = append(sections, "", summary)
		bottomSections = append(bottomSections, "", summary)
	}
	b.bottomHeight = lipgloss.Height(b.renderFormContent(bottomSections))

	// update form content
	b.formContent = b.renderFormContent(sections)
}

func (b *BaseScreen) renderFormContent(sections []string) string {
	content := strings.Join(sections, "\n")

	contentWidth, contentHeight := b.window.GetContentSize()
	viewportHeight := contentHeight - PaddingHeight // for final rendering
	approximateMaxHeight := lipgloss.Height(content)
	viewport := viewport.New(contentWidth, max(viewportHeight, approximateMaxHeight*3))

	if b.isVerticalLayout() {
		xAxisPadding := verticalLayoutPaddings[1] + verticalLayoutPaddings[3]
		verticalStyle := lipgloss.NewStyle().Width(contentWidth - xAxisPadding)
		content = verticalStyle.Render(content)
	} else {
		leftWidth, rightWidth := MinMenuWidth, MinInfoWidth
		extraWidth := contentWidth - leftWidth - rightWidth - PaddingWidth
		if extraWidth > 0 {
			leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)
		}
		xAxisPadding := horizontalLayoutPaddings[1] + horizontalLayoutPaddings[3]
		content = lipgloss.NewStyle().Width(leftWidth - xAxisPadding).Render(content)
	}

	viewport.SetContent(content)
	viewport.Height = viewport.VisibleLineCount()

	return viewport.View()
}

func (b *BaseScreen) renderHelpContent() string {
	helpContent := b.handler.GetHelpContent()
	if b.isVerticalLayout() {
		return b.renderFormContent([]string{helpContent})
	}

	contentWidth, _ := b.window.GetContentSize()
	leftWidth, rightWidth := MinMenuWidth, MinInfoWidth
	extraWidth := contentWidth - leftWidth - rightWidth - PaddingWidth
	if extraWidth > 0 {
		leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)
		rightWidth = contentWidth - leftWidth - PaddingWidth/2
	}

	return lipgloss.NewStyle().Width(rightWidth - 2).Render(helpContent)
}

// renderForm renders the left panel with the form
func (b *BaseScreen) renderForm() string {
	if !b.initialized {
		return locale.UILoading
	}
	return b.viewportForm.View()
}

// renderHelp renders the right panel with help content
func (b *BaseScreen) renderHelp() string {
	if !b.initialized {
		return ""
	}
	return b.viewportHelp.View()
}

// ensureFocusVisible scrolls the viewport to ensure focused field is visible
func (b *BaseScreen) ensureFocusVisible() {
	if b.focusedIndex >= len(b.fieldHeights) {
		return
	}

	// calculate y position of focused field
	focusY := 0
	if b.focusedIndex == len(b.fieldHeights)-1 {
		focusY = b.bottomHeight
	}
	for i := range b.focusedIndex {
		focusY += b.fieldHeights[i] + 1 // empty line between fields
	}

	// get viewport dimensions
	visibleRows := b.viewportForm.Height
	offset := b.viewportForm.YOffset

	// if focused field is above visible area, scroll up
	if focusY < offset {
		b.viewportForm.YOffset = focusY
	}

	// if focused field is below visible area, scroll down
	if focusY+b.fieldHeights[b.focusedIndex] >= offset+visibleRows {
		b.viewportForm.YOffset = focusY + b.fieldHeights[b.focusedIndex] - visibleRows + 1
	}
}

// Navigation methods

// focusNext moves focus to the next field
func (b *BaseScreen) focusNext() {
	totalElements := b.getTotalElements()
	if totalElements == 0 {
		return
	}

	// blur current field
	b.blurCurrentField()

	// move to next element (with wrapping)
	b.focusedIndex = (b.focusedIndex + 1) % totalElements

	// focus new field
	b.focusCurrentField()
	b.updateFormContent()
}

// focusPrev moves focus to the previous field
func (b *BaseScreen) focusPrev() {
	totalElements := b.getTotalElements()
	if totalElements == 0 {
		return
	}

	// blur current field
	b.blurCurrentField()

	// move to previous element (with wrapping)
	b.focusedIndex = (b.focusedIndex - 1 + totalElements) % totalElements

	// focus new field
	b.focusCurrentField()
	b.updateFormContent()
}

// getTotalElements returns the total number of navigable elements
func (b *BaseScreen) getTotalElements() int {
	total := len(b.fields)

	// add 1 for list if present
	if b.listHandler != nil && b.listHandler.GetList() != nil {
		total++
	}

	return total
}

// blurCurrentField removes focus from the currently focused field
func (b *BaseScreen) blurCurrentField() {
	fieldIndex := b.getFieldIndex()
	if fieldIndex >= 0 && fieldIndex < len(b.fields) {
		b.fields[fieldIndex].Input.Blur()
	}
}

// focusCurrentField sets focus on the currently focused field
func (b *BaseScreen) focusCurrentField() {
	fieldIndex := b.getFieldIndex()
	if fieldIndex >= 0 && fieldIndex < len(b.fields) {
		b.fields[fieldIndex].Input.Focus()
	}
}

// getFieldIndex returns the field index for the current focusedIndex (-1 if focused on list)
func (b *BaseScreen) getFieldIndex() int {
	if b.listHandler != nil && b.listHandler.GetList() != nil {
		// list is at index 0, fields start at index 1
		return b.focusedIndex - 1
	}
	// no list, fields start at index 0
	return b.focusedIndex
}

// toggleShowValues toggles visibility of masked values
func (b *BaseScreen) toggleShowValues() {
	b.showValues = !b.showValues
	b.updateFormContent()
}

// handleTabCompletion handles tab completion for focused field
func (b *BaseScreen) handleTabCompletion() {
	fieldIndex := b.getFieldIndex()

	// check if we're focused on a valid field
	if fieldIndex >= 0 && fieldIndex < len(b.fields) {
		field := &b.fields[fieldIndex]

		// only handle tab completion if field has suggestions
		if len(field.Suggestions) > 0 {
			// use textinput's built-in suggestion functionality
			if suggestion := field.Input.CurrentSuggestion(); suggestion != "" {
				oldValue := field.Input.Value()
				field.Input.SetValue(suggestion)
				field.Input.CursorEnd()
				field.Value = suggestion
				b.hasChanges = true

				// notify handler about the change
				b.handler.OnFieldChanged(fieldIndex, oldValue, suggestion)

				// update the fields array
				b.fields[fieldIndex] = *field
				b.handler.SetFormFields(b.fields)
				b.updateViewports()
			}
		}
	}
}

// resetForm resets the form to default values
func (b *BaseScreen) resetForm() {
	b.handler.HandleReset()
	b.fields = b.handler.GetFormFields()
	b.hasChanges = false
	b.updateFormContent()
}

// saveConfiguration saves the current configuration
func (b *BaseScreen) saveConfiguration() tea.Cmd {
	if err := b.handler.HandleSave(); err != nil {
		logger.Errorf("[BaseScreen] SAVE: error: %v", err)
		return nil
	}

	b.hasChanges = false
	b.updateViewports()

	return nil
}

// saveAndReturn saves and returns to previous screen
func (b *BaseScreen) saveAndReturn() tea.Cmd {
	// save first
	cmd := b.saveConfiguration()
	if cmd != nil {
		return cmd
	}

	// return to previous screen
	return func() tea.Msg {
		return NavigationMsg{GoBack: true}
	}
}

// Layout methods

// isVerticalLayout determines if vertical layout should be used
func (b *BaseScreen) isVerticalLayout() bool {
	contentWidth := b.window.GetContentWidth()
	return contentWidth < (MinMenuWidth + MinInfoWidth + PaddingWidth)
}

// renderVerticalLayout renders content in vertical layout
func (b *BaseScreen) renderVerticalLayout(leftPanel, rightPanel string, width, height int) string {
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
func (b *BaseScreen) renderHorizontalLayout(leftPanel, rightPanel string, width, height int) string {
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

// Helper methods for concrete implementations

// HandleFieldInput handles input for a specific field
func (b *BaseScreen) HandleFieldInput(msg tea.KeyMsg) tea.Cmd {
	fieldIndex := b.getFieldIndex()

	// all hotkeys are handled by handleKeyPress(msg tea.KeyMsg) method
	// inherit screen must call HandleUpdate(msg tea.Msg) for all uncaught messages
	if slices.Contains(b.GetFormHotKeys(), msg.String()) {
		return nil
	}

	// check if we're focused on a valid field
	if fieldIndex >= 0 && fieldIndex < len(b.fields) {
		var cmd tea.Cmd
		oldValue := b.fields[fieldIndex].Input.Value()
		b.fields[fieldIndex].Input, cmd = b.fields[fieldIndex].Input.Update(msg)
		newValue := b.fields[fieldIndex].Input.Value()

		if oldValue != newValue {
			b.fields[fieldIndex].Value = newValue
			b.hasChanges = true
			b.handler.OnFieldChanged(fieldIndex, oldValue, newValue)
		}

		b.updateViewports()
		return cmd
	}

	return nil
}

// HandleListInput handles input for the list component
func (b *BaseScreen) HandleListInput(msg tea.KeyMsg) tea.Cmd {
	// check if we have a list and we're focused on it (skip if not)
	if b.listHandler == nil {
		return nil
	}

	// check if focused on list (index 0 when list is present)
	isFocusedOnList := b.listHandler.GetList() != nil && b.focusedIndex == 0
	if !isFocusedOnList {
		return nil
	}

	// filter list input keys to slide the list
	switch msg.String() {
	case "left", "right":
		break
	default:
		return nil
	}

	listModel := b.listHandler.GetList()
	if listModel == nil {
		return nil
	}

	// get old selection
	oldSelection := ""
	if selectedItem := listModel.SelectedItem(); selectedItem != nil {
		oldSelection = selectedItem.FilterValue()
	}

	// update list
	var cmd tea.Cmd
	*listModel, cmd = listModel.Update(msg)

	// get new selection
	newSelection := ""
	if selectedItem := listModel.SelectedItem(); selectedItem != nil {
		newSelection = selectedItem.FilterValue()
	}

	// notify handler if selection changed
	if oldSelection != newSelection {
		b.listHandler.OnListSelectionChanged(oldSelection, newSelection)
		b.hasChanges = true
		b.updateViewports()
	}

	return cmd
}

// GetController returns the state controller
func (b *BaseScreen) GetController() controller.Controller {
	return b.controller
}

// GetStyles returns the styles
func (b *BaseScreen) GetStyles() styles.Styles {
	return b.styles
}

// GetWindow returns the window
func (b *BaseScreen) GetWindow() window.Window {
	return b.window
}

// SetHasChanges sets the hasChanges flag
func (b *BaseScreen) SetHasChanges(hasChanges bool) {
	b.hasChanges = hasChanges
}

// GetHasChanges returns the hasChanges flag
func (b *BaseScreen) GetHasChanges() bool {
	return b.hasChanges
}

// GetShowValues returns the showValues flag
func (b *BaseScreen) GetShowValues() bool {
	return b.showValues
}

// GetFocusedIndex returns the currently focused field index
func (b *BaseScreen) GetFocusedIndex() int {
	return b.focusedIndex
}

// SetFocusedIndex sets the focused field index
func (b *BaseScreen) SetFocusedIndex(index int) {
	b.focusedIndex = index
}

// GetListHelper returns the list helper utility
func (b *BaseScreen) GetListHelper() *BaseListHelper {
	return &b.listHelper
}
