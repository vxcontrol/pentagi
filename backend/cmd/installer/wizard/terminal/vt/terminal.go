package vt

import (
	"image/color"
	"io"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// Logger represents a logger interface.
type Logger interface {
	Printf(format string, v ...any)
}

// Terminal represents a virtual terminal.
type Terminal struct {
	handlers

	// The terminal's indexed 256 colors.
	colors [256]color.Color

	// Both main and alt screens and a pointer to the currently active screen.
	scrs [2]Screen
	scr  *Screen

	// Character sets
	charsets [4]CharSet

	// log is the logger to use.
	logger Logger

	// terminal default colors.
	defaultFg, defaultBg, defaultCur color.Color
	fgColor, bgColor, curColor       color.Color

	// Terminal modes.
	modes ansi.Modes

	// The last written character.
	lastChar rune // either ansi.Rune or ansi.Grapheme
	// A slice of runes to compose a grapheme.
	grapheme []rune

	// The ANSI parser to use.
	parser *ansi.Parser
	// The last parser state.
	lastState parser.State

	cb Callbacks

	// The terminal's icon name and title.
	iconName, title string
	// The current reported working directory. This is not validated.
	cwd string

	// tabstop is the list of tab stops.
	tabstops *uv.TabStops

	// I/O pipes.
	pw io.Writer

	// The GL and GR character set identifiers.
	gl, gr  int
	gsingle int // temporarily select GL or GR

	// atPhantom indicates if the cursor is out of bounds.
	// When true, and a character is written, the cursor is moved to the next line.
	atPhantom bool
}

// NewTerminal creates a new terminal.
func NewTerminal(w, h int, pw io.Writer) *Terminal {
	t := new(Terminal)
	t.scrs[0] = *NewScreen(w, h)
	t.scrs[1] = *NewScreen(w, h)
	t.scr = &t.scrs[0]
	t.scrs[0].cb = &t.cb
	t.scrs[1].cb = &t.cb
	t.parser = ansi.NewParser()
	t.parser.SetParamsSize(parser.MaxParamsSize)
	t.parser.SetDataSize(1024 * 1024 * 4) // 4MB data buffer
	t.parser.SetHandler(ansi.Handler{
		Print:     t.handlePrint,
		Execute:   t.handleControl,
		HandleCsi: t.handleCsi,
		HandleEsc: t.handleEsc,
		HandleDcs: t.handleDcs,
		HandleOsc: t.handleOsc,
		HandleApc: t.handleApc,
		HandlePm:  t.handlePm,
		HandleSos: t.handleSos,
	})
	t.pw = pw
	t.resetModes()
	t.tabstops = uv.DefaultTabStops(w)
	t.registerDefaultHandlers()

	return t
}

// SetLogger sets the terminal's logger.
func (t *Terminal) SetLogger(l Logger) {
	t.logger = l
}

// SetCallbacks sets the terminal's callbacks.
func (t *Terminal) SetCallbacks(cb Callbacks) {
	t.cb = cb
	t.scrs[0].cb = &t.cb
	t.scrs[1].cb = &t.cb
}

// Touched returns the touched lines in the current screen buffer.
func (t *Terminal) Touched() []*uv.LineData {
	return t.scr.Touched()
}

var _ uv.Screen = (*Terminal)(nil)

// Bounds returns the bounds of the terminal.
func (t *Terminal) Bounds() uv.Rectangle {
	return t.scr.Bounds()
}

// CellAt returns the current focused screen cell at the given x, y position.
// It returns nil if the cell is out of bounds.
func (t *Terminal) CellAt(x, y int) *uv.Cell {
	return t.scr.CellAt(x, y)
}

// SetCell sets the current focused screen cell at the given x, y position.
func (t *Terminal) SetCell(x, y int, c *uv.Cell) {
	t.scr.SetCell(x, y, c)
}

// WidthMethod returns the width method used by the terminal.
func (t *Terminal) WidthMethod() uv.WidthMethod {
	if t.isModeSet(ansi.UnicodeCoreMode) {
		return ansi.GraphemeWidth
	}
	return ansi.WcWidth
}

// Draw implements the [uv.Drawable] interface.
func (t *Terminal) Draw(scr uv.Screen, area uv.Rectangle) {
	bg := uv.EmptyCell
	bg.Style.Bg = t.bgColor
	screen.FillArea(scr, &bg, area)
	for y := range t.Touched() {
		if y < 0 || y >= t.Height() {
			continue
		}
		for x := 0; x < t.Width(); {
			w := 1
			cell := t.CellAt(x, y)
			if cell != nil {
				cell = cell.Clone()
				if cell.Width > 1 {
					w = cell.Width
				}
				if cell.Style.Bg == nil && t.bgColor != nil {
					cell.Style.Bg = t.bgColor
				}
				if cell.Style.Fg == nil && t.fgColor != nil {
					cell.Style.Fg = t.fgColor
				}
				scr.SetCell(x+area.Min.X, y+area.Min.Y, cell)
			}
			x += w
		}
	}
}

// Height returns the height of the terminal.
func (t *Terminal) Height() int {
	return t.scr.Height()
}

// Width returns the width of the terminal.
func (t *Terminal) Width() int {
	return t.scr.Width()
}

// CursorPosition returns the terminal's cursor position.
func (t *Terminal) CursorPosition() uv.Position {
	x, y := t.scr.CursorPosition()
	return uv.Pos(x, y)
}

// Resize resizes the terminal.
func (t *Terminal) Resize(width int, height int) {
	x, y := t.scr.CursorPosition()
	if t.atPhantom {
		if x < width-1 {
			t.atPhantom = false
			x++
		}
	}

	if y < 0 {
		y = 0
	}
	if y >= height {
		y = height - 1
	}
	if x < 0 {
		x = 0
	}
	if x >= width {
		x = width - 1
	}

	t.scrs[0].Resize(width, height)
	t.scrs[1].Resize(width, height)
	t.tabstops = uv.DefaultTabStops(width)

	t.setCursor(x, y)
}

// Write writes data to the terminal output buffer.
func (t *Terminal) Write(p []byte) (n int, err error) {
	for i := range p {
		t.parser.Advance(p[i])
		state := t.parser.State()
		// flush grapheme if we transitioned to a non-utf8 state or we have
		// written the whole byte slice.
		if len(t.grapheme) > 0 {
			if (t.lastState == parser.GroundState && state != parser.Utf8State) || i == len(p)-1 {
				t.flushGrapheme()
			}
		}
		t.lastState = state
	}
	return len(p), nil
}

// InputPipe returns the terminal's input pipe.
// This can be used to send input to the terminal.
func (t *Terminal) InputPipe() io.Writer {
	return t.pw
}

// Paste pastes text into the terminal.
// If bracketed paste mode is enabled, the text is bracketed with the
// appropriate escape sequences.
func (t *Terminal) Paste(text string) {
	if t.isModeSet(ansi.BracketedPasteMode) {
		_, _ = io.WriteString(t.pw, ansi.BracketedPasteStart)
		defer io.WriteString(t.pw, ansi.BracketedPasteEnd) //nolint:errcheck
	}

	_, _ = io.WriteString(t.pw, text)
}

// SendText sends arbitrary text to the terminal.
func (t *Terminal) SendText(text string) {
	_, _ = io.WriteString(t.pw, text)
}

// SendKeys sends multiple keys to the terminal.
func (t *Terminal) SendKeys(keys ...uv.KeyEvent) {
	for _, k := range keys {
		t.SendKey(k)
	}
}

// ForegroundColor returns the terminal's foreground color. This returns nil if
// the foreground color is not set which means the outer terminal color is
// used.
func (t *Terminal) ForegroundColor() color.Color {
	if t.fgColor == nil {
		return t.defaultFg
	}
	return t.fgColor
}

// SetForegroundColor sets the terminal's foreground color.
func (t *Terminal) SetForegroundColor(c color.Color) {
	if c == nil {
		c = t.defaultFg
	}
	t.fgColor = c
	if t.cb.ForegroundColor != nil {
		t.cb.ForegroundColor(c)
	}
}

// SetDefaultForegroundColor sets the terminal's default foreground color.
func (t *Terminal) SetDefaultForegroundColor(c color.Color) {
	t.defaultFg = c
}

// BackgroundColor returns the terminal's background color. This returns nil if
// the background color is not set which means the outer terminal color is
// used.
func (t *Terminal) BackgroundColor() color.Color {
	if t.bgColor == nil {
		return t.defaultBg
	}
	return t.bgColor
}

// SetBackgroundColor sets the terminal's background color.
func (t *Terminal) SetBackgroundColor(c color.Color) {
	if c == nil {
		c = t.defaultBg
	}
	t.bgColor = c
	if t.cb.BackgroundColor != nil {
		t.cb.BackgroundColor(c)
	}
}

// SetDefaultBackgroundColor sets the terminal's default background color.
func (t *Terminal) SetDefaultBackgroundColor(c color.Color) {
	t.defaultBg = c
}

// CursorColor returns the terminal's cursor color. This returns nil if the
// cursor color is not set which means the outer terminal color is used.
func (t *Terminal) CursorColor() color.Color {
	if t.curColor == nil {
		return t.defaultCur
	}
	return t.curColor
}

// SetCursorColor sets the terminal's cursor color.
func (t *Terminal) SetCursorColor(c color.Color) {
	if c == nil {
		c = t.defaultCur
	}
	t.curColor = c
	if t.cb.CursorColor != nil {
		t.cb.CursorColor(c)
	}
}

// SetDefaultCursorColor sets the terminal's default cursor color.
func (t *Terminal) SetDefaultCursorColor(c color.Color) {
	t.defaultCur = c
}

// IndexedColor returns a terminal's indexed color. An indexed color is a color
// between 0 and 255.
func (t *Terminal) IndexedColor(i int) color.Color {
	if i < 0 || i > 255 {
		return nil
	}

	c := t.colors[i]
	if c == nil {
		// Return the default color.
		return ansi.IndexedColor(i) //nolint:gosec
	}

	return c
}

// SetIndexedColor sets a terminal's indexed color.
// The index must be between 0 and 255.
func (t *Terminal) SetIndexedColor(i int, c color.Color) {
	if i < 0 || i > 255 {
		return
	}

	t.colors[i] = c
}

// resetTabStops resets the terminal tab stops to the default set.
func (t *Terminal) resetTabStops() {
	t.tabstops = uv.DefaultTabStops(t.Width())
}

func (t *Terminal) logf(format string, v ...any) {
	if t.logger != nil {
		t.logger.Printf(format, v...)
	}
}

// Dump returns the complete terminal content including scrollback for main screen
// If styled is true, includes ANSI escape sequences
// For main screen: excludes trailing empty lines and includes scrollback
// For alt screen: includes all lines, no scrollback
func (t *Terminal) Dump(styled bool) []string {
	isAltScreen := t.scr == &t.scrs[1]
	result := t.scr.Dump(styled, isAltScreen)
	return result
}
