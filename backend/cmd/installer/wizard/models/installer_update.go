package models

import (
	"context"
	"fmt"
	"strings"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/processor"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/terminal"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// InstallerUpdateModel is the screen behind the "Update Installer" entry.
//
// It replaced the generic operation form, which asked "are you sure?" about an operation
// it described wrongly: the help text promised to replace the running binary and exit,
// and no version of this installer has ever done that. What actually happens is a
// download of one verified file next to the installation.
//
// So this screen names the version, the size and the path FIRST, and then offers a single
// action. The confirmation prompt is gone with the old text: it existed to guard a
// destructive step that was never taken, and a y/n on top of a described download is
// ceremony, not safety.
type InstallerUpdateModel struct {
	*BaseScreen

	processor processor.ProcessorModel
	terminal  terminal.Terminal

	// offered is the server's description of the build; offerErr is why there is none.
	// Both nil means the answer has not arrived yet.
	offered  *processor.InstallerPackage
	offerErr error

	running    bool
	confirming bool
}

// NewInstallerUpdateModel creates the installer self-update screen.
func NewInstallerUpdateModel(
	c controller.Controller, s styles.Styles, w window.Window, p processor.ProcessorModel,
) *InstallerUpdateModel {
	m := &InstallerUpdateModel{processor: p}
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// InstallerPackageLoadedMsg carries the server's description of the offered build.
type InstallerPackageLoadedMsg struct {
	Package *processor.InstallerPackage
	Err     error
}

// Init implements tea.Model.
//
// The description is fetched on every entry rather than once: the offered version changes
// when an update check is re-run while the app is open. Asking again is nearly free — the
// processor keeps the last answer and only goes to the network when the version or the
// platform it describes no longer matches.
func (m *InstallerUpdateModel) Init() tea.Cmd {
	m.running = false
	m.confirming = false
	m.offered = nil
	m.offerErr = nil

	return tea.Batch(m.BaseScreen.Init(), m.loadPackage)
}

func (m *InstallerUpdateModel) loadPackage() tea.Msg {
	pkg, err := m.processor.InstallerPackage(context.Background())
	if err != nil {
		logger.Errorf("[InstallerUpdateModel] PACKAGE: %v", err)
	}

	return InstallerPackageLoadedMsg{Package: pkg, Err: err}
}

// BaseScreenHandler interface implementation

func (m *InstallerUpdateModel) BuildForm() tea.Cmd {
	// no form fields for this screen - it's an action screen
	m.SetFormFields([]FormField{})

	contentWidth, contentHeight := m.getViewportFormSize()

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

	m.terminal.Append(locale.InstallerUpdateAsking)

	// prevent re-initialization on View() calls
	if !m.initialized {
		m.initialized = true
		return m.terminal.Init()
	}

	return nil
}

func (m *InstallerUpdateModel) GetFormSummary() string {
	// terminal viewport takes all available space
	return ""
}

func (m *InstallerUpdateModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.InstallerUpdateHelpTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.InstallerUpdateFormOverview))
	sections = append(sections, "")
	sections = append(sections, m.GetCurrentConfiguration())

	return strings.Join(sections, "\n")
}

func (m *InstallerUpdateModel) HandleSave() error { return nil }

func (m *InstallerUpdateModel) HandleReset() {}

func (m *InstallerUpdateModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {}

func (m *InstallerUpdateModel) GetFormFields() []FormField { return m.fields }

func (m *InstallerUpdateModel) SetFormFields(fields []FormField) { m.fields = fields }

// Update implements tea.Model
func (m *InstallerUpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		if m.terminal != nil {
			if !m.isVerticalLayout() {
				contentWidth -= 2
			}
			m.terminal.SetSize(contentWidth-2, contentHeight-1)
		}

		m.updateViewports()

		return m, nil

	case InstallerPackageLoadedMsg:
		m.offered, m.offerErr = msg.Package, msg.Err
		m.describeOffer()
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
			if !m.canDownload() || m.confirming {
				return m, nil
			}
			// The one thing worth asking about: a file already occupies the name we
			// are about to write. Downloading over nothing is not a question.
			if m.offered.Downloaded {
				m.confirming = true
				m.askToOverwrite()
				m.updateViewports()
				return m, nil
			}

			return m, m.startDownload()

		case "y":
			if !m.confirming || m.running {
				return m, nil
			}
			m.confirming = false

			return m, m.startDownload()

		case "n":
			if !m.confirming || m.running {
				return m, nil
			}
			m.confirming = false
			// Declining is not a dead end: the file the user chose to keep is the very
			// build they came here for, so what they need next is the command that puts
			// it in place — the same one a fresh download would have printed.
			m.showLines(append([]string{locale.InstallerUpdateKeptExisting},
				installerMoveInstructions(m.offered)...))
			m.updateViewports()

			return m, nil
		}

		// pass other keys to terminal for scrolling etc.
		return handleTerminal(msg)

	default:
		return handleTerminal(msg)
	}
}

// startDownload runs the operation and reports it in the panel.
func (m *InstallerUpdateModel) startDownload() tea.Cmd {
	logger.Log("[InstallerUpdateModel] DOWNLOAD: %s", m.offered.Version)
	m.running = true
	m.showLines([]string{locale.InstallerUpdateInProgress})

	return m.processor.Update(context.Background(),
		processor.ProductStackInstaller, processor.WithTerminal(m.terminal))
}

// canDownload reports whether Enter has anything to do.
//
// The description has to be in hand: it carries the length, the sha256 and the signature
// the download is checked against, so without it there is nothing to verify against and
// the operation would only fail further in.
func (m *InstallerUpdateModel) canDownload() bool {
	return !m.running && m.offered != nil && m.offerErr == nil
}

// showLines replaces the panel contents.
func (m *InstallerUpdateModel) showLines(lines []string) {
	if m.terminal == nil {
		return
	}

	m.terminal.Clear()
	for _, line := range lines {
		m.terminal.Append(line)
	}
}

// describeOffer replaces the panel contents with what is on offer.
func (m *InstallerUpdateModel) describeOffer() {
	m.showLines(installerUpdateLines(checker.InstallerVersion, m.offered, m.offerErr))
}

func (m *InstallerUpdateModel) askToOverwrite() {
	m.showLines([]string{
		locale.InstallerUpdateAlreadyHere,
		"",
		fmt.Sprintf(locale.InstallerUpdateOverwritePrompt, m.offered.Path),
		"",
		locale.InstallerUpdatePressYN,
	})
}

func (m *InstallerUpdateModel) handleCompletion(msg processor.ProcessorCompletionMsg) {
	m.running = false
	if m.terminal == nil {
		return
	}

	if msg.Error != nil {
		m.terminal.Append(fmt.Sprintf("%s: %v\n", locale.InstallerUpdateFailed, msg.Error))
		m.updateViewports()

		return
	}

	// The operation already said where the file landed and that it verified, so nothing
	// here repeats that. What it could not say is what to do next, because only the screen
	// knows the file is now sitting there under a name nothing will run by itself.
	m.offered.Downloaded = true
	for _, line := range installerMoveInstructions(m.offered) {
		m.terminal.Append(line)
	}

	m.updateViewports()
}

// installerMoveInstructions is the last thing the screen says: where the build is, and the
// exact command that puts it in place.
//
// Printed rather than run. A process cannot reliably replace the binary it is executing —
// Linux refuses the write, Windows holds the file locked — and an installer that tries and
// fails halfway leaves the user with no installer at all. Handing over the command keeps
// the decision, and the timing, with the person who can make it.
func installerMoveInstructions(offered *processor.InstallerPackage) []string {
	if offered == nil {
		return nil
	}

	lines := []string{"", fmt.Sprintf(locale.InstallerUpdateFileIsAt, offered.Path)}

	if command := offered.MoveCommand(); command != "" {
		lines = append(lines, "", locale.InstallerUpdateMoveTitle, "",
			fmt.Sprintf(locale.InstallerUpdateMoveIndent, command))
	} else {
		// os.Executable told us nothing, so there is no path to put in a command. Saying
		// what has to happen beats printing a command with a guess in it.
		lines = append(lines, "", locale.InstallerUpdateMoveUnknown)
	}

	// Only Windows needs this: renaming over a running binary is ordinary on Linux and
	// macOS — the running process keeps the inode it already opened.
	if offered.OS == "windows" {
		lines = append(lines, "", locale.InstallerUpdateWindowsNote)
	}

	return lines
}

// installerUpdateLines is what the screen says before anything is downloaded.
//
// Pure so the wording can be tested: the wizard has no harness for a bubbletea screen, and
// the part worth testing is what the text CLAIMS — this screen exists because the text it
// replaced claimed the running binary would be swapped and the app would exit.
func installerUpdateLines(running string, offered *processor.InstallerPackage, err error) []string {
	switch {
	case err != nil:
		return []string{fmt.Sprintf(locale.InstallerUpdateLoadFailed, err)}
	case offered == nil:
		return []string{locale.InstallerUpdateAsking}
	}

	lines := []string{
		fmt.Sprintf(locale.InstallerUpdateCurrentVersion, running),
		fmt.Sprintf(locale.InstallerUpdateOfferedVersion, offered.Version),
		fmt.Sprintf(locale.InstallerUpdatePlatform, offered.OS, offered.Arch),
		fmt.Sprintf(locale.InstallerUpdateSize, processor.FormatSize(offered.Size)),
		fmt.Sprintf(locale.InstallerUpdateTarget, offered.Path),
		"",
		locale.InstallerUpdateManualNote,
	}

	// The entry only appears when the server says there is a newer build, but that answer
	// can be older than the screen — a re-check, or a strategy changed under it. Printing
	// the two versions without a word between them reads as an update either way.
	if offered.Version == running {
		lines = append(lines, "", locale.InstallerUpdateSameVersion)
	}

	// Said here as well as in the confirmation, so pressing Enter is not the moment the
	// user first learns something is about to be replaced.
	if offered.Downloaded {
		lines = append(lines, "", locale.InstallerUpdateAlreadyHere)
	}

	return append(lines, "", locale.InstallerUpdatePressEnter)
}

// Override View to use custom layout
func (m *InstallerUpdateModel) View() string {
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return locale.UILoading
	}

	if !m.initialized {
		m.handler.BuildForm()
		m.fields = m.GetFormFields()
		m.updateViewports()
	}

	leftPanel := m.GetStyles().Error.Render(locale.ProcessorOperationTerminalNotInitialized)
	if m.terminal != nil {
		leftPanel = m.terminal.View()
	}

	rightPanel := m.renderHelp()

	if m.isVerticalLayout() {
		return m.renderVerticalLayout(leftPanel, rightPanel, contentWidth, contentHeight)
	}

	return m.renderHorizontalLayout(leftPanel, rightPanel, contentWidth, contentHeight)
}

// BaseScreenModel interface implementation

func (m *InstallerUpdateModel) GetFormTitle() string {
	return locale.MaintenanceUpdateInstaller
}

func (m *InstallerUpdateModel) GetFormDescription() string {
	return locale.MaintenanceUpdateInstallerDesc
}

// GetFormName returns the name shown in the maintenance list. Deliberately the string the
// operation form used, so the entry reads the way it always did.
func (m *InstallerUpdateModel) GetFormName() string {
	return locale.MaintenanceUpdateInstaller
}

func (m *InstallerUpdateModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MaintenanceUpdateInstaller))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.MaintenanceUpdateInstallerDesc))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.InstallerUpdateFormOverview))

	return strings.Join(sections, "\n")
}

func (m *InstallerUpdateModel) GetCurrentConfiguration() string {
	if m.offerErr != nil {
		return m.GetStyles().Error.Render(fmt.Sprintf(locale.InstallerUpdateLoadFailed, m.offerErr))
	}

	check := m.GetController().GetChecker()
	if check != nil && check.CanUpdateInstaller() {
		return locale.InstallerUpdateConfigurationOK
	}

	return locale.InstallerUpdateConfigurationNo
}

// IsConfigured reports whether there is a newer build to fetch.
func (m *InstallerUpdateModel) IsConfigured() bool {
	check := m.GetController().GetChecker()
	return check != nil && check.CanUpdateInstaller()
}

func (m *InstallerUpdateModel) GetFormHotKeys() []string {
	if m.terminal != nil && m.terminal.IsRunning() {
		return nil
	}
	if m.confirming {
		return []string{"y|n"}
	}
	if !m.canDownload() {
		return nil
	}

	return []string{"enter"}
}

// Compile-time interface validation
var _ BaseScreenModel = (*InstallerUpdateModel)(nil)
var _ BaseScreenHandler = (*InstallerUpdateModel)(nil)
