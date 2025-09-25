package vt

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// nextTab moves the cursor to the next tab stop n times. This respects the
// horizontal scrolling region. This performs the same function as [ansi.CHT].
func (t *Terminal) nextTab(n int) {
	x, y := t.scr.CursorPosition()
	scroll := t.scr.ScrollRegion()
	for range n {
		ts := t.tabstops.Next(x)
		if ts < x {
			break
		}
		x = ts
	}

	if x >= scroll.Max.X {
		x = min(scroll.Max.X-1, t.Width()-1)
	}

	// NOTE: We use t.scr.setCursor here because we don't want to reset the
	// phantom state.
	t.scr.setCursor(x, y, false)
}

// prevTab moves the cursor to the previous tab stop n times. This respects the
// horizontal scrolling region when origin mode is set. If the cursor would
// move past the leftmost valid column, the cursor remains at the leftmost
// valid column and the operation completes.
func (t *Terminal) prevTab(n int) {
	x, _ := t.scr.CursorPosition()
	leftmargin := 0
	scroll := t.scr.ScrollRegion()
	if t.isModeSet(ansi.DECOM) {
		leftmargin = scroll.Min.X
	}

	for range n {
		ts := t.tabstops.Prev(x)
		if ts > x {
			break
		}
		x = ts
	}

	if x < leftmargin {
		x = leftmargin
	}

	// NOTE: We use t.scr.setCursorX here because we don't want to reset the
	// phantom state.
	t.scr.setCursorX(x, false)
}

// moveCursor moves the cursor by the given x and y deltas. If the cursor
// is at phantom, the state will reset and the cursor is back in the screen.
func (t *Terminal) moveCursor(dx, dy int) {
	t.scr.moveCursor(dx, dy)
	t.atPhantom = false
}

// setCursor sets the cursor position. This resets the phantom state.
func (t *Terminal) setCursor(x, y int) {
	t.scr.setCursor(x, y, false)
	t.atPhantom = false
}

// setCursorPosition sets the cursor position. This respects [ansi.DECOM],
// Origin Mode. This performs the same function as [ansi.CUP].
func (t *Terminal) setCursorPosition(x, y int) {
	mode, ok := t.modes[ansi.DECOM]
	margins := ok && mode.IsSet()
	t.scr.setCursor(x, y, margins)
	t.atPhantom = false
}

// carriageReturn moves the cursor to the leftmost column. If [ansi.DECOM] is
// set, the cursor is set to the left margin. If not, and the cursor is on or
// to the right of the left margin, the cursor is set to the left margin.
// Otherwise, the cursor is set to the leftmost column of the screen.
// This performs the same function as [ansi.CR].
func (t *Terminal) carriageReturn() {
	mode, ok := t.modes[ansi.DECOM]
	margins := ok && mode.IsSet()
	x, y := t.scr.CursorPosition()
	if margins {
		t.scr.setCursor(0, y, true)
	} else if region := t.scr.ScrollRegion(); uv.Pos(x, y).In(region) {
		t.scr.setCursor(region.Min.X, y, false)
	} else {
		t.scr.setCursor(0, y, false)
	}
	t.atPhantom = false
}

// repeatPreviousCharacter repeats the previous character n times. This is
// equivalent to typing the same character n times. This performs the same as
// [ansi.REP].
func (t *Terminal) repeatPreviousCharacter(n int) {
	if t.lastChar == 0 {
		return
	}
	for range n {
		t.handlePrint(t.lastChar)
	}
}
