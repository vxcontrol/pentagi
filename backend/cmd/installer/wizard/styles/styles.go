package styles

import (
	"os"
	"slices"
	"strings"
	"sync"

	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"

	"github.com/charmbracelet/glamour"
	glamourstyles "github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// Colors defines the color palette for the installer
var (
	Primary     = lipgloss.Color("#7D56F4") // Purple
	Secondary   = lipgloss.Color("#04B575") // Green
	Accent      = lipgloss.Color("#FFD700") // Gold
	Success     = lipgloss.Color("#00FF00") // Bright Green
	Error       = lipgloss.Color("#FF0000") // Red
	Warning     = lipgloss.Color("#FFA500") // Orange
	Info        = lipgloss.Color("#00BFFF") // Sky Blue
	Muted       = lipgloss.Color("#888888") // Gray
	Background  = lipgloss.Color("#1A1A1A") // Dark Gray
	Foreground  = lipgloss.Color("#FFFFFF") // White
	Border      = lipgloss.Color("#444444") // Dark Border
	Placeholder = lipgloss.Color("#666666") // Gray
	Black       = lipgloss.Color("#000000") // Black
)

// Styles contains all styled components for the installer
type Styles struct {
	// Layout styles
	Header  lipgloss.Style
	Content lipgloss.Style
	Footer  lipgloss.Style

	// Component styles
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Paragraph lipgloss.Style
	Logo      lipgloss.Style

	// Status styles
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style

	// Interactive styles
	Button          lipgloss.Style
	ButtonActive    lipgloss.Style
	List            lipgloss.Style
	ListItem        lipgloss.Style
	ListSelected    lipgloss.Style
	ListDisabled    lipgloss.Style
	ListHighlighted lipgloss.Style

	// Form styles
	FormField       lipgloss.Style
	FormInput       lipgloss.Style
	FormLabel       lipgloss.Style
	FormHelp        lipgloss.Style
	FormError       lipgloss.Style
	FormPlaceholder lipgloss.Style
	FormPagination  lipgloss.Style

	// Special components
	StatusCheck lipgloss.Style
	ASCIIArt    lipgloss.Style
	Markdown    lipgloss.Style

	// Additional styles
	Muted  lipgloss.Style
	Border lipgloss.Color

	// Markdown renderer
	renderer *glamour.TermRenderer
	// renderers memoises the width-aware renderers. A POINTER, deliberately:
	// Styles is copied by value into every screen, and building a glamour
	// renderer is not free, so the cache has to be shared rather than cloned.
	renderers *rendererCache
	// markdownStyle is the dark/light glamour style every renderer in this
	// process uses, resolved exactly once in New(). See resolveMarkdownStyle
	// for why this is not glamour.WithAutoStyle() called fresh per renderer.
	markdownStyle string
}

// Word-wrap bounds for rendered markdown.
//
// The floor keeps a narrow terminal from breaking every changelog line into two
// or three; the ceiling exists because a line running the full width of a
// maximised terminal is measurably harder to read than one that stops — the same
// reason every prose medium has a column width.
const (
	MinRenderWidth     = 40
	MaxRenderWidth     = 120
	DefaultRenderWidth = 80
)

// rendererCache hands out one renderer per wrap width.
//
// Locked because renderers are built inside tea.Cmd goroutines: a screen loads
// its markdown off the update loop, which is exactly where two screens can ask
// for the same width at once.
type rendererCache struct {
	mu sync.Mutex
	by map[int]*glamour.TermRenderer
}

// New creates a new styles instance with default values
func New() Styles {
	markdownStyle := resolveMarkdownStyle()

	// Create glamour renderer for markdown
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(markdownStyle),
		glamour.WithWordWrap(DefaultRenderWidth),
	)
	if err != nil {
		logger.Errorf("[Styles] NEW: error creating renderer: %v", err)
	}

	s := Styles{
		renderer:      renderer,
		renderers:     &rendererCache{by: map[int]*glamour.TermRenderer{}},
		markdownStyle: markdownStyle,
	}
	s.initializeStyles()

	return s
}

// resolveMarkdownStyle decides dark vs light exactly once, here, at New() —
// which the wizard calls before tea.NewProgram(...).Run() takes over the
// terminal.
//
// glamour.WithAutoStyle() makes this same decision by calling
// termenv.HasDarkBackground(), which queries the terminal directly (an OSC
// background-color request over stdout, read back from stdin) and — on
// termenv's default, uncached output — does that full round trip again on
// EVERY call, not just the first. GetRendererForWidth calls WithAutoStyle()
// fresh for every wrap width it has not seen yet, and it does so WHILE the
// program is running: bubbletea's own input reader already owns stdin at
// that point, so the query and the reader are pulling on the same bytes.
// Whichever one loses that race sits waiting — for termenv's read timeout, or
// for the next keypress to hand it something to (mis)read — and that is what
// the screen looked frozen for. Resolving the choice once, here, before
// bubbletea has claimed the terminal, and reusing the answer for every
// renderer afterwards removes the query from the running program entirely
// instead of hoping it loses the race quickly.
func resolveMarkdownStyle() string {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return glamourstyles.NoTTYStyle
	}
	if termenv.HasDarkBackground() {
		return glamourstyles.DarkStyle
	}
	return glamourstyles.LightStyle
}

// GetRenderer returns the markdown renderer at the default width.
//
// Prefer GetRendererForWidth on any screen that knows how much room it has: this
// one wraps at 80 columns whatever the terminal is doing.
func (s *Styles) GetRenderer() *glamour.TermRenderer {
	return s.renderer
}

// GetRendererForWidth returns a markdown renderer wrapping at the given width,
// clamped to a readable band and memoised.
//
// It never returns nil: a width that cannot be built falls back to the default
// renderer, and a caller that got nil would have to check for it on every render
// path — which is how a changelog ends up not being shown at all.
func (s *Styles) GetRendererForWidth(width int) *glamour.TermRenderer {
	width = min(max(width, MinRenderWidth), MaxRenderWidth)

	if s.renderers == nil {
		return s.renderer
	}

	s.renderers.mu.Lock()
	defer s.renderers.mu.Unlock()

	if cached, ok := s.renderers.by[width]; ok {
		return cached
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(s.markdownStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		logger.Errorf("[Styles] RENDERER: width %d: %v", width, err)
		return s.renderer
	}

	s.renderers.by[width] = renderer
	return renderer
}

// initializeStyles sets up all the style definitions
func (s *Styles) initializeStyles() {
	// Layout styles
	s.Header = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Height(1).
		PaddingLeft(2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(Border)

	s.Content = lipgloss.NewStyle().
		Padding(1).
		Margin(0)

	s.Footer = lipgloss.NewStyle().
		Foreground(Muted).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(Border)

	// Typography styles
	s.Title = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(1)

	s.Subtitle = lipgloss.NewStyle().
		Foreground(Secondary).
		Bold(false).
		MarginBottom(1)

	s.Paragraph = lipgloss.NewStyle().
		Foreground(Foreground).
		MarginBottom(1)

	s.Logo = lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true)

	// Status styles
	s.Success = lipgloss.NewStyle().
		Foreground(Success).
		Bold(true)

	s.Error = lipgloss.NewStyle().
		Foreground(Error).
		Bold(true)

	s.Warning = lipgloss.NewStyle().
		Foreground(Warning).
		Bold(true)

	s.Info = lipgloss.NewStyle().
		Foreground(Info)

	// Interactive styles
	s.Button = lipgloss.NewStyle().
		Foreground(Primary).
		Background(Background).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(0, 1)

	s.ButtonActive = s.Button.
		Foreground(Background).
		Background(Primary).
		Bold(true)

	s.List = lipgloss.NewStyle().
		MarginLeft(1)

	s.ListItem = lipgloss.NewStyle().
		Foreground(Foreground).
		PaddingLeft(2)

	s.ListSelected = s.ListItem.
		Foreground(Primary).
		Bold(true).
		PaddingLeft(0)

	s.ListDisabled = s.ListItem.
		Foreground(Muted)

	s.ListHighlighted = s.ListItem.
		Foreground(Accent).
		Bold(true)

	// Form styles
	s.FormField = lipgloss.NewStyle().
		MarginBottom(1)

	s.FormInput = lipgloss.NewStyle().
		Foreground(Foreground).
		Background(Background).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(0, 1)

	s.FormLabel = lipgloss.NewStyle().
		Foreground(Secondary).
		Bold(true).
		MarginBottom(0)

	s.FormHelp = lipgloss.NewStyle().
		Foreground(Muted).
		Italic(true)

	s.FormError = lipgloss.NewStyle().
		Foreground(Error).
		Bold(true)

	s.FormPlaceholder = s.FormInput.
		Foreground(Placeholder)

	s.FormPagination = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(Black)

	// Special components
	s.StatusCheck = lipgloss.NewStyle().
		Bold(true)

	s.ASCIIArt = lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(1)

	s.Markdown = lipgloss.NewStyle().
		Foreground(Foreground)

	// Additional styles
	s.Muted = lipgloss.NewStyle().
		Foreground(Muted)

	s.Border = Border
}

// RenderStatusIcon returns a styled status icon
func (s *Styles) RenderStatusIcon(success bool) string {
	if success {
		return s.Success.Render("✓")
	}
	return s.Error.Render("✗")
}

// RenderStatusText returns styled status text with icon
func (s *Styles) RenderStatusText(text string, success bool) string {
	icon := s.RenderStatusIcon(success)
	style := s.Success
	if !success {
		style = s.Error
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, icon, " ", style.Render(text))
}

// RenderMenuItem returns a styled menu item
func (s *Styles) RenderMenuItem(text string, selected bool, disabled bool, highlighted bool) string {
	if disabled {
		return s.ListDisabled.Render("  " + text)
	}
	if selected {
		return s.ListSelected.Render("> " + text)
	}
	if highlighted {
		return s.ListHighlighted.Render("  " + text)
	}
	return s.ListItem.Render("  " + text)
}

// RenderASCIILogo returns the PentAGI ASCII art logo
func (s *Styles) RenderASCIILogo(width int) string {
	logo := `
 ██████╗ ███████╗███╗   ██╗████████╗ █████╗  ██████╗ ██╗
 ██╔══██╗██╔════╝████╗  ██║╚══██╔══╝██╔══██╗██╔════╝ ██║
 ██████╔╝█████╗  ██╔██╗ ██║   ██║   ███████║██║  ███╗██║
 ██╔═══╝ ██╔══╝  ██║╚██╗██║   ██║   ██╔══██║██║   ██║██║
 ██║     ███████╗██║ ╚████║   ██║   ██║  ██║╚██████╔╝██║
 ╚═╝     ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚═╝
 `

	// cut logo to width if it's too wide otherwise use full width and center it
	return s.ASCIIArt.
		Width(max(width, lipgloss.Width(logo))).
		MarginTop(3).
		Render(logo)
}

// RenderFooter returns a styled footer
func (s *Styles) RenderFooter(actions []string, width int) string {
	footerText := strings.Join(actions, locale.NavSeparator)
	footerPadding := 2 // left and right padding
	footerWidth := lipgloss.Width(footerText) + footerPadding

	if footerWidth > width {
		for count := divCeil(footerWidth, width); count <= len(actions); count++ {
			footerLines := make([]string, 0, count)
			for chunk := range slices.Chunk(actions, divCeil(len(actions), count)) {
				footerLines = append(footerLines, strings.Join(chunk, locale.NavSeparator))
			}
			footerText = strings.Join(footerLines, "\n")
			if lipgloss.Width(footerText)+footerPadding <= width {
				break
			}
		}
	}

	return lipgloss.NewStyle().
		Width(max(width, lipgloss.Width(footerText)+footerPadding)).
		Background(s.Border).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1, 0, 1).
		Render(footerText)
}

func divCeil(a, b int) int {
	if a%b == 0 {
		return a / b
	}
	return a/b + 1
}
