package vt

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// lineCache stores cached line content
type lineCache struct {
	styled   string
	unstyled string
	isEmpty  bool
	valid    bool
}

// Screen represents a virtual terminal screen.
type Screen struct {
	// cb is the callbacks struct to use.
	cb *Callbacks
	// The buffer of the screen.
	buf uv.Buffer
	// The cur of the screen.
	cur, saved Cursor
	// scroll is the scroll region.
	scroll uv.Rectangle

	// scrollback stores lines that have scrolled off the top
	scrollback [][]uv.Cell
	// maxScrollback is the maximum number of scrollback lines to keep
	maxScrollback int

	// lineCache caches rendered lines for performance
	lineCache []lineCache
	// scrollbackCache caches rendered scrollback lines
	scrollbackCache []lineCache
}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	s := Screen{
		maxScrollback: 10000, // Default scrollback size
	}
	s.Resize(w, h) // This calls initCache internally
	return &s
}

// Reset resets the screen.
// It clears the screen, sets the cursor to the top left corner, reset the
// cursor styles, and resets the scroll region.
func (s *Screen) Reset() {
	s.buf.Clear()
	s.cur = Cursor{}
	s.saved = Cursor{}
	s.scroll = s.buf.Bounds()
	s.scrollback = nil
	s.clearCache()
}

// Bounds returns the bounds of the screen.
func (s *Screen) Bounds() uv.Rectangle {
	return s.buf.Bounds()
}

// Touched returns touched lines in the screen buffer.
func (s *Screen) Touched() []*uv.LineData {
	return s.buf.Touched
}

// CellAt returns the cell at the given x, y position.
func (s *Screen) CellAt(x int, y int) *uv.Cell {
	return s.buf.CellAt(x, y)
}

// SetCell sets the cell at the given x, y position.
func (s *Screen) SetCell(x, y int, c *uv.Cell) {
	s.buf.SetCell(x, y, c)
	s.invalidateLineCache(y)
}

// Height returns the height of the screen.
func (s *Screen) Height() int {
	return s.buf.Height()
}

// Resize resizes the screen.
func (s *Screen) Resize(width int, height int) {
	s.buf.Resize(width, height)
	s.scroll = s.buf.Bounds()
	s.initCache(height)
}

// Width returns the width of the screen.
func (s *Screen) Width() int {
	return s.buf.Width()
}

// Clear clears the screen with blank cells.
func (s *Screen) Clear() {
	s.ClearArea(s.Bounds())
}

// ClearArea clears the given area.
func (s *Screen) ClearArea(area uv.Rectangle) {
	s.buf.ClearArea(area)
	for y := area.Min.Y; y < area.Max.Y; y++ {
		s.invalidateLineCache(y)
	}
}

// Fill fills the screen or part of it.
func (s *Screen) Fill(c *uv.Cell) {
	s.FillArea(c, s.Bounds())
}

// FillArea fills the given area with the given cell.
func (s *Screen) FillArea(c *uv.Cell, area uv.Rectangle) {
	s.buf.FillArea(c, area)
	for y := area.Min.Y; y < area.Max.Y; y++ {
		s.invalidateLineCache(y)
	}
}

// setHorizontalMargins sets the horizontal margins.
func (s *Screen) setHorizontalMargins(left, right int) {
	s.scroll.Min.X = left
	s.scroll.Max.X = right
}

// setVerticalMargins sets the vertical margins.
func (s *Screen) setVerticalMargins(top, bottom int) {
	s.scroll.Min.Y = top
	s.scroll.Max.Y = bottom
}

// setCursorX sets the cursor X position. If margins is true, the cursor is
// only set if it is within the scroll margins.
func (s *Screen) setCursorX(x int, margins bool) {
	s.setCursor(x, s.cur.Y, margins)
}

// setCursorY sets the cursor Y position. If margins is true, the cursor is
// only set if it is within the scroll margins.
func (s *Screen) setCursorY(y int, margins bool) { //nolint:unused
	s.setCursor(s.cur.X, y, margins)
}

// setCursor sets the cursor position. If margins is true, the cursor is only
// set if it is within the scroll margins. This follows how [ansi.CUP] works.
func (s *Screen) setCursor(x, y int, margins bool) {
	old := s.cur.Position
	if !margins {
		y = clamp(y, 0, s.buf.Height()-1)
		x = clamp(x, 0, s.buf.Width()-1)
	} else {
		y = clamp(s.scroll.Min.Y+y, s.scroll.Min.Y, s.scroll.Max.Y-1)
		x = clamp(s.scroll.Min.X+x, s.scroll.Min.X, s.scroll.Max.X-1)
	}
	s.cur.X, s.cur.Y = x, y

	if s.cb.CursorPosition != nil && (old.X != x || old.Y != y) {
		s.cb.CursorPosition(old, uv.Pos(x, y))
	}
}

// moveCursor moves the cursor by the given x and y deltas. If the cursor
// position is inside the scroll region, it is bounded by the scroll region.
// Otherwise, it is bounded by the screen bounds.
// This follows how [ansi.CUU], [ansi.CUD], [ansi.CUF], [ansi.CUB], [ansi.CNL],
// [ansi.CPL].
func (s *Screen) moveCursor(dx, dy int) {
	scroll := s.scroll
	old := s.cur.Position
	if old.X < scroll.Min.X {
		scroll.Min.X = 0
	}
	if old.X >= scroll.Max.X {
		scroll.Max.X = s.buf.Width()
	}

	pt := uv.Pos(s.cur.X+dx, s.cur.Y+dy)

	var x, y int
	if old.In(scroll) {
		y = clamp(pt.Y, scroll.Min.Y, scroll.Max.Y-1)
		x = clamp(pt.X, scroll.Min.X, scroll.Max.X-1)
	} else {
		y = clamp(pt.Y, 0, s.buf.Height()-1)
		x = clamp(pt.X, 0, s.buf.Width()-1)
	}

	s.cur.X, s.cur.Y = x, y

	if s.cb.CursorPosition != nil && (old.X != x || old.Y != y) {
		s.cb.CursorPosition(old, uv.Pos(x, y))
	}
}

// Cursor returns the cursor.
func (s *Screen) Cursor() Cursor {
	return s.cur
}

// CursorPosition returns the cursor position.
func (s *Screen) CursorPosition() (x, y int) {
	return s.cur.X, s.cur.Y
}

// ScrollRegion returns the scroll region.
func (s *Screen) ScrollRegion() uv.Rectangle {
	return s.scroll
}

// SaveCursor saves the cursor.
func (s *Screen) SaveCursor() {
	s.saved = s.cur
}

// RestoreCursor restores the cursor.
func (s *Screen) RestoreCursor() {
	old := s.cur.Position
	s.cur = s.saved

	if s.cb.CursorPosition != nil && (old.X != s.cur.X || old.Y != s.cur.Y) {
		s.cb.CursorPosition(old, s.cur.Position)
	}
}

// setCursorHidden sets the cursor hidden.
func (s *Screen) setCursorHidden(hidden bool) {
	changed := s.cur.Hidden != hidden
	s.cur.Hidden = hidden
	if changed && s.cb.CursorVisibility != nil {
		s.cb.CursorVisibility(!hidden)
	}
}

// setCursorStyle sets the cursor style.
func (s *Screen) setCursorStyle(style CursorStyle, blink bool) {
	changed := s.cur.Style != style || s.cur.Steady != !blink
	s.cur.Style = style
	s.cur.Steady = !blink
	if changed && s.cb.CursorStyle != nil {
		s.cb.CursorStyle(style, !blink)
	}
}

// cursorPen returns the cursor pen.
func (s *Screen) cursorPen() uv.Style {
	return s.cur.Pen
}

// cursorLink returns the cursor link.
func (s *Screen) cursorLink() uv.Link {
	return s.cur.Link
}

// ShowCursor shows the cursor.
func (s *Screen) ShowCursor() {
	s.setCursorHidden(false)
}

// HideCursor hides the cursor.
func (s *Screen) HideCursor() {
	s.setCursorHidden(true)
}

// InsertCell inserts n blank characters at the cursor position pushing out
// cells to the right and out of the screen.
func (s *Screen) InsertCell(n int) {
	if n <= 0 {
		return
	}

	x, y := s.cur.X, s.cur.Y
	s.buf.InsertCellArea(x, y, n, s.blankCell(), s.scroll)
	s.invalidateLineCache(y)
}

// DeleteCell deletes n cells at the cursor position moving cells to the left.
// This has no effect if the cursor is outside the scroll region.
func (s *Screen) DeleteCell(n int) {
	if n <= 0 {
		return
	}

	x, y := s.cur.X, s.cur.Y
	s.buf.DeleteCellArea(x, y, n, s.blankCell(), s.scroll)
	s.invalidateLineCache(y)
}

// ScrollUp scrolls the content up n lines within the given region. Lines
// scrolled past the top margin are moved to scrollback buffer.
func (s *Screen) ScrollUp(n int) {
	if n <= 0 {
		return
	}

	x, y := s.CursorPosition()
	scroll := s.scroll

	// Save scrolled lines to scrollback buffer
	for i := 0; i < n && scroll.Min.Y < scroll.Max.Y; i++ {
		line := make([]uv.Cell, s.buf.Width())
		for x := 0; x < s.buf.Width(); x++ {
			if cell := s.buf.CellAt(x, scroll.Min.Y); cell != nil {
				line[x] = *cell
			}
		}
		s.addToScrollback(line)
	}

	s.setCursor(s.cur.X, 0, true)
	s.DeleteLine(n)
	s.setCursor(x, y, false)
}

// ScrollDown scrolls the content down n lines within the given region. Lines
// scrolled past the bottom margin are lost. This is equivalent to [ansi.SD]
// which moves the cursor to top margin and performs a [ansi.IL] operation.
func (s *Screen) ScrollDown(n int) {
	x, y := s.CursorPosition()
	s.setCursor(s.cur.X, 0, true)
	s.InsertLine(n)
	s.setCursor(x, y, false)
}

// InsertLine inserts n blank lines at the cursor position Y coordinate.
// Only operates if cursor is within scroll region. Lines below cursor Y
// are moved down, with those past bottom margin being discarded.
// It returns true if the operation was successful.
func (s *Screen) InsertLine(n int) bool {
	if n <= 0 {
		return false
	}

	x, y := s.cur.X, s.cur.Y

	// Only operate if cursor Y is within scroll region
	if y < s.scroll.Min.Y || y >= s.scroll.Max.Y ||
		x < s.scroll.Min.X || x >= s.scroll.Max.X {
		return false
	}

	s.buf.InsertLineArea(y, n, s.blankCell(), s.scroll)

	// Invalidate cache for affected lines
	for i := y; i < s.scroll.Max.Y; i++ {
		s.invalidateLineCache(i)
	}

	return true
}

// DeleteLine deletes n lines at the cursor position Y coordinate.
// Only operates if cursor is within scroll region. Lines below cursor Y
// are moved up, with blank lines inserted at the bottom of scroll region.
// It returns true if the operation was successful.
func (s *Screen) DeleteLine(n int) bool {
	if n <= 0 {
		return false
	}

	scroll := s.scroll
	x, y := s.cur.X, s.cur.Y

	// Only operate if cursor Y is within scroll region
	if y < scroll.Min.Y || y >= scroll.Max.Y ||
		x < scroll.Min.X || x >= scroll.Max.X {
		return false
	}

	s.buf.DeleteLineArea(y, n, s.blankCell(), scroll)

	// Invalidate cache for affected lines
	for i := y; i < scroll.Max.Y; i++ {
		s.invalidateLineCache(i)
	}

	return true
}

// blankCell returns the cursor blank cell with the background color set to the
// current pen background color. If the pen background color is nil, the return
// value is nil.
func (s *Screen) blankCell() *uv.Cell {
	if s.cur.Pen.Bg == nil {
		return nil
	}

	c := uv.EmptyCell
	c.Style.Bg = s.cur.Pen.Bg
	return &c
}

// initCache initializes the line cache with the given height
func (s *Screen) initCache(height int) {
	s.lineCache = make([]lineCache, height)
	for i := range s.lineCache {
		s.lineCache[i] = lineCache{isEmpty: true, valid: false}
	}
}

// clearCache clears all cached line data
func (s *Screen) clearCache() {
	for i := range s.lineCache {
		s.lineCache[i] = lineCache{isEmpty: true, valid: false}
	}
	for i := range s.scrollbackCache {
		s.scrollbackCache[i] = lineCache{isEmpty: true, valid: false}
	}
}

// invalidateLineCache marks a line's cache as invalid
func (s *Screen) invalidateLineCache(y int) {
	if y >= 0 && y < len(s.lineCache) {
		s.lineCache[y].valid = false
	}
}

// addToScrollback adds a line to the scrollback buffer
func (s *Screen) addToScrollback(line []uv.Cell) {
	s.scrollback = append(s.scrollback, line)

	// Maintain maximum scrollback size
	if len(s.scrollback) > s.maxScrollback {
		copy(s.scrollback, s.scrollback[1:])
		s.scrollback = s.scrollback[:s.maxScrollback]
		// Shift scrollback cache accordingly
		if len(s.scrollbackCache) > 0 {
			copy(s.scrollbackCache, s.scrollbackCache[1:])
			s.scrollbackCache = s.scrollbackCache[:len(s.scrollbackCache)-1]
		}
	}

	// Add cache entry for new scrollback line
	s.scrollbackCache = append(s.scrollbackCache, lineCache{isEmpty: true, valid: false})
}

// getCursorStyle returns ANSI style sequences for cursor based on its style and original cell
func (s *Screen) getCursorStyle(styled bool) (prefix, suffix string) {
	if !styled {
		// For unstyled output, use simple visual indicators
		switch s.cur.Style {
		case CursorBlock:
			return "[", "]" // Block cursor with brackets
		case CursorUnderline:
			return "", "_" // Underline cursor
		case CursorBar:
			return "|", "" // Bar cursor
		default:
			return "[", "]" // Default to brackets
		}
	}

	// For styled output, use ANSI escape sequences
	switch s.cur.Style {
	case CursorBlock:
		// Invert colors to create block cursor effect
		return "\033[7m", "\033[27m" // Reverse video on/off
	case CursorUnderline:
		// Add underline to the character
		return "\033[4m", "\033[24m" // Underline on/off
	case CursorBar:
		// Add a bar character before the original character
		return "\033[7m|\033[27m", "" // Inverted bar + original char
	default:
		// Default to reverse video
		return "\033[7m", "\033[27m"
	}
}

// renderLine renders a line to styled and unstyled strings
func (s *Screen) renderLine(cells []uv.Cell, width int) (styled, unstyled string, isEmpty bool) {
	var styledBuilder, unstyledBuilder strings.Builder

	isEmpty = true
	lastContentX := -1

	// First pass: build full strings and find last non-empty position
	for x := 0; x < width; x++ {
		var cell uv.Cell
		if x < len(cells) {
			cell = cells[x]
		}

		// Check if cell has actual content (not just whitespace)
		if cell.Content != "" && cell.Content != " " && cell.Content != "\t" {
			isEmpty = false
			lastContentX = x
		} else if cell.Content == " " || cell.Content == "\t" {
			// Whitespace is content for positioning but line can still be considered empty
			lastContentX = x
		}

		// Build styled string
		if cell.Style.Sequence() != "" {
			styledBuilder.WriteString(cell.Style.Sequence())
		}
		styledBuilder.WriteString(cell.Content)

		// Build unstyled string
		unstyledBuilder.WriteString(cell.Content)

		// Skip additional width for wide characters
		if cell.Width > 1 {
			x += cell.Width - 1
		}
	}

	// For styled output, don't trim - keep full ANSI sequences intact
	styled = styledBuilder.String()

	// For unstyled output, trim trailing whitespace
	unstyled = unstyledBuilder.String()
	if lastContentX >= 0 && lastContentX < len(unstyled) {
		// Trim trailing spaces/tabs but preserve intentional content
		unstyled = strings.TrimRightFunc(unstyled, func(r rune) bool {
			return r == ' ' || r == '\t'
		})
	}

	// Double-check: if unstyled content is empty or only whitespace, mark as empty
	if strings.TrimSpace(unstyled) == "" {
		isEmpty = true
	}

	return styled, unstyled, isEmpty
}

// renderLineWithCursor renders a line to styled and unstyled strings with semi-transparent cursor display
func (s *Screen) renderLineWithCursor(cells []uv.Cell, width int, showCursor bool, cursorX int, styled bool) (styledLine, unstyledLine string, isEmpty bool) {
	var styledBuilder, unstyledBuilder strings.Builder

	isEmpty = true
	lastContentX := -1

	// First pass: build full strings and find last non-empty position
	for x := 0; x < width; x++ {
		var cell uv.Cell
		if x < len(cells) {
			cell = cells[x]
		}

		// If this is cursor position and cursor should be shown
		if showCursor && x == cursorX {
			// Get original character (or space if empty)
			originalChar := cell.Content
			if originalChar == "" {
				originalChar = " "
			}

			// Get cursor style for this character
			prefix, suffix := s.getCursorStyle(styled)

			// Check if we have actual content (not just whitespace)
			if originalChar != " " && originalChar != "\t" {
				isEmpty = false
				lastContentX = x
			} else {
				// Even whitespace counts as cursor position
				lastContentX = x
			}

			if styled {
				// Build styled string with cursor style applied to original character
				if cell.Style.Sequence() != "" {
					styledBuilder.WriteString(cell.Style.Sequence())
				}
				styledBuilder.WriteString(prefix)
				styledBuilder.WriteString(originalChar)
				styledBuilder.WriteString(suffix)
			} else {
				// For unstyled output, show original char with simple cursor indicators
				unstyledBuilder.WriteString(prefix)
				unstyledBuilder.WriteString(originalChar)
				unstyledBuilder.WriteString(suffix)
			}
		} else {
			// Regular cell processing
			// Check if cell has actual content (not just whitespace)
			if cell.Content != "" && cell.Content != " " && cell.Content != "\t" {
				isEmpty = false
				lastContentX = x
			} else if cell.Content == " " || cell.Content == "\t" {
				// Whitespace is content for positioning but line can still be considered empty
				lastContentX = x
			}

			if styled {
				// Build styled string
				if cell.Style.Sequence() != "" {
					styledBuilder.WriteString(cell.Style.Sequence())
				}
				styledBuilder.WriteString(cell.Content)
			} else {
				// Build unstyled string
				unstyledBuilder.WriteString(cell.Content)
			}
		}

		// Skip additional width for wide characters
		if cell.Width > 1 {
			x += cell.Width - 1
		}
	}

	// Get final strings
	if styled {
		styledLine = styledBuilder.String()
		// For styled, also build unstyled version for return
		unstyledBuilder.Reset()
		for x := 0; x < width; x++ {
			var cell uv.Cell
			if x < len(cells) {
				cell = cells[x]
			}

			if showCursor && x == cursorX {
				originalChar := cell.Content
				if originalChar == "" {
					originalChar = " "
				}
				prefix, suffix := s.getCursorStyle(false)
				unstyledBuilder.WriteString(prefix)
				unstyledBuilder.WriteString(originalChar)
				unstyledBuilder.WriteString(suffix)
			} else {
				unstyledBuilder.WriteString(cell.Content)
			}

			if cell.Width > 1 {
				x += cell.Width - 1
			}
		}
		unstyledLine = unstyledBuilder.String()
	} else {
		unstyledLine = unstyledBuilder.String()
		styledLine = unstyledLine // For unstyled mode, both are the same
	}

	// Trim trailing whitespace for unstyled output
	if lastContentX >= 0 && lastContentX < len(unstyledLine) {
		unstyledLine = strings.TrimRightFunc(unstyledLine, func(r rune) bool {
			return r == ' ' || r == '\t'
		})
	}

	// Double-check: if unstyled content is empty or only whitespace, mark as empty
	if strings.TrimSpace(unstyledLine) == "" {
		isEmpty = true
	}

	return styledLine, unstyledLine, isEmpty
}

// updateLineCache updates the cache for a specific line
func (s *Screen) updateLineCache(y int) {
	if y < 0 || y >= len(s.lineCache) || y >= s.buf.Height() {
		return
	}

	line := make([]uv.Cell, s.buf.Width())
	for x := 0; x < s.buf.Width(); x++ {
		if cell := s.buf.CellAt(x, y); cell != nil {
			line[x] = *cell
		}
	}

	styled, unstyled, isEmpty := s.renderLine(line, s.buf.Width())
	s.lineCache[y] = lineCache{
		styled:   styled,
		unstyled: unstyled,
		isEmpty:  isEmpty,
		valid:    true,
	}
}

// updateScrollbackLineCache updates the cache for a specific scrollback line
func (s *Screen) updateScrollbackLineCache(idx int) {
	if idx < 0 || idx >= len(s.scrollback) || idx >= len(s.scrollbackCache) {
		return
	}

	line := s.scrollback[idx]
	styled, unstyled, isEmpty := s.renderLine(line, s.buf.Width())
	s.scrollbackCache[idx] = lineCache{
		styled:   styled,
		unstyled: unstyled,
		isEmpty:  isEmpty,
		valid:    true,
	}
}

// Dump returns the complete terminal content including scrollback
// If styled is true, includes ANSI escape sequences
// For main screen: excludes trailing empty lines and includes scrollback
// For alt screen: includes all lines, no scrollback
func (s *Screen) Dump(styled bool, isAltScreen bool) []string {
	var lines []string

	if !isAltScreen {
		// Add scrollback lines for main screen
		for i := range s.scrollback {
			if i >= len(s.scrollbackCache) {
				s.scrollbackCache = append(s.scrollbackCache, lineCache{isEmpty: true, valid: false})
			}

			if !s.scrollbackCache[i].valid {
				s.updateScrollbackLineCache(i)
			}

			cache := s.scrollbackCache[i]
			if styled {
				lines = append(lines, cache.styled)
			} else {
				lines = append(lines, cache.unstyled)
			}
		}
	}

	// Add current screen lines
	lastNonEmpty := -1
	screenLines := make([]string, s.buf.Height())

	// Check if cursor should be displayed for alt screen
	showCursor := isAltScreen && !s.cur.Hidden &&
		s.cur.Y >= 0 && s.cur.Y < s.buf.Height() &&
		s.cur.X >= 0 && s.cur.X < s.buf.Width()

	for y := 0; y < s.buf.Height(); y++ {
		var line string

		// If this is the cursor line and we should show cursor, render with cursor
		if showCursor && y == s.cur.Y {
			// Get line cells
			lineCells := make([]uv.Cell, s.buf.Width())
			for x := 0; x < s.buf.Width(); x++ {
				if cell := s.buf.CellAt(x, y); cell != nil {
					lineCells[x] = *cell
				}
			}

			// Render line with cursor
			styledLine, unstyledLine, isEmpty := s.renderLineWithCursor(lineCells, s.buf.Width(), true, s.cur.X, styled)
			if styled {
				line = styledLine
			} else {
				line = unstyledLine
			}

			// Track last non-empty line for main screen
			if !isAltScreen && !isEmpty {
				lastNonEmpty = y
			}
		} else {
			// Regular line rendering using cache
			// Ensure cache is large enough
			if y >= len(s.lineCache) {
				s.initCache(s.buf.Height())
			}

			if !s.lineCache[y].valid {
				s.updateLineCache(y)
			}

			cache := s.lineCache[y]
			if styled {
				line = cache.styled
			} else {
				line = cache.unstyled
			}

			// Track last non-empty line for main screen
			if !isAltScreen && !cache.isEmpty {
				lastNonEmpty = y
			}
		}

		screenLines[y] = line
	}

	if isAltScreen {
		// Alt screen: include all lines
		lines = append(lines, screenLines...)
	} else {
		// Main screen: exclude trailing empty lines
		if lastNonEmpty >= 0 {
			trimmedLines := screenLines[:lastNonEmpty+1]
			lines = append(lines, trimmedLines...)
		}
	}

	// Add ANSI reset sequence at the end of styled output to prevent style bleeding
	if styled && len(lines) > 0 {
		// Find the last non-empty line to add reset sequence
		for i := len(lines) - 1; i >= 0; i-- {
			if lines[i] != "" {
				lines[i] += "\033[0m" // ANSI reset sequence
				break
			}
		}
	}

	return lines
}
