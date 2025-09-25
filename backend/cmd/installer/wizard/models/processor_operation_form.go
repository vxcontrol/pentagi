package models

import (
	"context"
	"fmt"
	"strings"

	"pentagi/cmd/installer/processor"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/terminal"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// ProcessorOperationFormModel represents a generic processor operation form
type ProcessorOperationFormModel struct {
	*BaseScreen

	// operation details
	stack     processor.ProductStack
	operation processor.ProcessorOperation

	// processor integration
	processor processor.ProcessorModel
	running   bool

	// terminal integration
	terminal terminal.Terminal

	// confirmation state for operations that require it
	waitingForConfirmation bool

	// operation metadata
	operationInfo *processorOperationInfo
}

// processorOperationInfo contains localized information for operations
type processorOperationInfo struct {
	title                string
	description          string
	help                 string
	progressMessage      string
	requiresConfirmation bool
}

// NewProcessorOperationFormModel creates a new processor operation form model
func NewProcessorOperationFormModel(
	c controller.Controller, s styles.Styles, w window.Window,
	p processor.ProcessorModel, stack processor.ProductStack, operation processor.ProcessorOperation,
) *ProcessorOperationFormModel {
	m := &ProcessorOperationFormModel{
		processor: p,
		stack:     stack,
		operation: operation,
	}

	// create base screen with this model as handler
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	// initialize operation info
	m.operationInfo = m.getOperationInfo()

	return m
}

// BaseScreenHandler interface implementation

func (m *ProcessorOperationFormModel) BuildForm() tea.Cmd {
	// no form fields for this screen - it's an action screen
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

	// set initial message: always show welcome + press enter first
	m.terminal.Append(fmt.Sprintf(locale.ProcessorOperationNotStarted, strings.ToLower(m.operationInfo.title)))
	m.terminal.Append("")
	m.terminal.Append(fmt.Sprintf(locale.ProcessorOperationPressEnter, strings.ToLower(m.operationInfo.title)))

	// prevent re-initialization on View() calls
	if !m.initialized {
		m.initialized = true
	} else {
		return nil
	}

	// return terminal's init command
	return m.terminal.Init()
}

func (m *ProcessorOperationFormModel) GetFormTitle() string {
	return fmt.Sprintf(locale.ProcessorOperationFormTitle, m.operationInfo.title)
}

func (m *ProcessorOperationFormModel) GetFormDescription() string {
	return fmt.Sprintf(locale.ProcessorOperationFormDescription, strings.ToLower(m.operationInfo.title))
}

func (m *ProcessorOperationFormModel) GetFormName() string {
	return fmt.Sprintf(locale.ProcessorOperationFormName, m.operationInfo.title)
}

func (m *ProcessorOperationFormModel) GetFormSummary() string {
	// terminal viewport takes all available space
	return ""
}

func (m *ProcessorOperationFormModel) GetFormOverview() string {
	var sections []string

	// title and short purpose
	sections = append(sections, m.GetStyles().Subtitle.Render(m.operationInfo.title))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(m.operationInfo.description))
	sections = append(sections, "")

	// effects and guidance
	sections = append(sections, m.GetStyles().Paragraph.Render(m.operationInfo.help))

	return strings.Join(sections, "\n")
}

func (m *ProcessorOperationFormModel) GetCurrentConfiguration() string {
	var sections []string

	// echo current state and planned actions for clarity
	sections = append(sections, m.renderCurrentStateSummary())

	if planned := m.renderPlannedActions(); planned != "" {
		sections = append(sections, "")
		sections = append(sections, m.GetStyles().Subtitle.Render(locale.ProcessorSectionPlanned))
		sections = append(sections, planned)
	}

	if m.operationInfo.requiresConfirmation {
		// static notice without hotkeys; footer and prompt render exact keys
		sections = append(sections, "")
		sections = append(sections, m.GetStyles().Warning.Render(locale.ProcessorOperationRequiresConfirmationShort))
	}

	return strings.Join(sections, "\n")
}

func (m *ProcessorOperationFormModel) IsConfigured() bool {
	// always ready to execute
	return true
}

func (m *ProcessorOperationFormModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(fmt.Sprintf(locale.ProcessorOperationHelpTitle, m.operationInfo.title)))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(m.operationInfo.description))
	sections = append(sections, "")

	// explain practical effects
	sections = append(sections, m.GetStyles().Paragraph.Render(m.renderEffectsText()))
	sections = append(sections, m.GetCurrentConfiguration())

	return strings.Join(sections, "\n")
}

func (m *ProcessorOperationFormModel) HandleSave() error {
	// no configuration to save
	return nil
}

func (m *ProcessorOperationFormModel) HandleReset() {
	// no configuration to reset
}

func (m *ProcessorOperationFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// no fields to change
}

func (m *ProcessorOperationFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *ProcessorOperationFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// getOperationInfo returns localized information for the operation
func (m *ProcessorOperationFormModel) getOperationInfo() *processorOperationInfo {
	info := &processorOperationInfo{
		requiresConfirmation: false,
	}

	// determine title and description based on operation and stack
	switch m.operation {
	case processor.ProcessorOperationStart:
		info.title = locale.MaintenanceStartPentagi
		info.description = locale.MaintenanceStartPentagiDesc
		info.help = locale.ProcessorHelpStartPentagi
		info.progressMessage = locale.ProcessorOperationStarting

	case processor.ProcessorOperationStop:
		info.title = locale.MaintenanceStopPentagi
		info.description = locale.MaintenanceStopPentagiDesc
		info.help = locale.ProcessorHelpStopPentagi
		info.progressMessage = locale.ProcessorOperationStopping

	case processor.ProcessorOperationRestart:
		info.title = locale.MaintenanceRestartPentagi
		info.description = locale.MaintenanceRestartPentagiDesc
		info.help = locale.ProcessorHelpRestartPentagi
		info.progressMessage = locale.ProcessorOperationRestarting

	case processor.ProcessorOperationDownload:
		if m.stack == processor.ProductStackWorker {
			info.title = locale.MaintenanceDownloadWorkerImage
			info.description = locale.MaintenanceDownloadWorkerImageDesc
			info.help = locale.ProcessorHelpDownloadWorkerImage
		} else {
			info.title = fmt.Sprintf(locale.OperationTitleDownload, string(m.stack))
			info.description = fmt.Sprintf(locale.OperationDescDownloadComponents, string(m.stack))
			info.help = fmt.Sprintf(locale.ProcessorOperationHelpContentDownload, string(m.stack))
		}
		info.progressMessage = locale.ProcessorOperationDownloading

	case processor.ProcessorOperationUpdate:
		switch m.stack {
		case processor.ProductStackWorker:
			info.title = locale.MaintenanceUpdateWorkerImage
			info.description = locale.MaintenanceUpdateWorkerImageDesc
			info.help = locale.ProcessorHelpUpdateWorkerImage
		case processor.ProductStackInstaller:
			info.title = locale.MaintenanceUpdateInstaller
			info.description = locale.MaintenanceUpdateInstallerDesc
			info.help = locale.ProcessorHelpUpdateInstaller
			info.requiresConfirmation = true
		case processor.ProductStackAll, processor.ProductStackCompose:
			info.title = locale.MaintenanceUpdatePentagi
			info.description = locale.MaintenanceUpdatePentagiDesc
			info.help = locale.ProcessorHelpUpdatePentagi
			info.requiresConfirmation = true
		default:
			info.title = fmt.Sprintf(locale.OperationTitleUpdate, string(m.stack))
			info.description = fmt.Sprintf(locale.OperationDescUpdateToLatest, string(m.stack))
			info.help = fmt.Sprintf(locale.ProcessorOperationHelpContentUpdate, string(m.stack))
		}
		info.progressMessage = locale.ProcessorOperationUpdating

	case processor.ProcessorOperationFactoryReset:
		info.title = locale.MaintenanceFactoryReset
		info.description = locale.MaintenanceFactoryResetDesc
		info.help = locale.ProcessorHelpFactoryReset
		info.progressMessage = locale.ProcessorOperationResetting
		info.requiresConfirmation = true

	case processor.ProcessorOperationRemove:
		info.title = locale.MaintenanceRemovePentagi
		info.description = locale.MaintenanceRemovePentagiDesc
		info.help = locale.ProcessorHelpRemovePentagi
		info.progressMessage = locale.ProcessorOperationRemoving

	case processor.ProcessorOperationPurge:
		info.title = locale.MaintenancePurgePentagi
		info.description = locale.MaintenancePurgePentagiDesc
		info.help = locale.ProcessorHelpPurgePentagi
		info.progressMessage = locale.ProcessorOperationPurging
		info.requiresConfirmation = true

	case processor.ProcessorOperationInstall:
		info.title = locale.OperationTitleInstallPentagi
		info.description = locale.OperationDescInstallPentagi
		info.help = locale.ProcessorHelpInstallPentagi
		info.progressMessage = locale.ProcessorOperationInstalling
		info.requiresConfirmation = false

	case processor.ProcessorOperationApplyChanges:
		info.title = locale.ApplyChangesFormTitle
		info.description = locale.ApplyChangesFormDescription
		info.help = locale.ApplyChangesFormOverview
		info.progressMessage = locale.ApplyChangesInProgress
		info.requiresConfirmation = false

	default:
		info.title = fmt.Sprintf(locale.OperationTitleExecute, string(m.operation)+" "+string(m.stack))
		info.description = fmt.Sprintf(locale.OperationDescExecuteOn, string(m.operation), string(m.stack))
		info.help = fmt.Sprintf(locale.ProcessorOperationHelpContent, strings.ToLower(string(m.operation))+" "+string(m.stack))
		info.progressMessage = fmt.Sprintf(locale.OperationProgressExecuting, string(m.operation))
	}

	return info
}

// handleOperation starts the processor operation
func (m *ProcessorOperationFormModel) handleOperation() tea.Cmd {
	if m.terminal != nil {
		m.terminal.Clear()
		m.terminal.Append(fmt.Sprintf(locale.ProcessorOperationInProgress, strings.ToLower(m.operationInfo.title)))
	}

	// determine which operation to execute
	switch m.operation {
	case processor.ProcessorOperationStart:
		return m.processor.Start(context.Background(), m.stack, processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationStop:
		return m.processor.Stop(context.Background(), m.stack, processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationRestart:
		return m.processor.Restart(context.Background(), m.stack, processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationDownload:
		return m.processor.Download(context.Background(), m.stack, processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationUpdate:
		return m.processor.Update(context.Background(), m.stack, processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationFactoryReset:
		return m.processor.FactoryReset(context.Background(), processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationRemove:
		return m.processor.Remove(context.Background(), m.stack, processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationPurge:
		return m.processor.Purge(context.Background(), m.stack, processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationInstall:
		return m.processor.Install(context.Background(), processor.WithTerminal(m.terminal))

	case processor.ProcessorOperationApplyChanges:
		return m.processor.ApplyChanges(context.Background(), processor.WithTerminal(m.terminal))

	default:
		if m.terminal != nil {
			m.terminal.Append(fmt.Sprintf(locale.ProcessorOperationUnknown, m.operation))
		}
		return nil
	}
}

// handleCompletion handles operation completion
func (m *ProcessorOperationFormModel) handleCompletion(msg processor.ProcessorCompletionMsg) {
	m.running = false
	if msg.Error != nil {
		m.terminal.Append(fmt.Sprintf("%s: %v\n", locale.ProcessorOperationFailed, msg.Error))
	} else {
		m.terminal.Append(fmt.Sprintf(locale.ProcessorOperationCompleted, strings.ToLower(m.operationInfo.title)))
	}

	// rebuild display
	m.updateViewports()
}

// renderLeftPanel renders the terminal output
func (m *ProcessorOperationFormModel) renderLeftPanel() string {
	if m.terminal != nil {
		return m.terminal.View()
	}

	// fallback if terminal not initialized
	return m.GetStyles().Error.Render(locale.ProcessorOperationTerminalNotInitialized)
}

// Update method - handle screen-specific input and messages
func (m *ProcessorOperationFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case tea.KeyMsg:
		if m.terminal != nil && m.terminal.IsRunning() {
			return handleTerminal(msg)
		}

		switch msg.String() {
		case "enter":
			if !m.running && !m.waitingForConfirmation {
				if !m.isActionAvailable() {
					return m, nil
				}
				if m.operationInfo.requiresConfirmation {
					m.waitingForConfirmation = true
					if m.terminal != nil {
						m.terminal.Clear()
						m.terminal.Append(fmt.Sprintf(locale.ProcessorOperationConfirmation, strings.ToLower(m.operationInfo.title)))
						m.terminal.Append("")
						m.terminal.Append(locale.ProcessorOperationPressYN)
						m.updateViewports()
					}
					return m, nil
				}
				m.running = true
				return m, m.handleOperation()
			}
			return m, nil

		case "y":
			if !m.running && m.waitingForConfirmation {
				m.waitingForConfirmation = false
				m.running = true
				return m, m.handleOperation()
			}
			return m, nil

		case "n":
			if !m.running && m.waitingForConfirmation {
				m.waitingForConfirmation = false
				if m.terminal != nil {
					m.terminal.Clear()
					m.terminal.Append(locale.ProcessorOperationCancelled)
				}
				m.updateViewports()
				return m, nil
			}
			return m, nil
		}

		// pass other keys to terminal for scrolling etc.
		return handleTerminal(msg)

	default:
		return handleTerminal(msg)
	}
}

// Override View to use custom layout
func (m *ProcessorOperationFormModel) View() string {
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
func (m *ProcessorOperationFormModel) GetFormHotKeys() []string {
	var hotkeys []string
	if m.terminal != nil && !m.terminal.IsRunning() {
		if m.waitingForConfirmation {
			hotkeys = append(hotkeys, "y|n")
		} else if !m.running {
			if m.isActionAvailable() {
				hotkeys = append(hotkeys, "enter")
			}
		}
	}
	return hotkeys
}

// isActionAvailable checks availability using checker helpers per stack/operation
func (m *ProcessorOperationFormModel) isActionAvailable() bool {
	checker := m.GetController().GetChecker()
	switch m.operation {
	case processor.ProcessorOperationStart:
		return checker.CanStartAll()
	case processor.ProcessorOperationStop:
		return checker.CanStopAll()
	case processor.ProcessorOperationRestart:
		return checker.CanRestartAll()
	case processor.ProcessorOperationDownload:
		if m.stack == processor.ProductStackWorker {
			return checker.CanDownloadWorker()
		}
		return true
	case processor.ProcessorOperationUpdate:
		switch m.stack {
		case processor.ProductStackWorker:
			return checker.CanUpdateWorker()
		case processor.ProductStackInstaller:
			return checker.CanUpdateInstaller()
		case processor.ProductStackAll, processor.ProductStackCompose:
			return checker.CanUpdateAll()
		default:
			return true
		}
	case processor.ProcessorOperationFactoryReset:
		return checker.CanFactoryReset()
	case processor.ProcessorOperationRemove:
		return checker.CanRemoveAll()
	case processor.ProcessorOperationPurge:
		return checker.CanPurgeAll()
	case processor.ProcessorOperationInstall:
		return checker.CanInstallAll()
	case processor.ProcessorOperationApplyChanges:
		return m.GetController().IsDirty()
	default:
		return true
	}
}

// renderCurrentStateSummary composes concise current-state block
func (m *ProcessorOperationFormModel) renderCurrentStateSummary() string {
	c := m.GetController().GetChecker()
	var lines []string
	lines = append(lines, m.GetStyles().Subtitle.Render(locale.ProcessorSectionCurrentState))

	comp := func(label string, installed, running bool, modeEmbedded, connected, external bool) string {
		var states []string
		if installed {
			states = append(states, locale.ProcessorStateInstalled)
		} else {
			states = append(states, locale.ProcessorStateMissing)
		}
		if running {
			states = append(states, locale.ProcessorStateRunning)
		} else {
			states = append(states, locale.ProcessorStateStopped)
		}
		if external {
			states = append(states, locale.ProcessorStateExternal)
		} else if modeEmbedded {
			states = append(states, locale.ProcessorStateEmbedded)
		}
		if connected {
			states = append(states, locale.ProcessorStateConnected)
		}
		return "• " + label + ": " + strings.Join(states, ", ")
	}

	lines = append(lines, comp(locale.ProcessorComponentPentagi, c.PentagiInstalled, c.PentagiRunning, true, true, false))
	lines = append(lines, comp(locale.ProcessorComponentLangfuse, c.LangfuseInstalled, c.LangfuseRunning, !c.LangfuseExternal, c.LangfuseConnected, c.LangfuseExternal))
	lines = append(lines, comp(locale.ProcessorComponentObservability, c.ObservabilityInstalled, c.ObservabilityRunning, !c.ObservabilityExternal, c.ObservabilityConnected, c.ObservabilityExternal))

	return strings.Join(lines, "\n")
}

// renderPlannedActions describes high-level plan for the selected operation
func (m *ProcessorOperationFormModel) renderPlannedActions() string {
	c := m.GetController().GetChecker()
	var lines []string

	add := func(prefix, name string, cond bool) {
		if cond {
			lines = append(lines, "• "+prefix+" "+name)
		}
	}

	switch m.operation {
	case processor.ProcessorOperationStart:
		add(locale.PlannedWillStart, locale.ProcessorComponentObservability, c.CanStartAll() && !c.ObservabilityRunning)
		add(locale.PlannedWillStart, locale.ProcessorComponentLangfuse, c.CanStartAll() && !c.LangfuseRunning)
		add(locale.PlannedWillStart, locale.ProcessorComponentPentagi, c.CanStartAll() && !c.PentagiRunning)
	case processor.ProcessorOperationStop:
		add(locale.PlannedWillStop, locale.ProcessorComponentPentagi, c.PentagiRunning)
		add(locale.PlannedWillStop, locale.ProcessorComponentLangfuse, c.LangfuseRunning)
		add(locale.PlannedWillStop, locale.ProcessorComponentObservability, c.ObservabilityRunning)
	case processor.ProcessorOperationRestart:
		add(locale.PlannedWillRestart, locale.ProcessorComponentPentagi, c.PentagiRunning)
		add(locale.PlannedWillRestart, locale.ProcessorComponentLangfuse, c.LangfuseRunning)
		add(locale.PlannedWillRestart, locale.ProcessorComponentObservability, c.ObservabilityRunning)
	case processor.ProcessorOperationUpdate:
		add(locale.PlannedWillUpdate, locale.ProcessorComponentObservability, c.ObservabilityInstalled && !c.ObservabilityIsUpToDate)
		add(locale.PlannedWillUpdate, locale.ProcessorComponentLangfuse, c.LangfuseInstalled && !c.LangfuseIsUpToDate)
		add(locale.PlannedWillUpdate, locale.ProcessorComponentPentagi, c.PentagiInstalled && !c.PentagiIsUpToDate)
		if !(c.PentagiInstalled || c.LangfuseInstalled || c.ObservabilityInstalled) {
			return "" // nothing to show
		}
	case processor.ProcessorOperationDownload:
		add(locale.PlannedWillDownload, locale.ProcessorComponentWorkerImage, c.CanDownloadWorker())
	case processor.ProcessorOperationFactoryReset:
		add(locale.PlannedWillPurge, locale.ProcessorComponentComposeStacks, c.CanFactoryReset())
		add(locale.PlannedWillRestore, locale.ProcessorComponentDefaultFiles, true)
	case processor.ProcessorOperationRemove:
		add(locale.PlannedWillRemove, locale.ProcessorComponentComposeStacks, c.CanRemoveAll())
	case processor.ProcessorOperationPurge:
		add(locale.PlannedWillPurge, locale.ProcessorItemComposeStacksImagesVolumes, c.CanPurgeAll())
	case processor.ProcessorOperationInstall:
		add(locale.PlannedWillDownload, locale.ProcessorItemComposeFiles, c.CanInstallAll())
		add(locale.PlannedWillStart, locale.ProcessorComponentPentagi, true)
	}

	return strings.Join(lines, "\n")
}

// renderEffectsText returns operation-specific effect text
func (m *ProcessorOperationFormModel) renderEffectsText() string {
	switch m.operation {
	case processor.ProcessorOperationStart:
		return locale.EffectsStart
	case processor.ProcessorOperationStop:
		return locale.EffectsStop
	case processor.ProcessorOperationRestart:
		return locale.EffectsRestart
	case processor.ProcessorOperationUpdate:
		if m.stack == processor.ProductStackWorker {
			return locale.EffectsUpdateWorker
		}
		if m.stack == processor.ProductStackInstaller {
			return locale.EffectsUpdateInstaller
		}
		return locale.EffectsUpdateAll
	case processor.ProcessorOperationDownload:
		return locale.EffectsDownloadWorker
	case processor.ProcessorOperationFactoryReset:
		return locale.EffectsFactoryReset
	case processor.ProcessorOperationRemove:
		return locale.EffectsRemove
	case processor.ProcessorOperationPurge:
		return locale.EffectsPurge
	case processor.ProcessorOperationInstall:
		return locale.EffectsInstall
	case processor.ProcessorOperationApplyChanges:
		return locale.ApplyChangesFormOverview
	default:
		return ""
	}
}

// Compile-time interface validation
var _ BaseScreenModel = (*ProcessorOperationFormModel)(nil)
var _ BaseScreenHandler = (*ProcessorOperationFormModel)(nil)
