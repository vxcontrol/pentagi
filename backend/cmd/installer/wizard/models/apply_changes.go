package models

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/processor"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/terminal"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// ApplyChangesFormModel represents the Apply Changes form
type ApplyChangesFormModel struct {
	*BaseScreen

	// processor integration
	processor processor.ProcessorModel
	running   bool

	// terminal integration
	terminal terminal.Terminal

	// files access and integrity state
	collecting       bool
	waitingForChoice bool
	outdated         map[string]files.FileStatus
}

// NewApplyChangesFormModel creates a new Apply Changes form model
func NewApplyChangesFormModel(
	c controller.Controller, s styles.Styles, w window.Window, p processor.ProcessorModel,
) *ApplyChangesFormModel {
	m := &ApplyChangesFormModel{
		processor: p,
	}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// BaseScreenHandler interface implementation

func (m *ApplyChangesFormModel) BuildForm() tea.Cmd {
	// no form fields for this screen - it's a display/action screen
	m.SetFormFields([]FormField{})

	contentWidth, contentHeight := m.getViewportFormSize()

	// setup terminal
	if m.terminal == nil {
		if !m.isVerticalLayout() {
			contentWidth -= 2
		}
		m.terminal = terminal.NewTerminal(contentWidth-2, contentHeight-1,
			terminal.WithAutoScroll(),
			terminal.WithAutoPoll(),
			terminal.WithCurrentEnv(),
		)
	} else {
		m.terminal.Clear()
	}

	if m.getChangesCount() == 0 {
		m.terminal.Append(locale.ApplyChangesNoChanges)
	} else {
		m.terminal.Append(locale.ApplyChangesNotStarted)
		m.terminal.Append("")
		m.terminal.Append(locale.ApplyChangesInstructions)
	}

	// prevent re-initialization on View() calls
	if !m.initialized {
		m.initialized = true
	} else {
		return nil
	}

	// return terminal's init command to start listening for updates
	return m.terminal.Init()
}

func (m *ApplyChangesFormModel) GetFormTitle() string {
	return locale.ApplyChangesFormTitle
}

func (m *ApplyChangesFormModel) GetFormDescription() string {
	return locale.ApplyChangesFormDescription
}

func (m *ApplyChangesFormModel) GetFormName() string {
	return locale.ApplyChangesFormName
}

func (m *ApplyChangesFormModel) GetFormSummary() string {
	// terminal viewport takes all available space
	return ""
}

func (m *ApplyChangesFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ApplyChangesFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ApplyChangesFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ApplyChangesFormOverview))

	return strings.Join(sections, "\n")
}

func (m *ApplyChangesFormModel) GetCurrentConfiguration() string {
	var sections []string

	config := m.GetController().GetApplyChangesConfig()

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ApplyChangesChangesTitle))
	sections = append(sections, "")

	if config.ChangesCount == 0 {
		sections = append(sections, m.GetStyles().Warning.Render(locale.ApplyChangesChangesEmpty))
		return strings.Join(sections, "\n")
	}

	// show changes count
	sections = append(sections, m.GetStyles().Info.Render(
		fmt.Sprintf(locale.ApplyChangesChangesCount, config.ChangesCount)))
	sections = append(sections, "")

	// show warnings if applicable
	if config.HasCritical {
		sections = append(sections, m.GetStyles().Warning.Render(locale.ApplyChangesWarningCritical))
	}
	if config.HasSecrets {
		sections = append(sections, m.GetStyles().Info.Render(locale.ApplyChangesWarningSecrets))
	}
	if config.HasCritical || config.HasSecrets {
		sections = append(sections, "")
	}

	// show notes
	sections = append(sections, m.GetStyles().Muted.Render(locale.ApplyChangesNoteBackup))
	sections = append(sections, m.GetStyles().Muted.Render(locale.ApplyChangesNoteTime))
	sections = append(sections, "")

	getMaskedValue := func(value string) string {
		maskedValue := strings.Repeat("*", len(value))
		if len(value) > 15 {
			maskedValue = maskedValue[:15] + "..."
		} else if len(value) == 0 {
			maskedValue = locale.ApplyChangesChangesMasked
		}
		return maskedValue
	}

	// list all changes
	for _, change := range config.Changes {
		value := change.NewValue
		if change.Masked {
			value = getMaskedValue(value)
		}

		line := fmt.Sprintf("• %s: %s",
			change.Description,
			m.GetStyles().Info.Render(value))
		sections = append(sections, line)
	}

	return strings.Join(sections, "\n")
}

func (m *ApplyChangesFormModel) IsConfigured() bool {
	return m.getChangesCount() > 0
}

func (m *ApplyChangesFormModel) GetHelpContent() string {
	var sections []string

	config := m.GetController().GetApplyChangesConfig()

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ApplyChangesHelpTitle))
	sections = append(sections, "")

	// show installation or update description based on current state
	if !config.IsInstalled {
		sections = append(sections, locale.ApplyChangesInstallNotFound)
		sections = append(sections, "")

		// add additional components if selected
		if config.LangfuseEnabled {
			sections = append(sections, locale.ApplyChangesInstallFoundLangfuse)
		}
		if config.ObservabilityEnabled {
			sections = append(sections, locale.ApplyChangesInstallFoundObservability)
		}
	} else {
		sections = append(sections, locale.ApplyChangesUpdateFound)
	}

	sections = append(sections, "")
	sections = append(sections, locale.ApplyChangesHelpContent)
	sections = append(sections, "")
	sections = append(sections, m.GetCurrentConfiguration())

	return strings.Join(sections, "\n")
}

func (m *ApplyChangesFormModel) HandleSave() error {
	// saving is handled by the processor integration
	return nil
}

func (m *ApplyChangesFormModel) HandleReset() {
	// reset current changes
	m.GetController().Reset()
}

func (m *ApplyChangesFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// no fields to change in this screen
}

func (m *ApplyChangesFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *ApplyChangesFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

func (m *ApplyChangesFormModel) getChangesCount() int {
	if m.GetController().IsDirty() {
		return m.GetController().GetApplyChangesConfig().ChangesCount
	}

	return 0
}

func (m *ApplyChangesFormModel) handleCompletion(msg processor.ProcessorCompletionMsg) {
	if msg.Operation != processor.ProcessorOperationApplyChanges {
		return
	}

	m.running = false
	if msg.Error != nil {
		m.terminal.Append(fmt.Sprintf("%s: %v\n", locale.ApplyChangesFailed, msg.Error))
	} else {
		switch msg.Operation {
		case processor.ProcessorOperationFactoryReset:
			m.terminal.Append(locale.ApplyChangesResetCompleted)
		case processor.ProcessorOperationApplyChanges:
			m.terminal.Append(locale.ApplyChangesCompleted)
		}
	}

	// rebuild display
	m.updateViewports()
}

func (m *ApplyChangesFormModel) handleApplyChanges() tea.Cmd {
	if m.terminal != nil {
		m.terminal.Clear()
		m.terminal.Append(locale.ApplyChangesInProgress)
	}
	return m.processor.ApplyChanges(context.Background(), processor.WithTerminal(m.terminal))
}

func (m *ApplyChangesFormModel) handleResetChanges() tea.Cmd {
	if m.terminal != nil {
		m.terminal.Clear()
	}

	if err := m.GetController().Reset(); err != nil {
		if m.terminal != nil {
			m.terminal.Append(fmt.Sprintf("%s: %v\n", locale.ApplyChangesFailed, err))
		}
	} else {
		if m.terminal != nil {
			m.terminal.Append(locale.ApplyChangesResetCompleted)
		}
	}

	m.updateViewports()

	return nil
}

// renderLeftPanel renders the terminal output
func (m *ApplyChangesFormModel) renderLeftPanel() string {
	if m.terminal != nil {
		return m.terminal.View()
	}

	// fallback if terminal not initialized
	return m.GetStyles().Error.Render(locale.ApplyChangesTerminalIsNotInitialized)
}

// Update method - handle screen-specific input and messages
func (m *ApplyChangesFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	handleTerminal := func(msg tea.Msg) (tea.Model, tea.Cmd) {
		if m.terminal == nil {
			return m, nil
		}

		updatedModel, cmd := m.terminal.Update(msg)
		if terminalModel := terminal.RestoreModel(updatedModel); terminalModel != nil {
			m.terminal = terminalModel
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		contentWidth, contentHeight := m.getViewportFormSize()

		// update terminal size when window size changes
		if m.terminal != nil {
			if !m.isVerticalLayout() {
				contentWidth -= 2
			}
			m.terminal.SetSize(contentWidth-2, contentHeight-1)
		}

		m.updateViewports()

		return m, nil

	case terminal.TerminalUpdateMsg:
		return handleTerminal(msg)

	case processor.ProcessorCompletionMsg:
		m.handleCompletion(msg)

		return m, m.processor.HandleMsg(msg)

	case processor.ProcessorStartedMsg:
		return m, m.processor.HandleMsg(msg)

	case processor.ProcessorOutputMsg:
		// ignore (handled by terminal)
		return m, m.processor.HandleMsg(msg)

	case processor.ProcessorWaitMsg:
		return m, m.processor.HandleMsg(msg)

	case processor.ProcessorFilesCheckMsg:
		// finish collecting and process result
		m.collecting = false
		if msg.Error != nil {
			if m.terminal != nil {
				m.terminal.Append("")
				m.terminal.Append(fmt.Sprintf("%s: %v", locale.ApplyChangesFailed, msg.Error))
			}
			m.updateViewports()
			return m, m.processor.HandleMsg(msg)
		}
		// filter only modified files
		outdated := map[string]files.FileStatus{}
		for path, st := range msg.Result {
			if st == files.FileStatusModified {
				outdated[path] = st
			}
		}
		m.outdated = outdated
		if len(m.outdated) == 0 {
			if m.terminal != nil {
				m.terminal.Append("")
				m.terminal.Append(locale.ApplyChangesIntegrityNoOutdated)
			}
			m.running = true
			m.updateViewports()
			return m, m.handleApplyChanges()
		}
		m.waitingForChoice = true
		if m.terminal != nil {
			m.terminal.Append("")
			m.terminal.Append(locale.ApplyChangesIntegrityPromptMessage)
			m.terminal.Append("")
			m.terminal.Append(m.renderOutdatedFiles(m.outdated))
			m.terminal.Append("")
			m.terminal.Append("(y/n)")
		}
		m.updateViewports()
		return m, m.processor.HandleMsg(msg)

	case tea.KeyMsg:
		if m.terminal != nil && m.terminal.IsRunning() {
			return handleTerminal(msg)
		}
		switch msg.String() {
		case "enter":
			if !m.running && !m.collecting && !m.waitingForChoice && m.getChangesCount() != 0 {
				m.collecting = true
				if m.terminal != nil {
					m.terminal.Clear()
					m.terminal.Append(locale.ApplyChangesIntegrityPromptTitle)
					m.terminal.Append("")
					m.terminal.Append(locale.ApplyChangesIntegrityChecking)
				}
				return m, m.collectOutdatedFiles()
			}
			return m, nil
		case "y":
			if !m.running && m.waitingForChoice && m.getChangesCount() != 0 {
				m.waitingForChoice = false
				m.running = true
				return m, m.processor.ApplyChanges(context.Background(), processor.WithTerminal(m.terminal), processor.WithForce())
			}
			return m, nil
		case "n":
			if !m.running && m.waitingForChoice && m.getChangesCount() != 0 {
				m.waitingForChoice = false
				m.running = true
				return m, m.handleApplyChanges()
			}
			return m, nil
		case "ctrl+c":
			if (m.collecting || m.waitingForChoice) && !m.running {
				m.collecting = false
				m.waitingForChoice = false
				m.outdated = nil
				if m.terminal != nil {
					m.terminal.Clear()
					if m.getChangesCount() == 0 {
						m.terminal.Append(locale.ApplyChangesNoChanges)
					} else {
						m.terminal.Append(locale.ApplyChangesNotStarted)
						m.terminal.Append("")
						m.terminal.Append(locale.ApplyChangesInstructions)
					}
				}
				m.updateViewports()
				return m, nil
			}
			return m, nil
		case "ctrl+r":
			if !m.running && !m.collecting && !m.waitingForChoice && m.getChangesCount() != 0 {
				return m, m.handleResetChanges()
			}
			return m, nil
		}

		// then pass other keys to terminal for scrolling etc.
		return handleTerminal(msg)

	default:
		return handleTerminal(msg)
	}
}

// Override View to use custom layout
func (m *ApplyChangesFormModel) View() string {
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return locale.UILoading
	}

	if !m.initialized {
		m.handler.BuildForm()
		m.fields = m.GetFormFields()
		m.updateViewports()
	}

	leftPanel := m.renderLeftPanel()
	rightPanel := m.renderHelp()

	if m.isVerticalLayout() {
		return m.renderVerticalLayout(leftPanel, rightPanel, contentWidth, contentHeight)
	}

	return m.renderHorizontalLayout(leftPanel, rightPanel, contentWidth, contentHeight)
}

// GetFormHotKeys returns the hotkeys for this screen
func (m *ApplyChangesFormModel) GetFormHotKeys() []string {
	var hotkeys []string
	if m.terminal != nil && !m.terminal.IsRunning() && m.getChangesCount() != 0 {
		if m.collecting {
			hotkeys = append(hotkeys, "ctrl+c")
		} else if m.waitingForChoice {
			hotkeys = append(hotkeys, "y|n")
			hotkeys = append(hotkeys, "ctrl+c")
		} else if !m.running {
			hotkeys = append(hotkeys, "enter")
			hotkeys = append(hotkeys, "ctrl+r")
		}
	}
	return hotkeys
}

// integrity helpers
func (m *ApplyChangesFormModel) collectOutdatedFiles() tea.Cmd {
	// delegate file check to processor model; messages will be delivered as ProcessorFilesCheckMsg
	return m.processor.CheckFiles(context.Background(), processor.ProductStackAll)
}

func (m *ApplyChangesFormModel) renderOutdatedFiles(outdated map[string]files.FileStatus) string {
	if len(outdated) == 0 {
		return ""
	}

	var b strings.Builder
	list := make([]string, 0, len(outdated))
	for path, st := range outdated {
		if st == files.FileStatusModified {
			list = append(list, path)
		}
	}
	sort.Strings(list)

	for _, path := range list {
		b.WriteString("• ")
		b.WriteString(path)
		b.WriteString("\n")
	}

	return b.String()
}

// Compile-time interface validation
var _ BaseScreenModel = (*ApplyChangesFormModel)(nil)
var _ BaseScreenHandler = (*ApplyChangesFormModel)(nil)
