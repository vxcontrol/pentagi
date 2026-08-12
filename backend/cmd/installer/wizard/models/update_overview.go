package models

import (
	"fmt"
	"strings"
	"time"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UpdateOverviewModel shows what an update would change, before it is applied.
//
// It sits between the menu entry and the operation form: the entry used to lead
// straight into "are you sure?", which asked the user to agree to something the
// interface had never described. Everything shown here is already in the last
// check result — the screen makes no network call and cannot fail.
type UpdateOverviewModel struct {
	styles     styles.Styles
	window     window.Window
	controller controller.Controller

	viewport viewport.Model
	// source is the markdown, and content is that markdown rendered at
	// renderedAt columns. Both are kept because a resize has to re-render from
	// the source — re-wrapping already-rendered text would wrap the ANSI escapes
	// glamour put in it.
	source     string
	content    string
	renderedAt int
	ready      bool
	scrolled   bool
}

// UpdateOverviewRenderedMsg carries markdown already rendered at a given wrap
// width, produced off the update loop.
//
// This document can span every stack's full release history — every
// intermediate release, each with its own changelog and release notes — so it
// has no fixed size the way the EULA's single file does. Calling glamour's
// Render() straight from Update() would run that cost on the update loop
// itself: redraws and key handling both wait for it, so the screen would sit
// looking frozen for however long the render took, only to have the already
// finished content pop in on whatever the next keypress happened to be —
// which reads as "needs a keypress to show up" when what actually happened is
// the render finally returning.
type UpdateOverviewRenderedMsg struct {
	Width   int
	Content string
}

// NewUpdateOverviewModel creates the update overview screen.
func NewUpdateOverviewModel(
	c controller.Controller, s styles.Styles, w window.Window,
) *UpdateOverviewModel {
	return &UpdateOverviewModel{
		styles:     s,
		window:     w,
		controller: c,
	}
}

// Init implements tea.Model
func (m *UpdateOverviewModel) Init() tea.Cmd {
	m.reset()
	return m.loadOverview
}

func (m *UpdateOverviewModel) reset() {
	m.source = ""
	m.content = ""
	m.renderedAt = 0
	m.ready = false
	m.scrolled = false
}

// loadOverview builds the document from the last check result. It is a tea.Cmd
// for symmetry with the other document screens, not because it does any I/O.
func (m *UpdateOverviewModel) loadOverview() tea.Msg {
	check := m.controller.GetChecker()
	if check == nil {
		logger.Errorf("[UpdateOverviewModel] LOAD: no check result")
		return UpdateOverviewLoadedMsg{Source: overviewDocument(nil, false)}
	}
	// UpdateFailure is nil exactly when there is an answer, so it — and not the
	// emptiness of StackUpdates — is what separates "nothing to update" from
	// "we never found out".
	return UpdateOverviewLoadedMsg{
		Source: overviewDocument(check.StackUpdates, check.UpdateFailure == nil),
	}
}

// UpdateOverviewLoadedMsg carries the built markdown.
type UpdateOverviewLoadedMsg struct {
	Source string
}

// Update implements tea.Model
func (m *UpdateOverviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, m.resize()

	case UpdateOverviewLoadedMsg:
		m.source = msg.Source
		return m, m.startRender()

	case UpdateOverviewRenderedMsg:
		m.applyRender(msg)
		return m, func() tea.Msg { return nil }

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Continue. Deliberately NOT page-down as on the EULA screen: this
			// screen asks for no consent, so the one unambiguous key belongs to
			// the one action it offers. The operation form that follows keeps
			// its own confirmation.
			logger.Log("[UpdateOverviewModel] CONTINUE")
			return m, func() tea.Msg {
				// GoBack + Target, same as EULA on acceptance: this screen must
				// not stay in the stack, or ESC from the operation form would
				// bounce the user through it again instead of landing on
				// Maintenance.
				return NavigationMsg{GoBack: true, Target: UpdatePentagiScreen}
			}
		}

		if !m.ready {
			break
		}
		switch msg.String() {
		case "up":
			m.viewport.ScrollUp(1)
		case "down":
			m.viewport.ScrollDown(1)
		case "left":
			m.viewport.ScrollLeft(2)
		case "right":
			m.viewport.ScrollRight(2)
		case "pgup":
			m.viewport.PageUp()
		case "pgdown":
			m.viewport.PageDown()
		case "home":
			m.viewport.GotoTop()
		case "end":
			m.viewport.GotoBottom()
		}
		m.updateScrollStatus()
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		m.updateScrollStatus()
		return m, cmd
	}

	return m, nil
}

// resize applies the new terminal size to the viewport immediately —
// scrolling the content already on screen must keep working while a render is
// in flight — and starts a fresh one only when the wrap width the document
// was last rendered for no longer matches.
//
// GetRendererForWidth clamps, so most resizes land on the same renderer and
// this is a no-op, which matters because a drag-resize sends a WindowSizeMsg
// per column and glamour is not cheap.
func (m *UpdateOverviewModel) resize() tea.Cmd {
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return nil
	}

	if m.ready {
		m.viewport.Width = contentWidth
		m.viewport.Height = contentHeight
		m.updateScrollStatus()
	}

	if m.source == "" || m.renderedAt == contentWidth {
		return nil
	}

	return m.startRender()
}

// startRender renders the current source off the update loop.
//
// This is the fix for the screen appearing to freeze: rendering this
// document — every stack's full release history, concatenated — synchronously
// on Update() would block redraws and key handling for as long as glamour
// takes, which is not bounded the way the EULA's fixed file is. Running it in
// the returned tea.Cmd keeps the update loop free to redraw the screen the
// user is already looking at (and to react to ESC, scrolling, and so on)
// while the render is in flight; the result comes back as its own message
// through the ordinary Update/View cycle, no keypress required.
func (m *UpdateOverviewModel) startRender() tea.Cmd {
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 || m.source == "" {
		return nil
	}

	source := m.source
	sty := m.styles

	return func() tea.Msg {
		rendered, err := sty.GetRendererForWidth(contentWidth).Render(source)
		if err != nil {
			logger.Errorf("[UpdateOverviewModel] RENDER: %v", err)
			rendered = fmt.Sprintf(locale.UpdateOverviewRenderFallback, source, err)
		}
		return UpdateOverviewRenderedMsg{Width: contentWidth, Content: rendered}
	}
}

// applyRender installs a finished render, dropping it if the window has moved
// on to a different width since it was requested.
//
// A fast resize can have more than one render in flight at once; without this
// check, whichever one happened to finish last would win regardless of
// whether it matched the CURRENT width, and the content would flicker between
// wrap widths or settle on the wrong one.
func (m *UpdateOverviewModel) applyRender(msg UpdateOverviewRenderedMsg) {
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 || msg.Width != contentWidth {
		return
	}

	m.content = msg.Content
	m.renderedAt = msg.Width

	if !m.ready {
		m.viewport = viewport.New(contentWidth, contentHeight)
		m.viewport.Style = lipgloss.NewStyle()
		m.ready = true
	} else {
		m.viewport.Width = contentWidth
		m.viewport.Height = contentHeight
	}

	m.viewport.SetContent(m.content)
	m.updateScrollStatus()
}

func (m *UpdateOverviewModel) updateScrollStatus() {
	if m.ready {
		m.scrolled = m.viewport.ScrollPercent() > 0
	}
}

// View implements tea.Model
func (m *UpdateOverviewModel) View() string {
	contentWidth, contentHeight := m.window.GetContentSize()
	if contentWidth <= 0 || contentHeight <= 0 {
		return locale.UpdateOverviewLoading
	}

	if !m.ready || m.content == "" {
		loading := m.styles.Info.Render(locale.UpdateOverviewLoading)
		return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, loading)
	}

	return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Top, m.viewport.View())
}

// GetScrollInfo returns scroll information for footer display.
func (m *UpdateOverviewModel) GetScrollInfo() (scrolled bool, atEnd bool, percent int) {
	if !m.ready {
		return false, false, 0
	}
	return m.scrolled, m.viewport.AtBottom(), int(m.viewport.ScrollPercent() * 100)
}

// overviewDocument turns the last answer into the markdown the screen renders.
//
// Split out from the model so it can be tested without a terminal: the wizard
// has no harness for a bubbletea screen, and the part worth testing is what the
// document SAYS, not how it scrolls.
func overviewDocument(updates []checker.StackUpdate, checkSucceeded bool) string {
	var b strings.Builder
	b.WriteString(locale.UpdateOverviewHeading)
	b.WriteString("\n\n")

	if !checkSucceeded {
		// An update check that failed is not the same as an installation with
		// nothing to do, and saying "up to date" here would be a claim the
		// installer cannot support.
		b.WriteString(locale.UpdateOverviewCheckFailed)
		b.WriteString("\n")
		return b.String()
	}

	var changing, current []checker.StackUpdate
	for _, stack := range updates {
		if stack.HasUpdate {
			changing = append(changing, stack)
		} else {
			current = append(current, stack)
		}
	}

	if len(changing) == 0 {
		b.WriteString(locale.UpdateOverviewNothingToDo)
		b.WriteString("\n")
		return b.String()
	}

	for _, stack := range changing {
		writeStackSection(&b, stack)
	}

	if len(current) > 0 {
		b.WriteString(locale.UpdateOverviewUpToDateTitle)
		b.WriteString("\n\n")
		for _, stack := range current {
			fmt.Fprintf(&b, "- **%s**", stack.Stack)
			if stack.CurrentVersion != "" {
				fmt.Fprintf(&b, " %s", stack.CurrentVersion)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

func writeStackSection(b *strings.Builder, stack checker.StackUpdate) {
	fmt.Fprintf(b, locale.UpdateOverviewStackHeading, stack.Stack)
	b.WriteString("\n\n")

	switch {
	case stack.CurrentVersion != "" && stack.LatestVersion != "":
		fmt.Fprintf(b, locale.UpdateOverviewVersionMove, stack.CurrentVersion, stack.LatestVersion)
		b.WriteString("\n\n")
	case stack.LatestVersion != "":
		// No current version: С1/С6, an installation the server could not place.
		// Naming only the target is the honest form — an arrow from nowhere
		// would imply we know where it starts.
		fmt.Fprintf(b, locale.UpdateOverviewVersionTarget, stack.LatestVersion)
		b.WriteString("\n\n")
		b.WriteString(locale.UpdateOverviewVersionUnknown)
		b.WriteString("\n\n")
	}

	if stack.CurrentVersionMixed {
		b.WriteString(locale.UpdateOverviewVersionMixed)
		b.WriteString("\n\n")
	}

	if len(stack.Components) > 0 {
		b.WriteString(locale.UpdateOverviewComponentsTitle)
		b.WriteString("\n\n")
		for _, component := range stack.Components {
			fmt.Fprintf(b, locale.UpdateOverviewComponentChange,
				component.Component, component.OS, component.Arch,
				componentChangeText(component))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	writeReleaseNotes(b, stack)
}

// componentChangeText says what will happen to one component.
//
// "cannot verify" is its own answer and not a synonym for "unchanged": the
// server omits a digest it does not know, and reporting that as "nothing to do"
// is how an installation is told it is current on no evidence at all.
func componentChangeText(component checker.ComponentUpdate) string {
	switch {
	case !component.Verifiable:
		return locale.UpdateOverviewCannotVerify
	case component.Outdated:
		return locale.UpdateOverviewWillChange
	default:
		return locale.UpdateOverviewNoChange
	}
}

// writeReleaseNotes prints every release the installation crosses, oldest first.
//
// Falls back to the single changelog for an answer that carries no release list
// — a server older than the field, which is every server until the deployment
// that introduced it.
func writeReleaseNotes(b *strings.Builder, stack checker.StackUpdate) {
	if len(stack.Releases) == 0 {
		writeSection(b, stack.Changelog)
		writeSection(b, stack.ReleaseNotes)
		return
	}

	if stack.ReleasesTruncated {
		fmt.Fprintf(b, locale.UpdateOverviewReleasesCut, len(stack.Releases))
		b.WriteString("\n\n")
	}

	for _, release := range stack.Releases {
		heading := locale.UpdateOverviewReleaseHeading
		if !release.IsStable {
			heading = locale.UpdateOverviewReleasePreview
		}
		fmt.Fprintf(b, heading, release.Version)
		b.WriteString("\n\n")

		if released, err := time.Parse(time.RFC3339, release.ReleasedAt); err == nil {
			fmt.Fprintf(b, locale.UpdateOverviewReleaseDate, released.Format("2006-01-02"))
			b.WriteString("\n\n")
		}

		writeSection(b, release.Changelog)
		writeSection(b, release.ReleaseNotes)
	}
}

func writeSection(b *strings.Builder, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	b.WriteString(text)
	b.WriteString("\n\n")
}

// BaseScreenModel interface implementation

// GetFormTitle returns empty title — glamour renders its own heading from the top.
func (m *UpdateOverviewModel) GetFormTitle() string {
	return ""
}

// GetFormDescription returns the description for the form (right panel)
func (m *UpdateOverviewModel) GetFormDescription() string {
	return locale.UpdateOverviewFormDescription
}

// GetFormName returns the name for the form (right panel).
//
// Deliberately the same string the operation form uses: this screen replaced
// that one as the destination of the menu entry, and the entry must keep reading
// the way it always did.
func (m *UpdateOverviewModel) GetFormName() string {
	return locale.UpdateOverviewFormName
}

// GetFormOverview returns form overview for list screens (right panel)
func (m *UpdateOverviewModel) GetFormOverview() string {
	return locale.UpdateOverviewFormOverview
}

// GetCurrentConfiguration returns text with current configuration for list screens
func (m *UpdateOverviewModel) GetCurrentConfiguration() string {
	check := m.controller.GetChecker()
	if check != nil && check.CanUpdateAll() {
		return locale.UpdateOverviewConfigurationOK
	}
	return locale.UpdateOverviewConfigurationNo
}

// IsConfigured reports whether there is an update to apply.
func (m *UpdateOverviewModel) IsConfigured() bool {
	check := m.controller.GetChecker()
	return check != nil && check.CanUpdateAll()
}

// GetFormHotKeys returns the hotkeys for the form (layout footer)
func (m *UpdateOverviewModel) GetFormHotKeys() []string {
	hotkeys := []string{"enter"}
	if m.ready && m.content != "" {
		hotkeys = append(hotkeys, "up|down", "pgup|pgdown", "home|end")
	}
	return hotkeys
}

// Compile-time interface validation
var _ BaseScreenModel = (*UpdateOverviewModel)(nil)
